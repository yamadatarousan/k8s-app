package todoapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/k8s-app/backend/todoapi"
)

func TestTodoAPI_Todoを追加して一覧取得できる(t *testing.T) {
	t.Parallel()

	server := todoapi.NewServer(todoapi.Config{})

	作成リクエストボディ, err := json.Marshal(map[string]string{
		"title": "k8sでTodoアプリを動かす",
	})
	if err != nil {
		t.Fatalf("作成リクエストのJSON化に失敗: %v", err)
	}

	作成リクエスト := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewReader(作成リクエストボディ))
	作成リクエスト.Header.Set("Content-Type", "application/json")
	作成レスポンス := httptest.NewRecorder()

	server.ServeHTTP(作成レスポンス, 作成リクエスト)

	if 作成レスポンス.Code != http.StatusCreated {
		t.Fatalf("期待したステータスは201、実際は%d", 作成レスポンス.Code)
	}

	var 作成結果 struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}
	if err := json.Unmarshal(作成レスポンス.Body.Bytes(), &作成結果); err != nil {
		t.Fatalf("作成レスポンスのJSON解析に失敗: %v", err)
	}

	if 作成結果.ID == "" {
		t.Fatal("作成したTodoにIDが入っていない")
	}
	if 作成結果.Title != "k8sでTodoアプリを動かす" {
		t.Fatalf("期待したタイトルと異なる: %q", 作成結果.Title)
	}
	if 作成結果.Completed {
		t.Fatal("新規Todoは未完了で作成される想定")
	}

	一覧リクエスト := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	一覧レスポンス := httptest.NewRecorder()

	server.ServeHTTP(一覧レスポンス, 一覧リクエスト)

	if 一覧レスポンス.Code != http.StatusOK {
		t.Fatalf("期待したステータスは200、実際は%d", 一覧レスポンス.Code)
	}

	var 一覧結果 []struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}
	if err := json.Unmarshal(一覧レスポンス.Body.Bytes(), &一覧結果); err != nil {
		t.Fatalf("一覧レスポンスのJSON解析に失敗: %v", err)
	}

	if len(一覧結果) != 1 {
		t.Fatalf("Todo件数は1件の想定、実際は%d件", len(一覧結果))
	}
	if 一覧結果[0].ID != 作成結果.ID {
		t.Fatalf("一覧の先頭IDが作成結果と一致しない: %q", 一覧結果[0].ID)
	}
}

func TestTodoAPI_空タイトルは受け付けない(t *testing.T) {
	t.Parallel()

	server := todoapi.NewServer(todoapi.Config{})

	リクエスト := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewReader([]byte(`{"title":""}`)))
	リクエスト.Header.Set("Content-Type", "application/json")
	レスポンス := httptest.NewRecorder()

	server.ServeHTTP(レスポンス, リクエスト)

	if レスポンス.Code != http.StatusBadRequest {
		t.Fatalf("期待したステータスは400、実際は%d", レスポンス.Code)
	}
}

func TestTodoAPI_ヘルスチェックを返す(t *testing.T) {
	t.Parallel()

	server := todoapi.NewServer(todoapi.Config{
		ApplicationName: "todo-api",
	})

	healthRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthResponse := httptest.NewRecorder()

	server.ServeHTTP(healthResponse, healthRequest)

	if healthResponse.Code != http.StatusOK {
		t.Fatalf("期待したステータスは200、実際は%d", healthResponse.Code)
	}

	var healthResult struct {
		Status          string `json:"status"`
		ApplicationName string `json:"applicationName"`
	}
	if err := json.Unmarshal(healthResponse.Body.Bytes(), &healthResult); err != nil {
		t.Fatalf("healthレスポンスのJSON解析に失敗: %v", err)
	}

	if healthResult.Status != "ok" {
		t.Fatalf("期待した状態はok、実際は%q", healthResult.Status)
	}
	if healthResult.ApplicationName != "todo-api" {
		t.Fatalf("期待したアプリ名と異なる: %q", healthResult.ApplicationName)
	}

	readyRequest := httptest.NewRequest(http.MethodGet, "/ready", nil)
	readyResponse := httptest.NewRecorder()

	server.ServeHTTP(readyResponse, readyRequest)

	if readyResponse.Code != http.StatusOK {
		t.Fatalf("期待したステータスは200、実際は%d", readyResponse.Code)
	}
}
