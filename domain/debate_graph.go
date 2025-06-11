package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DebateGraphNode と DebateGraphEdge の定義は変更なし
type DebateGraphNode struct {
	Argument            string
	Causes              []*DebateGraphEdge
	Importance          []string
	Uniqueness          []string
	ImportanceRebuttals []string
	UniquenessRebuttals []string

	IsRebuttal bool
}

type DebateGraphEdge struct {
	Cause               *DebateGraphNode
	Effect              *DebateGraphNode
	Certainty           []string
	Uniqueness          []string
	CertaintyRebuttal   []string
	UniquenessRebuttals []string

	IsRebuttal bool
}

func NewDebateGraphNode(argument string, isRebuttal bool) *DebateGraphNode {
	return &DebateGraphNode{
		Argument:            argument,
		Causes:              make([]*DebateGraphEdge, 0),
		Importance:          make([]string, 0),
		Uniqueness:          make([]string, 0),
		ImportanceRebuttals: make([]string, 0),
		UniquenessRebuttals: make([]string, 0),
		IsRebuttal:          isRebuttal,
	}
}

func NewDebateGraphEdge(cause, effect *DebateGraphNode, isRebuttal bool) *DebateGraphEdge {
	return &DebateGraphEdge{
		Cause:               cause,
		Effect:              effect,
		Certainty:           make([]string, 0),
		Uniqueness:          make([]string, 0),
		CertaintyRebuttal:   make([]string, 0),
		UniquenessRebuttals: make([]string, 0),
		IsRebuttal:          isRebuttal,
	}
}

type DebateGraph struct {
	Nodes                    []*DebateGraphNode
	NodeRebuttals            []*DebateGraphNodeRebuttal
	EdgeRebuttals            []*DebateGraphEdgeRebuttal
	CounterArgumentRebuttals []*CounterArgumentRebuttal
	TurnArgumentRebuttals    []*TurnArgumentRebuttal

	nodeMap map[string]*DebateGraphNode // 小文字で非公開にし、メソッド経由でアクセス
	edgeMap map[string]*DebateGraphEdge // キー: "CauseArgument->EffectArgument"
}

func NewDebateGraph() *DebateGraph {
	return &DebateGraph{
		Nodes:         make([]*DebateGraphNode, 0),
		NodeRebuttals: make([]*DebateGraphNodeRebuttal, 0),
		EdgeRebuttals: make([]*DebateGraphEdgeRebuttal, 0),
		nodeMap:       make(map[string]*DebateGraphNode),
		edgeMap:       make(map[string]*DebateGraphEdge),
	}
}

// AddNode はグラフにノードを追加します。
// 同じArgumentを持つノードが既に存在する場合はエラーを返します。
func (dg *DebateGraph) AddNode(node *DebateGraphNode) error {
	if node == nil {
		return fmt.Errorf("cannot add a nil node to DebateGraph")
	}
	if _, exists := dg.nodeMap[node.Argument]; exists {
		// 既に存在する場合、エラーを返すか、既存ノードを返すか、何もしないかは設計次第。
		// ここではエラーとして、呼び出し元に重複を通知します。
		return fmt.Errorf("node with argument '%s' already exists in DebateGraph", node.Argument)
	}
	dg.Nodes = append(dg.Nodes, node)
	dg.nodeMap[node.Argument] = node
	return nil
}

// GetNode はArgument文字列によってノードを取得します。
func (dg *DebateGraph) GetNode(argument string) (*DebateGraphNode, bool) {
	node, exists := dg.nodeMap[argument]
	return node, exists
}

// generateEdgeKey はエッジマップ用のキーを生成する内部ヘルパー関数です。
func generateEdgeKey(causeArgument, effectArgument string) string {
	return fmt.Sprintf("%s->%s", causeArgument, effectArgument)
}

// AddEdge はグラフにエッジを追加します。
// エッジのCauseノードとEffectノードは事前にグラフに追加されている必要があります。
// EffectノードのCausesリストも更新します。
func (dg *DebateGraph) AddEdge(edge *DebateGraphEdge) error {
	if edge == nil {
		return fmt.Errorf("cannot add a nil edge to DebateGraph")
	}
	if edge.Cause == nil || edge.Effect == nil {
		return fmt.Errorf("edge must have valid cause and effect nodes")
	}

	// エッジが参照するノードがグラフに存在することを確認
	if _, exists := dg.nodeMap[edge.Cause.Argument]; !exists {
		return fmt.Errorf("cause node '%s' of the edge is not in the graph", edge.Cause.Argument)
	}
	effectNodeInMap, effectNodeExists := dg.nodeMap[edge.Effect.Argument]
	if !effectNodeExists {
		return fmt.Errorf("effect node '%s' of the edge is not in the graph", edge.Effect.Argument)
	}
	// edge.Effectがマップ内のインスタンスと同じであることを保証（通常、呼び出し側が正しく構築すれば問題ない）
	if edge.Effect != effectNodeInMap {
		return fmt.Errorf("edge's effect node instance does not match the instance in the graph's nodeMap for argument '%s'", edge.Effect.Argument)
	}

	edgeKey := generateEdgeKey(edge.Cause.Argument, edge.Effect.Argument)
	if _, exists := dg.edgeMap[edgeKey]; exists {
		return nil
	}

	dg.edgeMap[edgeKey] = edge
	// EffectノードのCausesリストにこのエッジを追加
	// edge.Effect はグラフ内の正しいインスタンスである前提
	edge.Effect.Causes = append(edge.Effect.Causes, edge)
	return nil
}

func (dg *DebateGraph) RemoveEdge(causeArgument, effectArgument string) error {
	edgeKey := generateEdgeKey(causeArgument, effectArgument)
	edge, exists := dg.edgeMap[edgeKey]
	if !exists {
		return fmt.Errorf("削除対象のエッジ '%s' がグラフ内に見つかりません", edgeKey)
	}

	// 1. edgeMapからエッジを削除します。
	delete(dg.edgeMap, edgeKey)

	// 2. EffectノードのCausesスライスから該当するエッジを削除します。
	//    スライスをループ処理し、削除対象以外の要素で新しいスライスを構築し直します。
	effectNode := edge.Effect
	newCauses := make([]*DebateGraphEdge, 0, len(effectNode.Causes)-1)
	for _, e := range effectNode.Causes {
		// ポインタを比較して、削除対象のエッジと同一でないものだけを新しいスライスに追加します。
		if e != edge {
			newCauses = append(newCauses, e)
		}
	}
	// EffectノードのCausesスライスを、更新された新しいスライスで上書きします。
	effectNode.Causes = newCauses

	return nil
}

func (dg *DebateGraph) GetEdge(causeArgument string, effectArgument string) (*DebateGraphEdge, bool) {
	edgeKey := generateEdgeKey(causeArgument, effectArgument)
	edge, exists := dg.edgeMap[edgeKey]
	return edge, exists
}

func (dg *DebateGraph) GetAllEdges() []*DebateGraphEdge {
	edges := make([]*DebateGraphEdge, 0, len(dg.edgeMap))
	for _, edge := range dg.edgeMap {
		edges = append(edges, edge)
	}
	return edges
}

func (dg *DebateGraph) DisplayGraph() {
	fmt.Println("--- Debate Graph ---")

	if len(dg.Nodes) == 0 {
		fmt.Println("Graph is empty.")
		fmt.Println("--------------------")
		return
	}

	fmt.Println("\n=== NODES ===")
	for i, node := range dg.Nodes {
		fmt.Printf("[%d] Argument: %s\n", i, node.Argument)
		if len(node.Importance) > 0 {
			fmt.Printf("    Importance: %s\n", strings.Join(node.Importance, ", "))
		}
		if len(node.Uniqueness) > 0 {
			fmt.Printf("    Uniqueness: %s\n", strings.Join(node.Uniqueness, ", "))
		}
		if len(node.ImportanceRebuttals) > 0 {
			fmt.Printf("    Importance Rebuttals: %s\n", strings.Join(node.ImportanceRebuttals, ", "))
		}
		if len(node.UniquenessRebuttals) > 0 {
			fmt.Printf("    Uniqueness Rebuttals: %s\n", strings.Join(node.UniquenessRebuttals, ", "))
		}
		// Display incoming edges (causes for this node)
		if len(node.Causes) > 0 {
			fmt.Printf("    Caused By (Incoming Edges):\n")
			for _, edge := range node.Causes {
				fmt.Printf("      - From: %s (Certainty: %s, Uniqueness: %s)\n",
					edge.Cause.Argument,
					strings.Join(edge.Certainty, ", "),
					strings.Join(edge.Uniqueness, ", "))
				if len(edge.CertaintyRebuttal) > 0 {
					fmt.Printf("        Certainty Rebuttals: %s\n", strings.Join(edge.CertaintyRebuttal, ", "))
				}
				if len(edge.UniquenessRebuttals) > 0 {
					fmt.Printf("        Uniqueness Rebuttals: %s\n", strings.Join(edge.UniquenessRebuttals, ", "))
				}
			}
		}
		fmt.Println("  ---")
	}

	fmt.Println("\n=== EDGES (Explicit Listing) ===")
	if len(dg.edgeMap) == 0 {
		fmt.Println("No edges in the graph.")
	} else {
		i := 0
		for key, edge := range dg.edgeMap {
			fmt.Printf("[%d] Edge: %s\n", i, key)
			fmt.Printf("    Cause: %s\n", edge.Cause.Argument)
			fmt.Printf("    Effect: %s\n", edge.Effect.Argument)
			if len(edge.Certainty) > 0 {
				fmt.Printf("    Certainty: %s\n", strings.Join(edge.Certainty, ", "))
			}
			if len(edge.Uniqueness) > 0 {
				fmt.Printf("    Uniqueness: %s\n", strings.Join(edge.Uniqueness, ", "))
			}
			if len(edge.CertaintyRebuttal) > 0 {
				fmt.Printf("    Certainty Rebuttals: %s\n", strings.Join(edge.CertaintyRebuttal, ", "))
			}
			if len(edge.UniquenessRebuttals) > 0 {
				fmt.Printf("    Uniqueness Rebuttals: %s\n", strings.Join(edge.UniquenessRebuttals, ", "))
			}
			fmt.Println("  ---")
			i++
		}
	}
	fmt.Println("--------------------")
}

type jsonNode struct {
	Argument            string   `json:"argument"`
	IsRebuttal          bool     `json:"is_rebuttal"`
	Importance          []string `json:"importance,omitempty"`
	Uniqueness          []string `json:"uniqueness,omitempty"`
	ImportanceRebuttals []string `json:"importance_rebuttals,omitempty"`
	UniquenessRebuttals []string `json:"uniqueness_rebuttals,omitempty"`
}

func (n *DebateGraphNode) ToJSON() (string, error) {
	if n == nil {
		return "", fmt.Errorf("cannot convert nil DebateGraphNode to JSON")
	}

	jNode := &jsonNode{
		Argument:            n.Argument,
		IsRebuttal:          n.IsRebuttal,
		Importance:          n.Importance,
		Uniqueness:          n.Uniqueness,
		ImportanceRebuttals: n.ImportanceRebuttals,
		UniquenessRebuttals: n.UniquenessRebuttals,
	}

	jsonData, err := json.MarshalIndent(jNode, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DebateGraphNode to JSON: %w", err)
	}

	return string(jsonData), nil
}

type jsonEdge struct {
	Cause               string   `json:"cause"`
	Effect              string   `json:"effect"`
	IsRebuttal          bool     `json:"is_rebuttal"`
	Certainty           []string `json:"certainty,omitempty"`
	Uniqueness          []string `json:"uniqueness,omitempty"`
	CertaintyRebuttal   []string `json:"certainty_rebuttal,omitempty"`
	UniquenessRebuttals []string `json:"uniqueness_rebuttals,omitempty"`
}

type jsonNodeRebuttal struct {
	TargetArgument   string `json:"target_argument"`
	RebuttalType     string `json:"rebuttal_type"`
	RebuttalArgument string `json:"rebuttal_argument"`
}

type jsonEdgeRebuttal struct {
	TargetCauseArgument  string `json:"target_cause_argument"`
	TargetEffectArgument string `json:"target_effect_argument"`
	RebuttalType         string `json:"rebuttal_type"`
	RebuttalArgument     string `json:"rebuttal_argument"`
}

func (e *DebateGraphEdge) ToJSON() (string, error) {
	if e == nil {
		return "", fmt.Errorf("cannot convert nil DebateGraphEdge to JSON")
	}
	if e.Cause == nil || e.Effect == nil {
		return "", fmt.Errorf("cannot marshal edge with nil cause or effect")
	}

	jEdge := &jsonEdge{
		Cause:               e.Cause.Argument,
		Effect:              e.Effect.Argument,
		IsRebuttal:          e.IsRebuttal,
		Certainty:           e.Certainty,
		Uniqueness:          e.Uniqueness,
		CertaintyRebuttal:   e.CertaintyRebuttal,
		UniquenessRebuttals: e.UniquenessRebuttals,
	}

	jsonData, err := json.MarshalIndent(jEdge, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DebateGraphEdge to JSON: %w", err)
	}

	return string(jsonData), nil
}

type jsonCounterArgumentRebuttal struct {
	RebuttalArgument string `json:"rebuttal_argument"`
	TargetArgument   string `json:"target_argument"`
}

type jsonTurnArgumentRebuttal struct {
	RebuttalArgument string `json:"rebuttal_argument"`
}

type jsonGraph struct {
	Nodes                    []*jsonNode                    `json:"nodes"`
	Edges                    []*jsonEdge                    `json:"edges"`
	NodeRebuttals            []*jsonNodeRebuttal            `json:"node_rebuttals,omitempty"`
	EdgeRebuttals            []*jsonEdgeRebuttal            `json:"edge_rebuttals,omitempty"`
	CounterArgumentRebuttals []*jsonCounterArgumentRebuttal `json:"counter_argument_rebuttals,omitempty"`
	TurnArgumentRebuttals    []*jsonTurnArgumentRebuttal    `json:"turn_argument_rebuttals,omitempty"`
}

// ToJSON はDebateGraphをJSON文字列に変換します。(修正)
func (dg *DebateGraph) ToJSON() (string, error) {
	if dg == nil {
		return "", fmt.Errorf("cannot convert nil DebateGraph to JSON")
	}

	jGraph := &jsonGraph{
		Nodes:                    make([]*jsonNode, 0, len(dg.Nodes)),
		Edges:                    make([]*jsonEdge, 0, len(dg.edgeMap)),
		NodeRebuttals:            make([]*jsonNodeRebuttal, 0, len(dg.NodeRebuttals)),
		EdgeRebuttals:            make([]*jsonEdgeRebuttal, 0, len(dg.EdgeRebuttals)),
		CounterArgumentRebuttals: make([]*jsonCounterArgumentRebuttal, 0, len(dg.CounterArgumentRebuttals)),
		TurnArgumentRebuttals:    make([]*jsonTurnArgumentRebuttal, 0, len(dg.TurnArgumentRebuttals)),
	}

	// ノードの変換
	for _, node := range dg.Nodes {
		jGraph.Nodes = append(jGraph.Nodes, &jsonNode{
			Argument:            node.Argument,
			IsRebuttal:          node.IsRebuttal,
			Importance:          node.Importance,
			Uniqueness:          node.Uniqueness,
			ImportanceRebuttals: node.ImportanceRebuttals,
			UniquenessRebuttals: node.UniquenessRebuttals,
		})
	}

	// エッジの変換
	for _, edge := range dg.edgeMap {
		jGraph.Edges = append(jGraph.Edges, &jsonEdge{
			Cause:               edge.Cause.Argument,
			Effect:              edge.Effect.Argument,
			IsRebuttal:          edge.IsRebuttal,
			Certainty:           edge.Certainty,
			Uniqueness:          edge.Uniqueness,
			CertaintyRebuttal:   edge.CertaintyRebuttal,
			UniquenessRebuttals: edge.UniquenessRebuttals,
		})
	}

	// ノード反論の変換
	for _, r := range dg.NodeRebuttals {
		jGraph.NodeRebuttals = append(jGraph.NodeRebuttals, &jsonNodeRebuttal{
			TargetArgument:   r.TargetNode.Argument,
			RebuttalType:     r.RebuttalType,
			RebuttalArgument: r.RebuttalNode.Argument,
		})
	}

	// エッジ反論の変換
	for _, r := range dg.EdgeRebuttals {
		jGraph.EdgeRebuttals = append(jGraph.EdgeRebuttals, &jsonEdgeRebuttal{
			TargetCauseArgument:  r.TargetEdge.Cause.Argument,
			TargetEffectArgument: r.TargetEdge.Effect.Argument,
			RebuttalType:         r.RebuttalType,
			RebuttalArgument:     r.RebuttalNode.Argument,
		})
	}

	// 反対意見の変換
	for _, r := range dg.CounterArgumentRebuttals {
		jGraph.CounterArgumentRebuttals = append(jGraph.CounterArgumentRebuttals, &jsonCounterArgumentRebuttal{
			RebuttalArgument: r.RebuttalNode.Argument,
			TargetArgument:   r.TargetNode.Argument,
		})
	}

	// ターンアラウンドの変換
	for _, r := range dg.TurnArgumentRebuttals {
		jGraph.TurnArgumentRebuttals = append(jGraph.TurnArgumentRebuttals, &jsonTurnArgumentRebuttal{
			RebuttalArgument: r.RebuttalNode.Argument,
		})
	}

	jsonData, err := json.MarshalIndent(jGraph, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DebateGraph to JSON: %w", err)
	}

	return string(jsonData), nil
}

// NewDebateGraphFromJSON はJSON文字列からDebateGraphを復元します。(新規追加)
func NewDebateGraphFromJSON(jsonData string) (*DebateGraph, error) {
	var jGraph jsonGraph
	if err := json.Unmarshal([]byte(jsonData), &jGraph); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to DebateGraph: %w", err)
	}

	dg := NewDebateGraph()

	// 1. ノードをすべて構築
	for _, jNode := range jGraph.Nodes {
		node := NewDebateGraphNode(jNode.Argument, jNode.IsRebuttal)
		node.Importance = jNode.Importance
		node.Uniqueness = jNode.Uniqueness
		node.ImportanceRebuttals = jNode.ImportanceRebuttals
		node.UniquenessRebuttals = jNode.UniquenessRebuttals
		if err := dg.AddNode(node); err != nil {
			return nil, fmt.Errorf("failed to add node '%s' from JSON: %w", jNode.Argument, err)
		}
	}

	// 2. エッジをすべて構築
	for _, jEdge := range jGraph.Edges {
		causeNode, causeExists := dg.GetNode(jEdge.Cause)
		if !causeExists {
			return nil, fmt.Errorf("cause node '%s' for edge not found in graph", jEdge.Cause)
		}
		effectNode, effectExists := dg.GetNode(jEdge.Effect)
		if !effectExists {
			return nil, fmt.Errorf("effect node '%s' for edge not found in graph", jEdge.Effect)
		}

		edge := NewDebateGraphEdge(causeNode, effectNode, jEdge.IsRebuttal)
		edge.Certainty = jEdge.Certainty
		edge.Uniqueness = jEdge.Uniqueness
		edge.CertaintyRebuttal = jEdge.CertaintyRebuttal
		edge.UniquenessRebuttals = jEdge.UniquenessRebuttals

		if err := dg.AddEdge(edge); err != nil {
			return nil, fmt.Errorf("failed to add edge '%s -> %s' from JSON: %w", jEdge.Cause, jEdge.Effect, err)
		}
	}

	// 3. ノード反論を再構築
	for _, jRebuttal := range jGraph.NodeRebuttals {
		targetNode, exists := dg.GetNode(jRebuttal.TargetArgument)
		if !exists {
			return nil, fmt.Errorf("target node '%s' for node rebuttal not found", jRebuttal.TargetArgument)
		}
		rebuttalNode, exists := dg.GetNode(jRebuttal.RebuttalArgument)
		if !exists {
			return nil, fmt.Errorf("rebuttal node '%s' for node rebuttal not found", jRebuttal.RebuttalArgument)
		}

		rebuttal := &DebateGraphNodeRebuttal{
			TargetNode:   targetNode,
			RebuttalType: jRebuttal.RebuttalType,
			RebuttalNode: rebuttalNode,
		}
		dg.NodeRebuttals = append(dg.NodeRebuttals, rebuttal)
	}

	// 4. エッジ反論を再構築
	for _, jRebuttal := range jGraph.EdgeRebuttals {
		targetEdge, exists := dg.GetEdge(jRebuttal.TargetCauseArgument, jRebuttal.TargetEffectArgument)
		if !exists {
			return nil, fmt.Errorf("target edge '%s -> %s' for edge rebuttal not found", jRebuttal.TargetCauseArgument, jRebuttal.TargetEffectArgument)
		}
		rebuttalNode, exists := dg.GetNode(jRebuttal.RebuttalArgument)
		if !exists {
			return nil, fmt.Errorf("rebuttal node '%s' for edge rebuttal not found", jRebuttal.RebuttalArgument)
		}

		rebuttal := &DebateGraphEdgeRebuttal{
			TargetEdge:   targetEdge,
			RebuttalType: jRebuttal.RebuttalType,
			RebuttalNode: rebuttalNode,
		}
		dg.EdgeRebuttals = append(dg.EdgeRebuttals, rebuttal)
	}

	// 5. 反対意見を再構築
	for _, jRebuttal := range jGraph.CounterArgumentRebuttals {
		rebuttalNode, exists := dg.GetNode(jRebuttal.RebuttalArgument)
		if !exists {
			return nil, fmt.Errorf("rebuttal node '%s' for counter argument rebuttal not found", jRebuttal.RebuttalArgument)
		}

		targetNode, exiexists := dg.GetNode(jRebuttal.TargetArgument)
		if !exiexists {
			return nil, fmt.Errorf("target node '%s' for counter argument rebuttal not found", jRebuttal.TargetArgument)
		}

		rebuttal := &CounterArgumentRebuttal{
			RebuttalNode: rebuttalNode,
			TargetNode:   targetNode,
		}
		dg.CounterArgumentRebuttals = append(dg.CounterArgumentRebuttals, rebuttal)
	}

	// 6. ターンアラウンドを再構築
	for _, jRebuttal := range jGraph.TurnArgumentRebuttals {
		rebuttalNode, exists := dg.GetNode(jRebuttal.RebuttalArgument)
		if !exists {
			return nil, fmt.Errorf("rebuttal node '%s' for turn argument rebuttal not found", jRebuttal.RebuttalArgument)
		}

		rebuttal := &TurnArgumentRebuttal{
			RebuttalNode: rebuttalNode,
		}
		dg.TurnArgumentRebuttals = append(dg.TurnArgumentRebuttals, rebuttal)
	}

	return dg, nil
}
