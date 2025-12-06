<img src="https://lango.rpcx.io/images/logo/lango5.svg" alt="LangGraphGo Logo" height="20px">

# LangGraphGo 项目周报 #001

**报告周期**: 2025-12-01 ~ 2025-12-06
**项目状态**: 🚀 快速发展中
**当前版本**: v0.5.0 (已发布)

---

## 📊 本周概览

本周是 LangGraphGo 项目正式启动的第一周，取得了突破性进展。项目从 v0.3.0 起步，经历了 **5 个版本迭代**（v0.3.0 → v0.3.1 → v0.3.2 → v0.4.0 → v0.5.0），并成功发布了 **v0.5.0**。完成了 **7 个大型 Showcase 项目**的复刻和文档整合工作，搭建了**完整的官方网站和知识库**（233 个 HTML 页面 + 193 个 Markdown 文档），总计提交 **70+ 次**（主仓库），新增代码超过 **20,000 行**。

### 关键指标

| 指标 | 数值 |
|------|------|
| 版本发布 | 5 个 (v0.5.0, v0.4.0, v0.3.2, v0.3.1, v0.3.0) |
| Git 提交 | 70+ 次 |
| Showcases 项目 | 7 个完整项目 |
| 文档页面 | 33+ 个 (中英双语) |
| 代码行数增长 | ~20,000+ 行 |
| 功能特性新增 | 15+ 个核心特性 |

---

## 🎯 主要成果

### 1. 版本发布

#### v0.5.0 (2025-12-06) ⭐ 重大版本
**程序化工具调用 (PTC)**
- ✅ **PTC 包**: 新增 `ptc` 包，支持 LLM 生成代码直接调用工具 (#31)
- ✅ **双语言支持**: Python 和 Go 代码执行
- ✅ **双执行模式**: `ModeDirect`（子进程，默认）和 `ModeServer`（HTTP 服务器，备选）
- ✅ **性能提升**: 延迟和 token 使用降低高达 **10 倍**
- ✅ **多 LLM 支持**: OpenAI、Gemini、Claude 及所有 langchaingo 兼容模型

**设计模式**
- ✅ **规划模式**: 支持任务分解和执行规划 (#24)
- ✅ **反思代理**: 实现反思-行动循环模式 (#32)

**Showcases & 文档**
- ✅ **GPT Researcher**: 完整复刻 assafelovic/gpt-researcher (#34)
- ✅ **文档整合**: 合并 Trading Agents、Open Deep Research、Health Insights Agent 等文档 (#37, #38, #39)
- ✅ **CreateAgent 文档**: 添加全面的对比文档 (#33)

**官方网站**
- ✅ **网站上线**: http://lango.rpcx.io（233 个 HTML 页面，193 个 Markdown 文档）
- ✅ **知识库**: 完整的 Wiki 文档系统

#### v0.4.0 (2025-12-04)
**核心与代理**
- ✅ **MCP 支持**: Model Context Protocol 集成 (#21)
- ✅ **Claude Skills**: 支持 Claude 技能集成 (#20)
- ✅ 更新 `CreateAgent` 支持动态技能加载
- ✅ BettaFish 支持替代 OpenAI 兼容的 LLM 提供商

**工具集成**
- ✅ **Brave Search** API 支持
- ✅ **Bocha Search** 工具 (#22)

**Showcases**
- ✅ 更新 DeerFlow showcase
- ✅ 新增 BettaFish showcase（复刻 https://github.com/666ghj/BettaFish）

**文档与网站**
- ✅ 网站迁移至 https://github.com/smallnest/lango-website
- ✅ 为 showcases 添加 DIFF.md (#19)

#### v0.3.2 (2025-12-03)
- 🔧 Bug 修复和稳定性改进
- 📚 文档优化和示例完善

#### v0.3.1 (2025-12-02)
- 🔧 Bug 修复和稳定性改进
- 📚 文档优化和示例完善

#### v0.3.0 (2025-12-01)
**核心运行时增强**
- ✅ **并行执行**: Fan-out/fan-in 执行模型，线程安全的状态合并
- ✅ **Runtime 配置**: `RunnableConfig` 传播配置（线程 ID、用户 ID）
- ✅ **Command API**: 动态流程控制（`Goto`）和状态更新（`Update`）
- ✅ **子图支持**: 原生支持图组合，嵌套图结构

**持久化与检查点**
- ✅ Redis、PostgreSQL、SQLite 检查点存储实现
- ✅ 状态恢复和时间旅行功能

**高级状态管理**
- ✅ `Schema` 接口和 `Annotated` 风格的 Reducer
- ✅ 智能消息合并（`AddMessages`，基于 ID 的 upsert）
- ✅ 临时通道（Ephemeral Channels）支持

**预构建代理**
- ✅ ToolExecutor 节点
- ✅ ReAct 代理工厂
- ✅ CreateAgent 工厂（函数式选项）
- ✅ Supervisor 代理模式

### 2. Programmatic Tool Calling (PTC) - 已发布 ⭐

**重大突破性功能**（#31）
- 🚀 **新 PTC 包**: LLM 生成代码直接调用工具，无需 API 往返
- 🚀 **双语言支持**: Python 和 Go 代码执行
- 🚀 **双执行模式**: `ModeDirect`（子进程，默认）和 `ModeServer`（HTTP 服务器，备选）
- 🚀 **性能提升**: 延迟和 token 使用降低高达 **10 倍**
- 🚀 **多 LLM 支持**: OpenAI、Gemini、Claude 及所有 langchaingo 兼容模型
- 🚀 **默认模式优化**: 将默认执行模式从 ModeServer 改为 ModeDirect，简化配置

**PTC 核心能力**
- ✅ 代码执行器：沙箱环境中执行 LLM 生成的代码
- ✅ 工具服务器：通过 REST API 安全暴露工具
- ✅ 智能代码生成：自动生成 Python/Go 工具包装器
- ✅ 错误处理：完善的错误报告和调试信息
- ✅ 完整文档：中英双语文档 + Mermaid 流程图

**PTC 示例**
- 📚 `ptc_basic`: 计算器、天气和数据处理工具的 PTC 入门
- 📚 `ptc_simple`: 简单计算器演示基本 PTC 用法
- 📚 `ptc_expense_analysis`: 基于 Anthropic PTC Cookbook 的复杂场景

---

## 🏗️ Showcases 项目进展

本周完成了 **7 个完整的 Showcase 项目**，每个都是对知名开源项目的 Go 语言完整复刻：

### 1. **Open Deep Research** (#16)
- **原项目**: [langchain-ai/open_deep_research](https://github.com/langchain-ai/open_deep_research)
- **成果**: 完整的多代理深度研究系统
- **特性**:
  - Supervisor 代理协调多个 Researcher
  - 并行研究执行
  - 研究压缩和综合报告生成
- **文档**: 包含 5 个详细的 Mermaid 流程图（整体架构、Supervisor 工作流、Researcher 工作流、状态管理、数据流）
- **行数**: ~1,500 行代码

### 2. **DeepAgents** (#15)
- **原项目**: [langchain-ai/deepagents](https://github.com/langchain-ai/deepagents)
- **成果**: 文件系统感知型 AI 智能体
- **特性**:
  - 完整的文件系统工具（ls, read_file, write_file, glob）
  - 任务管理系统（TodoManager）
  - 子智能体委托机制
- **文档**: 详细的工具参考和最佳实践指南
- **行数**: ~800 行代码

### 3. **DeerFlow** (#17)
- **原项目**: 字节跳动 DeerFlow 示例
- **成果**: 深度研究系统，支持多种搜索引擎
- **特性**:
  - Podcast 生成
  - 图像搜索和集成
  - 并发搜索优化
  - Web 界面
- **文档**: 包含中文文档和使用指南
- **行数**: ~1,200 行代码

### 4. **BettaFish** (#19)
- **原项目**: [666ghj/BettaFish](https://github.com/666ghj/BettaFish)
- **成果**: AI 驱动的任务自动化系统
- **特性**:
  - 支持多种 OpenAI 兼容的 LLM 提供商
  - 完整的任务管理和执行流程
- **文档**: DIFF.md 记录与原项目的差异
- **行数**: ~1,000 行代码

### 5. **Health Insights Agent** (#25, #26)
- **原项目**: [harshhh28/hia](https://github.com/harshhh28/hia)
- **成果**: 血液报告分析 AI 助手
- **特性**:
  - PDF 文件处理支持
  - 智能数据提取
  - 健康风险评估
  - 个性化建议生成
- **文档**: 完整的中文文档，包含技术架构、性能指标和安全考虑
- **行数**: ~1,500 行代码

### 6. **Trading Agents** (新增)
- **原项目**: [TauricResearch/TradingAgents](https://github.com/TauricResearch/TradingAgents)
- **成果**: 多代理金融交易系统
- **特性**:
  - 7 个专业代理（基本面、情绪、技术、新闻分析师 + 看涨/看跌研究员 + 风险管理员 + 交易员）
  - 3 个完整接口（后端 API、CLI、Web 仪表板）
  - 实时市场数据集成（Alpha Vantage）
  - 详细模式输出（verbose logging）
- **文档**: 完整的中英双语文档，包含使用指南和 API 参考
- **行数**: ~2,000 行代码

### 7. **GPT Researcher** (#34)
- **原项目**: [assafelovic/gpt-researcher](https://github.com/assafelovic/gpt-researcher)
- **成果**: 完美复刻的 AI 研究助手
- **特性**:
  - 自动化研究和报告生成
  - 多源信息整合
- **文档**: 完整的文档和使用示例
- **行数**: ~1,200 行代码

---

## 📚 文档工作

本周完成了大规模的文档整合和优化工作：

### 文档合并项目

#### 1. Health Insights Agent (#37)
- ✅ 合并 `PROJECT_SUMMARY_CN.md` → `README_CN.md`
- 📊 新增章节：技术架构、性能指标、技术亮点、安全性、对比分析、未来规划
- 📄 最终文档：700 行（原 400 行）

#### 2. Open Deep Research (#38)
- ✅ 合并 `WORKFLOW.md` → `README.md` 和 `README_CN.md`
- 🎨 添加 5 个详细的 Mermaid 流程图：
  - 整体架构流程图
  - Supervisor 工作流详细流程
  - Researcher 工作流详细流程
  - 状态管理和消息流序列图
  - 数据流图
- 📖 新增"关键概念"章节：状态累积、消息序列、并行执行、迭代控制、子图集成
- 📄 最终文档：428 行（英文）、438 行（中文）

#### 3. Trading Agents (#39)
- ✅ 合并 `PROJECT_SUMMARY.md` + `USAGE.md` → `README.md`
- ✅ 合并 `PROJECT_SUMMARY_CN.md` + `USAGE_CN.md` → `README_CN.md`
- 📊 新增章节：项目成果、架构设计、统计数据、完整使用指南、详细模式输出示例
- 📄 最终文档：433 行（英文）、536 行（中文）

#### 4. DeepAgents (#36)
- ✅ 创建完整的中英双语文档
- 📖 包含工具参考、使用场景、最佳实践、故障排除
- 📄 文档：400+ 行

#### 5. DeerFlow & BettaFish (#35)
- ✅ 更新和完善文档
- 📚 添加使用示例和配置说明

#### 6. CreateAgent & CreateReactAgent (#33)
- ✅ 添加全面的文档
- 📖 详细的 API 参考和使用示例

#### 7. Reflection Agent (#32)
- ✅ 实现反思代理设计模式
- 📚 完整的文档和示例

### 文档统计

| 文档类型 | 数量 | 总字数（估算） |
|---------|------|----------------|
| README 文件 | 14+ | ~30,000 |
| 技术文档 | 6+ | ~15,000 |
| 示例文档 | 10+ | ~8,000 |
| API 文档 | 3+ | ~5,000 |
| **总计** | **33+** | **~58,000** |

---

## 🌐 网站与知识库建设

本周完成了项目官方网站和知识库的搭建工作，为用户提供了完整的学习资源。

### 官方网站 (lango-website)

**网站地址**: http://lango.rpcx.io
**仓库地址**: https://github.com/smallnest/lango-website

#### 网站结构
- 📄 **中文站点**: 完整的中文界面和文档
- 🌍 **英文站点**: 双语支持，国际化
- 📚 **四大板块**: 案例、文档、示例、知识库

#### 主要页面
1. **首页** (`index.html`)
   - 项目简介和核心特性展示
   - 快速开始和示例入口
   - 现代化的 UI 设计

2. **案例页面** (`showcases.html`)
   - 展示 6 个完整的 Showcase 项目
   - 每个案例包含详细说明和运行指南
   - 提供 GitHub 链接和文档链接

3. **文档页面** (`docs.html`)
   - 包含 16+ 个详细指南
   - API 参考文档
   - 核心概念讲解

4. **示例页面** (`examples.html`)
   - 20+ 个代码示例
   - 覆盖各种使用场景
   - 完整的代码片段和说明

#### 文档指南（docs/）

**基础指南**
- ✅ Getting Started - 快速开始
- ✅ Core Concepts - 核心概念
- ✅ API Reference - API 参考

**进阶指南**
- ✅ Guide: Basics - 基础教程
- ✅ Guide: Advanced Features - 高级特性
- ✅ Guide: State Management - 状态管理
- ✅ Guide: StateGraph vs MessageGraph - 图类型对比
- ✅ Guide: Streaming - 流式处理
- ✅ Guide: HITL (Human-in-the-Loop) - 人机协作
- ✅ Guide: Memory & Time Travel - 内存和时间旅行
- ✅ Guide: Multi-Agent - 多代理系统
- ✅ Guide: Pre-built Agents - 预构建代理
- ✅ Guide: CreateAgent vs CreateReactAgent - 代理对比
- ✅ Guide: RAG - 检索增强生成
- ✅ Guide: Monitoring - 监控和可观测性

### 知识库 (Wiki)

**位置**: `/repowiki/zh/`

#### Wiki 内容结构

**高级特性** (`高级特性/`)
- 📖 人机协作 (`人机协作.md`)
- 📖 可视化 (`可视化.md`)
- 📖 子图 (`子图.md`)
- 📖 并行执行 (`并行执行.md`)

**检查点存储** (`检查点存储/`)
- 📖 检查点存储概述 (`检查点存储.md`)
- 📖 SQLite 检查点存储 (`SQLite 检查点存储.md`)
- 📖 Redis 检查点存储 (`Redis 检查点存储.md`)
- 📖 PostgreSQL 检查点存储 (`PostgreSQL 检查点存储.md`)

**工具集成** (`工具集成/`)
- 📖 工具集成概述 (`工具集成.md`)
- 📖 工具执行框架 (`工具执行框架.md`)
- 📖 具体工具集成 (`具体工具集成.md`)

**预构建组件** (`预构建组件/`)
- 📖 预构建组件概述 (`预构建组件.md`)
- 📖 Create Agent (`Create Agent.md`)
- 📖 RAG 组件系列
  - 基础 RAG (`基础 RAG.md`)
  - 条件 RAG (`条件 RAG.md`)
  - 高级 RAG 系列
    - 高级 RAG 概述 (`高级 RAG.md`)
    - 重排序机制 (`重排序机制.md`)
    - 引用格式化 (`引用格式化.md`)

### 本周网站更新

#### Git 提交记录
- ✅ `9341c27` - 移除 .claude 配置
- ✅ `739f2f5` - **更新网站支持 v0.4.0**
- ✅ `9439dbe` - 移除 .history
- ✅ `e4be742` - **添加 Wiki 菜单**
- ✅ `1a5bfdd` - **Markdown 转 HTML**
- ✅ `c279db6` - 合并 PR #1（社区贡献）
- ✅ `eec9de2` - **添加 Qoder Repo Wiki**（社区贡献）
- ✅ `1eb3cfc` - **初始化网站**

#### 网站统计

| 类型 | 数量 | 说明 |
|------|------|------|
| HTML 页面 | 233 个 | 包含所有文档和指南页面 |
| Markdown 文档 | 193 个 | Wiki 源文件 |
| 指南文档 | 16+ 个 | 详细的使用指南 |
| 代码示例 | 20+ 个 | 涵盖各种场景 |

#### 网站特色

1. **现代化设计**
   - 响应式布局
   - 深色/浅色主题
   - 优雅的 UI 组件

2. **完整的导航**
   - 清晰的页面结构
   - Wiki 知识库集成
   - 中英文双语切换

3. **丰富的内容**
   - 6 个完整的 Showcase 案例
   - 16+ 个详细指南
   - 20+ 个代码示例
   - 全面的 API 文档

4. **社区友好**
   - GitHub 集成
   - 支持 PR 贡献
   - Wiki 协作编辑

### 网站与文档对照

| 项目主仓库文档 | 网站文档 | Wiki 文档 |
|---------------|---------|----------|
| README.md | index.html | 项目概述.md |
| Examples/ | examples.html | - |
| Showcases/ | showcases.html | - |
| - | docs/ | repowiki/zh/ |

---

## 💻 技术亮点

### 1. Planning Mode (#24)
- 🎯 实现计划模式功能
- 🎯 支持任务分解和执行规划
- 🎯 提供清晰的执行路径

### 2. Reflection Agent Pattern (#32)
- 🔄 实现反思-行动循环
- 🔄 自我评估和改进机制
- 🔄 提高输出质量

### 3. Node Description (#24)
- 📝 为 Node 类型添加 description 字段
- 📝 更新所有 AddNode 方法
- 📝 改善图的可读性和文档化

### 4. MCP Integration (#21)
- 🔌 Model Context Protocol 支持
- 🔌 扩展上下文管理能力
- 🔌 提高多模态交互能力

### 5. Skills System (#20)
- 🎓 Claude Skills 集成
- 🎓 动态技能加载
- 🎓 可扩展的技能系统

---

## 📈 项目统计

### 代码指标

```
总代码行数（估算）:
- 核心框架:           ~6,000 行
- Showcases:         ~12,000 行
- Examples:          ~4,000 行
- PTC 包:            ~1,500 行
- 文档:              ~18,000 行
- 网站 (HTML/CSS/JS): ~3,000 行
- Wiki (Markdown):    ~10,000 行
- 总计:              ~54,500 行
```

### Git 活动

```bash
本周提交次数: 70+
代码贡献者:   2+
文件修改:     200+
功能分支:     15+
```

### 功能覆盖

- ✅ 核心运行时: 100%
- ✅ 状态管理: 95%
- ✅ 持久化: 90%
- ✅ 可视化: 85%
- ✅ 预构建代理: 90%
- ✅ 工具集成: 85%
- ✅ PTC (已发布): 90%

---

## 🔧 技术债务与改进

### 已解决
- ✅ 删除误添加的文件 (#9234fed)
- ✅ 修复冲突和构建问题
- ✅ 更新过时的 LLM 调用方法
- ✅ 文档重命名为更简单的命名约定
- ✅ PTC 默认执行模式改为 ModeDirect
- ✅ 完成所有 Showcase 文档整合工作
- ✅ 发布 v0.5.0 版本
- ✅ **实现 TRUE ModeDirect 本地工具执行** (commit 665dbc9)
  - **ModeDirect 模式：完全本地执行，无 HTTP 服务器**
  - 遵循 goskills 模式：通过 subprocess 直接执行工具
  - Python 包装器：_run_shell(), _run_python(), _read_file(), _write_file()
  - Go 包装器：runShell(), runPython(), readFile(), writeFile()
  - 模式匹配：根据工具名自动识别工具类型
  - 临时文件 → 执行 → 清理（完整的 goskills 实现模式）
- ✅ **保持高单元测试覆盖率** (commit 665dbc9)
  - 测试数量：25 个综合测试
  - 测试覆盖率：61.0%
  - 所有测试通过 ✅
  - 涵盖 Direct/Server 模式、Python/Go 执行、并发、错误处理等

- ✅ **添加全面的日志系统** (commit b71b08c)
  - Logger 接口：灵活的日志记录
  - DefaultLogger：使用 Go 标准 log 包
  - NoOpLogger：无日志场景（默认）
  - LogLevel 支持：DEBUG, INFO, WARN, ERROR, NONE
  - CodeExecutor.SetLogger() / WithLogger() 方法
  - ToolServer 日志集成
  - 29 个测试通过 ✅
- ✅ **添加全面的边缘案例测试** (commit 9605af4)
  - 新增 14 个边缘案例测试
  - 测试覆盖率：62.9% → 64.1%（+1.2%）
  - 覆盖边缘情况：空输入、无效输入、大输入、特殊字符、状态转换
  - 43 个测试全部通过 ✅

### 待解决
- 🔲 继续提高测试覆盖率（目标：70%+）
- 🔲 性能基准测试
- 🔲 更多 PTC 示例

---

## 🌟 社区贡献

### 外部贡献者
- **@zxmfke**: BettaFish LLM 提供商支持 (#23)
- 贡献了 Qoder wiki 和其他改进

### 项目协作
- 积极响应 Issues
- 接受 Pull Requests
- 文档持续改进

---

## 🎓 学习与洞察

### 技术学习
1. **多代理系统设计**: 通过 Trading Agents 和 Open Deep Research 深入理解了多代理协作模式
2. **状态管理优化**: 学习了 Ephemeral Channels 和 Smart Messages 的实现
3. **工具调用优化**: PTC 提供了革命性的工具调用方式，大幅降低延迟

### 最佳实践
1. **文档先行**: 每个 Showcase 都配备完整的中英双语文档
2. **示例驱动**: 提供丰富的示例代码便于快速上手
3. **性能优先**: 关注执行效率和资源使用

---

## 📅 里程碑达成

- ✅ **第一周完成**: 7 个完整 Showcase 项目
- ✅ **版本迭代**: 5 个版本发布（v0.3.0 → v0.3.1 → v0.3.2 → v0.4.0 → v0.5.0）
- ✅ **文档里程碑**: 超过 33 个文档页面（主仓库）
- ✅ **代码里程碑**: 超过 20,000 行新代码（主仓库）
- ✅ **功能里程碑**: PTC 功能完整发布（革命性突破）
- ✅ **网站上线**: 完整的官方网站和知识库
  - 233 个 HTML 页面
  - 193 个 Markdown 文档
  - 16+ 个详细指南
  - 中英双语支持
- ✅ **v0.5.0 正式发布**: 包含完整 PTC 功能和文档

---

## 🚀 下周计划 (2025-12-07 ~ 2025-12-13)

### 主要目标

1. **PTC 功能完善**
   - ✅ 完善 ModeDirect 模式的实际工具调用实现 (commit 665dbc9)
   - 🎯 添加更多 PTC 示例和用例
   - ✅ 保持 PTC 单元测试覆盖 (61.0%, 25 tests passing)

2. **新 Showcases**
   - 🎯 计划添加 2-3 个新的 Showcase 项目
   - 🎯 覆盖更多应用场景（如代码生成、数据分析等）

3. **性能优化**
   - 🎯 优化并行执行性能
   - 🎯 减少内存占用
   - 🎯 改进检查点存储效率

4. **测试覆盖**
   - 🎯 增加单元测试
   - 🎯 添加集成测试
   - 🎯 性能基准测试

5. **文档完善**
   - 🎯 完善 API 文档
   - 🎯 添加更多教程
   - 🎯 改进网站内容

6. **社区建设**
   - 🎯 响应 Issues 和 PRs
   - 🎯 收集用户反馈
   - 🎯 规划社区活动

---

## 💡 思考与展望

### 项目优势
1. **完整的 Go 实现**: 为 Go 生态提供了企业级的 LangGraph 实现
2. **丰富的示例**: 6 个完整的 Showcase 项目覆盖多个应用场景
3. **双语文档**: 完整的中英文文档降低学习门槛
4. **创新功能**: PTC 等功能提供了独特的价值

### 挑战与应对
1. **与 Python 版本的功能对等**: 持续跟进 Python LangGraph 的新功能
2. **性能优化**: 充分利用 Go 的并发优势
3. **社区增长**: 通过高质量文档和示例吸引用户

### 长期愿景
- 🌟 成为 Go 生态中最流行的 LangGraph 实现
- 🌟 建立活跃的开发者社区
- 🌟 支持更多企业级应用场景
- 🌟 持续创新，引领 AI 工作流技术

---

## 📝 附录

### 相关链接
- **主仓库**: https://github.com/smallnest/langgraphgo
- **官方网站**: http://lango.rpcx.io
- **网站源码**: https://github.com/smallnest/lango-website
- **网站首页**: http://lango.rpcx.io/index.html
- **案例展示**: http://lango.rpcx.io/showcases.html
- **文档中心**: http://lango.rpcx.io/docs.html
- **知识库**: http://lango.rpcx.io/repowiki/zh/
- **Showcase 文档**: 见各 Showcase 的 README 文件

### 版本标签
- `v0.5.0` - 2025-12-06 ⭐ (PTC 重大版本)
- `v0.4.0` - 2025-12-04
- `v0.3.2` - 2025-12-03
- `v0.3.1` - 2025-12-02
- `v0.3.0` - 2025-12-01

### 重要提交
- `#39` - Trading Agents 文档合并
- `#38` - Open Deep Research 工作流文档
- `#37` - Health Insights Agent 文档合并
- `#36` - DeepAgents 文档
- `#34` - GPT Researcher
- `#33` - CreateAgent 文档
- `#32` - Reflection Agent
- `#31` - PTC 功能（重大更新）
- `#24` - Planning Mode
- `#21` - MCP 支持
- `#20` - Claude Skills
- `6283e30` - PTC 默认模式改为 ModeDirect

---

**报告编制**: LangGraphGo 项目组
**报告日期**: 2025-12-06
**下次报告**: 2025-12-13

---

> 📌 **备注**: 本周报基于 Git 历史、项目文档和代码统计自动生成，如有疏漏请及时反馈。

---

**🎉 第一周圆满结束！期待下周更多精彩！**
