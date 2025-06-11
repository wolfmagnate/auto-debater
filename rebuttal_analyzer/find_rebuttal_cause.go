package rebuttal_analyzer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"

	_ "embed"

	"github.com/wolfmagnate/auto_debater/domain"
	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed find_rebuttal_cause_prompt.md
var findRebuttalCausePromptMarkdown string

type FindRebuttalCauseTemplateData struct {
	DebateGraphJSON string
	Rebuttal        string
	TargetArgument  string
}

type RebuttalCauseFinder struct {
	tmpl *template.Template
}

func CreateRebuttalCauseFinder() (*RebuttalCauseFinder, error) {
	tmpl, err := template.New("prompt").Parse(findRebuttalCausePromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &RebuttalCauseFinder{tmpl: tmpl}, nil
}

type FoundCauses struct {
	Causes []string `json:"causes"`
}

func (finder *RebuttalCauseFinder) FindRebuttalCauses(ctx context.Context, debateGraph *domain.DebateGraph, rebuttal string, targetArgument string) (*FoundCauses, error) {
	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("DebateGraphのJSON作成に失敗しました: %w", err)
	}
	data := FindRebuttalCauseTemplateData{
		DebateGraphJSON: debateGraphJSON,
		Rebuttal:        rebuttal,
		TargetArgument:  targetArgument,
	}

	var processedPrompt bytes.Buffer
	err = finder.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	thinkingBudget := int32(24_000)
	foundCauses, _, err := infra.ChatCompletionHandler[FoundCauses](ctx, promptString, &thinkingBudget)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return foundCauses, nil
}
