package builtin

import "github.com/wordflowlab/agentsdk/pkg/tools"

// RegisterAll 注册所有内置工具
func RegisterAll(registry *tools.Registry) {
	// 文件系统工具
	registry.Register("fs_read", NewFsReadTool)
	registry.Register("fs_write", NewFsWriteTool)

	// Bash工具
	registry.Register("bash_run", NewBashRunTool)
}

// FileSystemTools 返回文件系统工具列表
func FileSystemTools() []string {
	return []string{"fs_read", "fs_write"}
}

// BashTools 返回Bash工具列表
func BashTools() []string {
	return []string{"bash_run"}
}

// AllTools 返回所有内置工具列表
func AllTools() []string {
	return append(FileSystemTools(), BashTools()...)
}
