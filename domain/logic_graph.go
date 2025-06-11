package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

type LogicGraphNode struct {
	Argument string
	Causes   []*LogicGraphNode
}

type LogicGraph struct {
	Nodes   []*LogicGraphNode
	NodeMap map[string]*LogicGraphNode
}

func NewLogicGraphNode(argument string) *LogicGraphNode {
	return &LogicGraphNode{
		Argument: argument,
		Causes:   make([]*LogicGraphNode, 0),
	}
}

func NewLogicGraph(initialNodes []*LogicGraphNode) *LogicGraph {
	graph := &LogicGraph{
		Nodes:   make([]*LogicGraphNode, 0),
		NodeMap: make(map[string]*LogicGraphNode),
	}
	for _, node := range initialNodes {
		graph.AddNode(node)
	}
	return graph
}

func (lg *LogicGraph) AddNode(node *LogicGraphNode) {
	if node == nil {
		fmt.Println("Cannot add a nil node.")
		return
	}
	if _, exists := lg.NodeMap[node.Argument]; exists {
		fmt.Printf("Node with argument '%s' already exists. Skipping.\n", node.Argument)
		return
	}
	lg.Nodes = append(lg.Nodes, node)
	lg.NodeMap[node.Argument] = node
}

func ListAllCausalRelationships(graph *LogicGraph) []string {
	var relationships []string

	if graph == nil {
		return relationships // 空のグラフの場合は空のリストを返す
	}

	for _, effect := range graph.Nodes {
		if effect == nil {
			continue // ノードがnilの場合はスキップ
		}
		for _, cause := range effect.Causes {
			if cause == nil {
				continue // 結果ノードがnilの場合はスキップ
			}
			relationship := fmt.Sprintf("- 「%s」であることが「%s」を引き起こす", cause.Argument, effect.Argument)
			relationships = append(relationships, relationship)
		}
	}

	return relationships
}

// ToJSON は LogicGraph を指定されたカスタム形式のJSON文字列に変換します。
// nodes: ["Argument1", "Argument2", ...]
// edges: [["CauseArg1", "EffectArg1"], ["CauseArg2", "EffectArg2"], ...]
func (lg *LogicGraph) ToJSON() (string, error) {
	type CustomJSONOutput struct {
		Nodes []string   `json:"nodes"`
		Edges [][]string `json:"edges"`
	}

	outputData := CustomJSONOutput{
		Nodes: make([]string, 0),
		Edges: make([][]string, 0),
	}

	if lg == nil {
		jsonData, err := json.MarshalIndent(outputData, "", "  ")
		if err != nil {
			return "", fmt.Errorf("nil LogicGraphのカスタムJSONへのマーシャリング中に予期せぬエラーが発生しました: %w", err)
		}
		return string(jsonData), nil
	}

	for _, node := range lg.Nodes {
		outputData.Nodes = append(outputData.Nodes, node.Argument)
	}
	sort.Strings(outputData.Nodes)

	for _, effectNode := range lg.Nodes {
		for _, causeNode := range effectNode.Causes {
			outputData.Edges = append(outputData.Edges, []string{causeNode.Argument, effectNode.Argument})
		}
	}
	sort.Slice(outputData.Edges, func(i, j int) bool {
		edgeI := outputData.Edges[i]
		edgeJ := outputData.Edges[j]
		if edgeI[0] != edgeJ[0] {
			return edgeI[0] < edgeJ[0]
		}
		return edgeI[1] < edgeJ[1]
	})

	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("LogicGraphのカスタムJSONへのマーシャリング中にエラーが発生しました: %w", err)
	}
	return string(jsonData), nil
}
