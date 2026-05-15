package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// HandleGetTasks 获取所有任务
func HandleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := GetAllTasks()
	if err != nil {
		respondError(w, 500, "数据库错误")
		return
	}

	respondJSON(w, 200, "成功", tasks)
}

// HandleGetTask 获取单个任务
func HandleGetTask(w http.ResponseWriter, r *http.Request, id uint) {
	task, err := GetTaskByID(id)
	if err == gorm.ErrRecordNotFound {
		respondError(w, 404, "任务不存在")
		return
	}
	if err != nil {
		respondError(w, 500, "数据库错误")
		return
	}

	respondJSON(w, 200, "成功", task)
}

// HandleCreateTask 创建任务
func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, 400, "无效的请求体")
		return
	}

	if input.Title == "" {
		respondError(w, 400, "标题不能为空")
		return
	}

	task, err := CreateTask(input.Title, input.Done)
	if err != nil {
		respondError(w, 500, "数据库错误")
		return
	}

	respondJSON(w, http.StatusCreated, "创建成功", task)
}

// HandleUpdateTask 更新任务
func HandleUpdateTask(w http.ResponseWriter, r *http.Request, id uint) {
	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, 400, "无效的请求体")
		return
	}

	task, err := UpdateTask(id, input.Title, input.Done)
	if err == gorm.ErrRecordNotFound {
		respondError(w, 404, "任务不存在")
		return
	}
	if err != nil {
		respondError(w, 500, "更新失败")
		return
	}

	respondJSON(w, 200, "更新成功", task)
}

// HandleDeleteTask 删除任务
func HandleDeleteTask(w http.ResponseWriter, r *http.Request, id uint) {
	// 直接删除并检查影响行数
	result := DB.Delete(&Task{}, id)
	if result.Error != nil {
		respondError(w, 500, "删除失败")
		return
	}
	
	if result.RowsAffected == 0 {
		respondError(w, 404, "任务不存在")
		return
	}

	respondJSON(w, http.StatusOK, "删除成功", nil)
}

// HandleTasks 路由处理器
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	// 提取 ID
	parts := strings.Split(r.URL.Path, "/")
	var id uint
	if len(parts) > 4 && parts[4] != "" {
		idInt, err := strconv.Atoi(parts[4])
		if err != nil {
			respondError(w, 400, "无效的任务 ID")
			return
		}
		id = uint(idInt)
	}

	switch r.Method {
	case http.MethodGet:
		if id > 0 {
			HandleGetTask(w, r, id)
		} else {
			HandleGetTasks(w, r)
		}
	case http.MethodPost:
		HandleCreateTask(w, r)
	case http.MethodPut:
		if id > 0 {
			HandleUpdateTask(w, r, id)
		} else {
			respondError(w, 400, "更新需要指定任务 ID")
		}
	case http.MethodDelete:
		if id > 0 {
			HandleDeleteTask(w, r, id)
		} else {
			respondError(w, 400, "删除需要指定任务 ID")
		}
	default:
		respondError(w, 405, "方法不允许")
	}
}

// respondJSON 返回 JSON 响应
func respondJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, code int, message string) {
	resp := Response{
		Code:    code,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}
