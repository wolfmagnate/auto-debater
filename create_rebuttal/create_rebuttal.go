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

func (creator *RebuttalCreator) CreateRebuttal(ctx context.Context, debateGraph *domain.DebateGraph) {
	log.Println("--- Starting Evidence Rebuttal Creation ---")

	allEdges := debateGraph.GetAllEdges()

	if len(allEdges) == 0 {
		log.Println("No edges found in the graph. Nothing to rebut.")
		log.Println("--- Finished Evidence Rebuttal Creation ---")
		return
	}

	log.Printf("Found %d edges to analyze for rebuttals.", len(allEdges))
	fmt.Println()

	// 1. 全てのエッジへのevidenceRebuttalを作れないか試す
	for i, edge := range allEdges {
		log.Printf("Analyzing edge %d/%d: [%s] -> [%s]", i+1, len(allEdges), edge.Cause.Argument, edge.Effect.Argument)

		evidenceRebuttals, err := creator.evidenceRebuttalFinder.FindeEvidenceRebuttalFinder(
			ctx,
			debateGraph,
			edge.Cause,
			edge.Effect,
			edge,
		)
		if err != nil {
			log.Printf("ERROR: Could not find evidence rebuttals for edge [%s] -> [%s]: %v", edge.Cause.Argument, edge.Effect.Argument, err)
			continue
		}

		if evidenceRebuttals != nil && len(evidenceRebuttals.Rebuttals) > 0 {
			log.Printf("SUCCESS: Found %d rebuttal(s) for edge [%s] -> [%s]. Adding them to the graph.", len(evidenceRebuttals.Rebuttals), edge.Cause.Argument, edge.Effect.Argument)

			// 2. もしも作れたら、得られたすべての反論についてその反論に対応するノードを作成しdebateGraphに追加
			for _, rebuttal := range evidenceRebuttals.Rebuttals {
				rebuttalArgument := rebuttal.Rebuttal

				// 2a. 反論に対応するノードが既に存在するか確認
				rebuttalNode, exists := debateGraph.GetNode(rebuttalArgument)
				if !exists {
					// 存在しない場合のみ、新しいノードを作成して追加
					log.Printf("  - Creating new rebuttal node: [%s]", rebuttalArgument)
					rebuttalNode = domain.NewDebateGraphNode(rebuttalArgument, true) // isRebuttal = true
					if err := debateGraph.AddNode(rebuttalNode); err != nil {
						log.Printf("ERROR: Failed to add new rebuttal node to graph: %v", err)
						continue // この反論の処理をスキップして次に進む
					}
				} else {
					log.Printf("  - Using existing rebuttal node: [%s]", rebuttalArgument)
				}

				// 3. 追加したノードをEdgeRebuttalとして指定
				newEdgeRebuttal := &domain.DebateGraphEdgeRebuttal{
					TargetEdge:   edge,
					RebuttalType: rebuttal.RebuttalType,
					RebuttalNode: rebuttalNode,
				}

				// 作成したEdgeRebuttalをグラフのリストに追加
				debateGraph.EdgeRebuttals = append(debateGraph.EdgeRebuttals, newEdgeRebuttal)

				log.Printf("  - Added EdgeRebuttal: Attacks edge [%s] -> [%s] with node [%s] (type: %s)",
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

	log.Println("--- Finished Evidence Rebuttal Creation ---")
	log.Println("--- Starting PMF Rebuttal Creation ---")

	pmfRebuttals, err := creator.pmfRebuttalFinder.FindPMFRebuttal(ctx, debateGraph)
	if err != nil {
		log.Printf("ERROR: Could not find PMF rebuttals: %v", err)
	}

	if pmfRebuttals != nil {
		allPMFRebuttals := append(pmfRebuttals.StatusQuo, pmfRebuttals.AffirmativePlan...)

		if len(allPMFRebuttals) > 0 {
			log.Printf("SUCCESS: Found %d PMF rebuttal(s). Adding corresponding nodes and relations to the graph.", len(allPMFRebuttals))

			for _, pmfRebuttal := range allPMFRebuttals {
				rebuttalArgument := pmfRebuttal.Rebuttal

				// ステップ1: 反論ノードを作成または取得する
				rebuttalNode, exists := debateGraph.GetNode(rebuttalArgument)
				if !exists {
					log.Printf("  - Creating new PMF rebuttal node: [%s]", rebuttalArgument)
					rebuttalNode = domain.NewDebateGraphNode(rebuttalArgument, true)
					if err := debateGraph.AddNode(rebuttalNode); err != nil {
						log.Printf("ERROR: Failed to add new PMF rebuttal node to graph: %v", err)
						continue
					}
				} else {
					log.Printf("  - Using existing PMF rebuttal node: [%s]", rebuttalArgument)
				}

				// ステップ2: ターゲットノードを取得する
				targetNode, exists := debateGraph.GetNode(pmfRebuttal.TargetArgument)
				if !exists {
					log.Printf("WARN: Target node '%s' for PMF rebuttal not found in graph. Skipping NodeRebuttal creation.", pmfRebuttal.TargetArgument)
					continue // ターゲットノードがなければ関連付けできない
				}

				// ステップ3: NodeRebuttalを作成し、グラフに追加する
				// この反論は、ターゲットノードの「重要性(importance)」に対するものとする
				nodeRebuttal := &domain.DebateGraphNodeRebuttal{
					TargetNode:   targetNode,
					RebuttalType: "importance", // PMFの指摘は、その主張の重要性に対する反論と定義
					RebuttalNode: rebuttalNode,
				}
				debateGraph.NodeRebuttals = append(debateGraph.NodeRebuttals, nodeRebuttal)

				log.Printf("  - Added NodeRebuttal: Attacks node [%s]'s importance with node [%s]",
					targetNode.Argument,
					rebuttalNode.Argument)
			}
		} else {
			log.Println("INFO: No PMF rebuttals were found for the current graph.")
		}
	}

	log.Println("--- Finished PMF Rebuttal Creation ---")
}
