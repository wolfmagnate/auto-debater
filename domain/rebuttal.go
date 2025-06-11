package domain

import "fmt"

// ノードに対する反論
type DebateGraphNodeRebuttal struct {
	TargetNode   *DebateGraphNode // どのノードに反論するか
	RebuttalType string           // "importance"または"uniqueness"
	RebuttalNode *DebateGraphNode // 反論を行うノード
}

// エッジに対する反論
type DebateGraphEdgeRebuttal struct {
	TargetEdge   *DebateGraphEdge // どのエッジに反論するか
	RebuttalType string           // "certainty"または"uniqueness"
	RebuttalNode *DebateGraphNode // 反論を行うノード
}

type CounterArgumentRebuttal struct {
	TargetNode   *DebateGraphNode // 反論対象のノード
	RebuttalNode *DebateGraphNode // 反対の主張を行うノード
}

type TurnArgumentRebuttal struct {
	RebuttalNode *DebateGraphNode // ターンによりメリット・デメリットを主張するノード
}

func NewDebateGraphNodeRebuttal(debateGraph *DebateGraph, targetArgument string, rebuttalType string, argument string) (*DebateGraphNodeRebuttal, error) {
	targetNode, exists := debateGraph.GetNode(targetArgument)
	if !exists {
		return nil, fmt.Errorf("target node '%s' not found in debate graph", targetArgument)
	}

	if rebuttalType != "importance" && rebuttalType != "uniqueness" {
		return nil, fmt.Errorf("invalid rebuttal type '%s' for node rebuttal", rebuttalType)
	}

	rebuttalNode, exists := debateGraph.GetNode(argument)
	if !exists {
		return nil, fmt.Errorf("rebuttal node '%s' not found in debate graph", argument)
	}

	return &DebateGraphNodeRebuttal{
		TargetNode:   targetNode,
		RebuttalType: rebuttalType,
		RebuttalNode: rebuttalNode,
	}, nil
}

func NewDebateGraphEdgeRebuttal(debateGraph *DebateGraph, targetCauseArgument string, targetEffectArgument string, rebuttalType string, argument string) (*DebateGraphEdgeRebuttal, error) {
	targetEdge, exists := debateGraph.GetEdge(targetCauseArgument, targetEffectArgument)
	if !exists {
		return nil, fmt.Errorf("target edge '%s -> %s' not found in debate graph", targetCauseArgument, targetEffectArgument)
	}

	if rebuttalType != "certainty" && rebuttalType != "uniqueness" {
		return nil, fmt.Errorf("invalid rebuttal type '%s' for edge rebuttal", rebuttalType)
	}

	rebuttalNode, exists := debateGraph.GetNode(argument)
	if !exists {
		return nil, fmt.Errorf("rebuttal node '%s' not found in debate graph", argument)
	}

	return &DebateGraphEdgeRebuttal{
		TargetEdge:   targetEdge,
		RebuttalType: rebuttalType,
		RebuttalNode: rebuttalNode,
	}, nil
}

func NewCounterArgumentRebuttal(debateGraph *DebateGraph, targetArgument, rebuttalArgument string) (*CounterArgumentRebuttal, error) {
	rebuttalNode, exists := debateGraph.GetNode(rebuttalArgument)
	if !exists {
		return nil, fmt.Errorf("rebuttal node '%s' not found in debate graph", rebuttalArgument)
	}

	targetNode, exists := debateGraph.GetNode(targetArgument)
	if !exists {
		return nil, fmt.Errorf("target node '%s' not found in debate graph", targetArgument)
	}

	return &CounterArgumentRebuttal{
		TargetNode:   targetNode,
		RebuttalNode: rebuttalNode,
	}, nil
}

func NewTurnArgumentRebuttal(debateGraph *DebateGraph, argument string) (*TurnArgumentRebuttal, error) {
	rebuttalNode, exists := debateGraph.GetNode(argument)
	if !exists {
		return nil, fmt.Errorf("rebuttal node '%s' not found in debate graph", argument)
	}

	return &TurnArgumentRebuttal{
		RebuttalNode: rebuttalNode,
	}, nil
}
