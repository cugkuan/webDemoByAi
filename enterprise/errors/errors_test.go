package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(400, "bad request")
	if err.Code != 400 {
		t.Errorf("Code = %d, want 400", err.Code)
	}
	if err.Message != "bad request" {
		t.Errorf("Message = %s, want bad request", err.Message)
	}
	if err.Err != nil {
		t.Error("Err should be nil")
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	err := Wrap(500, "internal error", original)

	if err.Code != 500 {
		t.Errorf("Code = %d, want 500", err.Code)
	}
	if err.Message != "internal error" {
		t.Errorf("Message = %s, want internal error", err.Message)
	}
	if err.Err != original {
		t.Errorf("Err = %v, want %v", err.Err, original)
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{
			name: "without wrapped error",
			err:  New(404, "not found"),
			want: "not found",
		},
		{
			name: "with wrapped error",
			err:  Wrap(500, "server error", errors.New("db timeout")),
			want: "server error: db timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	original := errors.New("original")
	err := Wrap(500, "error", original)

	if !errors.Is(err, original) {
		t.Error("errors.Is should find the wrapped error")
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		code int
		want int
	}{
		{200, 200},
		{400, 400},
		{404, 404},
		{429, 429},
		{500, 500},
		{1000, http.StatusInternalServerError},
		{2000, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := HTTPStatus(tt.code); got != tt.want {
				t.Errorf("HTTPStatus(%d) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		wantMsg string
		wantCode int
	}{
		{"ErrNotFound", ErrNotFound, "资源不存在", 404},
		{"ErrBadRequest", ErrBadRequest, "无效的请求", 400},
		{"ErrInternalServer", ErrInternalServer, "服务器内部错误", 500},
		{"ErrTitleRequired", ErrTitleRequired, "标题不能为空", 400},
		{"ErrInvalidID", ErrInvalidID, "无效的任务 ID", 400},
		{"ErrRateLimited", ErrRateLimited, "请求过于频繁，请稍后再试", 429},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", tt.err.Code, tt.wantCode)
			}
			if tt.err.Message != tt.wantMsg {
				t.Errorf("Message = %s, want %s", tt.err.Message, tt.wantMsg)
			}
		})
	}
}
