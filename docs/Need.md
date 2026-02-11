# 产品需求文档 (PRD): Project "Notebit" (暂定名)

**版本:** v0.1 (Draft)
**日期:** 2026-02-11
**负责人:** imicola
**产品经理:** Gemini

## 1. 产品愿景 (Vision)

打造一款 **Local-First（本地优先）** 的 Markdown 笔记应用。它拒绝 AI 对书写过程的侵入，专注于在**书写后**通过语义分析、自动关联和回顾机制，将静态的笔记转化为动态的个人知识资产。

**核心哲学：**

* **Write for Humans:** 编辑器是纯净的，零干扰。
* **Manage by Silicon:** 后台是活跃的，AI 像通过神经突触一样连接你的笔记。

---

## 2. 技术栈 (Tech Stack) - Vibe Coding 架构

* **前端 (UI):** React + Tailwind CSS (追求极速渲染与现代化 UI)
* **后端 (Logic):** Wails (Go)
* 利用 Go 的 `goroutines` 处理高并发的向量计算。
* 利用 Go 的强类型系统保证数据层稳定性。


* **数据存储 (Data):**
* **源文件:** 本地 Markdown 文件 (兼容 Obsidian 格式)。
* **索引库:** SQLite (存储文件元数据 + 向量 Embeddings)。


* **AI 引擎 (Intelligence):**
* **Embeddings:** 云端模型(产品需要考虑对大部分人的可行性，在本地跑个小模型会显著增加软件臃肿和占用)
* **LLM:** 接口层支持 Ollama (本地) 或 OpenAI 兼容接口 (云端)。

---

## 3. 功能需求 (Functional Requirements)

### 模块 A: 纯净编辑器 (The Sanctuary)

* **P0 (最高优先级):**
* 支持标准 Markdown 语法的高亮渲染。
* 支持实时保存到本地文件系统。
* **反向特性：** 禁止任何形式的 AI 自动补全（Ghost Text），除非用户显式触发。


* **P1 (次高优先级):**
* 所见即所得 (WYSIWYG) 模式与源码模式切换。



### 模块 B: 隐形策展人 (The Silent Curator) - 核心差异点

这是后端 Go 程序的重头戏。

* **P0 - 实时向量化 (Live Embedding):**
* 监听文件保存事件 (`fsnotify`)。
* 当文件变更时，后台自动调用 Embedding 模型，将文本转化为向量并存入 SQLite。
* *策略：* 针对长文，按 `## 标题` 或固定 Token 窗口进行切片 (Chunking)。


* **P1 - 语义相关性推荐 (Semantic Sidebar):**
* 在编辑器右侧栏，实时显示“与当前段落最相关的 5 篇笔记”。
* *算法：* 余弦相似度计算 (Cosine Similarity)。
* *场景：* 你在写“Go 接口”，侧边栏自动推给你半年前写的“Java 多态”和“设计模式”。



### 模块 C: 知识反刍 (Review & RAG)

* **P1 - 与笔记对话 (Chat with Notes):**
* 提供一个独立的 Chat 界面。
* 用户提问 -> 检索相关笔记片段 -> 喂给 LLM -> 生成答案。
* *关键点：* 必须标注引用来源（如：*“参考自笔记 [[2025-10-12 Linux网络]]”*）。


* **P2 - 每日次日 (Daily Digest):**
* 基于艾宾浩斯遗忘曲线 + 语义聚类。
* 每天启动时，弹出一张卡片：“你以前写过关于 React 的笔记，最近你又学了 Vue，来看看它们的区别？”

### 模块 D: 知识图谱与架构构建分析

* 存在一个界面：
    * 类似于obsidian的关系图谱，由ai进行主动分析，将关联度大的文件进行连线，也兼容obsidian的双链，这个界面可以放和ai的聊天界面，当ai回答有引用或关联度较高的文章时候，自动高亮显示文章


---

## 4. 非功能需求 (Non-functional Requirements)

* **性能：** 应用冷启动时间 < 2秒。
* **隐私：** 默认情况下，所有数据（包括 Embedding 向量）不出本地。
* **兼容性：** 只要文件夹还在，哪怕软件删了，Markdown 文件必须完好无损。

---

## 5. UI/UX 原型草图 (Mental Mockup)

为了配合你的 Vibe Coding，我们想象一下界面布局：

| 区域 | 功能 | Vibe 描述 |
| --- | --- | --- |
| **左侧栏** | 文件树 | 极简，像 VS Code，支持拖拽。 |
| **中间** | 编辑区 | 纯粹的打字机体验，无干扰。 |
| **右侧栏** | **AI 策展区** | 平时隐藏。点击或快捷键唤出。显示“相关笔记”、“待办提取”、“自动标签”。 |
| **底部** | 状态栏 | 显示 Go 后端状态（如：“正在索引... 3篇待处理”）。 |

---

## 6. 开发路线图 (Vibe Coding Roadmap)

既然是用 Wails + AI 辅助开发，我们将项目拆解为三个“冲刺 (Sprint)”：

**Sprint 1: 骨架构建 (The Skeleton)**

* 初始化 Wails 项目。
* 实现前端 React 编辑器 (集成现成的 Markdown 库)。
* 实现 Go 后端的文件读写 (File I/O)。
* *目标：* 能打开文件夹，写字，保存。

**Sprint 2: 植入大脑 (The Brain)**

* 引入 Go 的向量库 (如 `chromem-go` 或绑定 SQLite-vss)。
* 实现“保存即索引”逻辑。
* 写一个简单的算法，计算两篇笔记的相似度并在控制台输出。

**Sprint 3: 交互层 (The Interface)**

* 开发右侧“相关笔记”UI。
* 对接 Ollama 接口，实现“与笔记对话”。