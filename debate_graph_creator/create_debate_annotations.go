package debate_graph_creator

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/wolfmagnate/auto_debater/domain"
	"github.com/wolfmagnate/auto_debater/infra"
)

//go:embed create_debate_annotations_prompt.md
var creteDebateAnnotationsPromptMarkdown string

type DebateAnnotationCreator struct {
	tmpl *template.Template
}

func CreateDebateAnnotationCreator() (*DebateAnnotationCreator, error) {
	tmpl, err := template.New("prompt").Parse(creteDebateAnnotationsPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &DebateAnnotationCreator{tmpl: tmpl}, nil
}

type CreateDebateAnnotationTemplateData struct {
	Document        string
	TargetParagraph string
	LogicGraphNodes string
	LogicGraphEdges string
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
	AnnotationType     string `json:"annotation_type"`     // "argument"または"importance"または"uniqueness"または"importance_rebuttal"または"uniqueness_rebuttal"のいずれか
	Argument           string `json:"argument"`            // アノテーションを行う対象の論理構造グラフのノード
	Importance         string `json:"importance"`          // なぜArgumentが重要であるかの理由を表す文章。AnnotationTypeが"importance"のときのみ有効
	Uniqueness         string `json:"uniqueness"`          // なぜArgumentがStatus QuoまたはAffirmative Planでのみ発生するのかの理由を表す文章。AnnotationTypeが"uniqueness"のときのみ有効
	ImportanceRebuttal string `json:"importance_rebuttal"` // なぜArgumentが重要ではないのかの理由を表す文章。AnnotationTypeが"importance_rebuttal"のときのみ有効
	UniquenessRebuttal string `json:"uniqueness_rebuttal"` // なぜArgumentがStatus QuoとAffirmative Planの両方で発生してしまうかの理由を表す文章。AnnotationTypeが
}

type EdgeAnnotation struct {
	AnnotationType     string `json:"annotation_type"`     // "certainty"または"uniqueness"または"certainty_rebuttal"または"uniqueness_rebuttal"のいずれか
	CauseArgument      string `json:"cause_argument"`      // エッジの原因に対応する論理構造グラフのノード
	EffectArgument     string `json:"effect_argument"`     // エッジの結果に対応する論理構造グラフのノード
	Certainty          string `json:"certainty"`           // なぜCauseArgumentがEffectArgumentを引き起こす可能性が高いのかの理由を表す文章。AnnotationTypeが"certainty"のときのみ有効
	Uniqueness         string `json:"uniqueness"`          // なぜCauseArgumentがStatus QuoまたはAffirmative Planでのみ発生するのかの理由を表す文章。AnnotationTypeが"uniqueness"のときのみ有効
	CertaintyRebuttal  string `json:"certainty_rebuttal"`  // なぜCauseArgumentがEffectArgumentを発生させる可能性が低いのかを表す文章。AnnotationTypeが"certainty_rebuttal"のときのみ有効
	UniquenessRebuttal string `json:"uniqueness_rebuttal"` // なぜCauseArgumentがStatus QuoとAffirmative Planの両方でEffectArgumentを引き起こすのかの理由を表す文章。AnnotationTypeが"uniqueness_rebuttal"のときのみ有効
}

func (analyzer *DebateAnnotationCreator) CreateDebateAnnotations(ctx context.Context, document string, targetParagraph string, logicGraph *domain.LogicGraph) (*LogicAnnotations, error) {
	nodes := make([]string, 0)
	for _, node := range logicGraph.Nodes {
		nodes = append(nodes, node.Argument)
	}

	causalRelationships := domain.ListAllCausalRelationships(logicGraph)

	data := CreateDebateAnnotationTemplateData{
		Document:        document,
		TargetParagraph: targetParagraph,
		LogicGraphNodes: strings.Join(nodes, "\n"),
		LogicGraphEdges: strings.Join(causalRelationships, "\n"),
	}

	var processedPrompt bytes.Buffer
	err := analyzer.tmpl.Execute(&processedPrompt, data)
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
