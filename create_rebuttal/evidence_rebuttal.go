package createrebuttal

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

//go:embed evidence_rebuttal_prompt.md
var evidenceRebuttalPromptMarkdown string

type EvidenceRebuttalFinder struct {
	tmpl *template.Template
}

func CreateEvidenceRebuttalFinder() (*EvidenceRebuttalFinder, error) {
	tmpl, err := template.New("prompt").Parse(evidenceRebuttalPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &EvidenceRebuttalFinder{tmpl: tmpl}, nil
}

type FindEvidenceRebuttalTemplateData struct {
	DebateGraphJSON  string
	TargetCauseNode  string
	TargetEffectNode string
	TargetEdge       string
}

type EvidenceRebuttals struct {
	Rebuttals []EvidenceRebuttal `json:"rebuttals"`
}
type EvidenceRebuttal struct {
	// RebuttalType は、指摘の種類を示します。"certainty"（確実性）または
	// "uniqueness"（独自性）のいずれかが入ります。
	RebuttalType string `json:"rebuttal_type"`

	// Rebuttal は、具体的に不足している証拠の内容と、
	// それを補うために必要とされる調査を簡潔に記述します。
	Rebuttal string `json:"rebuttal"`
}

func (finder *EvidenceRebuttalFinder) FindeEvidenceRebuttalFinder(
	ctx context.Context,
	debateGraph *domain.DebateGraph,
	targetCauseNode *domain.DebateGraphNode,
	targetEffectNode *domain.DebateGraphNode,
	targetEdge *domain.DebateGraphEdge) (*EvidenceRebuttals, error) {

	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("JSONに変換に失敗しました: %w", err)
	}

	targetCauseNodeJSON, err := targetCauseNode.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("JSONに変換に失敗しました: %w", err)
	}

	targetEffectNodeJSON, err := targetEffectNode.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("JSONに変換に失敗しました: %w", err)
	}

	targetEdgeJSON, err := targetEdge.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("JSONに変換に失敗しました: %w", err)
	}

	data := FindEvidenceRebuttalTemplateData{
		DebateGraphJSON:  debateGraphJSON,
		TargetCauseNode:  targetCauseNodeJSON,
		TargetEffectNode: targetEffectNodeJSON,
		TargetEdge:       targetEdgeJSON,
	}

	var processedPrompt bytes.Buffer
	err = finder.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	rebuttals, _, err := infra.ChatCompletionHandler[EvidenceRebuttals](ctx, promptString, nil)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return rebuttals, nil
}
