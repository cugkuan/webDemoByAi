package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"web-demo/enterprise/internal/model"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func newTestRepo(t *testing.T) *TaskRepo {
	t.Helper()
	return NewTaskRepo(newTestDB(t))
}

func TestTaskRepo_Create(t *testing.T) {
	repo := newTestRepo(t)

	task := &model.Task{Title: "测试任务", Done: false}
	if err := repo.Create(task); err != nil {
		t.Fatal(err)
	}
	if task.ID == 0 {
		t.Error("task ID should be set after create")
	}
}

func TestTaskRepo_FindAll(t *testing.T) {
	repo := newTestRepo(t)

	// 初始应该为空
	tasks, err := repo.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("tasks = %d, want 0", len(tasks))
	}

	// 创建 3 个任务
	for i := 0; i < 3; i++ {
		repo.Create(&model.Task{Title: "task"})
	}

	tasks, err = repo.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 3 {
		t.Errorf("tasks = %d, want 3", len(tasks))
	}
}

func TestTaskRepo_FindByID(t *testing.T) {
	repo := newTestRepo(t)

	task := &model.Task{Title: "test", Done: true}
	repo.Create(task)

	found, err := repo.FindByID(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Title != "test" {
		t.Errorf("Title = %s, want test", found.Title)
	}
	if !found.Done {
		t.Error("Done should be true")
	}
}

func TestTaskRepo_FindByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.FindByID(999)
	if err == nil {
		t.Error("FindByID should return error for non-existent ID")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("err = %v, want gorm.ErrRecordNotFound", err)
	}
}

func TestTaskRepo_Update(t *testing.T) {
	repo := newTestRepo(t)

	task := &model.Task{Title: "original", Done: false}
	repo.Create(task)

	task.Title = "updated"
	task.Done = true
	if err := repo.Update(task); err != nil {
		t.Fatal(err)
	}

	found, _ := repo.FindByID(task.ID)
	if found.Title != "updated" {
		t.Errorf("Title = %s, want updated", found.Title)
	}
	if !found.Done {
		t.Error("Done should be true")
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	repo := newTestRepo(t)

	task := &model.Task{Title: "to delete"}
	repo.Create(task)

	if err := repo.Delete(task.ID); err != nil {
		t.Fatal(err)
	}

	_, err := repo.FindByID(task.ID)
	if err != gorm.ErrRecordNotFound {
		t.Errorf("err = %v, want gorm.ErrRecordNotFound", err)
	}
}

func TestTaskRepo_Count(t *testing.T) {
	repo := newTestRepo(t)

	count, err := repo.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	repo.Create(&model.Task{Title: "a"})
	repo.Create(&model.Task{Title: "b"})

	count, err = repo.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestTaskRepo_AutoMigrate(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepo(db)

	// 迁移应该成功
	if err := repo.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	// 迁移后应该可以正常使用
	task := &model.Task{Title: "after migrate"}
	if err := repo.Create(task); err != nil {
		t.Fatal(err)
	}
}

func TestTaskRepo_FindAll_Order(t *testing.T) {
	repo := newTestRepo(t)

	// 创建任务，ID 递增
	repo.Create(&model.Task{Title: "first"})
	repo.Create(&model.Task{Title: "second"})
	repo.Create(&model.Task{Title: "third"})

	tasks, err := repo.FindAll()
	if err != nil {
		t.Fatal(err)
	}

	// 应该按 ID DESC 排序
	if len(tasks) >= 3 {
		if tasks[0].Title != "third" {
			t.Errorf("first result = %s, want third", tasks[0].Title)
		}
		if tasks[2].Title != "first" {
			t.Errorf("last result = %s, want first", tasks[2].Title)
		}
	}
}
