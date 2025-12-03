# BettaFish 实现对比报告

本文档对比了 BettaFish 的 Go 语言实现 (`showcases/BettaFish`) 与原始 Python 实现 (`showcases/BettaFish_Ref`)。

## 1. 高层架构 (High-Level Architecture)

| 特性 | Python 实现 (`BettaFish_Ref`) | Go 实现 (`BettaFish`) |
| :--- | :--- | :--- |
| **框架** | `LangGraph` (Python), `LangChain` (Python) | `langgraphgo`, `langchaingo` |
| **应用类型** | Web 应用 (`app.py`, 可能是 Flask/Streamlit) | CLI 命令行工具 (`main.go`) |
| **图结构** | 多智能体状态图 | 多智能体状态图 (当前为顺序边) |
| **状态管理** | 复杂的 `State` 对象，包含历史追踪 | `BettaFishState` 结构体，包含相似字段 |

## 2. 组件分析 (Component Analysis)

### 2.1 查询引擎 (QueryEngine)
*   **Python**: 完整实现，包括报告结构生成、段落处理（搜索、总结、反思）和最终报告生成。
*   **Go**: **高还原度 (High Fidelity)**。使用 `langchaingo` 完全复刻。
    *   **提示词**: 完全匹配 (`query_engine/prompts.go`)。
    *   **逻辑**: 复刻了 `_generate_report_structure`, `_process_paragraphs` (顺序执行), 和 `_generate_final_report`。
    *   **工具**: 使用 `Tavily` 进行搜索，能力与 Python 版本一致。

### 2.2 媒体引擎 (MediaEngine)
*   **Python**: 复杂的智能体，处理图像搜索并可能包含视觉分析。
*   **Go**: **中等还原度 (Medium Fidelity)**。
    *   **提示词**: 完全匹配 (`media_engine/prompts.go`)。
    *   **逻辑**: 实现了 `MediaEngineNode`，使用 `SystemPromptFirstSearch` 生成查询，并使用 `Tavily` 搜索图片。
    *   **差异**: 目前未实现深度视觉分析或复杂的图像处理，主要侧重于检索图像元数据。

### 2.3 洞察引擎 (InsightEngine)
*   **Python**: 集成了本地 `MindSpider` (爬虫) 和 `SentimentAnalysisModel` (机器学习模型)。
*   **Go**: **模拟/基于LLM (Simulated / LLM-based)**。
    *   **提示词**: 完全匹配 (`insight_engine/prompts.go`)。
    *   **逻辑**: 实现了 `InsightEngineNode`，具有结构化的分析流程（结构 -> 搜索 -> 总结）。
    *   **MindSpider**: 由 `mind_spider` 包替代，使用 `Tavily` 模拟社交媒体搜索。
    *   **SentimentModel**: 由 `sentiment_model` 包替代，使用 LLM (OpenAI) 分析情感，而非本地 ML 模型。

### 2.4 论坛引擎 (ForumEngine)
*   **Python**: 可能是事件驱动或多进程系统，智能体 (`Host`, `NewsAgent`, `MediaAgent`) 在共享环境中交互，由 `monitor.py` 监控。
*   **Go**: **模拟循环 (Simulated Loop)**。
    *   **提示词**: 主持人 (Moderator) 提示词完全匹配 (`forum_engine/agent.go`)。
    *   **逻辑**: 实现了 `ForumEngineNode` 作为单个节点，运行固定轮次的对话循环。它使用各自的系统提示词模拟主持人和其他智能体之间的交互。
    *   **差异**: 不是真正的多智能体并发系统，但在功能上复刻了讨论结果。

### 2.5 报告引擎 (ReportEngine)
*   **Python**: 复杂的报告系统，使用 `Jinja2` 模板、`Pandoc` 和多种渲染器 (HTML, PDF)。
*   **Go**: **简化/原生 (Simplified / Native)**。
    *   **提示词**: 完全匹配 (`report_engine/prompts.go`)。
    *   **逻辑**: 使用 Go 原生的 `text/template` 包实现了 `ReportEngineNode`。
    *   **输出**: 生成结构化的 Markdown 报告。目前不支持 Python 版本中的 PDF 或复杂的 HTML 样式。

## 3. 提示词同步状态 (Prompt Synchronization Status)

所有提示词均已同步，以确保 Go 智能体的行为尽可能接近 Python 智能体。

| 引擎 | 状态 | 源文件 (Python) | 目标文件 (Go) |
| :--- | :--- | :--- | :--- |
| **QueryEngine** | ✅ 已同步 | `QueryEngine/prompts/prompts.py` | `query_engine/prompts.go` |
| **MediaEngine** | ✅ 已同步 | `MediaEngine/prompts/prompts.py` | `media_engine/prompts.go` |
| **InsightEngine** | ✅ 已同步 | `InsightEngine/prompts/prompts.py` | `insight_engine/prompts.go` |
| **ForumEngine** | ✅ 已同步 | `ForumEngine/llm_host.py` | `forum_engine/agent.go` |
| **ReportEngine** | ✅ 已同步 | `ReportEngine/prompts/prompts.py` | `report_engine/prompts.go` |

## 4. 差异总结 (Summary of Gaps)

1.  **本地模型 vs API**: Python 版本依赖本地组件 (`MindSpider`, `SentimentAnalysisModel`)，可能提供更多控制或隐私。Go 版本利用 API (`Tavily`, `OpenAI`) 以便于实现和扩展。
2.  **UI/UX**: Python 版本包含 Web UI。Go 版本目前是 CLI 工具。
3.  **报告格式**: Python 版本具有更高级的报告格式 (PDF/HTML)。Go 版本生成 Markdown。
4.  **并发性**: Python ForumEngine 可能支持真正的异步智能体交互。Go 版本将其序列化为对话循环。

## 5. 结论 (Conclusion)

Go 实现成功复刻了 BettaFish 的核心 **智能体逻辑 (agentic logic)** 和 **工作流 (workflow)**。通过同步提示词，我们确保了智能体的认知行为是相同的。主要的区别在于基础设施（Web 与 CLI）以及用通用 LLM/搜索 API 替代本地专用模型，这对于云原生 Go 实现来说是一个合理的架构选择。
