package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/wordflowlab/agentsdk/pkg/sandbox/cloud"
	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// MCPServer MCP Server 连接管理器
type MCPServer struct {
	mu       sync.RWMutex
	client   *cloud.MCPClient
	serverID string
	tools    []cloud.MCPTool
	registry *tools.Registry
}

// MCPServerConfig MCP Server 配置
type MCPServerConfig struct {
	ServerID        string
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
}

// NewMCPServer 创建 MCP Server 连接
func NewMCPServer(config *MCPServerConfig, registry *tools.Registry) (*MCPServer, error) {
	if config.ServerID == "" {
		return nil, fmt.Errorf("server_id is required")
	}

	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// 创建 MCP 客户端
	client := cloud.NewMCPClient(&cloud.MCPClientConfig{
		Endpoint:        config.Endpoint,
		AccessKeyID:     config.AccessKeyID,
		AccessKeySecret: config.AccessKeySecret,
		SecurityToken:   config.SecurityToken,
	})

	return &MCPServer{
		client:   client,
		serverID: config.ServerID,
		tools:    make([]cloud.MCPTool, 0),
		registry: registry,
	}, nil
}

// Connect 连接到 MCP Server 并发现工具
func (s *MCPServer) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 列出服务端提供的工具
	mcpTools, err := s.client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("list mcp tools: %w", err)
	}

	s.tools = mcpTools
	return nil
}

// RegisterTools 将 MCP 工具注册到 AgentSDK Registry
func (s *MCPServer) RegisterTools() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.tools) == 0 {
		return fmt.Errorf("no tools available, call Connect() first")
	}

	// 为每个 MCP 工具创建工厂并注册
	for _, mcpTool := range s.tools {
		// 使用 server_id 作为前缀避免工具名冲突
		toolName := fmt.Sprintf("%s:%s", s.serverID, mcpTool.Name)

		// 创建工具工厂
		factory := ToolFactory(s.client, mcpTool)

		// 注册到 Registry
		s.registry.Register(toolName, factory)
	}

	return nil
}

// ListTools 返回已发现的工具列表
func (s *MCPServer) ListTools() []cloud.MCPTool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本
	tools := make([]cloud.MCPTool, len(s.tools))
	copy(tools, s.tools)
	return tools
}

// GetToolCount 获取工具数量
func (s *MCPServer) GetToolCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tools)
}

// GetServerID 获取 Server ID
func (s *MCPServer) GetServerID() string {
	return s.serverID
}

// GetClient 获取底层 MCP 客户端
func (s *MCPServer) GetClient() *cloud.MCPClient {
	return s.client
}
