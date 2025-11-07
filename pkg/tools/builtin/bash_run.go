package builtin

import (
	"context"
	"fmt"

	"github.com/wordflowlab/agentsdk/pkg/tools"
)

// BashRunTool Bash命令执行工具
type BashRunTool struct{}

// NewBashRunTool 创建Bash执行工具
func NewBashRunTool(config map[string]interface{}) (tools.Tool, error) {
	return &BashRunTool{}, nil
}

func (t *BashRunTool) Name() string {
	return "bash_run"
}

func (t *BashRunTool) Description() string {
	return "Execute bash commands in the sandbox environment"
}

func (t *BashRunTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cmd": map[string]interface{}{
				"type":        "string",
				"description": "Command to execute",
			},
			"timeout_ms": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in milliseconds (default: 120000)",
			},
		},
		"required": []string{"cmd"},
	}
}

func (t *BashRunTool) Execute(ctx context.Context, input map[string]interface{}, tc *tools.ToolContext) (interface{}, error) {
	// 获取参数
	cmd, ok := input["cmd"].(string)
	if !ok {
		return nil, fmt.Errorf("cmd must be a string")
	}

	// 执行命令
	result, err := tc.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return map[string]interface{}{
			"ok":    false,
			"error": fmt.Sprintf("failed to execute command: %v", err),
			"recommendations": []string{
				"检查命令语法是否正确",
				"确认命令在沙箱环境中可执行",
				"验证是否有执行权限",
			},
		}, nil
	}

	// 合并输出
	output := result.Stdout
	if result.Stderr != "" {
		output += "\n" + result.Stderr
	}

	if output == "" {
		output = "(no output)"
	}

	success := result.Code == 0

	response := map[string]interface{}{
		"ok":     success,
		"code":   result.Code,
		"output": output,
	}

	if !success {
		response["error"] = fmt.Sprintf("command exited with code %d", result.Code)
		response["recommendations"] = []string{
			"检查命令的标准错误输出",
			"验证命令参数是否正确",
			"确认所需的依赖是否已安装",
		}
	}

	return response, nil
}

func (t *BashRunTool) Prompt() string {
	return `Execute bash commands in the sandboxed workspace.

Guidelines:
- Commands run in the sandbox working directory.
- You may provide "timeout_ms" to override the default 120s timeout.
- The tool returns stdout, stderr, and exit code.

Safety/Limitations:
- Dangerous commands are automatically blocked (rm -rf /, curl|bash, etc.).
- Commands timeout after 120s by default.
- Non-zero exit codes indicate command failure.`
}
