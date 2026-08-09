package task

import (
	"sort"
	"sync"
	"time"
)

// defaultNow 是預設的時間來源;測試裡可把 nowFunc 換掉以固定時間。
func defaultNow() time.Time { return time.Now() }

// MemoryStore 是 Repository 介面的「記憶體實作」。
//
// 注意:我們「沒有」寫 `implements Repository`。只要 MemoryStore 具備
// Repository 要求的全部方法,它就自動滿足該介面(見 task.go 的說明)。
//
// 為什麼要有 mutex?因為 HTTP 伺服器會為「每個請求開一個 goroutine」並行處理,
// 多個 goroutine 可能同時讀寫 map,map 在 Go 裡「不是併發安全的」,會 panic 或資料錯亂。
// sync.Mutex 就像 Java 的 synchronized / ReentrantLock,用來保護共享狀態。
type MemoryStore struct {
	mu     sync.Mutex     // 保護底下的 tasks 與 nextID
	tasks  map[int]Task   // 用 map 當簡易資料表,key 是 ID(類似 HashMap<Integer, Task>)
	nextID int            // 自增主鍵
}

// NewMemoryStore 是慣用的「建構子」。
//
// Go 沒有 constructor 關鍵字,慣例是寫一個 NewXxx 函式回傳初始化好的值。
// map 必須先用 make 建立才能用(零值的 map 是 nil,寫入會 panic)——
// 這是 Java 工程師常踩的坑之一。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

// Create 新增一筆 task 並回傳。
func (s *MemoryStore) Create(title string) Task {
	// defer 會在函式「return 之前」執行,常用來成對地做「解鎖/關檔/關連線」。
	// 對照 Java 的 try-with-resources 或 finally,但更簡潔、就寫在 Lock 旁邊,不易漏。
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Task{
		ID:        s.nextID,
		Title:     title,
		Done:      false,
		CreatedAt: nowFunc(),
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t
}

// List 回傳所有 task,依 ID 排序(map 走訪順序在 Go 裡是隨機的,要排序才穩定)。
func (s *MemoryStore) List() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	// slice 類似 Java 的 ArrayList。make([]Task, 0, n) 預先配置容量,是小優化。
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks { // for range:底線 _ 表示「忽略 key,只要 value」
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Get 依 ID 取一筆。
//
// 這裡展示 Go 的「多重回傳值」:回傳 (Task, error)。
// Java 會回傳 Task 或 throw;Go 則把「結果」和「錯誤」一起回傳,
// 呼叫端必須顯式檢查 error——這是 Go 錯誤處理的核心風格。
func (s *MemoryStore) Get(id int) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 「comma ok」慣用法:從 map 取值時第二個回傳值 ok 表示 key 是否存在。
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound // 回傳「零值 Task」+ 我們定義的哨兵錯誤
	}
	return t, nil // nil 表示沒有錯誤(相當於 Java 的「沒 throw」)
}

// Update 更新既有 task 的標題與完成狀態。
func (s *MemoryStore) Update(id int, title string, done bool) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	t.Title = title
	t.Done = done
	s.tasks[id] = t
	return t, nil
}

// Delete 刪除一筆 task。
func (s *MemoryStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id) // delete 是內建函式,從 map 移除 key
	return nil
}

// nowFunc 讓時間可被測試替換(在測試裡可以固定時間)。
// 這是把「不確定性」抽出來的常見手法,對照 Java 的 Clock injection。
var nowFunc = defaultNow
