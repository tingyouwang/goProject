package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// httptest 是標準庫,讓你「不用真的開 port」就能測試 HTTP handler。
// 對照 Java 的 MockMvc,但更輕、內建、不需要 Spring Test。

// newTestServer 建立一個接好路由的測試伺服器。
func newTestServer() http.Handler {
	return NewHandler(NewMemoryStore()).Routes()
}

func TestHandler_CreateTask(t *testing.T) {
	srv := newTestServer()

	body := bytes.NewBufferString(`{"title":"prepare interview"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var got Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got.Title != "prepare interview" {
		t.Errorf("expected title %q, got %q", "prepare interview", got.Title)
	}
}

func TestHandler_CreateTask_MissingTitle(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"title":""}`))
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty title, got %d", rec.Code)
	}
}

func TestHandler_BatchCreate(t *testing.T) {
	srv := newTestServer()

	body := bytes.NewBufferString(`{"titles":["a","","c"]}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks/batch", body)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		Count   int               `json:"count"`
		Results []BatchItemResult `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Count != 3 {
		t.Errorf("expected count 3, got %d", resp.Count)
	}
	if resp.Results[1].Error == "" {
		t.Errorf("middle item (empty title) should have an error")
	}
	if resp.Results[0].Task == nil || resp.Results[2].Task == nil {
		t.Errorf("items 0 and 2 should have tasks")
	}
}

func TestHandler_GetNotFound(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_FullLifecycle(t *testing.T) {
	srv := newTestServer()

	// 1. 建立
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tasks",
		bytes.NewBufferString(`{"title":"learn Go"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", rec.Code)
	}

	// 2. 更新為完成
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/tasks/1",
		bytes.NewBufferString(`{"title":"learn Go","done":true}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("update failed: %d", rec.Code)
	}

	// 3. 刪除
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/tasks/1", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete failed: %d", rec.Code)
	}

	// 4. 確認已不存在
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tasks/1", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rec.Code)
	}
}
