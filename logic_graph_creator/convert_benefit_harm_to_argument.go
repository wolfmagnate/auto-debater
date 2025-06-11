package logic_graph_creator

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

//go:embed convert_benefit_harm_to_argument.md
var convertBenefitHarmToArgumentPromptMarkdown string

type BenefitHarmConverter struct {
	tmpl *template.Template
}

func CreateBenefitHarmConverter() (*BenefitHarmConverter, error) {
	tmpl, err := template.New("prompt").Parse(convertBenefitHarmToArgumentPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &BenefitHarmConverter{tmpl: tmpl}, nil
}

type ConvertBenefitHarmTemplateData struct {
	BenefitHarmJSON string
}

type ArgumentText struct {
	Argument string `json:"argument"`
}

func ConvertBenefitHarmToJSON(benefitHarm *BenefitHarm) (string, error) {
	jsonBytes, err := json.Marshal(benefitHarm)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (converter *BenefitHarmConverter) ConvertImpactAnalysisToArguments(ctx context.Context, impactAnalysis *ImpactAnalysis) ([]*domain.LogicGraphNode, error) {
	arguments := make([]*domain.LogicGraphNode, len(impactAnalysis.StatusQuo.Benefits)+len(impactAnalysis.StatusQuo.Harms)+len(impactAnalysis.AffirmativePlan.Benefits)+len(impactAnalysis.AffirmativePlan.Harms))
	index := 0

	// Status Quo Benefits
	for _, bh := range impactAnalysis.StatusQuo.Benefits {
		arg, err := converter.ConvertBenefitHarmToArgument(ctx, &bh)
		if err != nil {
			return nil, fmt.Errorf("status Quo BenefitのArgument変換に失敗しました: %w", err)
		}
		arguments[index] = domain.NewLogicGraphNode(arg)
		index++
	}

	// Status Quo Harms
	for _, bh := range impactAnalysis.StatusQuo.Harms {
		arg, err := converter.ConvertBenefitHarmToArgument(ctx, &bh)
		if err != nil {
			return nil, fmt.Errorf("status Quo HarmのArgument変換に失敗しました: %w", err)
		}
		arguments[index] = domain.NewLogicGraphNode(arg)
		index++
	}

	// Affirmative Plan Benefits
	for _, bh := range impactAnalysis.AffirmativePlan.Benefits {
		arg, err := converter.ConvertBenefitHarmToArgument(ctx, &bh)
		if err != nil {
			return nil, fmt.Errorf("affirmative Plan BenefitのArgument変換に失敗しました: %w", err)
		}
		arguments[index] = domain.NewLogicGraphNode(arg)
		index++
	}

	// Affirmative Plan Harms
	for _, bh := range impactAnalysis.AffirmativePlan.Harms {
		arg, err := converter.ConvertBenefitHarmToArgument(ctx, &bh)
		if err != nil {
			return nil, fmt.Errorf("affirmative Plan HarmのArgument変換に失敗しました: %w", err)
		}
		arguments[index] = domain.NewLogicGraphNode(arg)
		index++
	}

	return arguments, nil
}

func (converter *BenefitHarmConverter) ConvertBenefitHarmToArgument(ctx context.Context, benefitHarm *BenefitHarm) (string, error) {

	benefitHarmJSON, err := ConvertBenefitHarmToJSON(benefitHarm)
	if err != nil {
		return "", fmt.Errorf("benefitHarmのJSON文字列変換に失敗しました: %w", err)
	}

	data := ConvertBenefitHarmTemplateData{
		BenefitHarmJSON: benefitHarmJSON,
	}

	var processedPrompt bytes.Buffer
	err = converter.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return "", fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	argumentText, _, err := infra.ChatCompletionHandler[ArgumentText](ctx, promptString, nil)
	if err != nil {
		return "", fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return argumentText.Argument, nil
}
