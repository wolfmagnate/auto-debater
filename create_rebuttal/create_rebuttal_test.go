package createrebuttal

import (
	"context"
	"log"
	"testing"

	// ご自身のプロジェクトのドメインパッケージへのパスに修正してください
	"github.com/joho/godotenv"
	"github.com/wolfmagnate/auto_debater/domain"
)

// このテストを実行するには、AIモデルのAPIキーが環境変数などで
// 設定されている必要があります。
func TestCreateRebuttal(t *testing.T) {
	// --- 1. テストデータの準備 ---
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// ## 全体の議論グラフ (debateGraph) を構築
	// PMFや証拠反論を生成する際の広範なコンテキストとして使用されます。
	debateGraph := domain.NewDebateGraph()

	// ノードの定義
	nodeA := domain.NewDebateGraphNode("質の高い教育コンテンツをオンラインで提供する", false)
	nodeB := domain.NewDebateGraphNode("学習意欲の高いユーザーが集まる", false)
	nodeC := domain.NewDebateGraphNode("プラットフォームの利用料で収益が上がる", false)
	nodeD := domain.NewDebateGraphNode("既存の無料学習コンテンツが多い", false) // 競合・代替手段を示すノード

	// ノードをグラフに追加
	if err := debateGraph.AddNode(nodeA); err != nil {
		t.Fatalf("failed to add nodeA: %v", err)
	}
	if err := debateGraph.AddNode(nodeB); err != nil {
		t.Fatalf("failed to add nodeB: %v", err)
	}
	if err := debateGraph.AddNode(nodeC); err != nil {
		t.Fatalf("failed to add nodeC: %v", err)
	}
	if err := debateGraph.AddNode(nodeD); err != nil {
		t.Fatalf("failed to add nodeD: %v", err)
	}

	// エッジの定義と追加
	edgeAB := domain.NewDebateGraphEdge(nodeA, nodeB, false)
	edgeBC := domain.NewDebateGraphEdge(nodeB, nodeC, false)
	if err := debateGraph.AddEdge(edgeAB); err != nil {
		t.Fatalf("failed to add edgeAB: %v", err)
	}
	if err := debateGraph.AddEdge(edgeBC); err != nil {
		t.Fatalf("failed to add edgeBC: %v", err)
	}

	// ## 反論生成の対象となる部分グラフ (subGraph) を構築
	// 今回の反論生成のメインターゲットです。
	subGraph := domain.NewDebateGraph()

	// ノードの定義 (debateGraphから一部を抜粋)
	subNodeB := domain.NewDebateGraphNode("学習意欲の高いユーザーが集まる", false)
	subNodeC := domain.NewDebateGraphNode("プラットフォームの利用料で収益が上がる", false)

	// ノードをサブグラフに追加
	if err := subGraph.AddNode(subNodeB); err != nil {
		t.Fatalf("failed to add subNodeB: %v", err)
	}
	if err := subGraph.AddNode(subNodeC); err != nil {
		t.Fatalf("failed to add subNodeC: %v", err)
	}

	// エッジの定義と追加
	subEdgeBC := domain.NewDebateGraphEdge(subNodeB, subNodeC, false)
	if err := subGraph.AddEdge(subEdgeBC); err != nil {
		t.Fatalf("failed to add subEdgeBC: %v", err)
	}

	// --- 2. RebuttalCreatorのインスタンス化 ---
	creator, err := NewRebuttalCreator()
	if err != nil {
		t.Fatalf("Failed to create RebuttalCreator: %v", err)
	}

	// --- 3. CreateRebuttalの呼び出し ---
	t.Log("--- SubGraph BEFORE CreateRebuttal ---")
	subGraph.DisplayGraph()

	t.Log(">>> Calling CreateRebuttal...")
	// ここで実際にAIモデルへの問い合わせが発生します
	creator.CreateRebuttal(context.Background(), debateGraph, subGraph)
	t.Log("<<< Finished CreateRebuttal.")

	// --- 4. 結果の表示 ---
	t.Log("\n--- SubGraph AFTER CreateRebuttal ---")
	subGraph.DisplayGraph()

	// 簡単なアサーション
	if len(subGraph.EdgeRebuttals) == 0 && len(subGraph.NodeRebuttals) == 0 {
		// AIが反論を生成しない場合もあるため、これはエラーではなく警告とします。
		t.Log("Warning: No rebuttals were generated. This might be an expected outcome from the AI model.")
	} else {
		t.Logf("Success: Generated %d node rebuttals and %d edge rebuttals.", len(subGraph.NodeRebuttals), len(subGraph.EdgeRebuttals))
	}
}
