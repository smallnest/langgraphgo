package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/langgraphgo/llms"
	openai "github.com/smallnest/langgraphgo/llms/nativeopenai"
	"github.com/smallnest/langgraphgo/tool"

	"github.com/smallnest/langgraphgo/prebuilt"
)

func main() {
	// 检查API密钥
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("未设置OPENAI_API_KEY环境变量")
	}

	// 初始化LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// 创建复杂工具
	hotelTool := &SimpleToolWrapper{
		name:        "hotel_booking",
		description: HotelBookingTool{}.Description(),
		schema:      HotelBookingTool{}.Schema(),
		handler:     HotelBookingTool{}.Call,
	}

	mortgageTool := &SimpleToolWrapper{
		name:        "mortgage_calculator",
		description: MortgageCalculatorTool{}.Description(),
		schema:      MortgageCalculatorTool{}.Schema(),
		handler:     MortgageCalculatorTool{}.Call,
	}

	batchTool := &SimpleToolWrapper{
		name:        "batch_operation",
		description: BatchOperationTool{}.Description(),
		schema:      BatchOperationTool{}.Schema(),
		handler:     BatchOperationTool{}.Call,
	}

	allTools := []tool.Tool{hotelTool, mortgageTool, batchTool}

	// 打印工具schema用于演示
	fmt.Println("=== 复杂工具示例 ===")
	fmt.Println("\n可用工具及其参数schema：")
	fmt.Println()

	for _, tool := range allTools {
		fmt.Printf("工具名称: %s\n", tool.Name())
		fmt.Printf("描述: %s\n", tool.Description())

		if st, ok := tool.(prebuilt.ToolWithSchema); ok {
			if schema := st.Schema(); schema != nil {
				schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
				fmt.Printf("Schema:\n%s\n", string(schemaJSON))
			}
		}
		fmt.Println()
	}

	// 直接演示工具调用
	fmt.Println("=== 直接工具调用示例 ===")
	fmt.Println()

	ctx := context.Background()

	// 示例1：酒店预订
	fmt.Println("1. 酒店预订示例：")
	hotelInput := HotelBookingParams{
		CheckIn:         "2026-02-01",
		CheckOut:        "2026-02-05",
		Guests:          2,
		RoomType:        "deluxe",
		Breakfast:       true,
		Parking:         true,
		View:            "ocean",
		MaxPrice:        250.0,
		SpecialRequests: []string{"高层", "安静房间"},
	}
	hotelInputJSON, _ := json.Marshal(hotelInput)
	result, err := hotelTool.Call(ctx, string(hotelInputJSON))
	if err != nil {
		log.Printf("酒店预订错误: %v", err)
	} else {
		fmt.Printf("结果:\n%s\n\n", result)
	}

	// 示例2：房贷计算器
	fmt.Println("2. 房贷计算器示例：")
	mortgageInput := MortgageCalculationParams{
		Principal:    450000.0,
		InterestRate: 6.5,
		Years:        30,
		DownPayment:  90000.0,
		PropertyTax:  3600.0,
		Insurance:    1200.0,
		ExtraPayment: ExtraPaymentInfo{
			Enabled:   true,
			Amount:    200.0,
			Frequency: "monthly",
			StartYear: 1,
		},
	}
	mortgageInputJSON, _ := json.Marshal(mortgageInput)
	result, err = mortgageTool.Call(ctx, string(mortgageInputJSON))
	if err != nil {
		log.Printf("房贷计算错误: %v", err)
	} else {
		fmt.Printf("结果:\n%s\n\n", result)
	}

	// 示例3：批量操作
	fmt.Println("3. 批量操作示例：")
	batchInput := BatchOperationParams{
		Operation: "process",
		Items: []BatchOperationItem{
			{
				ID:       "ITEM-001",
				Action:   "create",
				Quantity: 100,
				Price:    15.99,
				Metadata: map[string]any{"category": "电子产品", "brand": "TechCo"},
			},
			{
				ID:       "ITEM-002",
				Action:   "update",
				Quantity: 50,
				Price:    29.99,
				Metadata: map[string]any{"category": "配件", "brand": "AccBrand"},
			},
			{
				ID:       "ITEM-003",
				Action:   "create",
				Quantity: 200,
				Price:    9.50,
				Metadata: map[string]any{"category": "用品", "urgent": true},
			},
		},
		DryRun:   false,
		Priority: "high",
	}
	batchInputJSON, _ := json.Marshal(batchInput)
	result, err = batchTool.Call(ctx, string(batchInputJSON))
	if err != nil {
		log.Printf("批量操作错误: %v", err)
	} else {
		fmt.Printf("结果:\n%s\n\n", result)
	}

	// 示例4：使用代理调用工具
	fmt.Println("=== 使用ReAct代理调用工具（支持复杂schema） ===")
	fmt.Println()

	agent, err := prebuilt.CreateAgentMap(llm, allTools, 10)
	if err != nil {
		log.Fatal(err)
	}

	// 代理测试查询
	queries := []string{
		"我想预订2026-02-10至2026-02-15的酒店房间。需要2位客人，豪华间海景房，含早餐，每晚预算300美元。",
		"计算一笔40万美元、30年期、利率6.5%的房贷月供，首付8万美元。包含3000美元年房产税和1200美元保险。",
		"处理这三个项目：创建ITEM-A（100件，每件10美元），创建ITEM-B（50件，每件25美元），更新ITEM-C（75件，每件15美元）。",
	}

	for i, query := range queries {
		fmt.Printf("查询 %d: %s\n", i+1, query)
		fmt.Println("---")

		resp, err := agent.Invoke(ctx, map[string]any{
			"messages": []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeHuman, query),
			},
		})
		if err != nil {
			log.Printf("代理错误: %v\n", err)
		} else {
			if msgs, ok := resp["messages"].([]llms.MessageContent); ok && len(msgs) > 0 {
				for _, msg := range msgs {
					if msg.Role == llms.ChatMessageTypeAI {
						for _, part := range msg.Parts {
							if textPart, ok := part.(llms.TextContent); ok {
								fmt.Printf("代理回复: %s\n\n", textPart.Text)
							}
						}
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("=== 示例结束 ===")
}

// SimpleToolWrapper 包装复杂工具以实现tool.Tool接口
type SimpleToolWrapper struct {
	name        string
	description string
	schema      map[string]any
	handler     func(ctx context.Context, input string) (string, error)
}

func (w *SimpleToolWrapper) Name() string {
	return w.name
}

func (w *SimpleToolWrapper) Description() string {
	return w.description
}

func (w *SimpleToolWrapper) Call(ctx context.Context, input string) (string, error) {
	return w.handler(ctx, input)
}

func (w *SimpleToolWrapper) Schema() map[string]any {
	return w.schema
}
