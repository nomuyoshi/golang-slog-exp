package main

import (
	"os"
	"slogexp/sdk"

	"log/slog"
)

func main() {
	logger := slog.New(sdk.NewMyHandler(os.Stdout, nil))
	logger.Info("slog体験", "count", 4)
	//userLogger := logger.WithGroup("user app")
	//userLogger.Info("slog user application", "user_id", 1)
	jlogger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).WithGroup("corp app")
	jlogger.Info("test", "company_id", 1000)
}
