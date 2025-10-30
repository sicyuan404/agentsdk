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
	return `## bash_run - 执行Bash命令

**用途**: 在沙箱环境中执行shell命令

**参数**:
- cmd (必填): 要执行的命令
- timeout_ms (可选): 超时时间(毫秒),默认120000

**返回**:
- ok: 是否成功(exit code = 0)
- code: 退出码
- output: 标准输出+标准错误

**示例**:
` + "```json\n" + `{
  "cmd": "ls -la",
  "timeout_ms": 5000
}
` + "```\n" + `

**安全限制**:
- 危险命令会被自动阻止(rm -rf /, curl|bash等)
- 命令在沙箱工作目录内执行
- 默认120秒超时

**注意事项**:
- 非零退出码表示命令失败
- stdout和stderr会合并到output字段
- 建议使用具体的命令而不是脚本
`
}
