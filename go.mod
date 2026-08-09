// go.mod 相當於 Java 的 pom.xml / build.gradle,但極簡。
// 「module」定義了這個專案的匯入路徑(import path)。
// 「go」那行是這個專案要求的最低 Go 版本。
// 這個專案只用「標準庫」,所以下面沒有任何第三方依賴——
// 這本身就是 Go 的特色:標準庫非常夠用,net/http 就能寫出完整的 web 服務。
module taskapi

go 1.22
