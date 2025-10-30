package core

import (
	"context"
	"testing"
	"time"
)

// TestRoom_JoinAndLeave 测试加入和离开 Room
func TestRoom_JoinAndLeave(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建 Agent
	config1 := createTestConfig("agent-1")
	_, err := pool.Create(ctx, config1)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// 加入 Room
	err = room.Join("alice", "agent-1")
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}

	// 验证成员
	if !room.IsMember("alice") {
		t.Error("Alice should be a member")
	}

	if room.GetMemberCount() != 1 {
		t.Errorf("Expected 1 member, got %d", room.GetMemberCount())
	}

	// 离开 Room
	err = room.Leave("alice")
	if err != nil {
		t.Fatalf("Failed to leave room: %v", err)
	}

	// 验证已离开
	if room.IsMember("alice") {
		t.Error("Alice should not be a member")
	}

	if room.GetMemberCount() != 0 {
		t.Errorf("Expected 0 members, got %d", room.GetMemberCount())
	}
}

// TestRoom_JoinDuplicate 测试重复加入
func TestRoom_JoinDuplicate(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建 Agent
	config := createTestConfig("agent-1")
	_, err := pool.Create(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// 第一次加入
	err = room.Join("alice", "agent-1")
	if err != nil {
		t.Fatalf("First join failed: %v", err)
	}

	// 第二次加入应该失败
	err = room.Join("alice", "agent-1")
	if err == nil {
		t.Error("Expected error when joining with duplicate name")
	}
}

// TestRoom_JoinNonexistentAgent 测试加入不存在的 Agent
func TestRoom_JoinNonexistentAgent(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)

	// 尝试加入不存在的 Agent
	err := room.Join("alice", "nonexistent-agent")
	if err == nil {
		t.Error("Expected error when joining with nonexistent agent")
	}
}

// TestRoom_GetMembers 测试获取成员列表
func TestRoom_GetMembers(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建多个 Agent
	agents := []struct {
		name    string
		agentID string
	}{
		{"alice", "agent-1"},
		{"bob", "agent-2"},
		{"charlie", "agent-3"},
	}

	for _, a := range agents {
		config := createTestConfig(a.agentID)
		_, err := pool.Create(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create agent %s: %v", a.agentID, err)
		}

		err = room.Join(a.name, a.agentID)
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}
	}

	// 获取成员
	members := room.GetMembers()
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// 验证成员信息
	memberMap := make(map[string]string)
	for _, m := range members {
		memberMap[m.Name] = m.AgentID
	}

	for _, a := range agents {
		if agentID, exists := memberMap[a.name]; !exists {
			t.Errorf("Member %s not found", a.name)
		} else if agentID != a.agentID {
			t.Errorf("Expected agent ID %s for %s, got %s", a.agentID, a.name, agentID)
		}
	}
}

// TestRoom_ExtractMentions 测试提取 @mentions
func TestRoom_ExtractMentions(t *testing.T) {
	room := NewRoom(nil) // 不需要 pool 来测试这个功能

	tests := []struct {
		text     string
		expected []string
	}{
		{
			text:     "Hello @alice and @bob",
			expected: []string{"alice", "bob"},
		},
		{
			text:     "@alice can you help with this?",
			expected: []string{"alice"},
		},
		{
			text:     "No mentions here",
			expected: []string{},
		},
		{
			text:     "@alice @alice @bob", // 重复的 mention
			expected: []string{"alice", "bob"},
		},
	}

	for _, tt := range tests {
		mentions := room.extractMentions(tt.text)

		if len(mentions) != len(tt.expected) {
			t.Errorf("For text %q, expected %d mentions, got %d",
				tt.text, len(tt.expected), len(mentions))
			continue
		}

		// 检查所有预期的 mention 都存在
		mentionMap := make(map[string]bool)
		for _, m := range mentions {
			mentionMap[m] = true
		}

		for _, exp := range tt.expected {
			if !mentionMap[exp] {
				t.Errorf("For text %q, expected mention %q not found", tt.text, exp)
			}
		}
	}
}

// TestRoom_Say 测试发送消息
func TestRoom_Say(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建 Agent
	config1 := createTestConfig("agent-1")
	_, err := pool.Create(ctx, config1)
	if err != nil {
		t.Fatalf("Failed to create agent-1: %v", err)
	}

	config2 := createTestConfig("agent-2")
	_, err = pool.Create(ctx, config2)
	if err != nil {
		t.Fatalf("Failed to create agent-2: %v", err)
	}

	// 加入 Room
	room.Join("alice", "agent-1")
	room.Join("bob", "agent-2")

	// 发送消息
	err = room.Say(ctx, "alice", "Hello @bob!")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待消息发送完成
	time.Sleep(100 * time.Millisecond)

	// 检查历史记录
	history := room.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message in history, got %d", len(history))
	}

	if len(history) > 0 {
		msg := history[0]
		if msg.From != "alice" {
			t.Errorf("Expected message from alice, got %s", msg.From)
		}
		if msg.Text != "Hello @bob!" {
			t.Errorf("Expected text 'Hello @bob!', got %s", msg.Text)
		}
	}
}

// TestRoom_Broadcast 测试广播消息
func TestRoom_Broadcast(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建多个 Agent
	for i := 1; i <= 3; i++ {
		config := createTestConfig("agent-" + string(rune('0'+i)))
		_, err := pool.Create(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create agent-%d: %v", i, err)
		}

		name := string(rune('a' + i - 1)) // a, b, c
		room.Join(name, "agent-"+string(rune('0'+i)))
	}

	// 广播消息
	err := room.Broadcast(ctx, "System announcement")
	if err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// 等待消息发送
	time.Sleep(100 * time.Millisecond)

	// 检查历史
	history := room.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message in history, got %d", len(history))
	}

	if len(history) > 0 {
		msg := history[0]
		if msg.From != "system" {
			t.Errorf("Expected message from system, got %s", msg.From)
		}
	}
}

// TestRoom_SendTo 测试点对点消息
func TestRoom_SendTo(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建 Agent
	config1 := createTestConfig("agent-1")
	_, err := pool.Create(ctx, config1)
	if err != nil {
		t.Fatalf("Failed to create agent-1: %v", err)
	}

	config2 := createTestConfig("agent-2")
	_, err = pool.Create(ctx, config2)
	if err != nil {
		t.Fatalf("Failed to create agent-2: %v", err)
	}

	// 加入 Room
	room.Join("alice", "agent-1")
	room.Join("bob", "agent-2")

	// 发送点对点消息
	err = room.SendTo(ctx, "alice", "bob", "Private message")
	if err != nil {
		t.Fatalf("Failed to send private message: %v", err)
	}

	// 检查历史
	history := room.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message in history, got %d", len(history))
	}

	if len(history) > 0 {
		msg := history[0]
		if msg.From != "alice" {
			t.Errorf("Expected message from alice, got %s", msg.From)
		}
		if len(msg.To) != 1 || msg.To[0] != "bob" {
			t.Errorf("Expected message to bob, got %v", msg.To)
		}
	}
}

// TestRoom_ClearHistory 测试清空历史
func TestRoom_ClearHistory(t *testing.T) {
	deps := createTestDeps(t)
	pool := NewPool(&PoolOptions{
		Dependencies: deps,
		MaxAgents:    10,
	})
	defer pool.Shutdown()

	room := NewRoom(pool)
	ctx := context.Background()

	// 创建 Agent 并发送消息
	config := createTestConfig("agent-1")
	_, err := pool.Create(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	room.Join("alice", "agent-1")
	room.Broadcast(ctx, "Test message")

	time.Sleep(50 * time.Millisecond)

	// 验证有历史
	history := room.GetHistory()
	if len(history) == 0 {
		t.Error("Expected messages in history")
	}

	// 清空历史
	room.ClearHistory()

	// 验证已清空
	history = room.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(history))
	}
}
