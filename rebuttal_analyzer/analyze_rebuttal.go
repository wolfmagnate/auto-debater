package rebuttal_analyzer

import (
	"context"
	"fmt"

	"github.com/wolfmagnate/auto_debater/domain"
)

type RebuttalAnalyzer struct {
	rebuttalFinder      *RebuttalFinder
	rebuttalCauseFinder *RebuttalCauseFinder
	newArgumentFinder   *NewArgumentFinder

	documentSplitter          *DocumentSplitter
	rebuttalAnnotationCreator *RebuttalAnnotationCreator
}

func CreateRebuttalAnalyzer() (*RebuttalAnalyzer, error) {
	rebuttalFinder, err := CreateRebuttalFinder()
	if err != nil {
		return nil, fmt.Errorf("failed to create RebuttalFinder: %w", err)
	}

	rebuttalCauseFinder, err := CreateRebuttalCauseFinder()
	if err != nil {
		return nil, fmt.Errorf("failed to create RebuttalCauseFinder: %w", err)
	}

	newArgumentFinder, err := CreateNewArgumentFinder()
	if err != nil {
		return nil, fmt.Errorf("failed to create NewArgumentFinder: %w", err)
	}

	documentSplitter, err := CreateDocumentSplitter()
	if err != nil {
		return nil, fmt.Errorf("failed to create DocumentSplitter: %w", err)
	}

	rebuttalAnnotationCreator, err := CreateRebuttalAnnotationCreator()
	if err != nil {
		return nil, fmt.Errorf("failed to create RebuttalAnnotationCreator: %w", err)
	}

	return &RebuttalAnalyzer{
		rebuttalFinder:      rebuttalFinder,
		rebuttalCauseFinder: rebuttalCauseFinder,
		newArgumentFinder:   newArgumentFinder,

		documentSplitter:          documentSplitter,
		rebuttalAnnotationCreator: rebuttalAnnotationCreator,
	}, nil
}

func (analyzer *RebuttalAnalyzer) AnalyzeRebuttal(ctx context.Context, debateGraph *domain.DebateGraph, rebuttal string) error {
	analyzedRebuttals, err := analyzer.rebuttalFinder.FindRebuttals(ctx, debateGraph, rebuttal)
	if err != nil {
		return fmt.Errorf("反論の発見に失敗しました: %w", err)
	}

	queue := make([]*domain.DebateGraphNode, 0)

	for _, rebuttal := range analyzedRebuttals.Rebuttals {
		addedNode, err := addRebuttalToDebateGraph(debateGraph, rebuttal)
		if err != nil {
			return fmt.Errorf("反論の追加に失敗しました: %w", err)
		}
		queue = append(queue, addedNode)
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		foundCauses, err := analyzer.rebuttalCauseFinder.FindRebuttalCauses(ctx, debateGraph, rebuttal, current.Argument)
		if err != nil {
			return fmt.Errorf("反論の発見に失敗しました: %w", err)
		}

		targetArgumentAndCause := &ArgumentAndCauses{
			Argument: current.Argument,
			Causes:   foundCauses.Causes,
		}

		findNewArgumentResult, err := analyzer.newArgumentFinder.FindNewArguments(ctx, debateGraph, targetArgumentAndCause)
		if err != nil {
			return fmt.Errorf("反論の発見に失敗しました: %w", err)
		}

		for _, newArgument := range findNewArgumentResult.NewNodes {
			newNode := domain.NewDebateGraphNode(newArgument, true)
			err := debateGraph.AddNode(newNode)
			if err != nil {
				return fmt.Errorf("反論の追加に失敗しました: %w", err)
			}

			queue = append(queue, newNode)
		}

		// エッジの追加
		effectNode, exists := debateGraph.GetNode(targetArgumentAndCause.Argument)
		if !exists {
			return fmt.Errorf("ノードが見つかりません: %s", targetArgumentAndCause.Argument)
		}

		for _, cause := range findNewArgumentResult.UsedCauses {
			causeNode, exists := debateGraph.GetNode(cause)
			if !exists {
				return fmt.Errorf("ノードが見つかりません: %s", cause)
			}
			err = debateGraph.AddEdge(domain.NewDebateGraphEdge(causeNode, effectNode, true))
			if err != nil {
				return fmt.Errorf("エッジの追加に失敗しました: %w", err)
			}
		}
	}

	// 3. 新しいグラフ構造に対してアノテーションを見つけて、グラフを装飾
	splittedDocument, err := analyzer.documentSplitter.SplitDocumentToParagraph(ctx, rebuttal)
	if err != nil {
		return fmt.Errorf("failed to split document: %w", err)
	}

	var annotations []LogicAnnotation
	for _, paragraph := range splittedDocument.Paragraphs {
		paragraphAnnotations, err := analyzer.rebuttalAnnotationCreator.CreateRebuttalAnnotations(ctx, debateGraph, rebuttal, paragraph)
		if err != nil {
			return fmt.Errorf("failed to create debate annotations: %w", err)
		}

		if paragraphAnnotations != nil {
			annotations = append(annotations, paragraphAnnotations.Annotations...)
		}
	}

	filteredAnnotations := []LogicAnnotation{}
	for _, ann := range annotations {
		if !(ann.TargetType == "node" && ann.NodeAnnotation.AnnotationType == "argument") {
			filteredAnnotations = append(filteredAnnotations, ann)
		}
	}

	for _, ann := range filteredAnnotations {
		if ann.TargetType == "node" {
			targetNode, exists := debateGraph.GetNode(ann.NodeAnnotation.Argument)
			if !exists {
				fmt.Printf("Warning: Annotation for non-existent node '%s' skipped.\n", ann.NodeAnnotation.Argument)
				continue
			}
			switch ann.NodeAnnotation.AnnotationType {
			case "importance":
				targetNode.Importance = append(targetNode.Importance, ann.NodeAnnotation.Importance)
			case "uniqueness":
				targetNode.Uniqueness = append(targetNode.Uniqueness, ann.NodeAnnotation.Uniqueness)
			}
		} else if ann.TargetType == "edge" {
			targetEdge, exists := debateGraph.GetEdge(ann.EdgeAnnotation.CauseArgument, ann.EdgeAnnotation.EffectArgument)
			if !exists {
				fmt.Printf("Warning: Annotation for non-existent edge '%s -> %s' skipped.\n", ann.EdgeAnnotation.CauseArgument, ann.EdgeAnnotation.EffectArgument)
				continue
			}
			switch ann.EdgeAnnotation.AnnotationType {
			case "certainty":
				targetEdge.Certainty = append(targetEdge.Certainty, ann.EdgeAnnotation.Certainty)
			case "uniqueness":
				targetEdge.Uniqueness = append(targetEdge.Uniqueness, ann.EdgeAnnotation.Uniqueness)
			}
		}
	}

	return nil
}

// debateGraphに対応する反論ノードを作成する
// その反論ノードに対応する反論アノテーションを作成する
// 作成したDebateGraphNodeを返す
func addRebuttalToDebateGraph(debateGraph *domain.DebateGraph, otherRebuttal RebuttalItem) (*domain.DebateGraphNode, error) {
	switch otherRebuttal.RebuttalKind {
	case "edge_rebuttal":
		edgeRebuttal := otherRebuttal.EdgeRebuttal
		switch edgeRebuttal.RebuttalType {
		case "certainty":
			newNode := domain.NewDebateGraphNode(edgeRebuttal.CertaintyRebuttal, true)
			err := debateGraph.AddNode(newNode)
			if err != nil {
				return nil, fmt.Errorf("エッジに対する反論の作成に失敗しました: %w", err)
			}
			debateGraphEdgeRebuttal, err := domain.NewDebateGraphEdgeRebuttal(
				debateGraph, edgeRebuttal.TargetEdgeCause, edgeRebuttal.TargetEdgeEffect,
				"certainty", edgeRebuttal.CertaintyRebuttal,
			)
			if err != nil {
				return nil, fmt.Errorf("エッジに対する反論の作成に失敗しました: %w", err)
			}
			debateGraph.EdgeRebuttals = append(debateGraph.EdgeRebuttals, debateGraphEdgeRebuttal)
			return newNode, nil
		case "uniqueness":
			newNode := domain.NewDebateGraphNode(edgeRebuttal.UniquenessRebuttal, true)
			err := debateGraph.AddNode(newNode)
			if err != nil {
				return nil, fmt.Errorf("エッジに対する反論の作成に失敗しました: %w", err)
			}
			debateGraphEdgeRebuttal, err := domain.NewDebateGraphEdgeRebuttal(
				debateGraph, edgeRebuttal.TargetEdgeCause, edgeRebuttal.TargetEdgeEffect,
				"uniqueness", edgeRebuttal.UniquenessRebuttal,
			)
			if err != nil {
				return nil, fmt.Errorf("エッジに対する反論の作成に失敗しました: %w", err)
			}
			debateGraph.EdgeRebuttals = append(debateGraph.EdgeRebuttals, debateGraphEdgeRebuttal)
			return newNode, nil
		}
	case "node_rebuttal":
		nodeRebuttal := otherRebuttal.NodeRebuttal
		switch nodeRebuttal.RebuttalType {
		case "importance":
			newNode := domain.NewDebateGraphNode(nodeRebuttal.ImportanceRebuttal, true)
			err := debateGraph.AddNode(newNode)
			if err != nil {
				return nil, fmt.Errorf("ノードに対する反論の作成に失敗しました: %w", err)
			}
			debateGraphNodeRebuttal, err := domain.NewDebateGraphNodeRebuttal(
				debateGraph, nodeRebuttal.TargetNode,
				"importance", nodeRebuttal.ImportanceRebuttal,
			)
			if err != nil {
				return nil, fmt.Errorf("ノードに対する反論の作成に失敗しました: %w", err)
			}
			debateGraph.NodeRebuttals = append(debateGraph.NodeRebuttals, debateGraphNodeRebuttal)
			return newNode, nil
		case "uniqueness":
			newNode := domain.NewDebateGraphNode(nodeRebuttal.UniquenessRebuttal, true)
			err := debateGraph.AddNode(newNode)
			if err != nil {
				return nil, fmt.Errorf("ノードに対する反論の作成に失敗しました: %w", err)
			}
			debateGraphNodeRebuttal, err := domain.NewDebateGraphNodeRebuttal(
				debateGraph, nodeRebuttal.TargetNode,
				"uniqueness", nodeRebuttal.UniquenessRebuttal,
			)
			if err != nil {
				return nil, fmt.Errorf("ノードに対する反論の作成に失敗しました: %w", err)
			}
			debateGraph.NodeRebuttals = append(debateGraph.NodeRebuttals, debateGraphNodeRebuttal)
			return newNode, nil
		}
	case "counter_argument":
		counterArgument := otherRebuttal.CounterArgument
		counterDebateGraphNode := domain.NewDebateGraphNode(counterArgument.Argument, true)
		err := debateGraph.AddNode(counterDebateGraphNode)
		if err != nil {
			return nil, fmt.Errorf("反論の追加に失敗しました: %w", err)
		}
		counterArgumentRebuttal, err := domain.NewCounterArgumentRebuttal(debateGraph, counterArgument.TargetNode, counterArgument.Argument)
		if err != nil {
			return nil, fmt.Errorf("反論の追加に失敗しました: %w", err)
		}
		debateGraph.CounterArgumentRebuttals = append(debateGraph.CounterArgumentRebuttals, counterArgumentRebuttal)
		return counterDebateGraphNode, nil
	case "turn_argument":
		turnArgument := otherRebuttal.TurnArgument
		turnDebateGraphNode := domain.NewDebateGraphNode(turnArgument.EffectArgument, true)
		err := debateGraph.AddNode(turnDebateGraphNode)
		if err != nil {
			return nil, fmt.Errorf("反論の追加に失敗しました: %w", err)
		}
		turnArgumentRebuttal, err := domain.NewTurnArgumentRebuttal(debateGraph, turnArgument.EffectArgument)
		if err != nil {
			return nil, fmt.Errorf("反論の追加に失敗しました: %w", err)
		}
		debateGraph.TurnArgumentRebuttals = append(debateGraph.TurnArgumentRebuttals, turnArgumentRebuttal)
		return turnDebateGraphNode, nil
	}
	return nil, fmt.Errorf("不明な反論の種類です: %s", otherRebuttal.RebuttalKind)
}
