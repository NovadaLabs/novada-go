# Novada Go SDK — 设计规范与目标调用示例（基于真实 OpenAPI）

> 已根据 `novada-openapi.json`（51 个接口）校准。本文件 + OpenAPI 文件一起作为 Claude Code 的输入。

---

## 0. 从 OpenAPI 得到的关键事实（设计基石）

| 事实 | 影响 |
|------|------|
| 鉴权：`Authorization: Bearer <API_KEY>` | 顶层 client 统一注入 |
| **两套传输机制完全不同** | proxy/wallet 与 scraper 必须分开实现 |
| Proxy/Wallet 管理接口（50 个）：全部 `POST`，`multipart/form-data` 请求 | 表单编码，非 JSON body |
| 管理接口统一响应包裹：`{code, data, msg, timestamp}`，**`code==0` 为成功** | 错误判断以 `code` 为准，不能只看 HTTP 200 |
| Scraper 抓取（`POST /request`）：`application/x-www-form-urlencoded` | 字段：`scraper_name` `scraper_id` `scraper_params`(JSON 字符串) `scraper_errors` |
| 所有 scraper 场景共用同一端点，靠 `scraper_id` 区分 | scraper 包做"通用驱动 + 可选场景封装" |
| 列表类接口 data 为 `{list:[...], ...}` | 统一 list 解包 helper |
| 无 servers 字段 | baseURL 需手动设定，见下 |

Base URL（共 **3 个**，由使用方提供并确认）：

| 用途 | Base URL | 适用接口 |
|------|----------|----------|
| 通用 | `https://api-m.novada.com` | 所有 `/v1/*` 接口（Proxy、Wallet、以及 Scraper 目录下查地区/余额/单价的 `/v1/*`，如 `unblocker_area`、`capture/get_balance`、`browser_area`）|
| Web Unblocker | `https://webunlocker.novada.com` | `Scraper/Web Unblocker` 目录下的 `POST /request` |
| Scraper API | `https://scraper.novada.com` | `Scraper/Scraper api/*` 目录下的 `POST /request`（YouTube/Amazon/Google/…）|

> ⚠️ 注意区分：`/v1/*` 一律走通用 host；只有真正的抓取调用 `/request` 才按目录分到 webunlocker / scraper 两个 host。
> client 需支持 **3 个 baseURL**，scraper 包内不同方法路由到不同 host。

---

## 1. 项目元信息

| 项 | 值 |
|----|----|
| 模块路径 | `github.com/novada/novada-go`（确认仓库）|
| 包结构 | 分子包：`proxy`、`scraper`、`wallet`；顶层 `novada` 提供 client/config/errors/transport |
| 最低 Go 版本 | 1.21 |
| License | MIT |
| 依赖 | 仅标准库（`net/http`、`net/url`、`mime/multipart`、`encoding/json`）|

---

## 2. 目录结构

```
novada-go/
├── go.mod
├── client.go         // 顶层 Client：双 baseURL、Bearer 注入、重试
├── config.go         // Functional Options
├── errors.go         // APIError（含业务 code）
├── transport.go      // doForm()（multipart）/ doFormURLEncoded()，统一 envelope 解包
├── envelope.go       // {code,data,msg,timestamp} 解包 + code!=0 转 APIError
├── version.go
├── proxy/            // 50 个管理接口分组实现
│   ├── proxy.go      // Service：account/whitelist/residential/static/mobile/...
│   ├── residential.go static.go mobile.go whitelist.go account.go ...
│   └── types.go
├── scraper/
│   ├── scraper.go    // 通用 Do(scraperID, params)；强类型场景封装在 sources.go
│   ├── sources.go    // YouTube 等已知 scraper_id 的强类型方法
│   └── types.go
├── wallet/
│   └── wallet.go     // balance / usage_record
├── internal/testutil/
└── examples/
```

---

## 3. 顶层 Client 与配置

```go
client, err := novada.NewClient("API_KEY",
    novada.WithBaseURL("https://api-m.novada.com"),              // 通用 /v1/*
    novada.WithWebUnblockerURL("https://webunlocker.novada.com"),// web unblocker /request
    novada.WithScraperURL("https://scraper.novada.com"),         // scraper api /request
    novada.WithTimeout(30*time.Second),
    novada.WithMaxRetries(2),
)
// client.Proxy / client.Scraper / client.Wallet
// 三个 URL 均有默认值，可不传；需要私有部署/测试时再覆盖
```

- Bearer 自动注入到所有请求 `Authorization` 头。
- 默认从 `NOVADA_API_KEY` 环境变量兜底。
- 重试仅针对网络错误、HTTP 429/5xx；不对业务 `code!=0` 重试。

---

## 4. 统一响应包裹与错误

```go
// envelope.go —— 所有 /v1 管理接口响应先过这里
type envelope struct {
    Code      int             `json:"code"`
    Msg       string          `json:"msg"`
    Data      json.RawMessage `json:"data"`
    Timestamp int64           `json:"timestamp"`
}

// errors.go
type APIError struct {
    HTTPStatus int    // HTTP 状态码
    Code       int    // 业务 code（0 以外即失败）
    Message    string // msg
}
func (e *APIError) Error() string { ... }

func IsAuthError(err error) bool      // HTTP 401/403
func IsRateLimited(err error) bool    // HTTP 429
func CodeOf(err error) (int, bool)    // 取业务 code
```

解包流程：HTTP 非 2xx → APIError(HTTPStatus)；2xx 但 `code!=0` → APIError(Code,Msg)；成功 → `json.Unmarshal(env.Data, &out)`。

---

## 5. Proxy 包 — 目标调用示例（基于真实接口）

```go
// 子账号
client.Proxy.Account.Create(ctx, proxy.CreateAccountParams{
    Product: proxy.ProductResidential, // 1=Residential 2=RotatingISP 3=RotatingDC 4=Unlimited 7=Unblocker 9=Mobile
    Account: "account11", Password: "pass11", Status: 1,
})
client.Proxy.Account.List(ctx, proxy.ListAccountParams{})

// 白名单（product: 1=Residential 5=StaticISP 4=Unlimited）
client.Proxy.Whitelist.Add(ctx, proxy.AddWhitelistParams{Product: 1, IP: "10.10.10.1", Remark: "test"})
client.Proxy.Whitelist.List(ctx, proxy.ListWhitelistParams{})
client.Proxy.Whitelist.Delete(ctx, ...)

// 住宅代理：地区/流量
client.Proxy.Residential.Countries(ctx)
client.Proxy.Residential.Cities(ctx, proxy.CityParams{...})
client.Proxy.Residential.Balance(ctx)                       // 剩余流量
client.Proxy.Residential.ConsumeLog(ctx, proxy.TimeRange{   // start_time/end_time "2006-01-02 15:04:05"
    Start: "2025-01-01 00:00:00", End: "2025-01-31 23:59:59"})

// 静态 ISP / 专用数据中心：open/list/export/region/renew/renew_setting
client.Proxy.StaticISP.Open(ctx, ...)
client.Proxy.DedicatedDC.List(ctx, ...)
```

> proxy 包按 OpenAPI 的 tag 分子服务：Account / Whitelist / Residential / StaticISP / RotatingISP /
> RotatingDC / Mobile / Unlimited / DedicatedDC / ProhibitDomain。每个方法对应一个 `/v1/*` 接口。

---

## 6. Scraper 包 — 目标调用示例

> **Host 路由**：`/request` 抓取调用分两个 host —— Scraper API（`scraper.novada.com`）与
> Web Unblocker（`webunlocker.novada.com`）。而该目录下的 `/v1/*` 查询接口（地区、余额、单价）
> 走通用 host。SDK 用不同子服务区分：`client.Scraper.API.*` vs `client.Scraper.Unblocker.*`。

通用驱动（覆盖所有 scraper_id），需指明走哪个 host：

```go
res, err := client.Scraper.Do(ctx, scraper.Request{
    Target:      scraper.TargetScraperAPI, // 或 scraper.TargetWebUnblocker
    ScraperName: "youtube.com",
    ScraperID:   "youtube_video-post_explore",
    Params: []map[string]any{
        {"url": "https://www.youtube.com/watch?v=HAwTwmzgNc4"},
    },
    ReturnErrors: true, // -> scraper_errors=true
})
// SDK 内部：scraper_params = json.Marshal(Params)，整体 urlencode；按 Target 选 host
fmt.Println(res.Raw)
```

强类型场景封装（已知 scraper_id，自动选 host）：

```go
res, err := client.Scraper.API.YouTube.VideoPost(ctx, scraper.YouTubeVideoParams{
    URL: "https://www.youtube.com/watch?v=HAwTwmzgNc4",
})
```

Universal / Browser 下的 `/v1/*` 查询接口（走通用 host）：

```go
client.Scraper.Universal.Balance(ctx)        // /v1/capture/get_balance
client.Scraper.Universal.Unit(ctx)           // /v1/capture/unit
client.Scraper.Unblocker.Countries(ctx)      // /v1/proxy/unblocker_area
client.Scraper.Browser.Countries(ctx)        // /v1/proxy/browser_area
```

> 当前 spec 仅含 YouTube 一个 scraper_id；其余平台（Amazon/Google/Instagram/LinkedIn/TikTok/Walmart）
> 待文档补全后按同样模式追加。通用 `Do` 保证未覆盖场景也能用。

---

## 7. Wallet 包

```go
client.Wallet.Balance(ctx)
client.Wallet.UsageRecord(ctx, wallet.TimeRange{...})
```

---

## 8. 测试 / CI / 文档

- 单测：`httptest.Server` 模拟 `{code,data,msg,timestamp}`，覆盖 code!=0 → APIError、multipart 与 urlencoded 编码、Bearer 注入、list 解包。
- 集成测试：`//go:build integration`，读 `NOVADA_API_KEY`。
- CI：GitHub Actions，`go vet` + `golangci-lint` + `go test ./...`，Go 1.21/1.22/1.23。
- README：go get 安装、proxy/scraper/wallet 各一例、错误处理、双 baseURL 说明。

---

## 9. Claude Code 实施顺序

1. 顶层：`Client`（双 baseURL + Bearer）、`config`、`errors`、`envelope`、`transport`（multipart + urlencoded 两个 helper）。先用 mock server 跑通一个管理接口（如 `white_list/list`）验证解包。
2. Proxy.Whitelist + Proxy.Account（最简单的 CRUD），配单测，确立 multipart 模式。
3. 按 tag 铺开 proxy 其余子服务（基于 OpenAPI 批量生成 types + 方法）。
4. Scraper.Do 通用驱动（urlencoded + scraper_params JSON 编码），单测验证编码正确。
5. Scraper.YouTube 强类型封装样例。
6. Wallet。
7. README + examples + CI。

> OpenAPI 最大价值在第 3、4 步：让 Claude Code 按 spec 的 requestBody properties 自动生成各接口
> 的参数结构体与必填校验，避免手抄 48 个表单的字段出错。
