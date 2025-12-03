# 动态技能代理示例

本示例演示了如何使用 `langgraphgo` 和 `goskills` 创建一个能够根据用户输入动态发现和选择技能的代理。

## 概述

代理配置了 `skillDir`。当它收到用户请求时，会执行以下步骤：
1.  **发现**：扫描指定目录以查找可用技能。
2.  **选择**：使用 LLM 根据用户的请求选择最相关的技能。
3.  **执行**：加载所选技能的工具并执行代理逻辑，其中可能涉及调用工具。

## 先决条件

- Go 1.22+
- OpenAI API Key (设置为 `OPENAI_API_KEY` 环境变量)

## 如何运行

```bash
export OPENAI_API_KEY="your-api-key"
go run examples/dynamic_skill_agent/main.go
```

该示例将：
1.  在 `skills` 目录中创建一个虚拟的 "hello_world" 技能。
2.  使用 `skills` 目录初始化代理。
3.  向代理发送请求 "Please run the hello world script."。
4.  代理将发现技能，选择它，并执行脚本。
