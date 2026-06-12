# Claude Code 实施 Prompt（连同 novada-openapi.json 与 spec 一起放入项目）

把以下内容作为初始指令，并将 `novada-openapi.json` 和 `novada-go-sdk-spec.md` 放进仓库根目录。

---

我要开发一个 Novada 的 Go SDK，可通过 `go get github.com/novada/novada-go` 使用。
项目根目录有 `novada-openapi.json`（51 个真实接口）和 `novada-go-sdk-spec.md`（设计规范）。
请严格以这两份文件为准，不要凭记忆编造接口、字段或端点。

## 硬性约束（来自真实 spec，务必遵守）

1. **鉴权**：`Authorization: Bearer <API_KEY>`，由顶层 client 统一注入。
2. **三个 base URL，按接口路由**：
   - 通用 `https://api-m.novada.com`：所有 `/v1/*` 接口（含 Scraper 目录下的 `unblocker_area`/`capture/*`/`browser_area` 等查询接口），`multipart/form-data`。
   - Web Unblocker `https://webunlocker.novada.com`：`Scraper/Web Unblocker` 的 `POST /request`，`application/x-www-form-urlencoded`。
   - Scraper API `https://scraper.novada.com`：`Scraper/Scraper api/*` 的 `POST /request`，`application/x-www-form-urlencoded`。
   - **区分要点**：`/v1/*` 一律走通用 host；只有 `/request` 抓取调用才按目录分到 webunlocker / scraper。Client 需支持 3 个 baseURL，scraper 包不同方法路由到不同 host（通过 `Target` 字段或子服务区分）。
   - `/request` 表单字段：`scraper_name`/`scraper_id`/`scraper_params`(JSON 字符串)/`scraper_errors`。
3. **统一响应包裹**：管理接口返回 `{code,data,msg,timestamp}`，**`code==0` 才是成功**；`code!=0` 必须转成 `*APIError`，不能因 HTTP 200 就当成功。
4. 列表接口的 `data` 形如 `{list:[...], ...}`，做统一解包 helper。
5. 仅用标准库，不引入第三方 HTTP 框架。最低 Go 1.21。
6. 包结构：分子包 `proxy` / `scraper` / `wallet`，顶层 `novada` 提供 client/config/errors/transport/envelope。
7. 所有对外方法首参为 `context.Context`。

## 实施顺序（每步做完跑测试再继续）

1. 顶层骨架：`go.mod`、`Client`（**3 个 baseURL** + Bearer + 重试）、`config.go`（Functional Options：`WithBaseURL`/`WithWebUnblockerURL`/`WithScraperURL`，三者均有默认值）、`errors.go`（`APIError` 含业务 code + `IsAuthError/IsRateLimited/CodeOf`）、`envelope.go`、`transport.go`（`doMultipart` 与 `doFormURLEncoded` 两个 helper，统一过 envelope）。用 `httptest.Server` mock `white_list/list` 跑通解包。
2. `proxy.Whitelist` + `proxy.Account` 的 CRUD，配单测，确立 multipart 模式。
3. 读 OpenAPI 的 tag 分组，按子服务（Residential/StaticISP/RotatingISP/RotatingDC/Mobile/Unlimited/DedicatedDC/ProhibitDomain）铺开 proxy 其余接口。**根据每个接口 requestBody 的 properties 生成参数结构体，required 字段在发请求前校验。**
4. `scraper.Do(ctx, Request)` 通用驱动：`scraper_params` 用 `json.Marshal(params)` 后放入表单。单测验证编码正确（URL 编码 + JSON 字符串嵌套）。
5. `scraper.YouTube.VideoPost` 强类型封装作为样例（scraper_id=`youtube_video-post_explore`）。
6. `wallet.Balance` / `wallet.UsageRecord`。
7. `README.md`（go get、三个包各一例、错误处理、**3 个 baseURL** 说明）+ `examples/` + GitHub Actions（`go vet` + `golangci-lint` + `go test ./...`，Go 1.21/1.22/1.23 矩阵）。

## 验收标准

- `go build ./...` 与 `go vet ./...` 通过。
- 核心包（transport/envelope/errors/proxy.Whitelist/scraper.Do）单测覆盖率 > 80%，不依赖真实网络。
- 每个导出符号有 godoc 注释。
- README 的快速开始示例可直接编译。

先只做第 1 步，完成后停下来给我看顶层骨架和那个 mock 单测，我确认后再继续。
