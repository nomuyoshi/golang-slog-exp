package sdk

import (
	"context"
	"io"
	"log/slog"
	"time"
)

type (
	MyHandler struct {
		opts Options
		w    io.Writer
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

// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - TODO: If r.PC is zero, ignore it.
//   - TODO: Attr's values should be resolved.
//   - TODO: If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - TODO: If a group's key is empty, inline the group's Attrs.
//   - TODO: If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
	output := newOutput()
	defer output.free()

	output.buf.WriteByte('{')

	if !r.Time.IsZero() {
		output.appendKey(slog.TimeKey)
		output.appendTime(r.Time)
	}
	output.appendKey(slog.LevelKey)
	output.appendString(r.Level.String())

	output.appendKey(slog.MessageKey)
	output.appendString(r.Message)
	output.buf.WriteByte('}')
	output.buf.WriteByte('\n')

	_, err := h.w.Write(*output.buf)
	return err
}

type output struct {
	buf *Buffer
}

func newOutput() *output {
	return &output{buf: NewBuffer()}
}

func (o *output) free() {
	o.buf.Free()
}

func (o *output) appendKey(key string) {
	o.appendString(key)
	o.buf.WriteByte(':')
}

func (o *output) appendString(str string) {
	o.buf.WriteByte('"')
	o.buf.WriteString(str) // TODO strをescape
	o.buf.WriteByte('"')
}

func (o *output) appendTime(t time.Time) {
	o.buf.WriteByte('"')
	o.buf.WriteString(t.Format(time.RFC3339Nano))
	o.buf.WriteByte('"')
}
