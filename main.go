package main

import (
	"fmt"
	"os"
	"slogexp/sdk"

	"log/slog"
)

func main() {
	logger := slog.New(sdk.NewMyHandler(os.Stdout, nil))
	logger = logger.With("trace_id", "adkjk290jwe3w03233=")
	logger.Info("slog体験", "count", 4)
	// => {"time":"2023-09-09T23:39:33.922167+09:00","level":"INFO","msg":"slog体験","trace_id":"adkjk290jwe3w03233=","count":4}
	userLogger := logger.WithGroup("user")
	userLogger = userLogger.With("service", "sns")
	userLogger.Info("コメント投稿完了", "user_id", 1, "comment_id", 1000)
	// => {"time":"2023-09-09T23:39:33.922331+09:00","level":"INFO","msg":"コメント投稿完了","trace_id":"adkjk290jwe3w03233=","user":{"service":"sns","user_id":1,"comment_id":1000}}
	userLogger.Error("コメント投稿エラー", "user_id", 1, "comment", "ああああ", "error", fmt.Errorf("hogehoge").Error())
	// => {"time":"2023-09-09T23:42:00.869847+09:00","level":"ERROR","msg":"コメント投稿エラー","trace_id":"adkjk290jwe3w03233=","user":{"service":"sns","user_id":1,"comment":"ああああ","error":"hogehoge"}}
}
