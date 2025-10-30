package sandbox

import (
	"context"
	"time"
)

// ExecOptions 命令执行选项
type ExecOptions struct {
	Timeout time.Duration
	WorkDir string
	Env     map[string]string
}

// ExecResult 命令执行结果
type ExecResult struct {
	Code   int
	Stdout string
	Stderr string
}

// FileChangeEvent 文件变更事件
type FileChangeEvent struct {
	Path  string
	Mtime time.Time
}

// FileChangeListener 文件变更监听器
type FileChangeListener func(event FileChangeEvent)

// SandboxFS 沙箱文件系统接口
type SandboxFS interface {
	// Resolve 解析路径为绝对路径
	Resolve(path string) string

	// IsInside 检查路径是否在沙箱内
	IsInside(path string) bool

	// Read 读取文件内容
	Read(ctx context.Context, path string) (string, error)

	// Write 写入文件内容
	Write(ctx context.Context, path string, content string) error

	// Temp 生成临时文件路径
	Temp(name string) string

	// Stat 获取文件状态
	Stat(ctx context.Context, path string) (FileInfo, error)

	// Glob 文件匹配
	Glob(ctx context.Context, pattern string, opts *GlobOptions) ([]string, error)
}

// FileInfo 文件信息
type FileInfo struct {
	Path       string
	Size       int64
	ModTime    time.Time
	IsDir      bool
	Mode       int
}

// GlobOptions Glob选项
type GlobOptions struct {
	CWD      string
	Ignore   []string
	Dot      bool
	Absolute bool
}

// Sandbox 沙箱接口
type Sandbox interface {
	// Kind 返回沙箱类型
	Kind() string

	// WorkDir 返回工作目录
	WorkDir() string

	// FS 返回文件系统接口
	FS() SandboxFS

	// Exec 执行命令
	Exec(ctx context.Context, cmd string, opts *ExecOptions) (*ExecResult, error)

	// Watch 监听文件变更
	Watch(paths []string, listener FileChangeListener) (watchID string, err error)

	// Unwatch 取消监听
	Unwatch(watchID string) error

	// Dispose 释放资源
	Dispose() error
}
