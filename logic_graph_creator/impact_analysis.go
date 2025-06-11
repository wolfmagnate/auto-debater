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

//go:embed impact_analysis_prompt.md
var impactAnalysisPromptMarkdown string

type ImpactAnalyzer struct {
	tmpl *template.Template
}

func CreateImpactAnalyzer() (*ImpactAnalyzer, error) {
	tmpl, err := template.New("prompt").Parse(impactAnalysisPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &ImpactAnalyzer{tmpl: tmpl}, nil
}

type ImpactAnalysisTemplateData struct {
	Document               string
	BasicArgumentStructure string
}

type BenefitHarm struct {
	Who  string `json:"who"`
	What string `json:"what"`
}

type PlanAnalysis struct {
	Benefits []BenefitHarm `json:"benefits"`
	Harms    []BenefitHarm `json:"harms"`
}

type ImpactAnalysis struct {
	StatusQuo       PlanAnalysis `json:"status_quo"`
	AffirmativePlan PlanAnalysis `json:"affirmative_plan"`
}

func (analyzer *ImpactAnalyzer) AnalyzeImpact(ctx context.Context, document string, basicArgumentStructure *BasicArgumentStructure) (*ImpactAnalysis, error) {
	basicArgumentStructureString, err := ConvertBasicArgumentStructureToJSON(basicArgumentStructure)
	if err != nil {
		return nil, fmt.Errorf("BasicArgumentStructureのJSON文字列変換に失敗しました: %w", err)
	}

	data := ImpactAnalysisTemplateData{
		Document:               document,
		BasicArgumentStructure: basicArgumentStructureString,
	}

	var processedPrompt bytes.Buffer
	err = analyzer.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	thinkingBudget := int32(24_000)
	analysisResult, _, err := infra.ChatCompletionHandler[ImpactAnalysis](ctx, promptString, &thinkingBudget)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return analysisResult, nil
}
