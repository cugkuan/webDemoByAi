package response

import (
	"encoding/json"
	"testing"
)

func TestSuccess(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := Success(data)

	if resp.Code != 200 {
		t.Errorf("Code = %d, want 200", resp.Code)
	}
	if resp.Message != "成功" {
		t.Errorf("Message = %s, want 成功", resp.Message)
	}
	if resp.Data == nil {
		t.Error("Data should not be nil")
	}

	// 验证 JSON 序列化
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Response
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Code != 200 {
		t.Errorf("JSON Code = %d, want 200", decoded.Code)
	}
}

func TestSuccess_NilData(t *testing.T) {
	resp := Success(nil)
	if resp.Code != 200 {
		t.Errorf("Code = %d, want 200", resp.Code)
	}
	// nil data 应该被 omitempty 忽略
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, exists := decoded["data"]; exists {
		t.Error("data field should be omitted for nil")
	}
}

func TestCreated(t *testing.T) {
	data := "new resource"
	resp := Created(data)

	if resp.Code != 201 {
		t.Errorf("Code = %d, want 201", resp.Code)
	}
	if resp.Message != "创建成功" {
		t.Errorf("Message = %s, want 创建成功", resp.Message)
	}
	if resp.Data != data {
		t.Errorf("Data = %v, want %v", resp.Data, data)
	}
}

func TestError(t *testing.T) {
	resp := Error(400, "无效的请求")

	if resp.Code != 400 {
		t.Errorf("Code = %d, want 400", resp.Code)
	}
	if resp.Message != "无效的请求" {
		t.Errorf("Message = %s, want 无效的请求", resp.Message)
	}
	if resp.Data != nil {
		t.Error("Data should be nil for error response")
	}
}

func TestResponse_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		resp Response
	}{
		{"success", Success([]string{"a", "b"})},
		{"created", Created(map[string]int{"id": 1})},
		{"error", Error(500, "服务器错误")},
		{"empty_success", Success(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.resp)
			if err != nil {
				t.Fatal(err)
			}

			var decoded Response
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatal(err)
			}

			if decoded.Code != tt.resp.Code {
				t.Errorf("Code = %d, want %d", decoded.Code, tt.resp.Code)
			}
			if decoded.Message != tt.resp.Message {
				t.Errorf("Message = %s, want %s", decoded.Message, tt.resp.Message)
			}
		})
	}
}
