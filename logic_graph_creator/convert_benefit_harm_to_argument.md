# 指示
与えられた「誰が」「どのような」メリット・デメリットを受けるという形式のJSONを、自然な日本語の主張に変換してください。

# 入力形式

```go
type BenefitHarm struct {
	Who  string `json:"who"`
	What string `json:"what"`
}
```

# 出力形式

```go
type ArgumentText struct {
    Argument string `json:"argument"`
}
```

# 注意点
生成される文章はメリット・デメリットを表現した文章です。そのため、なるべく文章は簡潔に保ちつつメリット・デメリットに見えるような文章にしてください。ただし、勝手に入力に無い内容を追加してはいけません。

# 具体例
入力

```json
{
    "who": "家庭",
    "what": "燃料費の削減"
}
```

期待される出力

```json
{
    "argument": "家庭の燃料費を削減できる"
}
```

メリットなので、「できる」という文体にしました。これが「家庭が燃料費を削減する」だと単なる動作になってしまいます。

# 変換対象のJSON

{{.BenefitHarmJSON}}
