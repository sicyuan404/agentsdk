package mcp

import (
	"context"
	"testing"

	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// TestMCPManager_Creation 测试 Manager 创建
func TestMCPManager_Creation(t *testing.T) {
	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	if manager.GetServerCount() != 0 {
		t.Errorf("Expected 0 servers, got %d", manager.GetServerCount())
	}
}

// TestMCPManager_AddServer 测试添加 Server
func TestMCPManager_AddServer(t *testing.T) {
	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	server, err := manager.AddServer(&MCPServerConfig{
		ServerID: "server-1",
		Endpoint: "http://localhost:8080/mcp",
	})

	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	if server == nil {
		t.Fatal("Server is nil")
	}

	if manager.GetServerCount() != 1 {
		t.Errorf("Expected 1 server, got %d", manager.GetServerCount())
	}

	// 测试重复添加
	_, err = manager.AddServer(&MCPServerConfig{
		ServerID: "server-1",
		Endpoint: "http://localhost:8081/mcp",
	})

	if err == nil {
		t.Error("Expected error when adding duplicate server")
	}
}

// TestMCPManager_GetServer 测试获取 Server
func TestMCPManager_GetServer(t *testing.T) {
	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	_, err := manager.AddServer(&MCPServerConfig{
		ServerID: "server-1",
		Endpoint: "http://localhost:8080/mcp",
	})

	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// 获取存在的 Server
	server, exists := manager.GetServer("server-1")
	if !exists {
		t.Error("Server should exist")
	}

	if server.GetServerID() != "server-1" {
		t.Errorf("Expected server ID 'server-1', got '%s'", server.GetServerID())
	}

	// 获取不存在的 Server
	_, exists = manager.GetServer("nonexistent")
	if exists {
		t.Error("Nonexistent server should not exist")
	}
}

// TestMCPManager_ListServers 测试列出 Server
func TestMCPManager_ListServers(t *testing.T) {
	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	// 添加多个 Server
	servers := []string{"server-1", "server-2", "server-3"}
	for _, id := range servers {
		_, err := manager.AddServer(&MCPServerConfig{
			ServerID: id,
			Endpoint: "http://localhost:8080/mcp",
		})
		if err != nil {
			t.Fatalf("Failed to add server %s: %v", id, err)
		}
	}

	// 列出所有 Server
	list := manager.ListServers()
	if len(list) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(list))
	}

	// 验证所有 Server ID 都在列表中
	serverMap := make(map[string]bool)
	for _, id := range list {
		serverMap[id] = true
	}

	for _, id := range servers {
		if !serverMap[id] {
			t.Errorf("Server %s not found in list", id)
		}
	}
}

// TestMCPManager_RemoveServer 测试移除 Server
func TestMCPManager_RemoveServer(t *testing.T) {
	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	_, err := manager.AddServer(&MCPServerConfig{
		ServerID: "server-1",
		Endpoint: "http://localhost:8080/mcp",
	})

	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// 移除 Server
	err = manager.RemoveServer("server-1")
	if err != nil {
		t.Fatalf("Failed to remove server: %v", err)
	}

	if manager.GetServerCount() != 0 {
		t.Errorf("Expected 0 servers after removal, got %d", manager.GetServerCount())
	}

	// 移除不存在的 Server
	err = manager.RemoveServer("nonexistent")
	if err == nil {
		t.Error("Expected error when removing nonexistent server")
	}
}

// TestMCPManager_ConnectServer 测试连接 Server (需要模拟 MCP Server)
func TestMCPManager_ConnectServer(t *testing.T) {
	t.Skip("Skipping ConnectServer test - requires mock MCP server")

	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	_, err := manager.AddServer(&MCPServerConfig{
		ServerID: "server-1",
		Endpoint: "http://localhost:8080/mcp",
	})

	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	ctx := context.Background()
	err = manager.ConnectServer(ctx, "server-1")
	if err != nil {
		t.Fatalf("Failed to connect server: %v", err)
	}

	// 验证工具已注册
	if manager.GetTotalToolCount() == 0 {
		t.Error("Expected tools after connect")
	}
}

// TestMCPManager_ConnectAll 测试连接所有 Server
func TestMCPManager_ConnectAll(t *testing.T) {
	t.Skip("Skipping ConnectAll test - requires mock MCP server")

	registry := tools.NewRegistry()
	manager := NewMCPManager(registry)

	// 添加多个 Server
	for i := 1; i <= 3; i++ {
		_, err := manager.AddServer(&MCPServerConfig{
			ServerID: "server-" + string(rune('0'+i)),
			Endpoint: "http://localhost:8080/mcp",
		})
		if err != nil {
			t.Fatalf("Failed to add server: %v", err)
		}
	}

	ctx := context.Background()
	err := manager.ConnectAll(ctx)
	if err != nil {
		t.Fatalf("Failed to connect all servers: %v", err)
	}
}
