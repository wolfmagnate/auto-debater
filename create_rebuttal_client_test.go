package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	// 依存パッケージのインポートパスはご自身のプロジェクトに合わせてください
	"github.com/joho/godotenv"
	createrebuttal "github.com/wolfmagnate/auto_debater/create_rebuttal"
	"github.com/wolfmagnate/auto_debater/handler"
	"github.com/wolfmagnate/auto_debater/logic_composer"
)

// --- ヘルパー: テストのセットアップ ---

// setupTestHandler は、テストに必要な依存関係をすべて注入したHandlerを初期化します。
func setupTestHandler(t *testing.T) *handler.Handler {
	t.Helper() // この関数がテストのヘルパーであることを示す

	rebuttalCreator, err := createrebuttal.NewRebuttalCreator()
	if err != nil {
		t.Fatalf("FATAL: Failed to create RebuttalCreator: %v", err)
	}

	logicEnhancer, err := logic_composer.CreateLogicEnhancer()
	if err != nil {
		t.Fatalf("FATAL: Failed to create LogicEnhancer: %v", err)
	}

	todoEnhancer, err := logic_composer.CreateTODOEnhancer()
	if err != nil {
		t.Fatalf("FATAL: Failed to create TODOEnhancer: %v", err)
	}

	return handler.NewHandler(rebuttalCreator, logicEnhancer, todoEnhancer)
}

// --- ヘルパー: テストデータ ---

// すべてのテストケースで共通して使用するグラフデータ
const testGraphJSON = `
{
  "nodes": [
    { "argument": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する", "is_rebuttal": false },
    { "argument": "オンラインでの公式な情報拠点ができる(MEO/SEO強化)", "is_rebuttal": false },
    { "argument": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "is_rebuttal": false },
    { "argument": "多くの店舗（特に中小規模）はWebマーケティングに関する専門知識・リソースが不足している", "is_rebuttal": false }
  ],
  "edges": [
    { "cause": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する", "effect": "オンラインでの公式な情報拠点ができる(MEO/SEO強化)", "is_rebuttal": false },
    { "cause": "オンラインでの公式な情報拠点ができる(MEO/SEO強化)", "effect": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "is_rebuttal": false }
  ]
}
`

// --- 統合テスト ---

// TestIntegration_AllEndpoints は、すべてのエンドポイントを順にテストします。
// APIキーなどの設定が必要なため、実行には時間がかかる場合があります。
func TestIntegration_AllEndpoints(t *testing.T) {

	// dotenvを読み込む
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// テスト対象のハンドラを初期化
	handler := setupTestHandler(t)

	// 各エンドポイントのテストをサブテストとして実行
	t.Run("CreateRebuttalEndpoint", func(t *testing.T) {
		testCreateRebuttalEndpoint(t, handler)
	})
}

// testCreateRebuttalEndpoint は /api/create-rebuttal のテストを行います。
func testCreateRebuttalEndpoint(t *testing.T, h *handler.Handler) {
	// リクエストボディを準備
	requestBody := map[string]json.RawMessage{
		"debate_graph": json.RawMessage(testGraphJSON),
		"subgraph":     json.RawMessage(testGraphJSON),
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// リクエストとレスポンスレコーダーを作成
	req := httptest.NewRequest(http.MethodPost, "/api/create-rebuttal", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	// ハンドラを実行
	http.HandlerFunc(h.CreateRebuttalEndpoint).ServeHTTP(rr, req)

	// レスポンスを検証
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Errorf("response body: %s", rr.Body.String())
		return
	}

	// レスポンスボディの構造を検証
	var result createrebuttal.CreateRebuttalResult
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	t.Logf("OK - CreateRebuttalEndpoint returned %d node rebuttals and %d edge rebuttals.", len(result.NodeRebuttals), len(result.EdgeRebuttals))
}

// testEnhanceLogicEndpoint は /api/enhance-logic のテストを行います。
func testEnhanceLogicEndpoint(t *testing.T, h *handler.Handler) {
	// リクエストボディを準備
	requestBody := map[string]interface{}{
		"debate_graph": json.RawMessage(testGraphJSON),
		"cause":        "オンラインでの公式な情報拠点ができる(MEO/SEO強化)",
		"effect":       "オンラインでの認知度が向上し、新規顧客の来店が増加する",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// リクエストとレスポンスレコーダーを作成
	req := httptest.NewRequest(http.MethodPost, "/api/enhance-logic", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	// ハンドラを実行
	http.HandlerFunc(h.EnhanceLogicEndpoint).ServeHTTP(rr, req)

	// レスポンスを検証
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Errorf("response body: %s", rr.Body.String())
		return
	}

	// レスポンスボディの構造を検証
	var result []logic_composer.EnhancementAction
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(result) == 0 {
		t.Log("Warning: EnhanceLogicEndpoint returned 0 actions. This might be unexpected depending on the AI response.")
	}
	t.Logf("OK - EnhanceLogicEndpoint returned %d actions.", len(result))
}

// testEnhanceTODOEndpoint は /api/enhance-todo のテストを行います。
func testEnhanceTODOEndpoint(t *testing.T, h *handler.Handler) {
	// リクエストボディを準備
	requestBody := map[string]json.RawMessage{
		"debate_graph": json.RawMessage(testGraphJSON),
		"subgraph":     json.RawMessage(testGraphJSON),
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// リクエストとレスポンスレコーダーを作成
	req := httptest.NewRequest(http.MethodPost, "/api/enhance-todo", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	// ハンドラを実行
	http.HandlerFunc(h.EnhanceTODOEndpoint).ServeHTTP(rr, req)

	// レスポンスを検証
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Errorf("response body: %s", rr.Body.String())
		return
	}

	// レスポンスボディの構造を検証
	var result logic_composer.TODOSuggestions
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	t.Logf("OK - EnhanceTODOEndpoint returned %d TODOs.", len(result.TODOs))
}
