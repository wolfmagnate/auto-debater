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

//go:embed pmf_rebuttal_prompt.md
var pmfRebuttalPromptMarkdown string

type PMFRebuttalFinder struct {
	tmpl *template.Template
}

func CreatePMFRebuttalFinder() (*PMFRebuttalFinder, error) {
	tmpl, err := template.New("prompt").Parse(pmfRebuttalPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &PMFRebuttalFinder{tmpl: tmpl}, nil
}

type FindPMFRebuttalTemplateData struct {
	DebateGraphJSON string
}

// PMFRebuttals は、事業計画のPMFに関する指摘事項を格納します。
type PMFRebuttals struct {
	StatusQuo       []PMFRebuttal `json:"status_quo"`
	AffirmativePlan []PMFRebuttal `json:"affirmative_plan"`
}

// PMFRebuttal は、論理構造グラフの特定の主張に対する具体的な指摘を格納します。
type PMFRebuttal struct {
	TargetArgument string `json:"target_argument"`
	Rebuttal       string `json:"rebuttal"`
}

func (finder *PMFRebuttalFinder) FindPMFRebuttal(
	ctx context.Context,
	debateGraph *domain.DebateGraph) (*PMFRebuttals, error) {

	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("JSONに変換に失敗しました: %w", err)
	}

	data := FindEvidenceRebuttalTemplateData{
		DebateGraphJSON: debateGraphJSON,
	}

	var processedPrompt bytes.Buffer
	err = finder.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	rebuttals, _, err := infra.ChatCompletionHandler[PMFRebuttals](ctx, promptString, nil)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return rebuttals, nil
}
