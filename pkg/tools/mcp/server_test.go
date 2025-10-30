package mcp

import (
	"context"
	"testing"

	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// TestMCPServer_Creation 测试 MCP Server 创建
func TestMCPServer_Creation(t *testing.T) {
	registry := tools.NewRegistry()

	server, err := NewMCPServer(&MCPServerConfig{
		ServerID: "test-server",
		Endpoint: "http://localhost:8080/mcp",
	}, registry)

	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}

	if server.GetServerID() != "test-server" {
		t.Errorf("Expected server ID 'test-server', got '%s'", server.GetServerID())
	}

	if server.GetToolCount() != 0 {
		t.Errorf("Expected 0 tools before connect, got %d", server.GetToolCount())
	}
}

// TestMCPServer_InvalidConfig 测试无效配置
func TestMCPServer_InvalidConfig(t *testing.T) {
	registry := tools.NewRegistry()

	// 测试缺少 ServerID
	_, err := NewMCPServer(&MCPServerConfig{
		Endpoint: "http://localhost:8080/mcp",
	}, registry)

	if err == nil {
		t.Error("Expected error for missing server_id")
	}

	// 测试缺少 Endpoint
	_, err = NewMCPServer(&MCPServerConfig{
		ServerID: "test",
	}, registry)

	if err == nil {
		t.Error("Expected error for missing endpoint")
	}
}

// TestMCPServer_Connect 测试连接 (需要模拟 MCP Server)
func TestMCPServer_Connect(t *testing.T) {
	t.Skip("Skipping Connect test - requires mock MCP server")

	registry := tools.NewRegistry()

	server, err := NewMCPServer(&MCPServerConfig{
		ServerID: "test-server",
		Endpoint: "http://localhost:8080/mcp",
	}, registry)

	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	if err := server.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if server.GetToolCount() == 0 {
		t.Error("Expected tools after connect")
	}
}

// TestMCPServer_RegisterTools 测试工具注册
func TestMCPServer_RegisterTools(t *testing.T) {
	registry := tools.NewRegistry()

	server, err := NewMCPServer(&MCPServerConfig{
		ServerID: "test-server",
		Endpoint: "http://localhost:8080/mcp",
	}, registry)

	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// 在没有连接的情况下注册应该失败
	err = server.RegisterTools()
	if err == nil {
		t.Error("Expected error when registering tools before connect")
	}
}
