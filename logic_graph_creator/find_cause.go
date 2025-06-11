package logic_graph_creator

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"text/template"

	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed find_cause_prompt.md
var findCausePromptMarkdown string

type CauseFinder struct {
	tmpl *template.Template
}

func CreateCauseFinder() (*CauseFinder, error) {
	tmpl, err := template.New("prompt").Parse(findCausePromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &CauseFinder{tmpl: tmpl}, nil
}

type FindCauseTemplateData struct {
	Document               string
	BasicArgumentStructure string
	TargetArgument         string
}

type FoundCauses struct {
	Causes []string `json:"causes"`
}

func (finder *CauseFinder) FindCauses(ctx context.Context, document string, basicArgumentStructure *BasicArgumentStructure, targetArgument string) (*FoundCauses, error) {
	basicArgumentStructureString, err := ConvertBasicArgumentStructureToJSON(basicArgumentStructure)
	if err != nil {
		return nil, fmt.Errorf("BasicArgumentStructureのJSON文字列変換に失敗しました: %w", err)
	}

	data := FindCauseTemplateData{
		Document:               document,
		BasicArgumentStructure: basicArgumentStructureString,
		TargetArgument:         targetArgument,
	}

	var processedPrompt bytes.Buffer
	err = finder.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	// 原因の解析は難しいタスクなので思考させる
	thinkingBudget := int32(24_000)
	foundCauses, _, err := infra.ChatCompletionHandler[FoundCauses](ctx, promptString, &thinkingBudget)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return foundCauses, nil
}
