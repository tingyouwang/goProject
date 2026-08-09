// Package task 是這個服務的「領域層」(domain layer)。
//
// Go 的 package 概念類似 Java 的 package,但有兩個關鍵差異:
//  1. 一個資料夾 = 一個 package,資料夾裡所有 .go 檔共享同一個 package。
//  2. 「可見性」不是靠 public/private 關鍵字,而是靠「識別字的第一個字母大小寫」:
//       - 大寫開頭 (Task, Store, Create) = 匯出的 (public),別的 package 可以用。
//       - 小寫開頭 (tasks, mu, nextID)   = 未匯出的 (private),只有本 package 能用。
package task

import (
	"errors"
	"time"
)

// ErrNotFound 是一個「哨兵錯誤」(sentinel error)。
//
// Java 會 throw 一個 NotFoundException;Go 沒有例外機制,錯誤是「值」。
// 呼叫端用 errors.Is(err, task.ErrNotFound) 來判斷是不是這個特定錯誤。
// 這就是 Go 的哲學:錯誤是普通的回傳值,不是控制流。
var ErrNotFound = errors.New("task not found")

// Task 是我們的領域模型(相當於 Java 的一個 POJO / entity)。
//
// 反引號裡的 `json:"..."` 叫做「struct tag」,類似 Java 的 @JsonProperty 註解,
// 告訴 JSON 編解碼器欄位對應的名字。encoding/json 標準庫會自動讀取它。
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository 是一個「介面」(interface)。
//
// 這是 Go 最重要的觀念之一,跟 Java 差很多:
//   - Java: class 要寫 `implements Repository` 明確宣告。
//   - Go:   只要某個型別「剛好」有這些方法,它就「自動」滿足這個介面,
//           不需要任何宣告。這叫「隱式實作 / 結構化型別 (duck typing)」。
//
// 好處:介面可以「事後」定義。我們在 handler 只依賴這個介面,
// 之後把記憶體版換成資料庫版時,handler 一行都不用改——這就是依賴反轉。
type Repository interface {
	Create(title string) Task
	List() []Task
	Get(id int) (Task, error)
	Update(id int, title string, done bool) (Task, error)
	Delete(id int) error
}
