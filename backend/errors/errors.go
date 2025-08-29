package errors

import (
	"fmt"
)

// ErrorType 定义错误类型
type ErrorType int

const (
	// NetworkError 网络相关错误
	NetworkError ErrorType = iota
	// ConfigError 配置相关错误
	ConfigError
	// PermissionError 权限相关错误
	PermissionError
	// SystemError 系统相关错误
	SystemError
	// UserError 用户操作错误
	UserError
)

// AppError 应用错误结构
type AppError struct {
	Type        ErrorType
	Message     string
	Err         error
	UserMessage string // 用户友好的错误信息
	Suggestion  string // 错误解决建议
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 支持错误链
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建新的应用错误
func New(errType ErrorType, message string, err error) *AppError {
	appErr := &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
	
	// 根据错误类型设置用户友好信息
	switch errType {
	case NetworkError:
		appErr.UserMessage = "网络连接出现问题"
		appErr.Suggestion = "请检查网络连接并重试"
	case ConfigError:
		appErr.UserMessage = "配置错误"
		appErr.Suggestion = "请检查配置文件是否正确"
	case PermissionError:
		appErr.UserMessage = "权限不足"
		appErr.Suggestion = "请以管理员权限运行程序"
	case SystemError:
		appErr.UserMessage = "系统错误"
		appErr.Suggestion = "请重启程序或联系技术支持"
	case UserError:
		appErr.UserMessage = "操作错误"
		appErr.Suggestion = "请检查您的操作是否正确"
	default:
		appErr.UserMessage = message
		appErr.Suggestion = "请重试或联系技术支持"
	}
	
	return appErr
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, err error) *AppError {
	return New(NetworkError, message, err)
}

// NewConfigError 创建配置错误
func NewConfigError(message string, err error) *AppError {
	return New(ConfigError, message, err)
}

// NewPermissionError 创建权限错误
func NewPermissionError(message string, err error) *AppError {
	return New(PermissionError, message, err)
}

// NewSystemError 创建系统错误
func NewSystemError(message string, err error) *AppError {
	return New(SystemError, message, err)
}

// NewUserError 创建用户错误
func NewUserError(message string, err error) *AppError {
	return New(UserError, message, err)
}

// WithUserMessage 设置用户友好消息
func (e *AppError) WithUserMessage(msg string) *AppError {
	e.UserMessage = msg
	return e
}

// WithSuggestion 设置解决建议
func (e *AppError) WithSuggestion(suggestion string) *AppError {
	e.Suggestion = suggestion
	return e
}