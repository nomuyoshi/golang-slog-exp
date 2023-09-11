package main

import (
	"context"
	"fmt"
	"os"

	"log/slog"
	"slogexp/custom"

	"github.com/google/uuid"
)

func main() {
	u := NewUser(1, "testpass", "exampleemail@test.com")
	fmt.Println("========= 自作Handler =========")
	// 自作 Handlerお試し
	clogger := slog.New(custom.NewMyHandler(os.Stdout, nil))
	clogger = clogger.With("trace_id", "adkjk290jwe3w03233=")
	clogger.Info("slog体験", "count", 4)
	// => {"time":"2023-09-09T23:39:33.922167+09:00","level":"INFO","msg":"slog体験","trace_id":"adkjk290jwe3w03233=","count":4}
	clogger = clogger.WithGroup("user")
	clogger = clogger.With("service", "sns")
	clogger.Info("コメント投稿完了", "user_id", 1, "comment_id", 1000, "user", u)
	// => {"time":"2023-09-09T23:39:33.922331+09:00","level":"INFO","msg":"コメント投稿完了","trace_id":"adkjk290jwe3w03233=","user":{"service":"sns","user_id":1,"comment_id":1000}}
	clogger.Error("コメント投稿エラー", "user_id", 1, "comment", "ああああ", "error", fmt.Errorf("hogehoge").Error())
	// => {"time":"2023-09-09T23:42:00.869847+09:00","level":"ERROR","msg":"コメント投稿エラー","trace_id":"adkjk290jwe3w03233=","user":{"service":"sns","user_id":1,"comment":"ああああ","error":"hogehoge"}}

	// 組み込みのJSONHandlerカスタマイズ
	fmt.Println("========= 組み込みJSONHandler =========")
	opts := slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "時間"
				return a
			}
			if a.Value.Kind() == slog.KindString {
				// いろいろできる
			}

			return a
		},
	}
	dlogger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	// Userをまるごとログに渡したときもPassword, Emailを伏せ字にしたいなら、UserもLogValuerを実装している必要がある
	dlogger.Info("ユーザー登録完了", "user", u)
	dlogger.Info("ユーザー登録完了", "pass", u.Password, "email", u.Email)
	// trace_idをつけてみる
	ctx := withTraceID(context.Background())
	dlogger = withTraceIDLogger(ctx, dlogger)
	dlogger.Info("/hogehoge API request")
	dlogger.Info("/hogehoge API response")
}

type _CtxKeyTypeTraceID struct{}

var ctxTraceIDKey = _CtxKeyTypeTraceID{}

func withTraceID(ctx context.Context) context.Context {
	traceID, _ := uuid.NewUUID()
	return context.WithValue(ctx, ctxTraceIDKey, traceID.String())
}

func withTraceIDLogger(ctx context.Context, l *slog.Logger) *slog.Logger {
	v := ctx.Value(ctxTraceIDKey).(string)
	return l.With("traceID", v)
}

type (
	User struct {
		ID       int
		Password Password
		Email    Email
	}
	Password string
	Email    string
)

func NewUser(id int, pass Password, email Email) *User {
	return &User{
		ID:       id,
		Password: pass,
		Email:    email,
	}
}

func (p Password) LogValue() slog.Value {
	return slog.StringValue("**********")
}

func (e Email) LogValue() slog.Value {
	return slog.StringValue("*****@****")
}

func (u *User) LogValue() slog.Value {
	return slog.AnyValue(User{
		ID:       u.ID,
		Password: "**********",
		Email:    "*****@****",
	})
}
