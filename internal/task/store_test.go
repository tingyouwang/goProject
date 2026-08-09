package task

import (
	"errors"
	"testing"
)

// Go 的測試慣例:
//   - 檔名以 _test.go 結尾。
//   - 測試函式命名為 TestXxx(t *testing.T)。
//   - 用 `go test ./...` 執行,不需要 JUnit 之類的框架,測試是語言內建的。

func TestMemoryStore_CreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	created := s.Create("write resume")
	if created.ID != 1 {
		t.Fatalf("expected first ID to be 1, got %d", created.ID)
	}

	got, err := s.Get(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "write resume" {
		t.Errorf("expected title %q, got %q", "write resume", got.Title)
	}
	if got.Done {
		t.Errorf("new task should not be done")
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.Get(999)
	// errors.Is 檢查回傳的是不是我們定義的哨兵錯誤。
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestMemoryStore_Update 展示「table-driven test」——Go 最具代表性的測試風格。
// 把多組輸入/期望放進一個 slice,用迴圈跑,每組是一個獨立的子測試 t.Run(...)。
// 相當於 JUnit 的 @ParameterizedTest,但不需要任何註解或框架。
func TestMemoryStore_Update(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		done      bool
		wantTitle string
		wantDone  bool
	}{
		{name: "mark done", title: "task A", done: true, wantTitle: "task A", wantDone: true},
		{name: "rename only", title: "renamed", done: false, wantTitle: "renamed", wantDone: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMemoryStore()
			s.Create("original")

			updated, err := s.Update(1, tc.title, tc.done)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if updated.Title != tc.wantTitle {
				t.Errorf("title: got %q, want %q", updated.Title, tc.wantTitle)
			}
			if updated.Done != tc.wantDone {
				t.Errorf("done: got %v, want %v", updated.Done, tc.wantDone)
			}
		})
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := NewMemoryStore()
	s.Create("temp")

	if err := s.Delete(1); err != nil {
		t.Fatalf("unexpected error deleting: %v", err)
	}
	if _, err := s.Get(1); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected task to be gone, got %v", err)
	}
	// 再刪一次應回報找不到。
	if err := s.Delete(1); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}
