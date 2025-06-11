package main

import (
	"log"
	"net/http"

	createrebuttal "github.com/wolfmagnate/auto_debater/create_rebuttal"
	"github.com/wolfmagnate/auto_debater/handler"
	"github.com/wolfmagnate/auto_debater/logic_composer"
)

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
	http.HandleFunc("/api/create-rebuttal", apiHandler.CreateRebuttalEndpoint)
	http.HandleFunc("/api/enhance-logic", apiHandler.EnhanceLogicEndpoint) // 新しいエンドポイント

	// 4. サーバーを起動
	port := ":8080"
	log.Printf("INFO: Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}
}
