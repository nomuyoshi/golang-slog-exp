package sdk

import (
	"context"
	"io"
	"log/slog"
	"time"
)

type (
	MyHandler struct {
		opts   Options
		groups []string // all groups started from WithGroup
		w      io.Writer
	}

	Options struct {
		Level slog.Leveler
	}
)

func NewMyHandler(opt *Options) *MyHandler {
	if opt == nil {
		opt = &Options{}
	}
	return &MyHandler{
		opts: *opt,
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
	state := newHandleState()
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

type handleState struct {
	buf *Buffer
}

func newHandleState() *handleState {
	return &handleState{buf: NewBuffer()}
}

func (s *handleState) free() {
	s.buf.Free()
}

func (s *handleState) appendKey(key string) {
	s.appendString(key)
	s.buf.WriteByte(':')
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
