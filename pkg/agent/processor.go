package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wordflowlab/agentsdk/pkg/provider"
	"github.com/wordflowlab/agentsdk/pkg/tools"
	"github.com/wordflowlab/agentsdk/pkg/types"
)

// processMessages 处理消息队列
func (a *Agent) processMessages(ctx context.Context) {
	a.mu.Lock()
	if a.state != types.AgentStateReady {
		a.mu.Unlock()
		return // 已经在处理中
	}
	a.state = types.AgentStateWorking
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.state = types.AgentStateReady
		a.mu.Unlock()
	}()

	// 发送状态变更事件
	a.eventBus.EmitMonitor(&types.MonitorStateChangedEvent{
		State: types.AgentStateWorking,
	})

	// 设置断点
	a.setBreakpoint(types.BreakpointPreModel)

	// 调用模型
	if err := a.runModelStep(ctx); err != nil {
		a.eventBus.EmitMonitor(&types.MonitorErrorEvent{
			Severity: "error",
			Phase:    "model",
			Message:  err.Error(),
		})
	}

	// 发送完成事件
	a.eventBus.EmitProgress(&types.ProgressDoneEvent{
		Step:   a.stepCount,
		Reason: "completed",
	})

	// 发送状态变更事件
	a.eventBus.EmitMonitor(&types.MonitorStateChangedEvent{
		State: types.AgentStateReady,
	})
}

// runModelStep 运行模型步骤
func (a *Agent) runModelStep(ctx context.Context) error {
	a.setBreakpoint(types.BreakpointStreamingModel)

	// 准备工具Schema
	toolSchemas := make([]provider.ToolSchema, 0, len(a.toolMap))
	for _, tool := range a.toolMap {
		toolSchemas = append(toolSchemas, provider.ToolSchema{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}

	// 调用模型
	streamOpts := &provider.StreamOptions{
		Tools:     toolSchemas,
		MaxTokens: 4096,
		System:    a.template.SystemPrompt,
	}

	stream, err := a.provider.Stream(ctx, a.messages, streamOpts)
	if err != nil {
		return fmt.Errorf("stream model: %w", err)
	}

	// 处理流式响应
	assistantContent := make([]types.ContentBlock, 0)
	currentBlockIndex := -1
	textBuffers := make(map[int]string)
	inputJSONBuffers := make(map[int]string)

	for chunk := range stream {
		switch chunk.Type {
		case "content_block_start":
			currentBlockIndex = chunk.Index
			if delta, ok := chunk.Delta.(map[string]interface{}); ok {
				blockType, _ := delta["type"].(string)
				if blockType == "text" {
					// 发送文本开始事件
					a.eventBus.EmitProgress(&types.ProgressTextChunkStartEvent{
						Step: a.stepCount,
					})
					// 初始化文本块
					for len(assistantContent) <= currentBlockIndex {
						assistantContent = append(assistantContent, nil)
					}
					assistantContent[currentBlockIndex] = &types.TextBlock{Text: ""}
					textBuffers[currentBlockIndex] = ""
				} else if blockType == "tool_use" {
					// 初始化工具调用块
					for len(assistantContent) <= currentBlockIndex {
						assistantContent = append(assistantContent, nil)
					}
					assistantContent[currentBlockIndex] = &types.ToolUseBlock{
						ID:    delta["id"].(string),
						Name:  delta["name"].(string),
						Input: make(map[string]interface{}),
					}
				}
			}

		case "content_block_delta":
			if delta, ok := chunk.Delta.(map[string]interface{}); ok {
				deltaType, _ := delta["type"].(string)
				if deltaType == "text_delta" {
					text, _ := delta["text"].(string)
					// 累积文本
					textBuffers[currentBlockIndex] += text
					if block, ok := assistantContent[currentBlockIndex].(*types.TextBlock); ok {
						block.Text = textBuffers[currentBlockIndex]
					}
					// 发送文本增量事件
					a.eventBus.EmitProgress(&types.ProgressTextChunkEvent{
						Step:  a.stepCount,
						Delta: text,
					})
				} else if deltaType == "input_json_delta" {
					partialJSON, _ := delta["partial_json"].(string)
					// 累积JSON字符串
					inputJSONBuffers[currentBlockIndex] += partialJSON
				}
			}

		case "content_block_stop":
			if block, ok := assistantContent[currentBlockIndex].(*types.TextBlock); ok {
				// 发送文本结束事件
				a.eventBus.EmitProgress(&types.ProgressTextChunkEndEvent{
					Step: a.stepCount,
					Text: block.Text,
				})
			} else if block, ok := assistantContent[currentBlockIndex].(*types.ToolUseBlock); ok {
				// 解析完整的工具输入JSON
				if jsonStr, exists := inputJSONBuffers[currentBlockIndex]; exists && jsonStr != "" {
					var input map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &input); err == nil {
						block.Input = input
					}
				}
			}

		case "message_delta":
			if chunk.Usage != nil {
				// 发送Token使用事件
				a.eventBus.EmitMonitor(&types.MonitorTokenUsageEvent{
					InputTokens:  chunk.Usage.InputTokens,
					OutputTokens: chunk.Usage.OutputTokens,
					TotalTokens:  chunk.Usage.InputTokens + chunk.Usage.OutputTokens,
				})
			}
		}
	}

	// 保存助手消息
	a.mu.Lock()
	a.messages = append(a.messages, types.Message{
		Role:    types.MessageRoleAssistant,
		Content: assistantContent,
	})
	a.mu.Unlock()

	// 持久化
	if err := a.deps.Store.SaveMessages(ctx, a.id, a.messages); err != nil {
		return fmt.Errorf("save messages: %w", err)
	}

	// 检查是否有工具调用
	toolUses := make([]*types.ToolUseBlock, 0)
	for _, block := range assistantContent {
		if tu, ok := block.(*types.ToolUseBlock); ok {
			toolUses = append(toolUses, tu)
		}
	}

	if len(toolUses) > 0 {
		a.setBreakpoint(types.BreakpointToolPending)
		return a.executeTools(ctx, toolUses)
	}

	return nil
}

// executeTools 执行工具
func (a *Agent) executeTools(ctx context.Context, toolUses []*types.ToolUseBlock) error {
	toolResults := make([]types.ContentBlock, 0, len(toolUses))

	for _, tu := range toolUses {
		result := a.executeSingleTool(ctx, tu)
		toolResults = append(toolResults, result)
	}

	// 保存工具结果
	a.mu.Lock()
	a.messages = append(a.messages, types.Message{
		Role:    types.MessageRoleUser,
		Content: toolResults,
	})
	a.stepCount++
	a.mu.Unlock()

	// 持久化
	if err := a.deps.Store.SaveMessages(ctx, a.id, a.messages); err != nil {
		return fmt.Errorf("save messages: %w", err)
	}

	// 持久化工具记录
	records := make([]types.ToolCallRecord, 0, len(a.toolRecords))
	for _, record := range a.toolRecords {
		records = append(records, *record)
	}
	if err := a.deps.Store.SaveToolCallRecords(ctx, a.id, records); err != nil {
		return fmt.Errorf("save tool records: %w", err)
	}

	// 继续处理
	return a.runModelStep(ctx)
}

// executeSingleTool 执行单个工具
func (a *Agent) executeSingleTool(ctx context.Context, tu *types.ToolUseBlock) types.ContentBlock {
	// 创建工具调用记录
	record := tools.NewToolCallRecord(tu.ID, tu.Name, tu.Input).Build()
	a.mu.Lock()
	a.toolRecords[tu.ID] = record
	a.mu.Unlock()

	// 发送工具开始事件
	a.eventBus.EmitProgress(&types.ProgressToolStartEvent{
		Call: types.ToolCallSnapshot{
			ID:    record.ID,
			Name:  record.Name,
			State: record.State,
		},
	})

	// 获取工具
	tool, ok := a.toolMap[tu.Name]
	if !ok {
		// 工具未找到
		errorMsg := fmt.Sprintf("tool not found: %s", tu.Name)
		a.updateToolRecord(tu.ID, types.ToolCallStateFailed, errorMsg)
		a.eventBus.EmitProgress(&types.ProgressToolErrorEvent{
			Call: types.ToolCallSnapshot{
				ID:    tu.ID,
				Name:  tu.Name,
				State: types.ToolCallStateFailed,
			},
			Error: errorMsg,
		})
		return &types.ToolResultBlock{
			ToolUseID: tu.ID,
			Content: map[string]interface{}{
				"ok":    false,
				"error": errorMsg,
			},
			IsError: true,
		}
	}

	// 设置断点
	a.setBreakpoint(types.BreakpointPreTool)

	// 执行工具
	a.updateToolRecord(tu.ID, types.ToolCallStateExecuting, "")
	a.setBreakpoint(types.BreakpointToolExecuting)

	startTime := time.Now()

	toolCtx := &tools.ToolContext{
		AgentID: a.id,
		Sandbox: a.sandbox,
		Signal:  ctx,
	}

	execResult := a.executor.Execute(ctx, &tools.ExecuteRequest{
		Tool:    tool,
		Input:   tu.Input,
		Context: toolCtx,
		Timeout: 60 * time.Second,
	})

	endTime := time.Now()

	// 更新记录
	if execResult.Success {
		a.updateToolRecord(tu.ID, types.ToolCallStateCompleted, "")
		a.mu.Lock()
		a.toolRecords[tu.ID].Result = execResult.Output
		a.toolRecords[tu.ID].StartedAt = &startTime
		a.toolRecords[tu.ID].CompletedAt = &endTime
		durationMs := execResult.DurationMs
		a.toolRecords[tu.ID].DurationMs = &durationMs
		a.mu.Unlock()
	} else {
		errorMsg := ""
		if execResult.Error != nil {
			errorMsg = execResult.Error.Error()
		}
		a.updateToolRecord(tu.ID, types.ToolCallStateFailed, errorMsg)
	}

	// 发送工具结束事件
	a.eventBus.EmitProgress(&types.ProgressToolEndEvent{
		Call: types.ToolCallSnapshot{
			ID:    tu.ID,
			Name:  tu.Name,
			State: a.toolRecords[tu.ID].State,
		},
	})

	// 设置断点
	a.setBreakpoint(types.BreakpointPostTool)

	// 构建工具结果
	if execResult.Success {
		return &types.ToolResultBlock{
			ToolUseID: tu.ID,
			Content:   execResult.Output,
			IsError:   false,
		}
	} else {
		errorMsg := ""
		if execResult.Error != nil {
			errorMsg = execResult.Error.Error()
		}
		return &types.ToolResultBlock{
			ToolUseID: tu.ID,
			Content: map[string]interface{}{
				"ok":    false,
				"error": errorMsg,
			},
			IsError: true,
		}
	}
}

// setBreakpoint 设置断点
func (a *Agent) setBreakpoint(state types.BreakpointState) {
	a.mu.Lock()
	previous := a.breakpoint
	a.breakpoint = state
	a.mu.Unlock()

	a.eventBus.EmitMonitor(&types.MonitorBreakpointChangedEvent{
		Previous:  previous,
		Current:   state,
		Timestamp: time.Now(),
	})
}

// updateToolRecord 更新工具记录
func (a *Agent) updateToolRecord(id string, state types.ToolCallState, errorMsg string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	record, ok := a.toolRecords[id]
	if !ok {
		return
	}

	now := time.Now()
	record.State = state
	record.UpdatedAt = now

	if errorMsg != "" {
		record.Error = errorMsg
		record.IsError = true
	}

	record.AuditTrail = append(record.AuditTrail, types.ToolCallAuditEntry{
		State:     state,
		Timestamp: now,
	})
}
