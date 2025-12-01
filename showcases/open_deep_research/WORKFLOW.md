# Open Deep Research 流程图

## 整体架构流程

```mermaid
graph TB
    Start([用户输入研究问题]) --> Init[初始化研究<br/>创建研究简报]
    Init --> Supervisor[Supervisor 代理<br/>分析并委派任务]
    
    Supervisor --> Think{使用 think_tool<br/>规划研究策略}
    Think --> Delegate{委派研究任务}
    
    Delegate -->|并行执行| R1[Researcher 1<br/>研究子任务 1]
    Delegate -->|并行执行| R2[Researcher 2<br/>研究子任务 2]
    Delegate -->|并行执行| R3[Researcher 3<br/>研究子任务 3]
    
    R1 --> C1[压缩研究结果 1]
    R2 --> C2[压缩研究结果 2]
    R3 --> C3[压缩研究结果 3]
    
    C1 --> Collect[收集所有研究结果]
    C2 --> Collect
    C3 --> Collect
    
    Collect --> Check{研究是否完整?}
    Check -->|需要更多信息| Supervisor
    Check -->|完成| Report[生成最终报告]
    
    Report --> End([输出综合研究报告])
    
    style Start fill:#e1f5e1
    style End fill:#e1f5e1
    style Supervisor fill:#fff4e1
    style R1 fill:#e1f0ff
    style R2 fill:#e1f0ff
    style R3 fill:#e1f0ff
    style Report fill:#ffe1f0
```

## Supervisor 工作流详细流程

```mermaid
graph TB
    S1([接收研究简报]) --> S2{是否首次迭代?}
    S2 -->|是| S3[添加系统提示词<br/>+ 研究简报]
    S2 -->|否| S4[添加系统提示词<br/>+ 对话历史]
    
    S3 --> S5[调用 LLM<br/>with tools]
    S4 --> S5
    
    S5 --> S6{LLM 响应类型?}
    
    S6 -->|think_tool| S7[记录反思内容]
    S6 -->|ConductResearch| S8[并行执行研究任务]
    S6 -->|ResearchComplete| S9([返回完成状态])
    
    S7 --> S10[返回工具消息]
    S8 --> S11[等待所有研究完成]
    
    S11 --> S12[收集压缩后的结果]
    S12 --> S10
    
    S10 --> S13{达到最大迭代?}
    S13 -->|否| S2
    S13 -->|是| S9
    
    style S1 fill:#e1f5e1
    style S9 fill:#ffe1e1
    style S5 fill:#fff4e1
    style S8 fill:#e1f0ff
```

## Researcher 工作流详细流程

```mermaid
graph TB
    R1([接收研究主题]) --> R2[添加系统提示词<br/>+ 研究主题]
    R2 --> R3[调用 LLM<br/>with search tools]
    
    R3 --> R4{LLM 响应类型?}
    
    R4 -->|tavily_search| R5[执行网络搜索]
    R4 -->|think_tool| R6[记录思考过程]
    R4 -->|无工具调用| R7[结束研究]
    
    R5 --> R8[存储原始搜索结果]
    R6 --> R9[返回反思消息]
    
    R8 --> R9
    R9 --> R10{达到工具调用限制?}
    
    R10 -->|否| R3
    R10 -->|是| R7
    
    R7 --> R11[压缩研究发现]
    R11 --> R12[生成结构化总结]
    R12 --> R13([返回压缩结果])
    
    style R1 fill:#e1f5e1
    style R13 fill:#e1f5e1
    style R3 fill:#fff4e1
    style R5 fill:#e1f0ff
    style R11 fill:#ffe1f0
```

## 状态管理和消息流

```mermaid
sequenceDiagram
    participant User as 用户
    participant Main as 主工作流
    participant Sup as Supervisor
    participant Res as Researcher
    participant LLM as LLM API
    participant Search as Tavily API
    
    User->>Main: 提交研究问题
    Main->>Main: 初始化状态<br/>[messages, supervisor_messages, notes]
    Main->>Sup: 调用 Supervisor 子图
    
    Sup->>LLM: [system, human: 研究简报]
    LLM->>Sup: [ai: tool_calls(think_tool, ConductResearch)]
    
    Sup->>Sup: 处理 think_tool
    Sup->>Res: 并行调用 Researcher 子图 x3
    
    par Researcher 1
        Res->>LLM: [system, human: 子任务1]
        LLM->>Res: [ai: tool_calls(tavily_search)]
        Res->>Search: 执行搜索
        Search->>Res: 搜索结果
        Res->>Res: 存储原始结果
        Res->>LLM: [system, ..., tool: 搜索结果]
        LLM->>Res: [ai: 继续或结束]
        Res->>LLM: 压缩请求
        LLM->>Res: 压缩后的总结
    and Researcher 2
        Res->>LLM: [system, human: 子任务2]
        Note over Res,Search: 类似流程...
    and Researcher 3
        Res->>LLM: [system, human: 子任务3]
        Note over Res,Search: 类似流程...
    end
    
    Res-->>Sup: 返回压缩结果 x3
    Sup->>Sup: 收集所有结果到 notes
    Sup->>LLM: [system, ai, tool, ai, tool]
    LLM->>Sup: [ai: ResearchComplete]
    
    Sup-->>Main: 返回 notes
    Main->>LLM: 生成最终报告请求
    LLM->>Main: 最终报告
    Main->>User: 输出报告
```

## 数据流图

```mermaid
graph LR
    subgraph 输入
        Q[用户查询]
    end
    
    subgraph 初始化
        Q --> Brief[研究简报]
        Brief --> State1[初始状态<br/>messages: []<br/>supervisor_messages: []<br/>notes: []]
    end
    
    subgraph Supervisor循环
        State1 --> SM1[supervisor_messages<br/>+ AI message]
        SM1 --> SM2[supervisor_messages<br/>+ tool messages]
        SM2 --> Notes[notes<br/>+ 研究结果]
    end
    
    subgraph Researcher并行
        SM1 -.委派.-> R1State[Researcher 1<br/>messages: []]
        SM1 -.委派.-> R2State[Researcher 2<br/>messages: []]
        SM1 -.委派.-> R3State[Researcher 3<br/>messages: []]
        
        R1State --> R1Notes[raw_notes<br/>+ 搜索结果]
        R2State --> R2Notes[raw_notes<br/>+ 搜索结果]
        R3State --> R3Notes[raw_notes<br/>+ 搜索结果]
        
        R1Notes --> R1Comp[compressed_research]
        R2Notes --> R2Comp[compressed_research]
        R3Notes --> R3Comp[compressed_research]
        
        R1Comp -.返回.-> SM2
        R2Comp -.返回.-> SM2
        R3Comp -.返回.-> SM2
    end
    
    subgraph 最终报告
        Notes --> Findings[所有研究发现]
        Findings --> FinalReport[最终报告]
    end
    
    subgraph 输出
        FinalReport --> Output[综合研究报告]
    end
    
    style Q fill:#e1f5e1
    style Output fill:#e1f5e1
    style SM1 fill:#fff4e1
    style R1Comp fill:#e1f0ff
    style R2Comp fill:#e1f0ff
    style R3Comp fill:#e1f0ff
    style FinalReport fill:#ffe1f0
```

## 关键概念说明

### 1. 状态累积
- 使用 `AppendReducer` 累积消息历史
- 每个节点返回的消息会追加到状态中
- 保持完整的对话上下文

### 2. 消息序列
正确的消息顺序：
```
[system] -> [human] -> [ai with tool_calls] -> [tool responses] -> [ai] -> ...
```

### 3. 并行执行
- Supervisor 使用 goroutines 并行调用多个 Researcher
- 使用 channels 收集结果
- 限制最大并发数量

### 4. 迭代控制
- Supervisor: `MAX_RESEARCHER_ITERATIONS` (默认 10)
- Researcher: `MAX_TOOL_CALL_ITERATIONS` (默认 20)
- 防止无限循环

### 5. 子图集成
- Supervisor 和 Researcher 都是独立的子图
- 每个子图有自己的 schema 和 reducers
- 主图协调子图的执行
