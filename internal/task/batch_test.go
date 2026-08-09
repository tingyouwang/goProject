package task

import (
	"context"
	"testing"
)

func TestProcessBatch_AllSucceed(t *testing.T) {
	repo := NewMemoryStore()
	titles := []string{"a", "b", "c", "d", "e"}

	results := ProcessBatch(context.Background(), repo, titles, 3)

	if len(results) != len(titles) {
		t.Fatalf("expected %d results, got %d", len(titles), len(results))
	}
	// 結果順序必須與輸入一致(worker pool 免鎖寫入各自 index 的保證)。
	for i, r := range results {
		if r.Error != "" {
			t.Errorf("item %d unexpected error: %s", i, r.Error)
		}
		if r.Task == nil {
			t.Fatalf("item %d expected a task, got nil", i)
		}
		if r.Task.Title != titles[i] {
			t.Errorf("item %d: got title %q, want %q", i, r.Task.Title, titles[i])
		}
	}
	// 全部應已寫進 store。
	if got := len(repo.List()); got != len(titles) {
		t.Errorf("expected %d tasks in store, got %d", len(titles), got)
	}
}

// TestProcessBatch_PartialFailure:批次不因單筆失敗而中止,錯誤項各自回報。
func TestProcessBatch_PartialFailure(t *testing.T) {
	repo := NewMemoryStore()
	titles := []string{"valid", "", "  ", "also valid"}

	results := ProcessBatch(context.Background(), repo, titles, 2)

	if results[0].Error != "" || results[0].Task == nil {
		t.Errorf("item 0 should succeed")
	}
	if results[1].Error == "" {
		t.Errorf("item 1 (empty title) should fail")
	}
	if results[2].Error == "" {
		t.Errorf("item 2 (whitespace title) should fail")
	}
	if results[3].Error != "" || results[3].Task == nil {
		t.Errorf("item 3 should succeed")
	}
	// 只有兩筆有效標題應被寫入。
	if got := len(repo.List()); got != 2 {
		t.Errorf("expected 2 tasks created, got %d", got)
	}
}

// TestProcessBatch_ContextCancelled:context 已取消時,未開工的項目標記為取消。
func TestProcessBatch_ContextCancelled(t *testing.T) {
	repo := NewMemoryStore()
	titles := []string{"a", "b", "c"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立刻取消

	results := ProcessBatch(ctx, repo, titles, 2)

	if len(results) != len(titles) {
		t.Fatalf("expected %d results, got %d", len(titles), len(results))
	}
	// 已取消的 context 下,所有項目都應帶錯誤且沒有 task。
	for i, r := range results {
		if r.Error == "" {
			t.Errorf("item %d expected cancellation error, got none", i)
		}
	}
}

func TestProcessBatch_Empty(t *testing.T) {
	repo := NewMemoryStore()
	results := ProcessBatch(context.Background(), repo, nil, 4)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(results))
	}
}
