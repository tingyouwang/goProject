# Go 學習計劃(Java 工程師視角)

> 這份是我從 Java 轉學 Go 的自學計劃與進度追蹤。搭配本專案的實作
> ([README.md](README.md))一起服用:每學一個觀念,就回到對應的程式碼看它怎麼落地。

## 心態:Go 是「刻意反 Java」的極簡設計

Go 拿掉了很多 Java 的東西,理解「拿掉什麼、為什麼」就抓到精髓了:

| Java | Go | 觀念轉換 |
|---|---|---|
| `try/catch` 例外 | 回傳 `error` 值,顯式檢查 | 錯誤是「值」,不是控制流 |
| `class` 繼承 | `struct` + 組合 (embedding) | **沒有繼承**,只有組合 |
| `implements` 明確宣告 | interface **隱式滿足** | 方法對得上就算實作 |
| Thread / Executor | `goroutine` + `channel` | 併發是語言內建,極輕量 |
| `null` | 零值 (zero value) | 宣告即有預設值,少很多 NPE |
| Maven / Gradle | `go mod` | 內建、極簡 |
| getter/setter、`@Annotation` | 大小寫控制可見性 | 大寫=public,小寫=private |
| Checked Exception | `if err != nil { return err }` | 會寫到手軟,但這就是 Go |

**一句話:Java 追求豐富的抽象,Go 追求極簡與明確。**

---

## 階段 1:語法速通(1–2 天)

目標:能讀懂、能寫基本 Go。有 Java 底子這階段會很快。

- [ ] 做完官方互動教學 **[A Tour of Go](https://go.dev/tour/)**(務必全做)
- [ ] 搞懂 `slice` vs array(`slice` ≈ `ArrayList`,但底層是 view)
- [ ] 搞懂 `map`(≈ `HashMap`;走訪順序隨機、nil map 不能寫)
- [ ] `struct`、指標(有指標但無指標運算)
- [ ] **多重回傳值**(Java 沒有,Go 的核心慣用法)

📌 對照本專案:[internal/task/task.go](internal/task/task.go) 的 `Task` struct 與 tag。

---

## 階段 2:Go 的靈魂(2–3 天)

面試考點,花時間吃透。

- [ ] **error handling**:`if err != nil`、`errors.Is` / `errors.As`、`fmt.Errorf("...: %w", err)` 包裝
- [ ] **interface 隱式實作**:寫一個 interface,用不同 struct 實作,體會「鴨子型別」
- [ ] **struct embedding**:用組合取代繼承
- [ ] **defer / panic / recover**:`defer` 用於關資源(≈ try-with-resources)
- [ ] 讀 **[Effective Go](https://go.dev/doc/effective_go)**(官方最佳實踐,面試常問)

📌 對照本專案:
- 哨兵錯誤 + `errors.Is` → [task.go](internal/task/task.go)、[handler.go](internal/task/handler.go)
- `Repository` 介面隱式實作 + 依賴反轉 → [store.go](internal/task/store.go)、[handler.go](internal/task/handler.go)
- `defer` 成對 Lock/Unlock → [store.go](internal/task/store.go)

---

## 階段 3:併發(2–3 天)⭐ 面試最愛

Go 的招牌。Java 併發知識能遷移,但寫法完全不同。

- [ ] `goroutine`:`go func()` 就開一條,比 Thread 輕量千百倍
- [ ] `channel`:goroutine 間用 channel 溝通
      → 名言 **"Don't communicate by sharing memory; share memory by communicating"**
- [ ] `sync.WaitGroup`(≈ CountDownLatch)、`sync.Mutex`(≈ synchronized)
- [ ] `select`:多路 channel 選擇
- [ ] `context.Context`:取消與逾時控制(實務必備)
- [ ] 用 `go test -race` 檢測資料競爭

📌 對照本專案(這是主打亮點):
- 有界 worker pool(goroutine + channel + WaitGroup + context)→ [batch.go](internal/task/batch.go)
- goroutine 啟動伺服器 + channel/context 優雅關閉 → [main.go](main.go)
- `sync.Mutex` 保護共享 map → [store.go](internal/task/store.go)

---

## 階段 4:工程實務 + 產出作品(3–5 天)⭐

- [ ] `go mod init` / `go mod tidy` 管理依賴
- [ ] `go test`、**table-driven tests**(Go 慣用測試風格)
- [ ] `go vet`、`gofmt`(格式化是強制慣例)
- [ ] 標準庫:`net/http`、`encoding/json`
- [ ] 完成一個可跑、有測試、有 README 的專案 → **本專案已達成 ✅**

📌 對照本專案:
- table-driven test → [store_test.go](internal/task/store_test.go)
- `httptest`(≈ MockMvc)→ [handler_test.go](internal/task/handler_test.go)

---

## 階段 5(選配,有時間再做)

- [ ] 把 `MemoryStore` 換成真實資料庫實作(驗證介面抽象:換 DB 時 handler 零改動)
- [ ] 加 middleware(logging / request ID)
- [ ] 熱門框架:`Gin`(web)、`GORM`(ORM,≈ JPA/Hibernate)、`chi`(路由)
- [ ] 讀一個知名開源專案的部分原始碼

---

## 壓縮版時程(一週內要面試)

| 天 | 內容 |
|---|---|
| Day 1–2 | 階段 1:Tour of Go + 語法 |
| Day 3–4 | 階段 2–3:error / interface / 併發 |
| Day 5–7 | 階段 4:完善專案 + 寫 README + 練面試話術 |

---

## 常用參考資源

- [A Tour of Go](https://go.dev/tour/) — 官方互動教學
- [Effective Go](https://go.dev/doc/effective_go) — 官方最佳實踐
- [Go by Example](https://gobyexample.com/) — 用範例查語法
- [Standard library docs](https://pkg.go.dev/std) — 標準庫文件
- [Go Proverbs](https://go-proverbs.github.io/) — Go 設計哲學金句(面試可引用)
