package logic_graph_creator

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	Document                string
	BasicArgumentStructure  string
	LogicGraphNodes         string
	LogicGraphEdges         string
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

func (finder *NewArgumentFinder) FindNewArguments(ctx context.Context, document string, basicArgumentStructure *BasicArgumentStructure, logicGraph *domain.LogicGraph, target *ArgumentAndCauses) (*FindNewArgumentsResult, error) {
	nodes := make([]string, 0)
	for _, node := range logicGraph.Nodes {
		nodes = append(nodes, node.Argument)
	}

	causalRelationships := domain.ListAllCausalRelationships(logicGraph)

	basicArgumentStructureString, err := ConvertBasicArgumentStructureToJSON(basicArgumentStructure)
	if err != nil {
		return nil, fmt.Errorf("BasicArgumentStructureのJSON文字列変換に失敗しました: %w", err)
	}

	targetArgumentAndCauseJSON, err := ConvertArgumentAndCausesToJSON(target)
	if err != nil {
		return nil, fmt.Errorf("ArgumentAndCausesのJSON文字列変換に失敗しました: %w", err)
	}

	data := FindNewArgumentsTemplateData{
		Document:                document,
		BasicArgumentStructure:  basicArgumentStructureString,
		LogicGraphNodes:         strings.Join(nodes, "\n"),
		LogicGraphEdges:         strings.Join(causalRelationships, "\n"),
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
