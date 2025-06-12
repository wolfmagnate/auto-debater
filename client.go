package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// サーバーに送信するグラフのJSONデータ
const graphJSON = `
{
  "nodes": [
    { "argument": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する", "is_rebuttal": false },
    { "argument": "店舗の基本情報を入力するだけで、デザイン性の高いWebサイトが自動生成される（SEO対策済）", "is_rebuttal": false },
    { "argument": "セール情報や季節のイベントに応じて、AIが複数のSNS（Instagram, X, Facebook）に最適化された投稿コンテンツ（画像・テキスト）を自動で作成・予約投稿する", "is_rebuttal": false },
    { "argument": "オンラインでの公式な情報拠点ができる(MEO/SEO強化)", "is_rebuttal": false },
    { "argument": "SNSでの情報発信が継続的かつ高品質になる", "is_rebuttal": false },
    { "argument": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "is_rebuttal": false },
    { "argument": "リピーター増加や客単価向上にも繋がる(売上向上)", "is_rebuttal": false },
    { "argument": "店舗はSaaSの月額/年額利用料を支払う", "is_rebuttal": false },
    { "argument": "多くの店舗（特に中小規模）はWebマーケティングに関する専門知識・リソースが不足している", "is_rebuttal": false },
    { "argument": "Webサイト制作や更新、SNS運用のための時間や人手を確保できない", "is_rebuttal": false },
    { "argument": "外部の制作会社やコンサルタントに依頼するとコストが高い", "is_rebuttal": false },
    { "argument": "Webサイトがない、または情報が古い", "is_rebuttal": false },
    { "argument": "SNSアカウントはあるが、投稿が不定期になったり、魅力的なコンテンツを作成できない", "is_rebuttal": false },
    { "argument": "オンラインでの情報発信が不十分・魅力的でなく、潜在顧客にリーチできていない", "is_rebuttal": false },
    { "argument": "近隣の競合店に顧客を奪われ、機会損失が発生している", "is_rebuttal": false }
  ],
  "edges": [
    { "cause": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する", "effect": "店舗の基本情報を入力するだけで、デザイン性の高いWebサイトが自動生成される（SEO対策済）", "is_rebuttal": false },
    { "cause": "AIが店舗情報から自動でWebサイトやSNS投稿を生成するSaaSを提供する", "effect": "セール情報や季節のイベントに応じて、AIが複数のSNS（Instagram, X, Facebook）に最適化された投稿コンテンツ（画像・テキスト）を自動で作成・予約投稿する", "is_rebuttal": false },
    { "cause": "店舗の基本情報を入力するだけで、デザイン性の高いWebサイトが自動生成される（SEO対策済）", "effect": "オンラインでの公式な情報拠点ができる(MEO/SEO強化)", "is_rebuttal": false },
    { "cause": "セール情報や季節のイベントに応じて、AIが複数のSNS（Instagram, X, Facebook）に最適化された投稿コンテンツ（画像・テキスト）を自動で作成・予約投稿する", "effect": "SNSでの情報発信が継続的かつ高品質になる", "is_rebuttal": false },
    { "cause": "オンラインでの公式な情報拠点ができる(MEO/SEO強化)", "effect": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "is_rebuttal": false },
    { "cause": "SNSでの情報発信が継続的かつ高品質になる", "effect": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "is_rebuttal": false },
    { "cause": "オンラインでの認知度が向上し、新規顧客の来店が増加する", "effect": "リピーター増加や客単価向上にも繋がる(売上向上)", "is_rebuttal": false },
    { "cause": "リピーター増加や客単価向上にも繋がる(売上向上)", "effect": "店舗はSaaSの月額/年額利用料を支払う", "is_rebuttal": false },
    { "cause": "多くの店舗（特に中小規模）はWebマーケティングに関する専門知識・リソースが不足している", "effect": "Webサイト制作や更新、SNS運用のための時間や人手を確保できない", "is_rebuttal": false },
    { "cause": "多くの店舗（特に中小規模）はWebマーケティングに関する専門知識・リソースが不足している", "effect": "外部の制作会社やコンサルタントに依頼するとコストが高い", "is_rebuttal": false },
    { "cause": "Webサイト制作や更新、SNS運用のための時間や人手を確保できない", "effect": "Webサイトがない、または情報が古い", "is_rebuttal": false },
    { "cause": "Webサイト制作や更新、SNS運用のための時間や人手を確保できない", "effect": "SNSアカウントはあるが、投稿が不定期になったり、魅力的なコンテンツを作成できない", "is_rebuttal": false },
    { "cause": "外部の制作会社やコンサルタントに依頼するとコストが高い", "effect": "Webサイトがない、または情報が古い", "is_rebuttal": false },
    { "cause": "外部の制作会社やコンサルタントに依頼するとコストが高い", "effect": "SNSアカウントはあるが、投稿が不定期になったり、魅力的なコンテンツを作成できない", "is_rebuttal": false },
    { "cause": "Webサイトがない、または情報が古い", "effect": "オンラインでの情報発信が不十分・魅力的でなく、潜在顧客にリーチできていない", "is_rebuttal": false },
    { "cause": "SNSアカウントはあるが、投稿が不定期になったり、魅力的なコンテンツを作成できない", "effect": "オンラインでの情報発信が不十分・魅力的でなく、潜在顧客にリーチできていない", "is_rebuttal": false },
    { "cause": "オンラインでの情報発信が不十分・魅力的でなく、潜在顧客にリーチできていない", "effect": "近隣の競合店に顧客を奪われ、機会損失が発生している", "is_rebuttal": false }
  ]
}
`

// RebuttalRequest は /api/create-rebuttal エンドポイントへのリクエストボディの構造を定義します。
type RebuttalRequest struct {
	DebateGraph json.RawMessage `json:"debate_graph"`
	Subgraph    json.RawMessage `json:"subgraph"`
}

func main() {
	// APIエンドポイントのURL
	url := "https://auto-debater.onrender.com/api/create-rebuttal"

	// --- 1. リクエストボディの準備 ---
	// 提供されたグラフデータを debate_graph と subgraph の両方に使用します。
	requestPayload := RebuttalRequest{
		DebateGraph: json.RawMessage(graphJSON),
		Subgraph:    json.RawMessage(graphJSON),
	}

	// ペイロードをJSONバイトスライスに変換（マーシャリング）します。
	requestBodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Fatalf("リクエストペイロードのJSON変換に失敗しました: %v", err)
	}

	// --- 2. HTTPリクエストの作成 ---
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		log.Fatalf("HTTPリクエストの作成に失敗しました: %v", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Go-AutoDebater-Client/1.0")

	// --- 3. リクエストの送信 ---
	// タイムアウトを設定したHTTPクライアントを作成します。
	client := &http.Client{Timeout: 600 * time.Second}

	log.Println("リクエストを送信中:", url)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("リクエストの送信に失敗しました: %v", err)
	}
	defer resp.Body.Close()

	// --- 4. レスポンスの処理 ---
	log.Println("レスポンスステータス:", resp.Status)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("レスポンスボディの読み込みに失敗しました: %v", err)
	}

	// ステータスコードが200 OKでない場合はエラーとして処理します。
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("正常なステータスコード(200 OK)が返されませんでした: %d\nレスポンス: %s", resp.StatusCode, string(responseBody))
	}

	// 返ってきたJSONを整形して表示します。
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, responseBody, "", "  "); err != nil {
		// 整形に失敗しても、元のレスポンスは表示します。
		log.Printf("レスポンスJSONの整形に失敗しました: %v", err)
		fmt.Println("--- 未整形のレスポンスボディ ---")
		fmt.Println(string(responseBody))
		return
	}

	fmt.Println("--- レスポンスボディ ---")
	fmt.Println(prettyJSON.String())
}
