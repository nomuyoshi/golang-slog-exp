package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"slices"
)

type (
	MyHandler struct {
		opts              Options
		preformattedAttrs []byte   // MEMO なにこれ??????????
		groups            []string // all groups started from WithGroup
		nOpenGroups       int      // the number of groups opened in preformattedAttrs
		mu                *sync.Mutex
		w                 io.Writer
	}

	Options struct {
		// どのログレベル以上なら出力するか
		Level slog.Leveler
	}
)

func NewMyHandler(w io.Writer, opt *Options) *MyHandler {
	if opt == nil {
		opt = &Options{}
	}
	return &MyHandler{
		opts: *opt,
		w:    w,
		mu:   &sync.Mutex{},
	}
}

func (h *MyHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return l >= minLevel
}

// Handle https://pkg.go.dev/golang.org/x/exp/slog#Handler
// まだGroupなどの扱いを実装していない
func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
	state := h.newHandleState(NewBuffer(), true)
	defer state.free()

	state.buf.WriteByte('{') // Open the top-level object

	if !r.Time.IsZero() {
		state.appendKey(slog.TimeKey)
		state.appendTime(r.Time)
	}
	state.appendKey(slog.LevelKey)
	state.appendString(r.Level.String())

	state.appendKey(slog.MessageKey)
	state.appendString(r.Message)
	state.appendNonBuiltIns(r)
	state.buf.WriteByte('}') // Close the top-level object
	state.buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(*state.buf)

	return err
}

func (h *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// 空のグループのみなら何もしない
	if countEmptyGroups(attrs) == len(attrs) {
		return h
	}
	h2 := h.clone()
	// state.bufはpreformattedAttrsのポインタ
	state := h2.newHandleState((*Buffer)(&h2.preformattedAttrs), false)
	defer state.free()

	if len(h2.preformattedAttrs) > 0 {
		state.sep = sep
	}
	state.openGroups()

	for _, a := range attrs {
		// newHandleStateにpreformattedAttrsのポインタを渡しているから
		// Attr をs.bufに書き込む = s.preformattedAttrs にも書き込まれる
		state.appendAttr(a)
	}
	h2.nOpenGroups = len(h2.groups)
	return h2
}

func (h *MyHandler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	h2.groups = append(h2.groups, name)

	return h2
}

func (h *MyHandler) clone() *MyHandler {
	// 1.21で追加されたslicesが使われている
	// slices.Clip スライスの不必要なcapを削ってくれる
	return &MyHandler{
		opts:              h.opts,
		preformattedAttrs: slices.Clip(h.preformattedAttrs),
		groups:            slices.Clip(h.groups),
		nOpenGroups:       h.nOpenGroups,
		w:                 h.w,
		mu:                h.mu,
	}
}

const sep = ","

type handleState struct {
	h       *MyHandler
	buf     *Buffer
	freeBuf bool
	sep     string // 各値の区切り文字
}

func (h *MyHandler) newHandleState(buf *Buffer, freeBuf bool) handleState {
	s := handleState{
		h:       h,
		buf:     buf,
		freeBuf: freeBuf,
	}
	return s
}

func (s *handleState) free() {
	if s.freeBuf {
		s.buf.Free()
	}
}

func (s *handleState) appendKey(key string) {
	// 初めて呼ばれるときいはs.sepは空文字
	// 関数の最後に s.sep に値を入れるので2回目以降は「,」で区切られる
	s.buf.WriteString(s.sep)
	s.appendString(key)
	s.buf.WriteByte(':')
	s.sep = sep
}

func (s *handleState) appendString(str string) {
	s.buf.WriteByte('"')
	s.buf.WriteString(str) // TODO strをescape
	s.buf.WriteByte('"')
}

func (s *handleState) appendTime(t time.Time) {
	s.buf.WriteByte('"')
	s.buf.WriteString(t.Format(time.RFC3339Nano))
	s.buf.WriteByte('"')
}

// appendNonBuiltIns 組み込みの値・項目（time, level, msg）以外を handleState.buf にappend
func (s *handleState) appendNonBuiltIns(r slog.Record) {
	if len(s.h.preformattedAttrs) > 0 {
		s.buf.WriteString(s.sep)
		s.buf.Write(s.h.preformattedAttrs)
		s.sep = sep
	}

	nOpenGroups := s.h.nOpenGroups
	if r.NumAttrs() > 0 {
		s.openGroups()
		nOpenGroups = len(s.h.groups)
		r.Attrs(func(a slog.Attr) bool {
			s.appendAttr(a)
			return true
		})
	}
	// close groups
	// オープンしたグループ分（nOpenGroups）括弧を閉じる
	for range s.h.groups[:nOpenGroups] {
		s.buf.WriteByte('}')
	}
}

func (s *handleState) appendAttr(a slog.Attr) {
	a.Value = a.Value.Resolve()
	if isEmpty(a) {
		return
	}

	// TODO Source Case
	if a.Value.Kind() == slog.KindGroup { // Groupだった場合
		attrs := a.Value.Group()
		if len(attrs) > 0 {
			if a.Key != "" {
				s.openGroup(a.Key) // Group開始（`{`)
			}
			for _, aa := range attrs {
				s.appendAttr(aa) // Group内にattrsを書き込む
			}
			if a.Key != "" {
				s.closeGroup()
			}

		}
	} else {
		s.appendKey(a.Key)
		s.appendValue(a.Value)
	}
}

func (s *handleState) appendError(err error) {
	s.appendString(fmt.Sprintf("!ERROR:%v", err))
}

func (s *handleState) appendValue(v slog.Value) {
	var err error
	switch v.Kind() {
	case slog.KindString:
		s.appendString(v.String())
	case slog.KindInt64:
		*s.buf = strconv.AppendInt(*s.buf, v.Int64(), 10)
	case slog.KindUint64:
		*s.buf = strconv.AppendUint(*s.buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*s.buf = strconv.AppendFloat(*s.buf, v.Float64(), 'E', -1, 64)
	case slog.KindBool:
		*s.buf = strconv.AppendBool(*s.buf, v.Bool())
	case slog.KindDuration:
		*s.buf = strconv.AppendInt(*s.buf, int64(v.Duration()), 10)
	case slog.KindTime:
		s.appendTime(v.Time())
	case slog.KindAny: // nilや元となる型が数値のNamed Typeを含む上記のcase以外の全ての型はKindAnyになる
		a := v.Any()
		_, jm := a.(json.Marshaler)
		if e, ok := a.(error); ok && !jm {
			err = e
		} else {
			err = appendJSONMarshal(s.buf, a) // 何が来るかわからんのでencoding/jsonを使ってEncode
		}
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}

	if err != nil {
		s.appendError(err)
	}
}

func (s *handleState) openGroup(name string) {
	s.appendKey(name)
	s.buf.WriteByte('{')
	s.sep = "" // 空にしないとGroup内の1番目のキーの前にカンマが入っちゃう {"group":{,"key1": 1}}
}

func (s *handleState) openGroups() {
	for _, n := range s.h.groups[s.h.nOpenGroups:] {
		s.openGroup(n)
	}
}

func (s *handleState) closeGroup() {
	s.buf.WriteByte('}')
	s.sep = sep
}

func appendJSONMarshal(buf *Buffer, v any) error {
	// Use a json.Encoder to avoid escaping HTML.
	var bb bytes.Buffer
	enc := json.NewEncoder(&bb)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	bs := bb.Bytes()
	buf.Write(bs[:len(bs)-1]) // 末尾の改行コードを削除
	return nil
}

func isEmpty(a slog.Attr) bool {
	return a.Equal(slog.Attr{})
}

// isEmptyGroup reports whether v is a group that has no attributes.
func isEmptyGroup(v slog.Value) bool {
	if v.Kind() != slog.KindGroup {
		return false
	}
	// We do not need to recursively examine the group's Attrs for emptiness,
	// because GroupValue removed them when the group was constructed, and
	// groups are immutable.
	attrs := v.Any().([]slog.Attr)
	return len(attrs) == 0
}

// countEmptyGroups returns the number of empty group values in its argument.
func countEmptyGroups(as []slog.Attr) int {
	n := 0
	for _, a := range as {
		if isEmptyGroup(a.Value) {
			n++
		}
	}
	return n
}
