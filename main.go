package main

import (
	"log"
	"net/http"

	createrebuttal "github.com/wolfmagnate/auto_debater/create_rebuttal"
	"github.com/wolfmagnate/auto_debater/handler"
	"github.com/wolfmagnate/auto_debater/logic_composer"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// すべてのオリジンからのリクエストを許可
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// 許可するHTTPメソッド
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		// 許可するヘッダー
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// プリフライトリクエストに対応
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 次のハンドラを実行
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. 依存関係の初期化
	rebuttalCreator, err := createrebuttal.NewRebuttalCreator()
	if err != nil {
		log.Fatalf("FATAL: Failed to create rebuttal creator: %v", err)
	}

	logicEnhancer, err := logic_composer.CreateLogicEnhancer()
	if err != nil {
		log.Fatalf("FATAL: Failed to create logic enhancer: %v", err)
	}

	// 2. ハンドラを初期化 (両方の依存を注入)
	apiHandler := handler.NewHandler(rebuttalCreator, logicEnhancer)

	// 3. エンドポイントを登録
	http.Handle("/api/create-rebuttal", corsMiddleware(http.HandlerFunc(apiHandler.CreateRebuttalEndpoint)))
	http.Handle("/api/enhance-logic", corsMiddleware(http.HandlerFunc(apiHandler.EnhanceLogicEndpoint)))

	// 4. サーバーを起動
	port := ":8080"
	log.Printf("INFO: Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}
}
