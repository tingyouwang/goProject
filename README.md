# Task API — 一個用 Go 標準庫寫的 REST 服務

> 這是我(Java 背景)自學 Go 的實作作品:一個待辦清單的 REST API。
> 全程只用 **Go 標準庫**,不依賴任何第三方框架,用來熟悉語言核心與工程慣例。

## 為什麼做這個

我原本主要用 Java。為了展示「快速自學新語言」的能力,我選了一個熟悉的題目
(CRUD REST API),把注意力放在**語言差異**而不是業務邏輯上,這樣最能逼自己
真正理解 Go 的設計哲學。原始碼裡的註解特地標了與 Java 的對照。

## 功能

一個 Task 資源的完整 CRUD:

| Method | Path          | 說明             |
|--------|---------------|------------------|
| POST   | `/tasks`      | 建立 task        |
| GET    | `/tasks`      | 列出全部          |
| GET    | `/tasks/{id}` | 取單筆            |
| PUT    | `/tasks/{id}` | 更新              |
| DELETE | `/tasks/{id}` | 刪除              |

## 專案結構

```
goProject/
├── go.mod                     # 模組定義(≈ pom.xml,但極簡、無第三方依賴)
├── main.go                    # 進入點:組裝依賴、啟動伺服器、優雅關閉
└── internal/task/             # 領域 + 應用層(internal 表示不對外匯出)
    ├── task.go                # 模型 Task + Repository 介面 + 哨兵錯誤
    ├── store.go               # 記憶體實作(Mutex 保護的併發安全 store)
    ├── handler.go             # HTTP handler:路由、JSON、錯誤處理
    ├── store_test.go          # 單元測試(含 table-driven test)
    └── handler_test.go        # HTTP 測試(httptest,≈ MockMvc)
```

## 如何執行

```bash
# 啟動伺服器(監聽 :8080)
go run .
```

另開一個終端機試打:

```bash
# 建立
curl -X POST localhost:8080/tasks -d '{"title":"prepare interview"}'

# 列出
curl localhost:8080/tasks

# 取單筆
curl localhost:8080/tasks/1

# 更新為完成
curl -X PUT localhost:8080/tasks/1 -d '{"title":"prepare interview","done":true}'

# 刪除
curl -X DELETE localhost:8080/tasks/1
```

## 如何測試

```bash
# 跑全部測試
go test ./...

# 顯示每個測試 + 覆蓋率
go test ./... -v -cover
```

## 我在這個專案裡刻意練到的 Go 特性(面試講點)

| Go 特性 | 在哪 | 對照 Java |
|---|---|---|
| **顯式 error 回傳** | 到處 `if err != nil` | 取代 try/catch;錯誤是「值」不是控制流 |
| **哨兵錯誤 + `errors.Is`** | `task.go` / `handler.go` | 取代自訂 Exception + instanceof |
| **介面隱式實作** | `Repository` 介面 | 不用 `implements`,方法對上就算實作 |
| **依賴反轉** | `Handler` 只依賴介面 | 換 DB 不用改 handler,但無需 DI 框架 |
| **goroutine 併發** | `main.go` 啟動伺服器 | 比 `new Thread()` 輕量千百倍 |
| **channel + signal** | `main.go` 優雅關閉 | 用 channel 等中斷訊號 |
| **`context` 逾時控制** | `main.go` Shutdown | 給收尾時間上限 |
| **`sync.Mutex`** | `store.go` | ≈ `synchronized`,保護共享 map |
| **`defer`** | 成對 Lock/Unlock | ≈ try-with-resources/finally |
| **多重回傳值** | `Get` 回 `(Task, error)` | Java 做不到,Go 的核心慣用法 |
| **零值 / comma-ok** | map 取值、struct 初始化 | 減少 NPE |
| **table-driven test** | `store_test.go` | ≈ @ParameterizedTest,但內建無框架 |
| **`httptest`** | `handler_test.go` | ≈ MockMvc,內建 |

## Java → Go 我踩過的坑 / 學到的差異

- **nil map 不能寫入**:map 必須先 `make` 才能用,否則 panic。Java 的 `new HashMap<>()` 是理所當然,Go 要記得。
- **map 走訪順序是隨機的**:要穩定輸出得自己排序(`List()` 有處理)。
- **沒有繼承**:一開始想找 `extends`,後來理解 Go 用「組合 / interface」達成多型。
- **可見性靠大小寫**:大寫 = public、小寫 = private,一開始很不習慣。
- **併發不安全的 map**:多 goroutine 同時寫 map 會直接 crash,所以要 Mutex。

## 接下來想深入的方向

- 把 `MemoryStore` 換成真正的資料庫實作(驗證介面抽象是否真的零改動於 handler)
- 加上 middleware(logging / request ID)
- 用 `errgroup` 或 worker pool 做批次併發處理
- 導入 `Gin` 或 `chi` 比較和標準庫的差異
