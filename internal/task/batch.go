package task

import (
	"context"
	"strings"
	"sync"
)

// BatchItemResult 是批次中「單一項目」的處理結果。
//
// 設計重點:批次建立時我們要「每一筆都試,並回報每一筆的成敗」,
// 而不是一有錯就整批中止。所以每個項目各自帶自己的 Task 或 error。
// (omitempty:欄位為零值時 JSON 就不輸出,讓回應更乾淨。)
type BatchItemResult struct {
	Index int    `json:"index"`           // 對應請求中第幾筆(順序保證)
	Task  *Task  `json:"task,omitempty"`  // 成功時的結果
	Error string `json:"error,omitempty"` // 失敗時的原因
}

// ProcessBatch 用「有界 worker pool」併發處理一批標題。
//
// 為什麼這樣設計(面試講點):
//   - 「有界」:用固定數量的 worker,避免一次湧入上萬筆時開出上萬個 goroutine
//     壓垮下游(資料庫連線、外部 API)。這是正式環境的關鍵考量。
//   - goroutine + channel:jobs channel 是「工作佇列」,worker 從裡面拉工作;
//     這就是 Go 名言 "share memory by communicating" 的實踐——
//     用 channel 傳遞工作,而不是共享一個佇列再加鎖。
//   - 免鎖收集結果:results 預先配置好長度,每個 worker 只寫「自己那一格」
//     (不同 index),不會有兩個 goroutine 寫同一格,因此不需要 Mutex,
//     同時天然保證輸出順序與輸入一致。這個小技巧在面試很加分。
//   - context:支援逾時/取消。上游(HTTP handler)設定的 deadline 到了,
//     還沒開始做的項目會直接標記為取消,不會白做工。
func ProcessBatch(ctx context.Context, repo Repository, titles []string, workers int) []BatchItemResult {
	results := make([]BatchItemResult, len(titles))
	if len(titles) == 0 {
		return results
	}

	// worker 數量做個保護:至少 1,且不超過工作量。
	if workers < 1 {
		workers = 1
	}
	if workers > len(titles) {
		workers = len(titles)
	}

	// job 是要處理的單一工作:第幾筆 + 標題內容。
	type job struct {
		index int
		title string
	}
	jobs := make(chan job) // 無緩衝 channel:送出的工作會被某個 worker 立即接走

	var wg sync.WaitGroup // ≈ Java 的 CountDownLatch,用來等所有 worker 收工

	// 啟動固定數量的 worker goroutine。
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done() // 這個 worker 結束時,計數減一
			for j := range jobs {
				results[j.index] = processOne(ctx, repo, j.index, j.title)
			}
		}()
	}

	// 派工:把所有工作送進 channel。
	// 若 context 已取消,剩下的工作就不派了,直接標記為取消。
	dispatch:
	for i, title := range titles {
		select {
		case <-ctx.Done():
			// context 逾時/取消:把「還沒派出去」的項目全部標記為取消後收工。
			for k := i; k < len(titles); k++ {
				results[k] = BatchItemResult{Index: k, Error: ctx.Err().Error()}
			}
			break dispatch
		case jobs <- job{index: i, title: title}:
			// 成功把工作交給某個 worker。
		}
	}
	close(jobs) // 關閉 channel:通知所有 worker「沒有更多工作了」,range 迴圈會結束
	wg.Wait()   // 等所有 worker 真正做完

	return results
}

// processOne 處理單一項目(此處為驗證 + 建立)。
//
// 抽成獨立函式讓 worker 邏輯清爽,也方便單獨測試。
// 真實情境下這裡可能是「呼叫外部 API 補資料」「寫入資料庫」等 I/O 工作——
// 那正是併發能大幅縮短總耗時的地方(I/O 等待時 CPU 可切去做別筆)。
func processOne(ctx context.Context, repo Repository, index int, title string) BatchItemResult {
	// 每筆開工前再檢查一次 context,逾時就不做。
	if err := ctx.Err(); err != nil {
		return BatchItemResult{Index: index, Error: err.Error()}
	}
	if strings.TrimSpace(title) == "" {
		return BatchItemResult{Index: index, Error: "title is required"}
	}
	t := repo.Create(title)
	return BatchItemResult{Index: index, Task: &t}
}
