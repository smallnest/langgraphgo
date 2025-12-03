# GoSkills 集成示例

本示例演示如何将 [GoSkills](https://github.com/smallnest/goskills) 与 LangGraphGo 集成，使代理能够将技能作为工具使用。

## 1. 背景

GoSkills 是一个用于创建和管理可重用技能的框架，可以执行 Python 脚本、Shell 命令，并提供各种内置工具，如网络搜索和文件操作。通过将 GoSkills 与 LangGraphGo 集成，您可以轻松地为代理扩展强大的预打包功能，而无需编写自定义工具实现。

## 2. 核心概念

- **GoSkills**: 一个技能管理框架，将脚本和工具打包成可重用的技能包。
- **技能包 (Skill Package)**: 包含 `SKILL.md` 元数据文件和可选 `scripts/` 子目录中脚本的目录。
- **SkillsToTools 适配器**: 一个便捷函数 (`adapter.SkillsToTools`)，将 GoSkills 技能包转换为与 LangGraphGo 兼容的 `tools.Tool` 接口。
- **多技能支持**: 适配器可以从目录加载并注册多个技能，聚合所有工具。

## 3. 工作原理

1.  **加载技能**: 使用 `goskills.ParseSkillPackages` 扫描目录（例如 `./skills`）以查找技能包。
2.  **转换为工具**: 对于每个技能包，调用 `adapter.SkillsToTools` 将技能的功能转换为 `tools.Tool` 实例。
3.  **聚合工具**: 将所有技能的工具组合到单个切片中。
4.  **创建 Agent**: 使用聚合的工具和包含所有技能描述的系统消息调用 `prebuilt.CreateAgent`。
5.  **调用**: 使用用户查询运行代理。代理将自动从加载的技能中选择并使用适当的工具。

## 4. 代码亮点

### 加载所有技能

```go
skillsDir := "skills"
packages, err := goskills.ParseSkillPackages(skillsDir)
if err != nil {
    log.Fatal(err)
}
```

### 将技能转换为工具

```go
var allTools []tools.Tool
var allSystemMessages strings.Builder

allSystemMessages.WriteString("You are a helpful assistant that can use skills.\n\n")

for _, skill := range packages {
    fmt.Printf("Loading skill: %s\n", skill.Meta.Name)
    skillTools, err := adapter.SkillsToTools(*skill)
    if err != nil {
        log.Printf("Failed to convert skill %s to tools: %v", skill.Meta.Name, err)
        continue
    }
    allTools = append(allTools, skillTools...)
    allSystemMessages.WriteString(fmt.Sprintf("Skill: %s\n%s\n\n", skill.Meta.Name, skill.Body))
}
```

### 创建 Agent

```go
agent, err := prebuilt.CreateAgent(llm, allTools, 
    prebuilt.WithSystemMessage(allSystemMessages.String()))
if err != nil {
    log.Fatal(err)
}
```

## 5. 技能结构

基本技能包结构：

```
skills/
└── hello_world/
    ├── SKILL.md          # 技能元数据和描述
    └── scripts/
        └── hello.py      # Python 脚本
```

**SKILL.md** 示例：
```markdown
---
name: hello_world
description: 一个打印问候语的简单技能
version: 1.0.0
---

## 概述
此技能演示基本脚本执行。
```

## 6. 运行示例

```bash
# 确保 ./skills 目录中有技能
# 如果不存在，示例将创建一个虚拟技能

export OPENAI_API_KEY=your_api_key
go run main.go
```

**预期输出:**
```text
Loading skill: hello_world
Tool: hello, Description: Run hello.py script
User: Please use the available skill to say hello to the world.
Agent: [代理使用 hello_world 技能并返回问候语]
```

## 7. 可用的技能工具

GoSkills 提供了几种可在技能中使用的内置工具类型：

- **脚本执行**: `run_python_script`, `run_shell_script`
- **代码执行**: `run_python_code`, `run_shell_code`
- **文件操作**: `read_file`, `write_file`
- **网络搜索**: `duckduckgo_search`, `wikipedia_search`, `tavily_search`
- **网页抓取**: `web_fetch`

## 8. 了解更多

- [GoSkills 文档](https://github.com/smallnest/goskills)
- [LangGraphGo 预构建 Agent](../../prebuilt/)
- [Create Agent 示例](../create_agent/)
