package rebuttal_analyzer

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

//go:embed create_rebuttal_annotations_prompt.md
var creteRebuttalAnnotationsPromptMarkdown string

type RebuttalAnnotationCreator struct {
	tmpl *template.Template
}

func CreateRebuttalAnnotationCreator() (*RebuttalAnnotationCreator, error) {
	tmpl, err := template.New("prompt").Parse(creteRebuttalAnnotationsPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &RebuttalAnnotationCreator{tmpl: tmpl}, nil
}

type CreateRebuttalAnnotationTemplateData struct {
	Rebuttal        string
	TargetParagraph string
	DebateGraphJSON string
}

type LogicAnnotations struct {
	Annotations []LogicAnnotation `json:"annotations"` // 分析対象の段落に含まれる全ての論理構造グラフの要素の分析結果
}

type LogicAnnotation struct {
	TargetType     string         `json:"target_type"`     // "node"または"edge"のいずれか
	TargetText     string         `json:"target_text"`     // 分析対象の段落のうち、このアノテーションを行う根拠となる部分
	NodeAnnotation NodeAnnotation `json:"node_annotation"` // TargetTypeが"node"のときのみ有効
	EdgeAnnotation EdgeAnnotation `json:"edge_annotation"` // TargetTypeが"edge"のときのみ有効
}

type NodeAnnotation struct {
	AnnotationType string `json:"annotation_type"` // "argument"または"importance"または"uniqueness"または"importance_rebuttal"または"uniqueness_rebuttal"のいずれか
	Argument       string `json:"argument"`        // アノテーションを行う対象の論理構造グラフのノード
	Importance     string `json:"importance"`      // なぜArgumentが重要であるかの理由を表す文章。AnnotationTypeが"importance"のときのみ有効
	Uniqueness     string `json:"uniqueness"`      // なぜArgumentがStatus QuoまたはAffirmative Planでのみ発生するのかの理由を表す文章。AnnotationTypeが"uniqueness"のときのみ有効
}

type EdgeAnnotation struct {
	AnnotationType string `json:"annotation_type"` // "certainty"または"uniqueness"または"certainty_rebuttal"または"uniqueness_rebuttal"のいずれか
	CauseArgument  string `json:"cause_argument"`  // エッジの原因に対応する論理構造グラフのノード
	EffectArgument string `json:"effect_argument"` // エッジの結果に対応する論理構造グラフのノード
	Certainty      string `json:"certainty"`       // なぜCauseArgumentがEffectArgumentを引き起こす可能性が高いのかの理由を表す文章。AnnotationTypeが"certainty"のときのみ有効
	Uniqueness     string `json:"uniqueness"`      // なぜCauseArgumentがStatus QuoまたはAffirmative Planでのみ発生するのかの理由を表す文章。AnnotationTypeが"uniqueness"のときのみ有効
}

func (analyzer *RebuttalAnnotationCreator) CreateRebuttalAnnotations(ctx context.Context, debateGraph *domain.DebateGraph, rebuttal, targetParagraph string) (*LogicAnnotations, error) {
	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("ディベートグラフのJSON化に失敗しました: %w", err)
	}
	data := CreateRebuttalAnnotationTemplateData{
		Rebuttal:        rebuttal,
		TargetParagraph: targetParagraph,
		DebateGraphJSON: debateGraphJSON,
	}

	var processedPrompt bytes.Buffer
	err = analyzer.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	thinkingBudget := int32(24_000)
	annotations, _, err := infra.ChatCompletionHandler[LogicAnnotations](ctx, promptString, &thinkingBudget)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}
	return annotations, nil
}
