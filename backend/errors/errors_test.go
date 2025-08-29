package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestAppError(t *testing.T) {
	t.Run("NewAppError", func(t *testing.T) {
		baseErr := errors.New("base error")
		appErr := New(NetworkError, "network failed", baseErr)
		
		if appErr.Type != NetworkError {
			t.Errorf("Expected NetworkError type, got %v", appErr.Type)
		}
		
		if appErr.Message != "network failed" {
			t.Errorf("Expected message 'network failed', got %s", appErr.Message)
		}
		
		if appErr.Err != baseErr {
			t.Error("Base error not set correctly")
		}
		
		if appErr.UserMessage != "网络连接出现问题" {
			t.Errorf("User message not set correctly: %s", appErr.UserMessage)
		}
	})
	
	t.Run("ErrorTypes", func(t *testing.T) {
		tests := []struct {
			name        string
			constructor func(string, error) *AppError
			errType     ErrorType
			userMsg     string
			suggestion  string
		}{
			{
				name:        "NetworkError",
				constructor: NewNetworkError,
				errType:     NetworkError,
				userMsg:     "网络连接出现问题",
				suggestion:  "请检查网络连接并重试",
			},
			{
				name:        "ConfigError",
				constructor: NewConfigError,
				errType:     ConfigError,
				userMsg:     "配置错误",
				suggestion:  "请检查配置文件是否正确",
			},
			{
				name:        "PermissionError",
				constructor: NewPermissionError,
				errType:     PermissionError,
				userMsg:     "权限不足",
				suggestion:  "请以管理员权限运行程序",
			},
			{
				name:        "SystemError",
				constructor: NewSystemError,
				errType:     SystemError,
				userMsg:     "系统错误",
				suggestion:  "请重启程序或联系技术支持",
			},
			{
				name:        "UserError",
				constructor: NewUserError,
				errType:     UserError,
				userMsg:     "操作错误",
				suggestion:  "请检查您的操作是否正确",
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.constructor("test error", nil)
				
				if err.Type != tt.errType {
					t.Errorf("Expected type %v, got %v", tt.errType, err.Type)
				}
				
				if err.UserMessage != tt.userMsg {
					t.Errorf("Expected user message '%s', got '%s'", 
						tt.userMsg, err.UserMessage)
				}
				
				if err.Suggestion != tt.suggestion {
					t.Errorf("Expected suggestion '%s', got '%s'", 
						tt.suggestion, err.Suggestion)
				}
			})
		}
	})
	
	t.Run("ErrorInterface", func(t *testing.T) {
		baseErr := errors.New("base error")
		appErr := NewNetworkError("connection failed", baseErr)
		
		errStr := appErr.Error()
		if !strings.Contains(errStr, "connection failed") {
			t.Errorf("Error string should contain message: %s", errStr)
		}
		
		if !strings.Contains(errStr, "base error") {
			t.Errorf("Error string should contain base error: %s", errStr)
		}
	})
	
	t.Run("ErrorWithoutBaseError", func(t *testing.T) {
		appErr := NewNetworkError("connection failed", nil)
		
		errStr := appErr.Error()
		if errStr != "connection failed" {
			t.Errorf("Expected 'connection failed', got '%s'", errStr)
		}
	})
	
	t.Run("Unwrap", func(t *testing.T) {
		baseErr := errors.New("base error")
		appErr := NewNetworkError("wrapper", baseErr)
		
		unwrapped := appErr.Unwrap()
		if unwrapped != baseErr {
			t.Error("Unwrap should return base error")
		}
	})
	
	t.Run("WithUserMessage", func(t *testing.T) {
		appErr := NewNetworkError("error", nil)
		customMsg := "自定义用户消息"
		
		appErr.WithUserMessage(customMsg)
		if appErr.UserMessage != customMsg {
			t.Errorf("Expected '%s', got '%s'", customMsg, appErr.UserMessage)
		}
	})
	
	t.Run("WithSuggestion", func(t *testing.T) {
		appErr := NewNetworkError("error", nil)
		customSuggestion := "自定义建议"
		
		appErr.WithSuggestion(customSuggestion)
		if appErr.Suggestion != customSuggestion {
			t.Errorf("Expected '%s', got '%s'", customSuggestion, appErr.Suggestion)
		}
	})
	
	t.Run("MethodChaining", func(t *testing.T) {
		appErr := NewNetworkError("error", nil).
			WithUserMessage("custom message").
			WithSuggestion("custom suggestion")
		
		if appErr.UserMessage != "custom message" {
			t.Error("Method chaining failed for WithUserMessage")
		}
		
		if appErr.Suggestion != "custom suggestion" {
			t.Error("Method chaining failed for WithSuggestion")
		}
	})
}

// BenchmarkAppErrorCreation 性能测试
func BenchmarkAppErrorCreation(b *testing.B) {
	baseErr := errors.New("base error")
	
	b.Run("NewAppError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New(NetworkError, "test error", baseErr)
		}
	})
	
	b.Run("NewNetworkError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewNetworkError("test error", baseErr)
		}
	})
	
	b.Run("WithCustomization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewNetworkError("test error", baseErr).
				WithUserMessage("custom").
				WithSuggestion("suggestion")
		}
	})
}

// TestErrorUsage 测试实际使用场景
func TestErrorUsage(t *testing.T) {
	t.Run("NetworkScenario", func(t *testing.T) {
		// 模拟网络连接失败
		baseErr := errors.New("dial tcp: connection refused")
		appErr := NewNetworkError("Failed to connect to proxy server", baseErr).
			WithUserMessage("无法连接到代理服务器").
			WithSuggestion("请检查服务器地址和端口是否正确，以及网络是否通畅")
		
		// 检查错误信息
		if !strings.Contains(appErr.Error(), "connection refused") {
			t.Error("Should contain original error message")
		}
		
		if appErr.UserMessage != "无法连接到代理服务器" {
			t.Error("User message not set correctly")
		}
	})
	
	t.Run("PermissionScenario", func(t *testing.T) {
		// 模拟权限不足
		baseErr := errors.New("operation not permitted")
		appErr := NewPermissionError("Cannot create TUN interface", baseErr)
		
		if appErr.Type != PermissionError {
			t.Error("Should be PermissionError type")
		}
		
		if !strings.Contains(appErr.Suggestion, "管理员") {
			t.Error("Suggestion should mention admin privileges")
		}
	})
}