# Architecture Layer Analysis Report (Round 2)

## Date: 2026-02-15
## Status: Analysis Complete ✅

---

## 1. 错误处理分层违规

### 问题：三种错误处理模式并存
| 模式 | 使用位置 | 
|------|----------|
| 自定义 Error 类型（定义但未用） | `pkg/ai/errors.go`, `pkg/database/errors.go` |
| `fmt.Errorf()` | 全局主流 |
| GORM 原始错误 | `repository.go` |

---

## 2. 返回类型分层违规

### `map[string]interface{}` 在公共 API 中泛滥
涉及: app_ai.go, app_search.go, app_files.go, app_chat.go, knowledge/service.go

---

## 3. Context 传播断裂
- `ai.Initialize()` 使用 `context.TODO()`
- `indexing.processJob()` 使用 `context.Background()`
- `knowledge.*` 大多方法不接受 context
- `chat.*` 绝大多数方法不接受 context

---

## 4. 三重维度映射表重复
- `pkg/ai/openai.go:GetModelDimension`
- `pkg/ai/ollama.go:getKnownModelDimension`
- `pkg/config/config.go:ModelDimensions`

---

## 5. 服务初始化重复
`OpenFolder()` / `SetFolder()` / `startup()` 三处相似的初始化序列

---

## Architecture Layer 任务清单

| ID | 任务 | 严重度 |
|----|------|--------|
| A-201 | 统一模型维度映射到单一来源 | P2 |
| A-202 | 提取 initializeServices 消除初始化重复 | P2 |
