package mcp

import (
	"context"
	"testing"

	"github.com/wordflowlab/agentsdk/pkg/sandbox/cloud"
	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// TestMCPToolAdapter_Interface 测试 MCPToolAdapter 实现 Tool 接口
func TestMCPToolAdapter_Interface(t *testing.T) {
	// 创建模拟的 MCP 客户端
	client := cloud.NewMCPClient(&cloud.MCPClientConfig{
		Endpoint: "http://mock-mcp-server",
	})

	// 创建适配器
	adapter := NewMCPToolAdapter(&MCPToolAdapterConfig{
		Client:      client,
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type":        "string",
					"description": "Test input",
				},
			},
		},
		Prompt: "Use this tool for testing",
	})

	// 验证接口实现
	var _ tools.Tool = adapter

	// 测试 Name
	if adapter.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", adapter.Name())
	}

	// 测试 Description
	if adapter.Description() != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", adapter.Description())
	}

	// 测试 Prompt
	if adapter.Prompt() != "Use this tool for testing" {
		t.Errorf("Expected prompt 'Use this tool for testing', got '%s'", adapter.Prompt())
	}

	// 测试 InputSchema
	schema := adapter.InputSchema()
	if schema == nil {
		t.Fatal("InputSchema returned nil")
	}

	if schema["type"] != "object" {
		t.Errorf("Expected schema type 'object', got '%v'", schema["type"])
	}
}

// TestToolFactory 测试工具工厂函数
func TestToolFactory(t *testing.T) {
	client := cloud.NewMCPClient(&cloud.MCPClientConfig{
		Endpoint: "http://mock-mcp-server",
	})

	mcpTool := cloud.MCPTool{
		Name:        "calculator",
		Description: "A simple calculator",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type": "string",
				},
				"a": map[string]interface{}{
					"type": "number",
				},
				"b": map[string]interface{}{
					"type": "number",
				},
			},
		},
	}

	// 创建工厂
	factory := ToolFactory(client, mcpTool)

	// 使用工厂创建工具
	tool, err := factory(map[string]interface{}{
		"prompt": "Custom prompt for calculator",
	})

	if err != nil {
		t.Fatalf("Factory failed: %v", err)
	}

	if tool.Name() != "calculator" {
		t.Errorf("Expected name 'calculator', got '%s'", tool.Name())
	}

	if tool.Prompt() != "Custom prompt for calculator" {
		t.Errorf("Expected custom prompt, got '%s'", tool.Prompt())
	}
}

// TestMCPToolAdapter_Execute 测试工具执行 (需要模拟 MCP Server)
func TestMCPToolAdapter_Execute(t *testing.T) {
	t.Skip("Skipping Execute test - requires mock MCP server")

	client := cloud.NewMCPClient(&cloud.MCPClientConfig{
		Endpoint: "http://localhost:8080/mcp",
	})

	adapter := NewMCPToolAdapter(&MCPToolAdapterConfig{
		Client:      client,
		Name:        "echo",
		Description: "Echo tool",
		InputSchema: map[string]interface{}{},
	})

	ctx := context.Background()
	result, err := adapter.Execute(ctx, map[string]interface{}{
		"message": "hello",
	}, &tools.ToolContext{
		AgentID: "test-agent",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}
}
