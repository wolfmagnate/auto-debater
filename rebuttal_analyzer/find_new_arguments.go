package rebuttal_analyzer

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"text/template"

	"github.com/wolfmagnate/auto_debater/domain"
	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed find_new_arguments_prompt.md
var findNewArgumentPromptMarkdown string

type NewArgumentFinder struct {
	tmpl *template.Template
}

func CreateNewArgumentFinder() (*NewArgumentFinder, error) {
	tmpl, err := template.New("prompt").Parse(findNewArgumentPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &NewArgumentFinder{tmpl: tmpl}, nil
}

type FindNewArgumentsTemplateData struct {
	DebateGraphJSON         string
	TargetArgumentAndCauses string
}

type ArgumentAndCauses struct {
	Argument string   `json:"argument"` // 主張
	Causes   []string `json:"causes"`   // 主張の原因が1つ以上ある
}

func ConvertArgumentAndCausesToJSON(item *ArgumentAndCauses) (string, error) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

type FindNewArgumentsResult struct {
	NewNodes   []string `json:"new_nodes"`
	UsedCauses []string `json:"used_causes"`
}

func (finder *NewArgumentFinder) FindNewArguments(ctx context.Context, debateGraph *domain.DebateGraph, target *ArgumentAndCauses) (*FindNewArgumentsResult, error) {
	targetArgumentAndCauseJSON, err := ConvertArgumentAndCausesToJSON(target)
	if err != nil {
		return nil, fmt.Errorf("ArgumentAndCausesのJSON文字列変換に失敗しました: %w", err)
	}

	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("DebateGraphのJSON文字列変換に失敗しました: %w", err)
	}

	data := FindNewArgumentsTemplateData{
		DebateGraphJSON:         debateGraphJSON,
		TargetArgumentAndCauses: targetArgumentAndCauseJSON,
	}

	var processedPrompt bytes.Buffer
	err = finder.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	thinkingBudget := int32(24_000)
	argumentText, _, err := infra.ChatCompletionHandler[FindNewArgumentsResult](ctx, promptString, &thinkingBudget)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return argumentText, nil
}
