package debate_graph_creator

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"text/template"

	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed split_document_to_paragraph_prompt.md
var splitDocumentToParagraphPromptMarkdown string

type DocumentSplitter struct {
	tmpl *template.Template
}

func CreateDocumentSplitter() (*DocumentSplitter, error) {
	tmpl, err := template.New("prompt").Parse(splitDocumentToParagraphPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &DocumentSplitter{tmpl: tmpl}, nil
}

type SplitDocumentToParagraphTemplateData struct {
	Document string
}

type SplittedDocument struct {
	Paragraphs []string `json:"paragraphs"`
}

func (splitter *DocumentSplitter) SplitDocumentToParagraph(ctx context.Context, document string) (*SplittedDocument, error) {

	data := SplitDocumentToParagraphTemplateData{
		Document: document,
	}

	var processedPrompt bytes.Buffer
	err := splitter.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	SplittedDocument, _, err := infra.ChatCompletionHandler[SplittedDocument](ctx, promptString, nil)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return SplittedDocument, nil
}
