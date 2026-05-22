package service

import (
	"testing"

	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/model"
	"web-demo/enterprise/internal/repository"
)

func newTestService(t *testing.T) (*TaskService, *cache.Cache, *repository.TaskRepo) {
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
	svc := NewTaskService(repo, c, log)

	return svc, c, repo
}

func TestTaskService_CreateTask(t *testing.T) {
	svc, c, _ := newTestService(t)

	// 先设置一个列表缓存
	c.SetL1(cache.AllTasksCacheKey, []model.Task{{ID: 1, Title: "old"}})

	task, err := svc.CreateTask("新任务", false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "新任务" {
		t.Errorf("Title = %s, want 新任务", task.Title)
	}
	if task.ID == 0 {
		t.Error("ID should be set")
	}

	// 验证列表缓存被清除
	if _, ok := c.GetL1(cache.AllTasksCacheKey); ok {
		t.Error("AllTasksCache should be invalidated after create")
	}
}

func TestTaskService_GetTaskByID_FromDB(t *testing.T) {
	svc, _, repo := newTestService(t)

	repo.Create(&model.Task{Title: "test task", Done: true})

	task, err := svc.GetTaskByID(1)
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "test task" {
		t.Errorf("Title = %s, want test task", task.Title)
	}
	if !task.Done {
		t.Error("Done should be true")
	}
}

func TestTaskService_GetTaskByID_FromL1Cache(t *testing.T) {
	svc, c, _ := newTestService(t)

	expected := &model.Task{ID: 1, Title: "cached task"}
	c.SetL1("task:1", expected)

	task, err := svc.GetTaskByID(1)
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "cached task" {
		t.Errorf("Title = %s, want cached task", task.Title)
	}
}

func TestTaskService_GetTaskByID_NotFound(t *testing.T) {
	svc, _, _ := newTestService(t)

	_, err := svc.GetTaskByID(999)
	if err == nil {
		t.Error("GetTaskByID should return error for non-existent ID")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("err = %v, want gorm.ErrRecordNotFound", err)
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	svc, c, repo := newTestService(t)

	repo.Create(&model.Task{Title: "original", Done: false})

	// 设置缓存
	c.SetL1("task:1", &model.Task{ID: 1, Title: "original"})
	c.SetL1(cache.AllTasksCacheKey, []model.Task{{ID: 1}})

	task, err := svc.UpdateTask(1, "updated", true)
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "updated" {
		t.Errorf("Title = %s, want updated", task.Title)
	}
	if !task.Done {
		t.Error("Done should be true")
	}

	// 验证缓存被清除
	if _, ok := c.GetL1("task:1"); ok {
		t.Error("task:1 cache should be invalidated")
	}
	if _, ok := c.GetL1(cache.AllTasksCacheKey); ok {
		t.Error("AllTasksCache should be invalidated")
	}
}

func TestTaskService_UpdateTask_NotFound(t *testing.T) {
	svc, _, _ := newTestService(t)

	_, err := svc.UpdateTask(999, "updated", false)
	if err == nil {
		t.Error("UpdateTask should return error for non-existent ID")
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	svc, c, repo := newTestService(t)

	repo.Create(&model.Task{Title: "to delete"})

	// 设置缓存
	c.SetL1("task:1", &model.Task{ID: 1})
	c.SetL1(cache.AllTasksCacheKey, []model.Task{{ID: 1}})

	if err := svc.DeleteTask(1); err != nil {
		t.Fatal(err)
	}

	// 验证缓存被清除
	if _, ok := c.GetL1("task:1"); ok {
		t.Error("task:1 cache should be invalidated")
	}
	if _, ok := c.GetL1(cache.AllTasksCacheKey); ok {
		t.Error("AllTasksCache should be invalidated")
	}

	// 验证数据库已删除
	_, err := repo.FindByID(1)
	if err != gorm.ErrRecordNotFound {
		t.Errorf("err = %v, want gorm.ErrRecordNotFound", err)
	}
}

func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	svc, _, _ := newTestService(t)

	err := svc.DeleteTask(999)
	if err == nil {
		t.Error("DeleteTask should return error for non-existent ID")
	}
}
