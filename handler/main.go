package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	createrebuttal "github.com/wolfmagnate/auto_debater/create_rebuttal"
	"github.com/wolfmagnate/auto_debater/domain"
	"github.com/wolfmagnate/auto_debater/logic_composer"
)

// Handler は、アプリケーションのHTTPハンドラと依存関係を保持します。
// RebuttalCreatorはポインタで保持するのが一般的です。
type Handler struct {
	RebuttalCreator *createrebuttal.RebuttalCreator
	LogicEnhancer   *logic_composer.LogicEnhancer
	TODOEnhancer    *logic_composer.TODOEnhancer
}

// NewHandler は、依存関係を注入して新しいHandlerを生成します。
func NewHandler(creator *createrebuttal.RebuttalCreator, logicEnhancer *logic_composer.LogicEnhancer, todoEnhancer *logic_composer.TODOEnhancer) *Handler {
	return &Handler{
		RebuttalCreator: creator,
		LogicEnhancer:   logicEnhancer,
		TODOEnhancer:    todoEnhancer,
	}
}

// CreateRebuttalRequest は、反論生成エンドポイントへのリクエストボディの構造を定義します。
type CreateRebuttalRequest struct {
	DebateGraphJSON json.RawMessage `json:"debate_graph"`
	SubgraphJSON    json.RawMessage `json:"subgraph"`
}

// CreateRebuttalEndpoint は、JSONでDebateGraphとsubGraphを受け取り、
// subGraphに反論を追加して、更新されたsubGraphを返すHTTPハンドラです。
func (h *Handler) CreateRebuttalEndpoint(w http.ResponseWriter, r *http.Request) {
	// 1. HTTPメソッドがPOSTであることを確認します。
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 2. リクエストボディからJSONデータを読み込みます。
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Could not read request body: %v", err)
		http.Error(w, "Could not read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 3. リクエストJSONをデコードします。
	var req CreateRebuttalRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("ERROR: Could not unmarshal request JSON: %v", err)
		http.Error(w, "Bad request: invalid JSON format", http.StatusBadRequest)
		return
	}

	// 必須フィールドの存在を検証します。
	if len(req.DebateGraphJSON) == 0 || len(req.SubgraphJSON) == 0 {
		http.Error(w, "Bad request: 'debate_graph' and 'subgraph' fields are required", http.StatusBadRequest)
		return
	}

	// 4. JSONから内部のDebateGraphオブジェクトに組み立てます。
	debateGraph, err := domain.NewDebateGraphFromJSON(string(req.DebateGraphJSON))
	if err != nil {
		log.Printf("ERROR: Could not create main graph from JSON: %v", err)
		http.Error(w, "Bad request: invalid debate_graph structure", http.StatusBadRequest)
		return
	}

	subGraph, err := domain.NewDebateGraphFromJSON(string(req.SubgraphJSON))
	if err != nil {
		log.Printf("ERROR: Could not create subgraph from JSON: %v", err)
		http.Error(w, "Bad request: invalid subgraph structure", http.StatusBadRequest)
		return
	}

	log.Println("INFO: Successfully created graphs from JSON. Starting rebuttal creation for subgraph...")

	// 5. RebuttalCreatorを使って、subGraphに反論を追加する処理を実行します。
	h.RebuttalCreator.CreateRebuttal(r.Context(), debateGraph, subGraph)

	log.Println("INFO: Rebuttal creation finished. Converting updated subgraph to JSON...")

	// 6. 更新されたsubGraphをJSONに変換します。
	updatedSubgraphJSON, err := subGraph.ToJSON()
	if err != nil {
		log.Printf("ERROR: Could not convert updated subgraph to JSON: %v", err)
		http.Error(w, "Internal server error while processing the graph", http.StatusInternalServerError)
		return
	}

	// 7. 成功したレスポンスとして、更新後のsubGraphのJSONを返します。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(updatedSubgraphJSON)); err != nil {
		log.Printf("ERROR: Could not write response: %v", err)
	}

	log.Println("INFO: Successfully sent updated subgraph as response.")
}

type EnhanceLogicRequest struct {
	DebateGraphJSON json.RawMessage `json:"debate_graph"`
	Cause           string          `json:"cause"`
	Effect          string          `json:"effect"`
}

// EnhanceLogicEndpoint は、二つのノード間の因果関係を強化する提案を生成するHTTPハンドラです。
func (h *Handler) EnhanceLogicEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Could not read request body: %v", err)
		http.Error(w, "Could not read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// リクエストのJSONペイロードをデコードします。
	var req EnhanceLogicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("ERROR: Could not unmarshal request JSON: %v", err)
		http.Error(w, "Bad request: invalid JSON format", http.StatusBadRequest)
		return
	}

	// 必須フィールドの存在を検証します。
	if len(req.DebateGraphJSON) == 0 || req.Cause == "" || req.Effect == "" {
		http.Error(w, "Bad request: 'debate_graph', 'cause', and 'effect' fields are required", http.StatusBadRequest)
		return
	}

	// 受け取ったJSONからメインのDebateGraphオブジェクトを構築します。
	debateGraph, err := domain.NewDebateGraphFromJSON(string(req.DebateGraphJSON))
	if err != nil {
		log.Printf("ERROR: Could not create graph from JSON: %v", err)
		http.Error(w, "Bad request: invalid debate_graph structure", http.StatusBadRequest)
		return
	}

	// (任意ですが推奨) CauseノードとEffectノードがグラフ内に存在するかを検証します。
	if _, exists := debateGraph.GetNode(req.Cause); !exists {
		http.Error(w, "Bad request: 'cause' node does not exist in the provided graph", http.StatusBadRequest)
		return
	}
	if _, exists := debateGraph.GetNode(req.Effect); !exists {
		http.Error(w, "Bad request: 'effect' node does not exist in the provided graph", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Starting logic enhancement for: [%s] -> [%s]", req.Cause, req.Effect)

	// コア機能であるLogicEnhancerを呼び出します。
	enhancements, err := h.LogicEnhancer.EnhanceLogic(r.Context(), debateGraph, req.Cause, req.Effect)
	if err != nil {
		log.Printf("ERROR: Logic enhancement process failed: %v", err)
		http.Error(w, "Internal server error during logic enhancement", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Successfully generated %d enhancement actions.", len(enhancements))

	// 結果の[]EnhancementActionスライスをJSONに変換します。
	responseJSON, err := json.Marshal(enhancements)
	if err != nil {
		log.Printf("ERROR: Failed to marshal enhancement actions to JSON: %v", err)
		http.Error(w, "Internal server error while formatting response", http.StatusInternalServerError)
		return
	}

	// 成功レスポンスを返します。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(responseJSON); err != nil {
		log.Printf("ERROR: Could not write response: %v", err)
	}

	log.Println("INFO: Successfully sent enhancement actions as response.")
}

type EnhanceTODORequest struct {
	DebateGraphJSON json.RawMessage `json:"debate_graph"`
	SubgraphJSON    json.RawMessage `json:"subgraph"`
}

// EnhanceTODOEndpoint は、サブグラフを改善するためのTODOリストを提案するHTTPハンドラです。
func (h *Handler) EnhanceTODOEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Could not read request body: %v", err)
		http.Error(w, "Could not read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// リクエストのJSONペイロードをデコードします。
	var req EnhanceTODORequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("ERROR: Could not unmarshal request JSON: %v", err)
		http.Error(w, "Bad request: invalid JSON format", http.StatusBadRequest)
		return
	}

	// 必須フィールドの存在を検証します。
	if len(req.DebateGraphJSON) == 0 || len(req.SubgraphJSON) == 0 {
		http.Error(w, "Bad request: 'debate_graph' and 'subgraph' fields are required", http.StatusBadRequest)
		return
	}

	// 受け取ったJSONからメインのDebateGraphオブジェクトを構築します。
	debateGraph, err := domain.NewDebateGraphFromJSON(string(req.DebateGraphJSON))
	if err != nil {
		log.Printf("ERROR: Could not create main graph from JSON: %v", err)
		http.Error(w, "Bad request: invalid debate_graph structure", http.StatusBadRequest)
		return
	}

	// 受け取ったJSONからサブグラフのDebateGraphオブジェクトを構築します。
	subGraph, err := domain.NewDebateGraphFromJSON(string(req.SubgraphJSON))
	if err != nil {
		log.Printf("ERROR: Could not create subgraph from JSON: %v", err)
		http.Error(w, "Bad request: invalid subgraph structure", http.StatusBadRequest)
		return
	}

	log.Println("INFO: Starting TODO enhancement...")

	// コア機能であるTODOEnhancerを呼び出します。
	suggestions, err := h.TODOEnhancer.EnhanceTODO(r.Context(), debateGraph, subGraph)
	if err != nil {
		log.Printf("ERROR: TODO enhancement process failed: %v", err)
		http.Error(w, "Internal server error during TODO enhancement", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Successfully generated %d TODO suggestions.", len(suggestions.TODOs))

	// 結果のTODOSuggestionsをJSONに変換します。
	responseJSON, err := json.Marshal(suggestions)
	if err != nil {
		log.Printf("ERROR: Failed to marshal TODO suggestions to JSON: %v", err)
		http.Error(w, "Internal server error while formatting response", http.StatusInternalServerError)
		return
	}

	// 成功レスポンスを返します。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(responseJSON); err != nil {
		log.Printf("ERROR: Could not write response: %v", err)
	}

	log.Println("INFO: Successfully sent TODO suggestions as response.")
}
