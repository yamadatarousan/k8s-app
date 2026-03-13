package todoapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

type Todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Server struct {
	mux   *http.ServeMux
	mutex sync.Mutex
	// 学習用の最小構成なので永続化はまだ持たず、Pod 再作成で消える前提にしている。
	// StatefulSet や外部 DB を導入する前の段階でも API の疎通確認に集中できる。
	todos      []Todo
	nextTodoID uint64
}

func NewServer() *Server {
	server := &Server{
		mux: http.NewServeMux(),
	}

	server.mux.HandleFunc("/api/todos", server.handleTodos)

	return server
}

func (server *Server) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	// k8s 配下では Frontend と Backend を別 Pod / Service に分けやすいため、
	// API 側で CORS を明示しておくとローカル確認と Ingress 経由の検証がしやすい。
	responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if request.Method == http.MethodOptions {
		responseWriter.WriteHeader(http.StatusNoContent)
		return
	}

	server.mux.ServeHTTP(responseWriter, request)
}

func (server *Server) handleTodos(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.handleListTodos(responseWriter)
	case http.MethodPost:
		server.handleCreateTodo(responseWriter, request)
	default:
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (server *Server) handleListTodos(responseWriter http.ResponseWriter) {
	server.mutex.Lock()
	// 内部スライスをそのまま返すと将来の変更で共有参照バグを埋め込みやすいため、
	// レスポンス用にコピーして外へ出す。
	todos := append([]Todo(nil), server.todos...)
	server.mutex.Unlock()

	writeJSON(responseWriter, http.StatusOK, todos)
}

func (server *Server) handleCreateTodo(responseWriter http.ResponseWriter, request *http.Request) {
	var createTodoRequest struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(request.Body).Decode(&createTodoRequest); err != nil {
		http.Error(responseWriter, "invalid json", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(createTodoRequest.Title)
	if title == "" {
		http.Error(responseWriter, "title is required", http.StatusBadRequest)
		return
	}

	todo := Todo{
		ID:        generateTodoID(&server.nextTodoID),
		Title:     title,
		Completed: false,
	}

	server.mutex.Lock()
	server.todos = append(server.todos, todo)
	server.mutex.Unlock()

	writeJSON(responseWriter, http.StatusCreated, todo)
}

func generateTodoID(nextTodoID *uint64) string {
	// 分散一意性までは不要なので、単一 Pod 前提で単調増加の連番に留める。
	sequence := atomic.AddUint64(nextTodoID, 1)
	return fmt.Sprintf("todo-%d", sequence)
}

func writeJSON(responseWriter http.ResponseWriter, statusCode int, value any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	_ = json.NewEncoder(responseWriter).Encode(value)
}
