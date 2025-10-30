package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wordflowlab/agentsdk/pkg/sandbox/cloud"
	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// MCPToolAdapter 将 MCP 工具适配为 AgentSDK Tool 接口
type MCPToolAdapter struct {
	client      *cloud.MCPClient
	name        string
	description string
	inputSchema map[string]interface{}
	prompt      string
}

// MCPToolAdapterConfig MCP 工具适配器配置
type MCPToolAdapterConfig struct {
	Client      *cloud.MCPClient
	Name        string
	Description string
	InputSchema map[string]interface{}
	Prompt      string
}

// NewMCPToolAdapter 创建 MCP 工具适配器
func NewMCPToolAdapter(config *MCPToolAdapterConfig) *MCPToolAdapter {
	return &MCPToolAdapter{
		client:      config.Client,
		name:        config.Name,
		description: config.Description,
		inputSchema: config.InputSchema,
		prompt:      config.Prompt,
	}
}

// Name 返回工具名称
func (m *MCPToolAdapter) Name() string {
	return m.name
}

// Description 返回工具描述
func (m *MCPToolAdapter) Description() string {
	return m.description
}

// InputSchema 返回输入 JSON Schema
func (m *MCPToolAdapter) InputSchema() map[string]interface{} {
	return m.inputSchema
}

// Prompt 返回工具使用说明
func (m *MCPToolAdapter) Prompt() string {
	return m.prompt
}

// Execute 执行 MCP 工具调用
func (m *MCPToolAdapter) Execute(ctx context.Context, input map[string]interface{}, tc *tools.ToolContext) (interface{}, error) {
	// 调用远程 MCP 工具
	result, err := m.client.CallTool(ctx, m.name, input)
	if err != nil {
		return nil, fmt.Errorf("mcp tool call failed: %w", err)
	}

	// 解析结果
	var output interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		// 如果无法解析为通用接口,返回原始 JSON
		return string(result), nil
	}

	return output, nil
}

// ToolFactory 创建 MCP 工具工厂函数
func ToolFactory(mcpClient *cloud.MCPClient, mcpTool cloud.MCPTool) tools.ToolFactory {
	return func(config map[string]interface{}) (tools.Tool, error) {
		// 从配置中提取自定义 prompt (可选)
		prompt := ""
		if p, ok := config["prompt"].(string); ok {
			prompt = p
		}

		return NewMCPToolAdapter(&MCPToolAdapterConfig{
			Client:      mcpClient,
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			InputSchema: mcpTool.InputSchema,
			Prompt:      prompt,
		}), nil
	}
}
