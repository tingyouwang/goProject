package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler 負責把 HTTP 請求轉成對 Repository 的呼叫,再把結果寫回 HTTP 回應。
//
// 關鍵:它只依賴「Repository 介面」,不依賴 MemoryStore 這個具體型別。
// 這就是依賴反轉——之後換成 PostgresStore,這裡完全不用改。
// (對照 Java 的 @Autowired Repository interface 但不需要框架/DI 容器。)
type Handler struct {
	repo Repository
}

// NewHandler 建構子:把依賴(repo)注入進來。
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// Routes 回傳掛好路由的 http.Handler。
//
// http.ServeMux 是標準庫的路由器。Go 1.22 起支援 "METHOD /path/{id}" 這種
// 帶方法與路徑參數的樣式,不再需要第三方框架就能做基本 REST。
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", h.create)
	mux.HandleFunc("POST /tasks/batch", h.batchCreate)
	mux.HandleFunc("GET /tasks", h.list)
	mux.HandleFunc("GET /tasks/{id}", h.get)
	mux.HandleFunc("PUT /tasks/{id}", h.update)
	mux.HandleFunc("DELETE /tasks/{id}", h.delete)
	return mux
}

// taskInput 是「請求本體」的 DTO(對照 Java 的 @RequestBody DTO)。
type taskInput struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in taskInput
	// json.NewDecoder(...).Decode(...) 把請求 body 解析進 struct。
	// 這一行同時示範了 Go 最常見的錯誤處理節奏:做一件事、馬上檢查 err。
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	t := h.repo.Create(in.Title)
	writeJSON(w, http.StatusCreated, t)
}

// batchInput 是批次建立的請求本體:一組標題。
type batchInput struct {
	Titles []string `json:"titles"`
}

// defaultBatchWorkers 是批次處理的併發度(同時最多幾個 worker)。
const defaultBatchWorkers = 4

// batchCreate 併發建立一批 task,回報每一筆的成敗。
//
// 這個端點是整個專案的併發亮點:它把請求交給 ProcessBatch,
// 由一個有界 worker pool 併發處理,並用 r.Context() 傳遞逾時/取消。
// r.Context() 會在「客戶端斷線」或「伺服器 WriteTimeout 到」時自動取消——
// 免費得到「客戶端都走了就別再白做工」的行為。
func (h *Handler) batchCreate(w http.ResponseWriter, r *http.Request) {
	var in batchInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(in.Titles) == 0 {
		writeError(w, http.StatusBadRequest, "titles must not be empty")
		return
	}

	results := ProcessBatch(r.Context(), h.repo, in.Titles, defaultBatchWorkers)
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(results),
		"results": results,
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.repo.List())
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	t, err := h.repo.Get(id)
	if err != nil {
		// errors.Is 用來比對哨兵錯誤——這比 Java 的 instanceof 更明確、可包裝。
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in taskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	t, err := h.repo.Update(id, in.Title, in.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204,無內容
}

// parseID 從路徑參數 {id} 取出整數 ID。
func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

// writeJSON 是共用小工具:設好 header、狀態碼,再把資料編碼成 JSON 寫出。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 統一的錯誤回應格式。
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
