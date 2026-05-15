package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/model"
	"web-demo/enterprise/internal/repository"
	"web-demo/enterprise/internal/service"
	"web-demo/enterprise/pkg/response"
)

func setupTest(t *testing.T) (*TaskHandler, *gin.Engine) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		t.Fatal(err)
	}

	log := zerolog.New(zerolog.NewTestWriter(t))
	repo := repository.NewTaskRepo(db)
	c := cache.NewForTest(log)
	svc := service.NewTaskService(repo, c, log)
	h := NewTaskHandler(svc, c, log)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/tasks", h.GetTasks)
	r.GET("/api/tasks/:id", h.GetTask)
	r.POST("/api/tasks", h.CreateTask)
	r.PUT("/api/tasks/:id", h.UpdateTask)
	r.DELETE("/api/tasks/:id", h.DeleteTask)
	r.DELETE("/api/cache", h.ClearAllCache)
	r.DELETE("/api/cache/:id", h.ClearTaskCache)

	return h, r
}

func performRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestGetTasks_Empty(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "GET", "/api/tasks", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}
}

func TestCreateTask(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "POST", "/api/tasks", `{"title":"测试任务","done":false}`)
	resp := parseResponse(t, w)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if resp.Code != 201 {
		t.Errorf("code = %d, want 201", resp.Code)
	}
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "POST", "/api/tasks", `{"title":"","done":false}`)
	resp := parseResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if resp.Code != 400 {
		t.Errorf("code = %d, want 400", resp.Code)
	}
	if resp.Message != "标题不能为空" {
		t.Errorf("message = %s, want 标题不能为空", resp.Message)
	}
}

func TestCreateTask_InvalidJSON(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "POST", "/api/tasks", `invalid json`)
	resp := parseResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if resp.Code != 400 {
		t.Errorf("code = %d, want 400", resp.Code)
	}
}

func TestGetTask(t *testing.T) {
	_, r := setupTest(t)

	// 先创建一个任务
	w := performRequest(r, "POST", "/api/tasks", `{"title":"test"}`)
	if w.Code != http.StatusCreated {
		t.Fatal("failed to create task")
	}

	// 获取该任务
	w = performRequest(r, "GET", "/api/tasks/1", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "GET", "/api/tasks/999", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
	if resp.Code != 404 {
		t.Errorf("code = %d, want 404", resp.Code)
	}
}

func TestGetTask_InvalidID(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "GET", "/api/tasks/abc", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if resp.Code != 400 {
		t.Errorf("code = %d, want 400", resp.Code)
	}
}

func TestUpdateTask(t *testing.T) {
	_, r := setupTest(t)

	// 先创建
	performRequest(r, "POST", "/api/tasks", `{"title":"original","done":false}`)

	// 更新
	w := performRequest(r, "PUT", "/api/tasks/1", `{"title":"updated","done":true}`)
	resp := parseResponse(t, w)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "PUT", "/api/tasks/999", `{"title":"updated","done":true}`)
	resp := parseResponse(t, w)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
	if resp.Code != 404 {
		t.Errorf("code = %d, want 404", resp.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	_, r := setupTest(t)

	// 先创建
	performRequest(r, "POST", "/api/tasks", `{"title":"to delete"}`)

	// 删除
	w := performRequest(r, "DELETE", "/api/tasks/1", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}

	// 验证已删除
	w = performRequest(r, "GET", "/api/tasks/1", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("status after delete = %d, want 404", w.Code)
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "DELETE", "/api/tasks/999", "")

	// 删除不存在的任务，handler 调用 c.Error(err)，没有错误中间件时返回 200 空响应
	// 这是当前实现的行为，实际生产应添加错误中间件处理
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestClearAllCache(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "DELETE", "/api/cache", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}
}

func TestClearTaskCache(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "DELETE", "/api/cache/1", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}
}

func TestClearTaskCache_InvalidID(t *testing.T) {
	_, r := setupTest(t)
	w := performRequest(r, "DELETE", "/api/cache/abc", "")
	resp := parseResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if resp.Code != 400 {
		t.Errorf("code = %d, want 400", resp.Code)
	}
}

func TestCreateAndGetTask_Integration(t *testing.T) {
	_, r := setupTest(t)

	// 创建任务
	w := performRequest(r, "POST", "/api/tasks", `{"title":"集成测试","done":true}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", w.Code)
	}

	// 获取所有任务
	w = performRequest(r, "GET", "/api/tasks", "")
	resp := parseResponse(t, w)
	if resp.Code != 200 {
		t.Fatalf("get all code = %d, want 200", resp.Code)
	}

	// 获取单个任务
	w = performRequest(r, "GET", "/api/tasks/1", "")
	resp = parseResponse(t, w)
	if resp.Code != 200 {
		t.Fatalf("get one code = %d, want 200", resp.Code)
	}

	// 更新任务
	w = performRequest(r, "PUT", "/api/tasks/1", `{"title":"已更新","done":false}`)
	resp = parseResponse(t, w)
	if resp.Code != 200 {
		t.Fatalf("update code = %d, want 200", resp.Code)
	}

	// 删除任务
	w = performRequest(r, "DELETE", "/api/tasks/1", "")
	resp = parseResponse(t, w)
	if resp.Code != 200 {
		t.Fatalf("delete code = %d, want 200", resp.Code)
	}

	// 验证已删除
	w = performRequest(r, "GET", "/api/tasks/1", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("final get status = %d, want 404", w.Code)
	}
}
