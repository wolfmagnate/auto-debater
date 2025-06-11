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

//go:embed enhance_logic_prompt.md
var enhanceLogicPromptMarkdown string

type LogicEnhancer struct {
	tmpl *template.Template
}

func CreateLogicEnhancer() (*LogicEnhancer, error) {
	tmpl, err := template.New("prompt").Parse(enhanceLogicPromptMarkdown)

	if err != nil {
		return nil, fmt.Errorf("起動時のテンプレート解析に失敗しました: %w", err)
	}

	return &LogicEnhancer{tmpl: tmpl}, nil
}

type EnhanceLogicTemplateData struct {
	DebateGraphJSON       string
	TargetDebateGraphJSON string
}

type EnhancementAction struct {
	StrengthenEdge *StrengthenEdgePayload `json:"strengthen_edge,omitempty"`
	InsertNode     *InsertNodePayload     `json:"insert_node,omitempty"`
}

type StrengthenEdgePayload struct {
	CauseArgument  string `json:"cause_argument"`
	EffectArgument string `json:"effect_argument"`
	// EnhancementType は "uniqueness" または "certainty" のいずれかです。
	EnhancementType string `json:"enhancement_type"`
	Content         string `json:"content"`
}

type InsertNodePayload struct {
	CauseArgument        string `json:"cause_argument"`
	EffectArgument       string `json:"effect_argument"`
	IntermediateArgument string `json:"intermediate_argument"`
}

func (enhancer *LogicEnhancer) EnhanceLogic(ctx context.Context, debateGraph *domain.DebateGraph, cause, effect string) ([]EnhancementAction, error) {
	subGraph := domain.NewDebateGraph()
	causeNode := domain.NewDebateGraphNode(cause, false)
	effectNode := domain.NewDebateGraphNode(effect, false)
	if err := subGraph.AddNode(causeNode); err != nil {
		return nil, err
	}
	if err := subGraph.AddNode(effectNode); err != nil {
		return nil, err
	}
	edge := domain.NewDebateGraphEdge(causeNode, effectNode, false)
	if err := subGraph.AddEdge(edge); err != nil {
		return nil, err
	}

	// 全体の議論グラフはループ内で不変なので、最初にJSON化します。
	debateGraphJSON, err := debateGraph.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("グラフのJSON化に失敗しました: %w", err)
	}

	// 成功した強化策を格納するためのスライスを初期化します。
	enhancements := make([]EnhancementAction, 0, 3)

	// 強化のプロセスを3回繰り返します。
	for i := 0; i < 3; i++ {
		// ループの都度、更新されたサブグラフをJSON化します。
		subGraphJSON, err := subGraph.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("ループ%d回目のサブグラフのJSON化に失敗しました: %w", i+1, err)
		}

		data := EnhanceLogicTemplateData{
			DebateGraphJSON:       debateGraphJSON,
			TargetDebateGraphJSON: subGraphJSON,
		}

		var processedPrompt bytes.Buffer
		if err := enhancer.tmpl.Execute(&processedPrompt, data); err != nil {
			log.Printf("ループ%d回目のテンプレートの実行に失敗しました: %v", i+1, err)
			return nil, fmt.Errorf("ループ%d回目のテンプレートの実行に失敗しました: %w", i+1, err)
		}

		// AIに次の強化策を問い合わせます。
		enhancement, _, err := infra.ChatCompletionHandler[EnhancementAction](ctx, processedPrompt.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("ループ%d回目のAIモデルの呼び出しに失敗しました: %w", i+1, err)
		}

		if enhancement == nil {
			return nil, fmt.Errorf("ループ%d回目にAIから有効な強化策が返されませんでした", i+1)
		}

		// AIから受け取った強化策をサブグラフに適用します。
		if payload := enhancement.InsertNode; payload != nil {
			// --- 中間ノードの挿入 ---
			causeNode, _ := subGraph.GetNode(payload.CauseArgument)
			effectNode, _ := subGraph.GetNode(payload.EffectArgument)

			intermediateNode := domain.NewDebateGraphNode(payload.IntermediateArgument, false)
			if err := subGraph.AddNode(intermediateNode); err != nil {
				return nil, fmt.Errorf("ループ%d回目, 中間ノード '%s' の追加に失敗しました: %w", i+1, payload.IntermediateArgument, err)
			}
			if err := subGraph.RemoveEdge(payload.CauseArgument, payload.EffectArgument); err != nil {
				return nil, fmt.Errorf("ループ%d回目, 元のエッジ '%s -> %s' の削除に失敗しました: %w", i+1, payload.CauseArgument, payload.EffectArgument, err)
			}
			edge1 := domain.NewDebateGraphEdge(causeNode, intermediateNode, false)
			if err := subGraph.AddEdge(edge1); err != nil {
				return nil, fmt.Errorf("ループ%d回目, 新しいエッジ '%s -> %s' の追加に失敗しました: %w", i+1, payload.CauseArgument, payload.IntermediateArgument, err)
			}
			edge2 := domain.NewDebateGraphEdge(intermediateNode, effectNode, false)
			if err := subGraph.AddEdge(edge2); err != nil {
				return nil, fmt.Errorf("ループ%d回目, 新しいエッジ '%s -> %s' の追加に失敗しました: %w", i+1, payload.IntermediateArgument, payload.EffectArgument, err)
			}
		} else if payload := enhancement.StrengthenEdge; payload != nil {
			// --- 既存エッジの強化 ---
			edge, exists := subGraph.GetEdge(payload.CauseArgument, payload.EffectArgument)
			if !exists {
				return nil, fmt.Errorf("ループ%d回目, 強化対象のエッジ '%s -> %s' が見つかりません", i+1, payload.CauseArgument, payload.EffectArgument)
			}
			switch payload.EnhancementType {
			case "uniqueness":
				edge.Uniqueness = append(edge.Uniqueness, payload.Content)
			case "certainty":
				edge.Certainty = append(edge.Certainty, payload.Content)
			default:
				return nil, fmt.Errorf("ループ%d回目, 不明なエッジ強化タイプです: '%s'", i+1, payload.EnhancementType)
			}
		} else {
			return nil, fmt.Errorf("ループ%d回目にAIから返された強化策の形式が不正です", i+1)
		}

		// 適用に成功した強化策を結果のスライスに追加します。
		enhancements = append(enhancements, *enhancement)
	}

	// 3回の強化策をすべて返します。
	return enhancements, nil
}
