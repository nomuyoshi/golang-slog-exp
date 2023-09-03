package sdk

import (
	"context"
	"log/slog"
)

type (
	MyHandler struct {
		opts Options
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
