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

func (creator *RebuttalCreator) CreateRebuttal(
	ctx context.Context,
	debateGraph *domain.DebateGraph, // コンテキストとして使用
	subGraph *domain.DebateGraph, // 反論の生成と保存の対象
) {
	// --- ステップ1: subGraphの全てのエッジに対して証拠に基づく反論を生成 ---
	log.Println("--- Starting Evidence Rebuttal Creation for subGraph edges ---")
	subGraphEdges := subGraph.GetAllEdges()

	if len(subGraphEdges) == 0 {
		log.Println("No edges found in the subGraph. Nothing to rebut.")
		log.Println("--- Finished Evidence Rebuttal Creation for subGraph edges ---")
		return
	}

	log.Printf("Found %d edges in subGraph to analyze for rebuttals.", len(subGraphEdges))
	fmt.Println()

	for i, edge := range subGraphEdges {
		log.Printf("Analyzing subGraph edge %d/%d: [%s] -> [%s]", i+1, len(subGraphEdges), edge.Cause.Argument, edge.Effect.Argument)

		// 全体のdebateGraphをコンテキストとして渡し、subGraphのedgeを対象に反論を探す
		evidenceRebuttals, err := creator.evidenceRebuttalFinder.FindeEvidenceRebuttalFinder(
			ctx,
			debateGraph, // Full graph for context
			edge.Cause,
			edge.Effect,
			edge,
		)
		if err != nil {
			log.Printf("ERROR: Could not find evidence rebuttals for edge [%s] -> [%s]: %v", edge.Cause.Argument, edge.Effect.Argument, err)
			continue
		}

		if evidenceRebuttals != nil && len(evidenceRebuttals.Rebuttals) > 0 {
			log.Printf("SUCCESS: Found %d rebuttal(s) for edge [%s] -> [%s]. Adding them to the subGraph.", len(evidenceRebuttals.Rebuttals), edge.Cause.Argument, edge.Effect.Argument)

			for _, rebuttal := range evidenceRebuttals.Rebuttals {
				rebuttalArgument := rebuttal.Rebuttal

				// 2a. 反論ノードをsubGraph内で作成または取得
				rebuttalNode, exists := subGraph.GetNode(rebuttalArgument)
				if !exists {
					log.Printf("  - Creating new rebuttal node in subGraph: [%s]", rebuttalArgument)
					rebuttalNode = domain.NewDebateGraphNode(rebuttalArgument, true)
					if err := subGraph.AddNode(rebuttalNode); err != nil {
						log.Printf("ERROR: Failed to add new rebuttal node to subGraph: %v", err)
						continue
					}
				} else {
					log.Printf("  - Using existing rebuttal node in subGraph: [%s]", rebuttalArgument)
				}

				// 2b. EdgeRebuttalを作成し、subGraphに追加
				newEdgeRebuttal := &domain.DebateGraphEdgeRebuttal{
					TargetEdge:   edge,
					RebuttalType: rebuttal.RebuttalType,
					RebuttalNode: rebuttalNode,
				}
				subGraph.EdgeRebuttals = append(subGraph.EdgeRebuttals, newEdgeRebuttal)

				log.Printf("  - Added EdgeRebuttal to subGraph: Attacks edge [%s] -> [%s] with node [%s] (type: %s)",
					edge.Cause.Argument,
					edge.Effect.Argument,
					rebuttalNode.Argument,
					rebuttal.RebuttalType)
			}
		} else {
			log.Printf("INFO: No rebuttals found for edge [%s] -> [%s].", edge.Cause.Argument, edge.Effect.Argument)
		}
		fmt.Println()
	}

	log.Println("--- Finished Evidence Rebuttal Creation for subGraph edges ---")

	// --- ステップ2: subGraphに対するPMFの反論を生成 ---
	log.Println("--- Starting PMF Rebuttal Creation for subGraph ---")

	pmfRebuttals, err := creator.pmfRebuttalFinder.FindPMFRebuttal(ctx, debateGraph, subGraph)
	if err != nil {
		log.Printf("ERROR: Could not find PMF rebuttals for the subGraph: %v", err)
	}

	if pmfRebuttals != nil {
		allPMFRebuttals := append(pmfRebuttals.StatusQuo, pmfRebuttals.AffirmativePlan...)

		if len(allPMFRebuttals) > 0 {
			log.Printf("SUCCESS: Found %d PMF rebuttal(s) for the subGraph. Adding them to the subGraph.", len(allPMFRebuttals))
			fmt.Println()

			for _, pmfRebuttal := range allPMFRebuttals {
				rebuttalArgument := pmfRebuttal.Rebuttal

				// 1a. 反論ノードをsubGraph内で作成または取得
				rebuttalNode, exists := subGraph.GetNode(rebuttalArgument)
				if !exists {
					log.Printf("  - Creating new PMF rebuttal node in subGraph: [%s]", rebuttalArgument)
					rebuttalNode = domain.NewDebateGraphNode(rebuttalArgument, true)
					if err := subGraph.AddNode(rebuttalNode); err != nil {
						log.Printf("ERROR: Failed to add new PMF rebuttal node to subGraph: %v", err)
						continue
					}
				} else {
					log.Printf("  - Using existing PMF rebuttal node in subGraph: [%s]", rebuttalArgument)
				}

				// 1b. ターゲットノードをsubGraphから取得
				targetNode, exists := subGraph.GetNode(pmfRebuttal.TargetArgument)
				if !exists {
					log.Printf("WARN: Target node '%s' for PMF rebuttal not found in subGraph. Skipping NodeRebuttal creation.", pmfRebuttal.TargetArgument)
					continue
				}

				// 1c. NodeRebuttalを作成し、subGraphに追加
				nodeRebuttal := &domain.DebateGraphNodeRebuttal{
					TargetNode:   targetNode,
					RebuttalType: "importance", // PMFの指摘は、その主張の重要性に対する反論と定義
					RebuttalNode: rebuttalNode,
				}
				subGraph.NodeRebuttals = append(subGraph.NodeRebuttals, nodeRebuttal)

				log.Printf("  - Added NodeRebuttal to subGraph: Attacks node [%s]'s importance with node [%s]",
					targetNode.Argument,
					rebuttalNode.Argument)
			}
		} else {
			log.Println("INFO: No PMF rebuttals were found for the current subGraph.")
		}
	}
	fmt.Println()
	log.Println("--- Finished PMF Rebuttal Creation for subGraph ---")
}
