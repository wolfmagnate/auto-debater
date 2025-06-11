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

//go:embed find_rebuttals_prompt.md
var findRebuttalsPromptMarkdown string

type FindRebuttalsTemplateData struct {
	DebateGraphJSON string
	Rebuttal        string
}

type RebuttalFinder struct {
	tmpl *template.Template
}

func CreateRebuttalFinder() (*RebuttalFinder, error) {
	tmpl, err := template.New("prompt").Parse(findRebuttalsPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &RebuttalFinder{tmpl: tmpl}, nil
}

// エッジに対する反論
type EdgeRebuttal struct {
	RebuttalType       string `json:"rebuttal_type"`      // "certainty"または"uniqueness"のいずれか
	TargetEdgeCause    string `json:"target_edge_cause"`  // どのエッジに反論するか
	TargetEdgeEffect   string `json:"target_edge_effect"` // どのエッジに反論するか
	CertaintyRebuttal  string `json:"certainty_rebuttal"`
	UniquenessRebuttal string `json:"uniqueness_rebuttal"`
}

// ノードに対する反論
type NodeRebuttal struct {
	RebuttalType       string `json:"rebuttal_type"` // "importance"または"uniqueness"のいずれか
	TargetNode         string `json:"target_node"`   // どのノードに反論するか
	ImportanceRebuttal string `json:"importance_rebuttal"`
	UniquenessRebuttal string `json:"uniqueness_rebuttal"`
}

// あるノードと逆の主張をする議論
// 例えば"男女共学は生徒の学力を改善する"に対応して"男女共学は生徒の学力を低下させる"という議論を行う
type CounterArgument struct {
	TargetNode string `json:"target_node"` // どの主張に反対しているか。例：原発再稼働は経済的に良い
	Argument   string `json:"argument"`    // 反論に含まれる主張。例：原発再稼働は経済に悪影響をもたらす
}

// 相手の議論を逆利用した議論。ディベートにおけるTurnと呼ばれる反論
// 例えば、小さな政府か大きな政府かという論題において、大きな政府側の「法人税減税は、企業の内部留保を増やすだけで、格差を拡大させる」に対応して「法人税減税は投資や雇用を活発にし経済を活性化させる」という反論をすると、途中まで相手の「法人税減税」という理屈を認めつつ、途中から新しいノード「投資や雇用の活発化」を追加してメリットを生みだしている
type TurnArgument struct {
	TargetCauseNodes []string `json:"target_cause_nodes"` // 途中まで元の主張の議論構造グラフのノードを認めている。どこまで認めているか。
	EffectArgument   string   `json:"effect_argument"`    // 反論では最終的にどのようなメリット・デメリットを主張しているか
}
type RebuttalItem struct {
	// RebuttalKind はこのアイテムがどの種類の反論であるかを示します。以下のいずれかの値のみをとります。
	// "edge_rebuttal", "node_rebuttal", "counter_argument", "turn_argument"
	RebuttalKind string `json:"rebuttal_kind"`

	// RebuttalKindに対応するものが1つだけ設定されます。
	EdgeRebuttal    *EdgeRebuttal    `json:"edge_rebuttal,omitempty"`
	NodeRebuttal    *NodeRebuttal    `json:"node_rebuttal,omitempty"`
	CounterArgument *CounterArgument `json:"counter_argument,omitempty"`
	TurnArgument    *TurnArgument    `json:"turn_argument,omitempty"`
}

// AnalyzedRebuttals は、すべての反論分析結果を格納するトップレベルの構造体です。
type AnalyzedRebuttals struct {
	Rebuttals []RebuttalItem `json:"rebuttals"`
}

func (finder *RebuttalFinder) FindRebuttals(ctx context.Context, debateGraph *domain.DebateGraph, rebuttal string) (*AnalyzedRebuttals, error) {
	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("DebateGraphのJSON作成に失敗しました: %w", err)
	}
	data := FindRebuttalsTemplateData{
		DebateGraphJSON: debateGraphJSON,
		Rebuttal:        rebuttal,
	}

	var processedPrompt bytes.Buffer
	err = finder.tmpl.Execute(&processedPrompt, data)
	if err != nil {
		log.Printf("テンプレートの実行に失敗しました: %v", err)
		return nil, fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	promptString := processedPrompt.String()

	thinkingBudget := int32(24_000)
	analyzedRebuttals, _, err := infra.ChatCompletionHandler[AnalyzedRebuttals](ctx, promptString, &thinkingBudget)
	if err != nil {
		return nil, fmt.Errorf("AIモデルの呼び出しに失敗しました: %w", err)
	}

	return analyzedRebuttals, nil
}
