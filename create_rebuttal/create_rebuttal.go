package createrebuttal

import (
	"context"
	"fmt"
	"log"

	"github.com/wolfmagnate/auto_debater/domain"
)

type RebuttalCreator struct {
	evidenceRebuttalFinder *EvidenceRebuttalFinder
	pmfRebuttalFinder      *PMFRebuttalFinder
}

type NodeRebuttalResult struct {
	TargetArgument   string `json:"target_argument"`
	RebuttalType     string `json:"rebuttal_type"`
	RebuttalArgument string `json:"rebuttal_argument"`
}

// EdgeRebuttalResult は、エッジへの反論提案を文字列ベースで保持します。
type EdgeRebuttalResult struct {
	TargetCauseArgument  string `json:"target_cause_argument"`
	TargetEffectArgument string `json:"target_effect_argument"`
	RebuttalType         string `json:"rebuttal_type"`
	RebuttalArgument     string `json:"rebuttal_argument"`
}

// CreateRebuttalResult は、生成された全ての反論提案を保持するトップレベルの構造体です。
type CreateRebuttalResult struct {
	NodeRebuttals []NodeRebuttalResult `json:"node_rebuttals"`
	EdgeRebuttals []EdgeRebuttalResult `json:"edge_rebuttals"`
}

func NewRebuttalCreator() (*RebuttalCreator, error) {
	finder, err := CreateEvidenceRebuttalFinder()
	if err != nil {
		return nil, fmt.Errorf("failed to create evidenceRebuttalFinder: %w", err)
	}

	pmfFinder, err := CreatePMFRebuttalFinder()
	if err != nil {
		return nil, fmt.Errorf("failed to create pmfRebuttalFinder: %w", err)
	}

	return &RebuttalCreator{
		evidenceRebuttalFinder: finder,
		pmfRebuttalFinder:      pmfFinder,
	}, nil
}

// CreateRebuttal は、subGraphに対する反論を生成し、その結果を構造体で返します。
// グラフ自体は変更しません。
func (creator *RebuttalCreator) CreateRebuttal(
	ctx context.Context,
	debateGraph *domain.DebateGraph,
	subGraph *domain.DebateGraph,
) (*CreateRebuttalResult, error) {
	result := &CreateRebuttalResult{
		NodeRebuttals: make([]NodeRebuttalResult, 0),
		EdgeRebuttals: make([]EdgeRebuttalResult, 0),
	}
	// --- ステップ1: subGraphの全てのエッジに対して証拠に基づく反論を生成 ---
	log.Println("--- Starting Evidence Rebuttal Creation for subGraph edges ---")
	subGraphEdges := subGraph.GetAllEdges()

	if len(subGraphEdges) > 0 {
		// ... (ログ出力部分は変更なし) ...
		for _, edge := range subGraphEdges {
			evidenceRebuttals, err := creator.evidenceRebuttalFinder.FindeEvidenceRebuttalFinder(ctx, debateGraph, edge.Cause, edge.Effect, edge)
			if err != nil {
				log.Printf("ERROR: Could not find evidence rebuttals for edge [%s] -> [%s]: %v", edge.Cause.Argument, edge.Effect.Argument, err)
				continue
			}

			if evidenceRebuttals != nil && len(evidenceRebuttals.Rebuttals) > 0 {
				log.Printf("SUCCESS: Found %d rebuttal(s) for edge [%s] -> [%s].", len(evidenceRebuttals.Rebuttals), edge.Cause.Argument, edge.Effect.Argument)
				for _, rebuttal := range evidenceRebuttals.Rebuttals {
					edgeResult := EdgeRebuttalResult{
						TargetCauseArgument:  edge.Cause.Argument,
						TargetEffectArgument: edge.Effect.Argument,
						RebuttalType:         rebuttal.RebuttalType,
						RebuttalArgument:     rebuttal.Rebuttal,
					}
					result.EdgeRebuttals = append(result.EdgeRebuttals, edgeResult)
					log.Printf("  - Proposing EdgeRebuttal for edge [%s] -> [%s] with argument [%s]", edge.Cause.Argument, edge.Effect.Argument, rebuttal.Rebuttal)
				}
			}
		}
	}
	log.Println("--- Finished Evidence Rebuttal Creation for subGraph edges ---")

	// --- ステップ2: subGraphに対するPMFの反論を生成 ---
	log.Println("--- Starting PMF Rebuttal Creation for subGraph ---")
	pmfRebuttals, err := creator.pmfRebuttalFinder.FindPMFRebuttal(ctx, debateGraph, subGraph)
	if err != nil {
		return nil, fmt.Errorf("could not find PMF rebuttals for the subGraph: %w", err)
	}

	if pmfRebuttals != nil {
		allPMFRebuttals := append(pmfRebuttals.StatusQuo, pmfRebuttals.AffirmativePlan...)
		if len(allPMFRebuttals) > 0 {
			log.Printf("SUCCESS: Found %d PMF rebuttal(s) for the subGraph.", len(allPMFRebuttals))
			// ...
			for _, pmfRebuttal := range allPMFRebuttals {
				rebuttalArgument := pmfRebuttal.Rebuttal
				// ターゲットノードの存在確認
				targetNode, exists := subGraph.GetNode(pmfRebuttal.TargetArgument)
				if !exists {
					log.Printf("WARN: Target node '%s' for PMF rebuttal not found in subGraph. Skipping.", pmfRebuttal.TargetArgument)
					continue
				}

				nodeResult := NodeRebuttalResult{
					TargetArgument:   targetNode.Argument,
					RebuttalType:     "importance",
					RebuttalArgument: rebuttalArgument,
				}
				result.NodeRebuttals = append(result.NodeRebuttals, nodeResult)
				log.Printf("  - Proposing NodeRebuttal for node [%s] with argument [%s]", targetNode.Argument, rebuttalArgument)
			}
		}
	}
	log.Println("--- Finished PMF Rebuttal Creation for subGraph ---")

	return result, nil
}
