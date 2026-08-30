# 复杂工具示例

本示例演示如何在 LangGraphGo 中创建具有**复杂参数 schema** 的工具，支持多种参数类型，包括字符串、整数、浮点数、布尔值、数组和嵌套对象。

## 1. 背景介绍

标准的工具实现通常使用简单的 `input: string` 参数。然而，实际应用中经常需要具有复杂、结构化参数的工具，例如：

- **多个字段**，类型各不相同
- **嵌套对象**，用于层级数据
- **数组**，用于批量操作
- **布尔标志**，用于选项控制
- **数值约束**（最小值、最大值）

LangGraphGo 现在通过 `ToolWithSchema` 接口支持具有自定义 JSON Schema 定义的工具。

## 2. 核心概念

### ToolWithSchema 接口

工具可以选择性地实现此接口以提供其参数 schema：

```go
type ToolWithSchema interface {
    Schema() map[string]any
}
```

如果工具实现了此接口，代理在调用 LLM 时将使用提供的 schema，使模型能够正确地结构化复杂的工具调用。

### Schema 格式

Schema 遵循 JSON Schema 格式：
- `type`: 始终为 "object"
- `properties`: 参数名称到其定义的映射
- `required`: 必需参数名称的数组
- 每个属性可以包含：`type`、`description`、`enum`、`minimum`、`maximum` 等

### 自动参数处理

框架会自动检测工具是否具有自定义 schema：
- **有 schema**: 直接将完整的 JSON 参数传递给工具
- **无 schema**: 提取 "input" 字段以保持向后兼容性

## 3. 示例工具

本示例包含三个工具，展示不同复杂程度：

### 3.1 酒店预订工具

演示：字符串、整数、布尔值、数组

**参数：**
- `check_in` (string): 入住日期（YYYY-MM-DD 格式）
- `check_out` (string): 退房日期（YYYY-MM-DD 格式）
- `guests` (integer): 客人数量（1-10人）
- `room_type` (string): 房间类型，带枚举值
- `breakfast` (boolean): 是否包含早餐
- `parking` (boolean): 是否需要停车位
- `view` (string): 房间景观偏好
- `max_price` (number): 每晚最高价格
- `special_requests` (array): 特殊要求列表

### 3.2 房贷计算器工具

演示：浮点数、嵌套对象

**参数：**
- `principal` (number): 贷款本金
- `interest_rate` (number): 年利率百分比
- `years` (integer): 贷款年限
- `down_payment` (number): 首付金额
- `property_tax` (number): 年房产税
- `insurance` (number): 年保险费
- `extra_payment` (object): 嵌套对象，包含：
  - `enabled` (boolean): 是否启用
  - `amount` (number): 额外还款金额
  - `frequency` (string): 还款频率
  - `start_year` (integer): 开始年份

### 3.3 批量操作工具

演示：带嵌套对象的数组

**参数：**
- `operation` (string): 操作类型
- `items` (array): 项目列表，每个项目包含：
  - `id` (string): 唯一标识符
  - `action` (string): 要执行的操作
  - `quantity` (integer): 数量
  - `price` (number): 单价
  - `metadata` (object): 额外的键值对
- `dry_run` (boolean): 是否仅验证不执行
- `priority` (string): 操作优先级

## 4. 代码要点

### 实现带 Schema 的工具

```go
type HotelBookingTool struct{}

func (t HotelBookingTool) Name() string {
    return "hotel_booking"
}

func (t HotelBookingTool) Description() string {
    return `预订酒店房间，支持多种选项配置。
    参数说明：
    - check_in: 入住日期（格式：YYYY-MM-DD）
    - guests: 客人数量（1-10人）
    ...`
}

func (t HotelBookingTool) Schema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "check_in": map[string]any{
                "type":        "string",
                "description": "入住日期，格式为YYYY-MM-DD",
                "format":      "date",
            },
            "guests": map[string]any{
                "type":        "integer",
                "description": "客人数量（1-10人）",
                "minimum":     1,
                "maximum":     10,
            },
            // ... 更多属性
        },
        "required": []string{"check_in", "check_out", "guests", "room_type"},
    }
}

func (t HotelBookingTool) Call(ctx context.Context, input string) (string, error) {
    var params HotelBookingParams
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", fmt.Errorf("输入参数无效: %w", err)
    }
    // ... 处理参数
}
```

### 创建代理

```go
agent, err := prebuilt.CreateAgentMap(llm, allTools, 10)
if err != nil {
    log.Fatal(err)
}
```

代理会自动使用工具的 schema（如果有的话）！

## 5. 运行示例

```bash
cd examples/complex_tools
export OPENAI_API_KEY=your_key
go run main.go tools.go
```

## 6. 预期输出

示例演示了：
1. **直接工具调用** - 显示每个工具的结构化结果
2. **代理交互** - LLM 能够正确地使用复杂参数结构化工具调用

**酒店预订的示例输出：**
```json
{
  "booking_id": "HTL-1234567890",
  "check_in": "2026-02-01",
  "check_out": "2026-02-05",
  "nights": 4,
  "guests": 2,
  "room_type": "deluxe",
  "price_per_night": 270.0,
  "total_price": 1080.0,
  "special_requests": ["高层", "安静房间"]
}
```

## 7. 使用场景

复杂工具 schema 适用于：

- **电商**: 带过滤器的产品搜索（价格区间、类别、评分）
- **旅游**: 带日期、人数、偏好的预订
- **金融**: 带多个参数的贷款计算器
- **数据处理**: 带数组输入的批量操作
- **API 封装**: 镜像复杂 API 签名的工具

## 8. 向后兼容性

未实现 `ToolWithSchema` 的工具继续使用默认的简单 schema（`input: string`），确保与现有代码完全向后兼容。

## 9. 相关文档

- [ReAct Agent 示例](../react_agent/README.md) - 基础 ReAct 代理使用
- [工具定义](../../docs/REACTAGENT.md) - 代理和工具的详细文档
- [预构建代理](../../docs/AGENT.md) - 所有预构建代理的说明
