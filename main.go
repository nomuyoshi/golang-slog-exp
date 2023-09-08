package main

import (
	"os"
	"slogexp/sdk"
	"time"

	"log/slog"
)

func main() {
	logger := slog.New(sdk.NewMyHandler(os.Stdout, nil))
	logger.Info("slog体験", "count", 4)
	userLogger := logger.WithGroup("user")
	userLogger.Info("slog user application", "user_id", 1, "req_id", "a24jladjkelr30")

	accoutLogger := userLogger.WithGroup("accout")
	accoutLogger.Info("アカウント作成完了", "user_id", 1, "registered_at", time.Now())
	jlogger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("trace_id", "7uoirlad03").WithGroup("admin")
	jlogger.Info("test", "user_id", 1000)
}
