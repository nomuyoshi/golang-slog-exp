package main

import (
	"os"
	"slogexp/sdk"

	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(sdk.NewMyHandler(os.Stdout, nil))
	logger.Info("slog体験", "count", 4)
}
