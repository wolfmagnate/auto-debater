package logic_composer

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"text/template"

	"github.com/wolfmagnate/auto_debater/domain"
	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed enhance_todo_prompt.md
var enhanceTODOPromptMarkdown string

type TODOEnhancer struct {
	tmpl *template.Template
}

func CreateTODOEnhancer() (*TODOEnhancer, error) {
	tmpl, err := template.New("prompt").Parse(enhanceTODOPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &TODOEnhancer{tmpl: tmpl}, nil
}

type EnhanceTODOTemplateData struct {
	DebateGraphJSON       string
	TargetDebateGraphJSON string
}

type TODOSuggestions struct {
	TODOs []EnhancementTODO `json:"todo"`
}

type EnhancementTODO struct {
	Title string `json:"title"`

	StrengthenEdge *StrengthenEdgePayload `json:"strengthen_edge,omitempty"`
	StrengthenNode *StrengthenNodePayload `json:"strengthen_node,omitempty"`
	InsertNode     *InsertNodePayload     `json:"insert_node,omitempty"`
}

// ノードの強化
type StrengthenNodePayload struct {
	TargetArgument string `json:"target_argument"`
	// 具体的にPMFを達成するに値する課題とメリットなのか
	Content string `json:"content"`
}

func (enhancer *TODOEnhancer) EnhanceTODO(
	ctx context.Context,
	debateGraph *domain.DebateGraph,
	subGraph *domain.DebateGraph) (*TODOSuggestions, error) {
	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("グラフのJSON化に失敗しました: %w", err)
	}

	targetDebateGraphJSON, err := subGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("グラフのJSON化に失敗しました: %w", err)
	}

	data := EnhanceTODOTemplateData{
		DebateGraphJSON:       debateGraphJSON,
		TargetDebateGraphJSON: targetDebateGraphJSON,
	}

	var processedPrompt bytes.Buffer
	err = enhancer.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	todo, _, err := infra.ChatCompletionHandler[TODOSuggestions](ctx, promptString, nil)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return todo, nil
}
