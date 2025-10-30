# Agent Pool 示例

演示如何使用 `core.Pool` 管理多个 Agent 的生命周期。

## 功能展示

### 1. 创建 Agent 池

```go
pool := core.NewPool(&core.PoolOptions{
    Dependencies: deps,
    MaxAgents:    10, // 最大容量
})
defer pool.Shutdown()
```

### 2. 创建和管理 Agent

```go
// 创建 Agent
config := createAgentConfig("agent-1")
ag, err := pool.Create(ctx, config)

// 获取 Agent
ag, exists := pool.Get("agent-1")

// 列出所有 Agent
allAgents := pool.List("")

// 列出特定前缀的 Agent
userAgents := pool.List("user-")
```

### 3. Agent 状态查询

```go
status, err := pool.Status("agent-1")
fmt.Printf("Agent ID: %s\n", status.AgentID)
fmt.Printf("状态: %s\n", status.State)
fmt.Printf("步骤数: %d\n", status.StepCount)
```

### 4. 移除 Agent

```go
// 从池中移除 (不删除存储)
err := pool.Remove("agent-1")

// 删除 Agent (包括存储)
err := pool.Delete(ctx, "agent-1")
```

### 5. 遍历 Agent

```go
pool.ForEach(func(agentID string, ag *agent.Agent) error {
    status := ag.Status()
    fmt.Printf("Agent: %s, 状态: %s\n", agentID, status.State)
    return nil
})
```

### 6. 并发安全

Pool 使用 `sync.RWMutex` 保证线程安全,支持并发创建、获取、移除操作。

```go
// 并发创建 Agent
for i := 0; i < 10; i++ {
    go func(idx int) {
        config := createAgentConfig(fmt.Sprintf("agent-%d", idx))
        pool.Create(ctx, config)
    }(i)
}
```

## 运行示例

### 设置环境变量

```bash
export ANTHROPIC_API_KEY="your-api-key"
```

### 运行

```bash
go run examples/pool/main.go
```

## 输出示例

```
=== Agent SDK - Agent Pool 示例 ===

Agent 池已创建,最大容量: 10

--- 示例 1: 创建多个 Agent ---
✓ Agent 创建成功: agent-alice
✓ Agent 创建成功: agent-bob
✓ Agent 创建成功: agent-charlie

当前池大小: 3

--- 示例 2: 列出和获取 Agent ---
所有 Agent: [agent-alice agent-bob agent-charlie]
agent- 前缀的 Agent: [agent-alice agent-bob agent-charlie]
✓ 成功获取 Agent: agent-alice

--- 示例 3: Agent 状态查询 ---
Agent ID: agent-bob
状态: idle
步骤数: 0
消息数: 0
创建时间: 2025-10-30T12:34:56Z

--- 示例 4: 移除 Agent ---
移除前池大小: 3
✓ Agent 已移除
移除后池大小: 2

--- 示例 5: 遍历所有 Agent ---
遍历所有 Agent:
  - agent-alice (状态: idle, 步骤: 0)
  - agent-bob (状态: idle, 步骤: 0)

--- 示例 6: 并发访问 ---
✓ 并发创建成功: worker-1
✓ 并发创建成功: worker-2
✓ 并发创建成功: worker-3

最终池大小: 5

=== 所有示例完成 ===
```

## Pool API 详解

### 创建和销毁

- **Create(ctx, config)**: 创建新 Agent 并加入池
- **Resume(ctx, agentID, config)**: 从存储恢复 Agent
- **Remove(agentID)**: 从池中移除 Agent (关闭但不删除存储)
- **Delete(ctx, agentID)**: 删除 Agent (包括存储数据)
- **Shutdown()**: 关闭所有 Agent 并清空池

### 查询

- **Get(agentID)**: 获取指定 Agent
- **List(prefix)**: 列出所有 Agent ID (可选前缀过滤)
- **Status(agentID)**: 获取 Agent 状态
- **Size()**: 返回池中 Agent 数量

### 遍历

- **ForEach(fn)**: 遍历所有 Agent,执行回调函数

## 使用场景

### 1. 多租户系统

```go
// 为每个用户创建独立的 Agent
userID := "user-123"
config := createAgentConfig(userID)
ag, _ := pool.Create(ctx, config)

// 后续请求获取该用户的 Agent
ag, exists := pool.Get(userID)
```

### 2. 任务队列

```go
// 创建多个 Worker Agent 处理任务
for i := 0; i < 10; i++ {
    workerID := fmt.Sprintf("worker-%d", i)
    pool.Create(ctx, createAgentConfig(workerID))
}

// 分配任务给空闲的 Agent
pool.ForEach(func(id string, ag *agent.Agent) error {
    if ag.Status().State == types.AgentStateIdle {
        ag.Send(ctx, task)
    }
    return nil
})
```

### 3. 会话管理

```go
// 创建会话 Agent
sessionID := generateSessionID()
pool.Create(ctx, createAgentConfig(sessionID))

// 会话超时后清理
if isSessionExpired(sessionID) {
    pool.Remove(sessionID)
}
```

## 容量管理

Pool 支持设置最大容量,防止资源耗尽:

```go
pool := core.NewPool(&core.PoolOptions{
    MaxAgents: 50, // 最多 50 个 Agent
})

// 当池满时,Create 会返回错误
_, err := pool.Create(ctx, config)
if err != nil {
    // 处理池满的情况
    log.Printf("Pool is full: %v", err)
}
```

## 注意事项

1. **资源管理**: 及时调用 `Shutdown()` 释放资源
2. **并发安全**: Pool 内部已实现线程安全,可以安全地并发调用
3. **容量限制**: 根据系统资源合理设置 MaxAgents
4. **Agent ID**: 确保 Agent ID 唯一,重复 ID 会导致创建失败
5. **存储持久化**: Remove 不会删除存储数据,Delete 会删除

## 相关文档

- [Agent 基础示例](../agent)
- [ROADMAP.md](../../ROADMAP.md) - Phase 4 多 Agent 协作
