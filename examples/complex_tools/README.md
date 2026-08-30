# Complex Tools Example

This example demonstrates how to create tools with **complex parameter schemas** in LangGraphGo, supporting multiple parameter types including strings, integers, floats, booleans, arrays, and nested objects.

## 1. Background

Standard tool implementations often use a simple `input: string` parameter. However, real-world applications frequently require tools with complex, structured parameters such as:

- **Multiple fields** with different types
- **Nested objects** for hierarchical data
- **Arrays** for batch operations
- **Boolean flags** for options
- **Numeric constraints** (minimum, maximum values)

LangGraphGo now supports tools with custom JSON Schema definitions through the `ToolWithSchema` interface.

## 2. Key Concepts

### ToolWithSchema Interface

Tools can optionally implement this interface to provide their parameter schema:

```go
type ToolWithSchema interface {
    Schema() map[string]any
}
```

If a tool implements this interface, the agent will use the provided schema when calling the LLM, allowing the model to properly structure complex tool calls.

### Schema Format

The schema follows JSON Schema format:
- `type`: Always "object"
- `properties`: Map of parameter names to their definitions
- `required`: Array of required parameter names
- Each property can have: `type`, `description`, `enum`, `minimum`, `maximum`, etc.

### Automatic Parameter Handling

The framework automatically detects whether a tool has a custom schema:
- **With schema**: Passes the complete JSON arguments directly to the tool
- **Without schema**: Extracts the "input" field for backward compatibility

## 3. Example Tools

This example includes three tools demonstrating different complexity levels:

### 3.1 Hotel Booking Tool

Demonstrates: strings, integers, booleans, arrays

**Parameters:**
- `check_in` (string): Check-in date (YYYY-MM-DD format)
- `check_out` (string): Check-out date (YYYY-MM-DD format)
- `guests` (integer): Number of guests (1-10)
- `room_type` (string): Room type with enum values
- `breakfast` (boolean): Whether breakfast is included
- `parking` (boolean): Whether parking is needed
- `view` (string): Room view preference
- `max_price` (number): Maximum price per night
- `special_requests` (array): List of special requests

### 3.2 Mortgage Calculator Tool

Demonstrates: floats, nested objects

**Parameters:**
- `principal` (number): Loan amount
- `interest_rate` (number): Annual interest rate percentage
- `years` (integer): Loan term in years
- `down_payment` (number): Down payment amount
- `property_tax` (number): Annual property tax
- `insurance` (number): Annual insurance
- `extra_payment` (object): Nested object with:
  - `enabled` (boolean)
  - `amount` (number)
  - `frequency` (string)
  - `start_year` (integer)

### 3.3 Batch Operation Tool

Demonstrates: arrays with nested objects

**Parameters:**
- `operation` (string): Operation type
- `items` (array): List of items, each containing:
  - `id` (string): Unique identifier
  - `action` (string): Action to perform
  - `quantity` (integer): Item quantity
  - `price` (number): Unit price
  - `metadata` (object): Additional key-value pairs
- `dry_run` (boolean): Whether to only validate
- `priority` (string): Operation priority

## 4. Code Highlights

### Implementing a Tool with Schema

```go
type HotelBookingTool struct{}

func (t HotelBookingTool) Name() string {
    return "hotel_booking"
}

func (t HotelBookingTool) Description() string {
    return `Book a hotel room with various options.
    Parameters:
    - check_in: Check-in date (format: YYYY-MM-DD)
    - guests: Number of guests (1-10)
    ...`
}

func (t HotelBookingTool) Schema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "check_in": map[string]any{
                "type":        "string",
                "description": "Check-in date in YYYY-MM-DD format",
                "format":      "date",
            },
            "guests": map[string]any{
                "type":        "integer",
                "description": "Number of guests (1-10)",
                "minimum":     1,
                "maximum":     10,
            },
            // ... more properties
        },
        "required": []string{"check_in", "check_out", "guests", "room_type"},
    }
}

func (t HotelBookingTool) Call(ctx context.Context, input string) (string, error) {
    var params HotelBookingParams
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", fmt.Errorf("invalid input: %w", err)
    }
    // ... process parameters
}
```

### Creating the Agent

```go
agent, err := prebuilt.CreateAgentMap(llm, allTools, 10)
if err != nil {
    log.Fatal(err)
}
```

The agent automatically uses the tool's schema when available!

## 5. Running the Example

```bash
cd examples/complex_tools
export OPENAI_API_KEY=your_key
go run main.go tools.go
```

## 6. Expected Output

The example demonstrates:
1. **Direct tool calls** - Shows structured results from each tool
2. **Agent interaction** - The LLM properly structures tool calls with complex parameters

**Sample output for hotel booking:**
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
  "special_requests": ["high floor", "quiet room"]
}
```

## 7. Use Cases

Complex tool schemas are ideal for:

- **E-commerce**: Product search with filters (price range, categories, ratings)
- **Travel**: Booking with dates, number of travelers, preferences
- **Finance**: Loan calculators with multiple parameters
- **Data Processing**: Batch operations with array inputs
- **API Wrappers**: Tools that mirror complex API signatures

## 8. Backward Compatibility

Tools that don't implement `ToolWithSchema` continue to work with the default simple schema (`input: string`), ensuring full backward compatibility with existing code.
