package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	createrebuttal "github.com/wolfmagnate/auto_debater/create_rebuttal"
	"github.com/wolfmagnate/auto_debater/handler"
	"github.com/wolfmagnate/auto_debater/logic_composer"
)

// findProjectRootは、go:embedで埋め込まれたファイルをテストで正しく読み込むために、
// 'go.mod'ファイルを目印にプロジェクトのルートディレクトリを探します。
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// TestEnhanceLogicEndpoint_Integration は、/api/enhance-logicエンドポイントの統合テストです。
// 注意: このテストは実際にAIモデルへのAPI呼び出しを行う可能性があります。
// CI/CD環境で実行する場合は、infra.ChatCompletionHandlerをモックに差し替えることを推奨します。
func TestEnhanceLogicEndpoint_Integration(t *testing.T) {
	// --- 1. テストの準備 ---

	// dotenvを読み込む
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// go:embedが正しく機能するように、カレントディレクトリをプロジェクトルートに設定
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	projectRoot, err := findProjectRoot()
	require.NoError(t, err, "go.modファイルが見つかりません。プロジェクトのルートでテストを実行してください。")
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(originalWD) // テスト終了時にカレントディレクトリを元に戻す

	// 依存関係を初期化
	rebuttalCreator, err := createrebuttal.NewRebuttalCreator()
	if err != nil {
		log.Fatalf("FATAL: Failed to create rebuttal creator: %v", err)
	}

	logicEnhancer, err := logic_composer.CreateLogicEnhancer()
	if err != nil {
		log.Fatalf("FATAL: Failed to create logic enhancer: %v", err)
	}

	// テスト対象のハンドラとテストサーバーをセットアップ
	apiHandler := handler.NewHandler(rebuttalCreator, logicEnhancer)
	testServer := httptest.NewServer(http.HandlerFunc(apiHandler.EnhanceLogicEndpoint))
	defer testServer.Close()

	// --- 2. リクエストの準備と実行 ---

	// テスト用のリクエストボディを定義
	requestJSON := `{
		"debate_graph": {
			"nodes": [
				{ "argument": "再生可能エネルギーの導入が増加する", "is_rebuttal": false },
				{ "argument": "CO2排出量が削減される", "is_rebuttal": false },
				{ "argument": "地球温暖化の進行が緩和される", "is_rebuttal": false }
			],
			"edges": [
				{ "cause": "再生可能エネルギーの導入が増加する", "effect": "CO2排出量が削減される", "is_rebuttal": false },
				{ "cause": "CO2排出量が削減される", "effect": "地球温暖化の進行が緩和される", "is_rebuttal": false }
			]
		},
		"cause": "再生可能エネルギーの導入が増加する",
		"effect": "CO2排出量が削減される"
	}`

	// APIにPOSTリクエストを送信
	res, err := http.Post(testServer.URL, "application/json", bytes.NewBufferString(requestJSON))
	require.NoError(t, err, "HTTPリクエストの送信に失敗しました。")
	defer res.Body.Close()

	// --- 3. レスポンスの検証 ---

	// ステータスコードを検証
	assert.Equal(t, http.StatusOK, res.StatusCode, "期待されるHTTPステータスコードは200 OKです。")

	// Content-Typeヘッダーを検証
	assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"), "Content-Typeヘッダーが正しくありません。")

	// レスポンスボディをデコード
	responseBodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err, "レスポンスボディの読み込みに失敗しました。")

	log.Printf("Raw JSON Response Body:\n%s", string(responseBodyBytes))

	var enhancementActions []logic_composer.EnhancementAction
	err = json.Unmarshal(responseBodyBytes, &enhancementActions)
	require.NoError(t, err, "レスポンスボディのJSONデコードに失敗しました。")

	// AIの応答は非決定的だが、基本的な構造は検証可能
	assert.NotEmpty(t, enhancementActions, "レスポンスの配列が空であってはいけません。")
	t.Logf("AIから %d 件のロジック強化アクションが提案されました。", len(enhancementActions))

	// 各アクションが期待される構造を持っているか検証
	for i, action := range enhancementActions {
		isValidAction := action.InsertNode != nil || action.StrengthenEdge != nil
		assert.True(t, isValidAction, "インデックス %d のアクションに有効なペイロードが含まれていません。", i)

		// 内容をログに出力して確認
		if action.InsertNode != nil {
			log.Printf("  - アクション %d: [ノード挿入] 中間ノード: '%s'", i, action.InsertNode.IntermediateArgument)
		} else if action.StrengthenEdge != nil {
			log.Printf("  - アクション %d: [エッジ強化] 種類: %s, 内容: '%s'", i, action.StrengthenEdge.EnhancementType, action.StrengthenEdge.Content)
		}
	}
}

func TestCreateRebuttalEndpoint_Integration(t *testing.T) {
	// --- 1. テストの準備 ---

	// dotenvを読み込む
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// go:embedが正しく機能するように、カレントディレクトリをプロジェクトルートに設定
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	projectRoot, err := findProjectRoot()
	require.NoError(t, err, "go.modファイルが見つかりません。プロジェクトのルートでテストを実行してください。")
	require.NoError(t, os.Chdir(projectRoot))
	defer os.Chdir(originalWD) // テスト終了時にカレントディレクトリを元に戻す

	// 依存関係を初期化
	rebuttalCreator, err := createrebuttal.NewRebuttalCreator()
	if err != nil {
		log.Fatalf("FATAL: Failed to create rebuttal creator: %v", err)
	}

	logicEnhancer, err := logic_composer.CreateLogicEnhancer()
	if err != nil {
		log.Fatalf("FATAL: Failed to create logic enhancer: %v", err)
	}

	// テスト対象のハンドラとテストサーバーをセットアップ
	apiHandler := handler.NewHandler(rebuttalCreator, logicEnhancer)
	testServer := httptest.NewServer(http.HandlerFunc(apiHandler.CreateRebuttalEndpoint))
	defer testServer.Close()

	// --- 2. リクエストの準備と実行 ---

	// テスト用のリクエストボディ(初期グラフ)を定義
	requestJSON := `{
		"nodes": [
			{ "argument": "多くの小規模飲食店は、専門知識や時間不足から効果的なオンライン集客ができていない", "is_rebuttal": false },
			{ "argument": "潜在顧客にリーチできず、機会損失が発生している", "is_rebuttal": false },
			{ "argument": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する", "is_rebuttal": false },
			{ "argument": "オーナーは本来の調理・接客業務に集中できる", "is_rebuttal": false },
			{ "argument": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "is_rebuttal": false }
		],
		"edges": [
			{
				"cause": "多くの小規模飲食店は、専門知識や時間不足から効果的なオンライン集客ができていない",
				"effect": "潜在顧客にリーチできず、機会損失が発生している",
				"is_rebuttal": false
			},
			{
				"cause": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する",
				"effect": "オーナーは本来の調理・接客業務に集中できる",
				"is_rebuttal": false
			},
			{
				"cause": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する",
				"effect": "オンラインでの認知度が向上し、新規顧客の来店が増加する",
				"is_rebuttal": false
			}
		]
	}`

	// APIにPOSTリクエストを送信
	res, err := http.Post(testServer.URL, "application/json", bytes.NewBufferString(requestJSON))
	require.NoError(t, err, "HTTPリクエストの送信に失敗しました。")
	defer res.Body.Close()

	// --- 3. レスポンスの検証 ---

	assert.Equal(t, http.StatusOK, res.StatusCode, "期待されるHTTPステータスコードは200 OKです。")
	assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"), "Content-Typeヘッダーが正しくありません。")

	responseBodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err, "レスポンスボディの読み込みに失敗しました。")

	// レスポンスの生JSONをログに出力
	t.Logf("Raw JSON Response Body:\n%s", string(responseBodyBytes))

	// レスポンスJSONをマップにデコードして構造を検証
	var updatedGraphData map[string]interface{}
	err = json.Unmarshal(responseBodyBytes, &updatedGraphData)
	require.NoError(t, err, "レスポンスボディのJSONデコードに失敗しました。")
}
