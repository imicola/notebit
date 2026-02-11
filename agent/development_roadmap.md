# Notebit 全局开发路线图 (AI Agent 专用版)

**版本**: 1.0  
**最后更新**: 2026-02-11  
**基于文档**: `docs/Need.md`, `docs/notebit-prd-zh_cn.md`

---

## 1. 项目概览

**Notebit** 是一款 **Local-First（本地优先）** 的 Markdown 笔记应用，旨在通过本地 AI 能力（LLM + Embeddings）成为用户的“静默策展人”。

*   **核心理念**: "Write for Humans, Manage by Silicon"（为人而写，由硅管理）。
*   **关键差异**: 编辑器纯净无干扰（无自动补全），AI 在后台静默工作（自动关联、标签、RAG）。
*   **隐私策略**: 数据（笔记 + 向量库）完全本地化。

---

## 2. 技术架构栈

| 层级 | 技术选型 | 备注 |
| :--- | :--- | :--- |
| **应用框架** | **Wails (Go)** | 提供系统级绑定，高性能后端逻辑，生成原生应用。 |
| **前端 UI** | **React + Tailwind CSS** | 使用 Vite 构建，追求现代化 UI 和极速渲染。 |
| **数据存储** | **SQLite** | 存储元数据、标签关系及 **Vector Embeddings**。 |
| **文件系统** | **Native File System** | 直接读写本地 `.md` 文件，兼容 Obsidian 等。 |
| **AI 引擎** | **Ollama (Local)** / OpenAI API | 本地优先，支持切换。用于 Embedding 和 Chat。 |
| **向量处理** | **Go (Native/Bindings)** | 使用 Go 处理向量计算和存储交互（如 `sqlite-vec` 或纯 Go 实现）。 |

---

## 3. 推荐目录结构

```text
notebit/
├── frontend/           # React 前端代码
│   ├── src/
│   │   ├── components/ # UI 组件 (Editor, Sidebar, Graph)
│   │   ├── hooks/      # React Hooks (useNotes, useAI)
│   │   └── ...
│   └── ...
├── app.go              # Wails 应用主逻辑
├── main.go             # 入口文件
├── pkg/                # 后端核心逻辑包
│   ├── ai/             # AI 交互层 (Ollama/OpenAI 客户端, Embedding 生成)
│   ├── db/             # SQLite 数据库操作 (Gorm 或 sqlx)
│   ├── files/          # 文件系统操作 (Watcher, CRUD)
│   └── search/         # 向量搜索与相似度计算算法
├── data/               # (运行时生成) 存放 db.sqlite
└── go.mod
```

---

## 4. 分阶段开发路线图 (Phased Roadmap)

### 🔴 阶段一：骨架搭建与纯净编辑器 (MVP / The Sanctuary)
**目标**: 完成一个可用的、纯本地的 Markdown 编辑器，无 AI 功能。

*   **任务 1.1: 项目初始化**
    *   初始化 Wails + React + Tailwind CSS 项目结构。
    *   配置 ESLint/Prettier 及 Go Linter。
*   **任务 1.2: 文件系统交互 (Go Backend)**
    *   实现 `OpenFolder`：选择并读取本地文件夹。
    *   实现 `ReadFile` / `SaveFile`：读写 `.md` 内容。
    *   实现 `ListFiles`：递归获取文件树结构。
*   **任务 1.3: 前端编辑器实现**
    *   集成 Markdown 编辑器组件 (推荐 `EasyMDE`, `Milkdown` 或 `CodeMirror`)。
    *   实现左侧文件树侧边栏 (File Explorer)。
    *   实现基础的所见即所得 (WYSIWYG) 与源码模式切换。
    *   **约束**: 确保无 AI 自动补全干扰。

### 🟡 阶段二：核心智能基础设施 (The Brain)
**目标**: 建立后台数据处理流水线，实现“保存即索引”。

*   **任务 2.1: SQLite 数据库集成**
    *   设计数据库 Schema (`files` 表: path, last_modified; `chunks` 表: content, embedding)。
    *   集成 SQLite 驱动及向量扩展 (如 `sqlite-vec` 或利用 Go 库做纯内存向量计算，视数据量定，建议优先 SQLite 方案)。
*   **任务 2.2: AI 服务层 (Go)**
    *   实现 OpenAI 接口(支持 Embedding 接口)
    *   封装 Ollama API 客户端 (支持 Embedding 接口)。
    *   实现 Text Chunking (文本分块) 策略 (按标题或固定窗口)。
*   **任务 2.3: 实时索引流水线**
    *   集成 `fsnotify` 监听文件变更。
    *   实现 `Watcher` 逻辑：文件保存 -> 触发 Diff -> 调用 Embedding API -> 更新 SQLite。
    *   实现全量索引 (首次启动时扫描)。

### 🟢 阶段三：交互层与静默策展 (The Curator)
**目标**: 在 UI 上展示 AI 的分析结果，实现“语义关联”。

*   **任务 3.1: 语义侧边栏 (UI + Logic)**
    *   后端：实现 `FindSimilar(content string)` 接口，基于余弦相似度查询 Top 5。
    *   前端：开发右侧“相关笔记”面板。
    *   逻辑：编辑器内容变动/保存后，自动刷新右侧推荐。
*   **任务 3.2: 优雅降级处理**
    *   处理 ai 离线/未安装的情况。
    *   UI 显示状态指示器 (如：AI 离线图标)，确保不阻塞主编辑器功能。

### 🔵 阶段四：高级回顾与知识图谱 (Review & Graph)
**目标**: 深度利用数据进行对话和可视化。

*   **任务 4.1: RAG 对话系统 (Chat)**
    *   开发独立的 Chat 界面。
    *   后端实现 RAG 流程：Query -> Vector Search -> Context Assembly -> LLM Prompt -> Response。
    *   实现引用标注功能 (解析 `[[Link]]`)。
*   **任务 4.2: 知识图谱 (Visualization)**
    *   前端集成力导向图库 (如 `react-force-graph`)。
    *   后端聚合数据：显式链接 (Wiki Links) + 隐式链接 (语义相似度 > 阈值)。
    *   实现图谱节点的交互跳转。

---

## 5. 开发规范与注意事项

1.  **错误处理**: Go 后端错误需通过 Wails Events 或 Error 返回传递给前端，前端需做 Toast 提示，严禁弹窗阻断。
2.  **性能**:
    *   向量计算和 Embedding 请求必须在 Go 的 `goroutine` 中异步执行，**绝不能阻塞主线程**。
    *   搜索响应时间目标 < 200ms。
3.  **兼容性**:
    *   所有元数据存储在 `.sqlite` 中，不修改用户 Markdown 文件的原始内容（除非用户明确操作）。
    *   文件路径处理需考虑 Windows/Mac/Linux 差异。
4.  **测试**:
    *   关键算法 (Chunking, Similarity) 需编写 Go Unit Tests。
    *   文件监听逻辑需进行手动或集成测试验证。