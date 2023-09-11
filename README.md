# 概要
golangに追加されたslogを試してみるリポジトリ

# custom
組み込みのJSONHandlerを参考に自作した Handler  
色々考慮することが多いので実務でslogのHandlerを自作することはなさそう。  
JSONHandlerで十分。  

# main.go
自作のHanlderや組み込みのJSONHandlerを使ったログ出力処理

# slog感想
Hanlderを自作するのはとっても大変。組み込みのHanlderで色々できるのでゼロから自作することはないはず。

よく使いそうなカスタマイズ

- `Logger.With` を使って traceIDやsessionIDのような共通項目を出力させることができる
- `LogValuer` インターフェースを実装（`LogValue()` メソッド）してパスワードとか機密性の高い情報をマスクする
- `ReplaceAttr` オプションで出力内容をなんかいじる

# その他
[Go 1.21連載始まります＆slogをどう使うべきか](https://future-architect.github.io/articles/20230731a/) が勉強になった

何か凝ったカスタマイズが必要になったら、記事にあるようにJSONHandlerのラッパーを作ると良いかもしれない。

以下 [Go 1.21連載始まります＆slogをどう使うべきか](https://future-architect.github.io/articles/20230731a/) より引用

```
// ハンドラーのラッパー
type WriteTraceIDHandler struct {
	parent slog.Handler
}

func WithWriteTraceIDHandler(parent slog.Handler) *WriteTraceIDHandler {
	return &WriteTraceIDHandler{
		parent: parent,
	}
}

// ログ出力に情報を付与するメソッド
func (h *WriteTraceIDHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Add(slog.String("traceID", ctx.Value(ctxKey).(string)))
	return h.parent.Handle(ctx, record)
}

func (h *WriteTraceIDHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.parent.Enabled(ctx, level)
}

func (h *WriteTraceIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &WriteTraceIDHandler{h.parent.WithAttrs(attrs)}
}

func (h *WriteTraceIDHandler) WithGroup(name string) slog.Handler {
	return &WriteTraceIDHandler{h.parent.WithGroup(name)}
}
```
