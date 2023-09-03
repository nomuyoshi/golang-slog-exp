package sdk

import "log/slog"

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
