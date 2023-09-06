package main

import (
	"os"
	"slogexp/sdk"

	"log/slog"
)

func main() {
	logger := slog.New(sdk.NewMyHandler(os.Stdout, nil))
	logger.Info("slog体験", "count", 4)
}
