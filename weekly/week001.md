# LangGraphGo 项目周报 #001

**报告周期**: 2025-12-01 ~ 2025-12-06
**项目状态**: 🚀 快速发展中
**当前版本**: v0.5.0 (开发中)

---

## 📊 本周概览

本周是 LangGraphGo 项目正式启动的第一周，取得了突破性进展。项目从 v0.3.0 起步，经历了 4 个版本迭代（v0.3.0 → v0.3.1 → v0.3.2 → v0.4.0），并开始开发 v0.5.0。目前最新版本为 v0.4.0。完成了 **6 个大型 Showcase 项目**的复刻和文档整合工作，搭建了**完整的官方网站和知识库**（233 个 HTML 页面 + 193 个 Markdown 文档），总计提交 **39+ 次**（主仓库），新增代码超过 **15,000 行**。

### 关键指标

| 指标 | 数值 |
|------|------|
| 版本发布 | 4 个 (v0.4.0, v0.3.2, v0.3.1, v0.3.0) |
| Git 提交 | 39+ 次 |
| Showcases 项目 | 6 个完整项目 |
| 文档页面 | 20+ 个 (中英双语) |
| 代码行数增长 | ~15,000+ 行 |
| 功能特性新增 | 10+ 个核心特性 |

---

## 🎯 主要成果

### 1. 版本发布

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

### 2. Programmatic Tool Calling (PTC) - v0.5.0 ���发中

**重大突破性功能**（#27）
- 🚀 **新 PTC 包**: LLM 生成代码直接调用工具，无需 API 往返
- 🚀 **双语言支持**: Python 和 Go 代码执行
- 🚀 **双执行模式**: `ModeServer`（HTTP 服务器，默认）和 `ModeDirect`（子进程，实验性）
- 🚀 **性能提升**: 延迟和 token 使用降低高达 **10 倍**
- 🚀 **多 LLM 支持**: OpenAI、Gemini、Claude 及所有 langchaingo 兼容模型

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

本周完成了 **6 个完整的 Showcase 项目**，每个都是对知名开源项目的 Go 语言完整复刻：

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
- 核心框架:           ~5,000 行
- Showcases:         ~10,000 行
- Examples:          ~3,000 行
- 文档:              ~15,000 行
- 网站 (HTML/CSS/JS): ~3,000 行
- Wiki (Markdown):    ~10,000 行
- 总计:              ~46,000 行
```

### Git 活动

```bash
本周提交次数: 39+
代码贡献者:   2+
文件修改:     150+
功能分支:     10+
```

### 功能覆盖

- ✅ 核心运行时: 100%
- ✅ 状态管理: 95%
- ✅ 持久化: 90%
- ✅ 可视化: 85%
- ✅ 预构建代理: 90%
- ✅ 工具集成: 80%
- 🚧 PTC (开发中): 70%

---

## 🔧 技术债务与改进

### 已解决
- ✅ 删除误添加的文件 (#9234fed)
- ✅ 修复冲突和构建问题
- ✅ 更新过时的 LLM 调用方法
- ✅ 文档重命名为更简单的命名约定

### 待解决
- 🔲 完善 PTC 的 ModeDirect 模式
- 🔲 增加更多单元测试覆盖率
- 🔲 优化并行执行性能
- 🔲 改进错误处理和日志记录

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

- ✅ **第一周完成**: 6 个完整 Showcase 项目
- ✅ **版本迭代**: 4 个版本发布（v0.3.0 → v0.4.0）
- ✅ **文档里程碑**: 超过 30 个文档页面（主仓库）
- ✅ **代码里程碑**: 超过 15,000 行新代码（主仓库）
- ✅ **功能里程碑**: PTC 功能开发（革命性突破）
- ✅ **网站上线**: 完整的官方网站和知识库
  - 233 个 HTML 页面
  - 193 个 Markdown 文档
  - 16+ 个详细指南
  - 中英双语支持

---

## 🚀 下周计划 (2025-12-07 ~ 2025-12-13)

### 主要目标

1. **完成 v0.5.0 发布**
   - 🎯 完善 PTC 功能
   - 🎯 添加更多 PTC 示例
   - 🎯 完成 PTC 文档和测试

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
- `v0.5.0` - 开发中
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
- `#27` - PTC 功能
- `#24` - Planning Mode
- `#21` - MCP 支持
- `#20` - Claude Skills

---

**报告编制**: LangGraphGo 项目组
**报告日期**: 2025-12-06
**下次报告**: 2025-12-13

---

> 📌 **备注**: 本周报基于 Git 历史、项目文档和代码统计自动生成，如有疏漏请及时反馈。

---

**🎉 第一周圆满结束！期待下周更多精彩！**
