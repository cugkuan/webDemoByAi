package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	apperrors "web-demo/enterprise/errors"
	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/service"
	"web-demo/enterprise/pkg/response"
)

// TaskHandler 任务 HTTP Handler
type TaskHandler struct {
	svc *service.TaskService
	c   *cache.Cache
	log zerolog.Logger
}

// NewTaskHandler 创建任务 Handler
func NewTaskHandler(svc *service.TaskService, c *cache.Cache, log zerolog.Logger) *TaskHandler {
	return &TaskHandler{
		svc: svc,
		c:   c,
		log: log,
	}
}

// GetTasks 获取所有任务
func (h *TaskHandler) GetTasks(c *gin.Context) {
	tasks, err := h.svc.GetAllTasks()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(tasks))
}

// GetTask 获取单个任务
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.ErrInvalidID)
		return
	}

	task, err := h.svc.GetTaskByID(uint(id))
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, apperrors.ErrNotFound)
		return
	}
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(task))
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.ErrBadRequest)
		return
	}

	if input.Title == "" {
		c.JSON(http.StatusBadRequest, apperrors.ErrTitleRequired)
		return
	}

	task, err := h.svc.CreateTask(input.Title, input.Done)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response.Created(task))
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.ErrInvalidID)
		return
	}

	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.ErrBadRequest)
		return
	}

	task, err := h.svc.UpdateTask(uint(id), input.Title, input.Done)
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, apperrors.ErrNotFound)
		return
	}
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(task))
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.ErrInvalidID)
		return
	}

	if err := h.svc.DeleteTask(uint(id)); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// CountTasks 统计任务总数
func (h *TaskHandler) CountTasks(c *gin.Context) {
	count, err := h.svc.CountTasks()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total": count,
	}))
}

// ClearAllCache 清除所有缓存
func (h *TaskHandler) ClearAllCache(c *gin.Context) {
	h.c.InvalidateAllTasks(context.Background())
	c.JSON(http.StatusOK, response.Success(nil))
}

// ClearTaskCache 清除指定任务缓存
func (h *TaskHandler) ClearTaskCache(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.ErrInvalidID)
		return
	}

	h.c.InvalidateTask(context.Background(), uint(id))
	c.JSON(http.StatusOK, response.Success(nil))
}
