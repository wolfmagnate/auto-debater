package debate_graph_creator

import (
	"context"
	"fmt"

	"github.com/wolfmagnate/auto_debater/domain"
)

type DebateGraphCreator struct {
	DebateAnnotationCreator *DebateAnnotationCreator
	DocumentSplitter        *DocumentSplitter
}

func (creator *DebateGraphCreator) CreateDebateGraph(ctx context.Context, document string, logicGraph *domain.LogicGraph) (*domain.DebateGraph, error) {
	splittedDocument, err := creator.DocumentSplitter.SplitDocumentToParagraph(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to split document: %w", err)
	}

	var annotations []LogicAnnotation
	for _, paragraph := range splittedDocument.Paragraphs {
		paragraphAnnotations, err := creator.DebateAnnotationCreator.CreateDebateAnnotations(ctx, document, paragraph, logicGraph)
		if err != nil {
			return nil, fmt.Errorf("failed to create debate annotations: %w", err)
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

	debateGraph := domain.NewDebateGraph()

	// 1. LogicGraph からノードを DebateGraph にコピー
	for _, lgNode := range logicGraph.Nodes {
		dgNode := domain.NewDebateGraphNode(lgNode.Argument, false) // domainのコンストラクタ使用
		if err := debateGraph.AddNode(dgNode); err != nil {
			// LogicGraphが整合性を持っていれば、通常このエラーは発生しないはず
			return nil, fmt.Errorf("failed to add node '%s' to DebateGraph: %w", lgNode.Argument, err)
		}
	}

	// 2. LogicGraph からエッジを DebateGraph にコピー
	for _, lgEffectNode := range logicGraph.Nodes {
		for _, lgCauseNode := range lgEffectNode.Causes {
			dgCauseNode, causeOk := debateGraph.GetNode(lgCauseNode.Argument)
			if !causeOk {
				return nil, fmt.Errorf("internal consistency error: cause node '%s' (from LogicGraph) not found in DebateGraph when creating edge", lgCauseNode.Argument)
			}
			dgEffectNode, effectOk := debateGraph.GetNode(lgEffectNode.Argument)
			if !effectOk {
				return nil, fmt.Errorf("internal consistency error: effect node '%s' (from LogicGraph) not found in DebateGraph when creating edge", lgEffectNode.Argument)
			}

			dgEdge := domain.NewDebateGraphEdge(dgCauseNode, dgEffectNode, false) // domainのコンストラクタ使用
			if err := debateGraph.AddEdge(dgEdge); err != nil {
				return nil, fmt.Errorf("failed to add edge '%s -> %s' to DebateGraph: %w", lgCauseNode.Argument, lgEffectNode.Argument, err)
			}
		}
	}

	// 3. filteredAnnotations を DebateGraph に適用
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
			case "importance_rebuttal":
				targetNode.ImportanceRebuttals = append(targetNode.ImportanceRebuttals, ann.NodeAnnotation.ImportanceRebuttal)
			case "uniqueness_rebuttal":
				targetNode.UniquenessRebuttals = append(targetNode.UniquenessRebuttals, ann.NodeAnnotation.UniquenessRebuttal)
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
			case "certainty_rebuttal":
				targetEdge.CertaintyRebuttal = append(targetEdge.CertaintyRebuttal, ann.EdgeAnnotation.CertaintyRebuttal)
			case "uniqueness_rebuttal":
				targetEdge.UniquenessRebuttals = append(targetEdge.UniquenessRebuttals, ann.EdgeAnnotation.UniquenessRebuttal)
			}
		}
	}

	return debateGraph, nil
}
