package logic_graph_creator

import (
	"context"
	"fmt"

	"github.com/wolfmagnate/auto_debater/domain"
)

type LogicGraphCreator struct {
	BasicStructureAnalyzer *BasicStructureAnalyzer
	ImpactAnalyzer         *ImpactAnalyzer
	BenefitHarmConverter   *BenefitHarmConverter
	LogicGraphCompleter    *LogicGraphCompleter
}

type LogicGraphCompleter struct {
	CauseFinder       *CauseFinder
	NewArgumentFinder *NewArgumentFinder
}

func (creator *LogicGraphCreator) CreateLogicGraph(ctx context.Context, document string) (*domain.LogicGraph, error) {
	basicArgumentStructure, err := creator.BasicStructureAnalyzer.AnalyzeBasicArgumentStructure(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("基本構造の分析に失敗しました : %w", err)
	}

	impactAnalysis, err := creator.ImpactAnalyzer.AnalyzeImpact(ctx, document, basicArgumentStructure)

	if err != nil {
		return nil, fmt.Errorf("影響分析に失敗しました : %w", err)
	}

	initialArguments, err := creator.BenefitHarmConverter.ConvertImpactAnalysisToArguments(ctx, impactAnalysis)
	if err != nil {
		return nil, fmt.Errorf("初期議論の生成に失敗しました : %w", err)
	}

	logicGraph := domain.NewLogicGraph(initialArguments)

	err = creator.LogicGraphCompleter.CompleteLogicNodes(ctx, document, basicArgumentStructure, logicGraph)
	if err != nil {
		return nil, fmt.Errorf("論理グラフの生成に失敗しました : %w", err)
	}

	return logicGraph, nil
}

func (completer *LogicGraphCompleter) CompleteLogicNodes(ctx context.Context, document string, basicArgumentStructure *BasicArgumentStructure, logicGraph *domain.LogicGraph) error {
	queue := []*domain.LogicGraphNode{}
	visited := make(map[string]bool)
	for _, node := range logicGraph.Nodes {
		queue = append(queue, node)
		visited[node.Argument] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		addedNodes, err := completer.CompleteTargetLogicNode(ctx, document, basicArgumentStructure, logicGraph, current)
		if err != nil {
			return fmt.Errorf("ノードの追加に失敗しました : %w", err)
		}

		for _, node := range addedNodes {
			if !visited[node.Argument] {
				queue = append(queue, node)
				visited[node.Argument] = true
			}
		}
	}

	return nil
}

func (completer *LogicGraphCompleter) CompleteTargetLogicNode(ctx context.Context, document string, basicArgumentStructure *BasicArgumentStructure, logicGraph *domain.LogicGraph, targetNode *domain.LogicGraphNode) ([]*domain.LogicGraphNode, error) {

	foundCauses, err := completer.CauseFinder.FindCauses(ctx, document, basicArgumentStructure, targetNode.Argument)
	if err != nil {
		return nil, fmt.Errorf("原因の探索に失敗しました : %w", err)
	}
	targetArgumentAndCause := &ArgumentAndCauses{
		Argument: targetNode.Argument,
		Causes:   foundCauses.Causes,
	}
	findNewArgumentsResult, err := completer.NewArgumentFinder.FindNewArguments(ctx, document, basicArgumentStructure, logicGraph, targetArgumentAndCause)
	if err != nil {
		return nil, fmt.Errorf("新規議論の探索に失敗しました : %w", err)
	}

	newNodes := []*domain.LogicGraphNode{}
	for _, newArgument := range findNewArgumentsResult.NewNodes {
		newNode := domain.NewLogicGraphNode(newArgument)
		newNodes = append(newNodes, newNode)
		logicGraph.AddNode(newNode)
	}

	for _, cause := range findNewArgumentsResult.UsedCauses {
		targetNode.Causes = append(targetNode.Causes, logicGraph.NodeMap[cause])
	}

	return newNodes, nil
}
