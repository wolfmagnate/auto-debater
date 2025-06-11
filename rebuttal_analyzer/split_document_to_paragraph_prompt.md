# タスク
与えれらた文章は何らかの反論を行うためのものです。
文章を意味的なまとまりに区切り、数百文字程度の段落に分割して下さい。

# 入力
分析対象の文章が与えられます。

# 出力
以下の形式のJSONにしてください。
```go
type SplittedDocument struct {
    Paragraphs []string `json:"paragraphs"`
}
```

# 注意点
Paragraphsを結合した内容は入力した文章を完全に一致するようにしてください。

## 分析対象の反論

{{.Rebuttal}}
