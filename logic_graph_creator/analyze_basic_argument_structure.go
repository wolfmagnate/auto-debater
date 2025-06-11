package logic_graph_creator

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"text/template"

	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed analyze_basic_argument_structure_prompt.md
var basicAnalysisPromptMarkdown string

type BasicArgumentStructure struct {
	IsArgument      bool   `json:"is_argument"`
	StatusQuo       string `json:"status_quo"`
	AffirmativePlan string `json:"affirmative_plan"`
	Position        string `json:"position"` // "status_quo" または "affirmative_plan"
}

func ConvertBasicArgumentStructureToJSON(bas *BasicArgumentStructure) (string, error) {
	jsonBytes, err := json.Marshal(bas)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

type BasicStructureAnalysisTemplateData struct {
	Document string
}

type BasicStructureAnalyzer struct {
	tmpl *template.Template
}

func CreateBasicStructureAnalyzer() (*BasicStructureAnalyzer, error) {
	tmpl, err := template.New("prompt").Parse(basicAnalysisPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &BasicStructureAnalyzer{tmpl: tmpl}, nil
}

func (analyzer *BasicStructureAnalyzer) AnalyzeBasicArgumentStructure(ctx context.Context, document string) (*BasicArgumentStructure, error) {
	data := BasicStructureAnalysisTemplateData{
		Document: document,
	}

	var processedPrompt bytes.Buffer
	err := analyzer.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	analysisResult, _, err := infra.ChatCompletionHandler[BasicArgumentStructure](ctx, promptString, nil)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return analysisResult, nil
}
