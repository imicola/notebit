# Notebit

**Notebit** 是一款面向 PKM（个人知识管理）爱好者和研究者的本地优先 Markdown 笔记应用。它将纯净专注的写作环境与后台 AI 驱动的知识管理进行结合。

基于 **Wails** 框架，使用 **Go** 和 **React** 构建，Notebit 提供了原生桌面体验。

## 设计理念

Notebit 秉持这样的理念：写作应该是一种纯粹、专注的活动，而组织和连接则可以由 AI 增强。您只需专注于写作，Notebit 会自动整理、连接，并在您需要时呈现相关信息。

## 功能特性

- **编辑器**：
  - 由 **CodeMirror 6** 驱动的无干扰 Markdown 编辑器。
  - 语法高亮、代码检查和自动补全。
  
- **AI 智能检索**：
  - **RAG（检索增强生成）**：通过上下文感知的 AI 与您的笔记对话。
  - **多提供商支持**：可在 **OpenAI API** 和 **Ollama**之间无缝切换。
  - **智能分块**：支持固定大小、基于标题、句子感知等多种策略，保留完整语义上下文。
  - **向量搜索**：原生集成 `sqlite-vec`，实现高性能相似度搜索。

- **知识图谱**： 
  - 交互式力导向图可视化（`react-force-graph-2d`），探索笔记间的连接。
  - 创建链接时实时更新。

- **企业级日志**：
  - 支持文件轮转和 Kafka 的异步批量日志记录。
  - 结构化、上下文感知的日志，深度可观测性。

- **文件系统**：
  - 基于 `fsnotify` 的健壮文件管理，监听外部文件变更。
  - 快速递归文件索引。

##  技术栈

**后端**
- **Go 1.24**：核心逻辑和系统交互。
- **Wails v2.11**：桌面应用程序框架。
- **SQLite + sqlite-vec**：具有原生向量搜索能力的本地数据库。
- **GORM**：数据库 ORM。
- **Kafka-go**：异步事件处理。

**前端**
- **React 18**：UI 框架。
- **Vite 3**：快速构建工具和开发服务器。
- **Tailwind CSS**：实用优先的 CSS 框架。
- **CodeMirror 6**：可扩展代码编辑器。
- **React Force Graph**：2D 图形可视化。

## 快速开始

### 下载安装

前往本仓库的 releases 进行获取


## 从源码构建

### 前置要求

- **Go**：版本 1.24 或更高。
- **Node.js**：版本 18 或更高。
- **NPM**：包管理器。
- **CGO**：SQLite 驱动所需（Windows 上需 GCC/MinGW）。

### 安装

1.  **克隆仓库**：
    ```bash
    git clone https://github.com/yourusername/notebit.git
    cd notebit
    ```

2.  **安装前端依赖**：
    ```bash
    cd frontend
    npm install
    cd ..
    ```

3.  **安装 Wails**：
    请参阅 [官方 Wails 安装指南](https://wails.io/docs/gettingstarted/installation)。
    ```bash
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    ```

### 开发模式

以实时开发模式运行应用程序：

```bash
wails dev
```

该命令将：
- 构建 Go 后端。
- 为前端启动 Vite 开发服务器（支持热重载）。
- 启动桌面应用程序窗口。

您也可以在浏览器中访问 `http://localhost:34115`。

### 生产构建

创建可分发的应用程序包：

```bash
wails build
```

输出二进制文件将位于 `build/bin` 目录中。

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

## 📄 许可证

[许可信息待定]

## 其他

> Icon来自网络，如有侵犯请联系仓库作者 **imicola**

> 本项目 90% 代码来自于vibe coding