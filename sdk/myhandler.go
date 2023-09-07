package sdk

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"
)

type (
	MyHandler struct {
		opts   Options
		groups []string // all groups started from WithGroup
		w      io.Writer
		mu     *sync.Mutex
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
	state := h.newHandleState(NewBuffer())
	defer state.free()

	state.buf.WriteByte('{')

	if !r.Time.IsZero() {
		state.appendKey(slog.TimeKey)
		state.appendTime(r.Time)
	}
	state.appendKey(slog.LevelKey)
	state.appendString(r.Level.String())

	state.appendKey(slog.MessageKey)
	state.appendString(r.Message)
	state.buf.WriteByte('}')
	state.buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(*state.buf)

	return err
}

func (h *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *MyHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &MyHandler{
		opts:   h.opts,
		groups: append(h.groups, name),
		w:      h.w,
	}
}

const sep = ","

type handleState struct {
	buf *Buffer
	sep string // 各値の区切り文字
}

func (h *MyHandler) newHandleState(buf *Buffer) handleState {
	s := handleState{
		buf: buf,
	}
	return s
}

func (s *handleState) free() {
	s.buf.Free()
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
