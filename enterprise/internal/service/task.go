package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/model"
	"web-demo/enterprise/internal/repository"
)

// TaskService 任务业务逻辑
type TaskService struct {
	repo  *repository.TaskRepo
	cache *cache.Cache
	log   zerolog.Logger
}

// NewTaskService 创建任务服务
func NewTaskService(repo *repository.TaskRepo, c *cache.Cache, log zerolog.Logger) *TaskService {
	return &TaskService{
		repo:  repo,
		cache: c,
		log:   log,
	}
}

// GetTasksPage 分页获取任务
func (s *TaskService) GetTasksPage(page, pageSize int) ([]model.Task, int64, error) {
	s.log.Debug().Int("page", page).Int("pageSize", pageSize).Msg("SERVICE: 分页获取任务")
	return s.repo.FindPage(page, pageSize)
}

// GetAllTasks 获取所有任务（带缓存穿透）
func (s *TaskService) GetAllTasks() ([]model.Task, error) {
	ctx := context.Background()

	// L1 缓存检查
	if val, ok := s.cache.GetL1(cache.AllTasksCacheKey); ok {
		if tasks, ok := val.([]model.Task); ok {
			s.log.Debug().Msg("CACHE HIT: L1 - 获取所有任务")
			return tasks, nil
		}
		s.log.Debug().Msg("CACHE INVALID: L1 缓存类型异常，跳过")
	}
	s.log.Debug().Msg("CACHE MISS: L1 - 获取所有任务")

	// L2 缓存检查
	var tasks []model.Task
	if found, err := s.cache.GetL2(ctx, cache.AllTasksCacheKey, &tasks); found && err == nil {
		s.log.Debug().Msg("CACHE HIT: L2 - 获取所有任务，回写到 L1")
		s.cache.SetL1(cache.AllTasksCacheKey, tasks)
		return tasks, nil
	}
	s.log.Debug().Msg("CACHE MISS: L2 - 获取所有任务")

	// 从数据库查询
	s.log.Debug().Msg("DB QUERY: SELECT 所有任务")
	tasks, err := s.repo.FindAll()
	if err != nil {
		s.log.Error().Err(err).Msg("数据库查询失败")
		return nil, err
	}
	s.log.Debug().Int("count", len(tasks)).Msg("DB SUCCESS: 查询到任务记录")

	// 写入缓存（L1 和 L2）
	s.cache.SetL1(cache.AllTasksCacheKey, tasks)
	s.cache.SetL2(ctx, cache.AllTasksCacheKey, tasks)

	return tasks, nil
}

// GetTaskByID 根据 ID 获取任务（带缓存穿透）
func (s *TaskService) GetTaskByID(id uint) (*model.Task, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", cache.TaskCacheKeyPrefix, id)

	// L1 缓存检查
	if val, ok := s.cache.GetL1(key); ok {
		if task, ok := val.(*model.Task); ok {
			s.log.Debug().Uint("id", id).Msg("CACHE HIT: L1 - 获取任务")
			return task, nil
		}
		s.log.Debug().Uint("id", id).Msg("CACHE INVALID: L1 缓存类型异常，跳过")
	}
	s.log.Debug().Uint("id", id).Msg("CACHE MISS: L1 - 获取任务")

	// L2 缓存检查
	var task model.Task
	if found, err := s.cache.GetL2(ctx, key, &task); found && err == nil {
		s.log.Debug().Uint("id", id).Msg("CACHE HIT: L2 - 获取任务，回写到 L1")
		s.cache.SetL1(key, &task)
		return &task, nil
	}
	s.log.Debug().Uint("id", id).Msg("CACHE MISS: L2 - 获取任务")

	// 从数据库查询
	s.log.Debug().Uint("id", id).Msg("DB QUERY: SELECT 任务")
	taskPtr, err := s.repo.FindByID(id)
	if err != nil {
		s.log.Error().Uint("id", id).Err(err).Msg("数据库查询失败")
		return nil, err
	}
	s.log.Debug().Uint("id", id).Str("title", taskPtr.Title).Msg("DB SUCCESS: 查询到任务")

	// 写入缓存（L1 和 L2）
	s.cache.SetL1(key, taskPtr)
	s.cache.SetL2(ctx, key, taskPtr)

	return taskPtr, nil
}

// CreateTask 创建任务（清除缓存）
func (s *TaskService) CreateTask(title string, done bool) (*model.Task, error) {
	s.log.Debug().Str("title", title).Bool("done", done).Msg("SERVICE: 创建任务")
	task := &model.Task{Title: title, Done: done}
	if err := s.repo.Create(task); err != nil {
		s.log.Error().Err(err).Msg("创建任务失败")
		return nil, err
	}

	s.log.Debug().Uint("id", task.ID).Msg("SERVICE: 任务创建成功")
	// 清除任务列表缓存
	s.log.Debug().Msg("CACHE INVALIDATE: 清除所有任务列表缓存")
	s.cache.InvalidateAllTasks(context.Background())

	return task, nil
}

// UpdateTask 更新任务（清除缓存）
func (s *TaskService) UpdateTask(id uint, title string, done bool) (*model.Task, error) {
	ctx := context.Background()
	s.log.Debug().Uint("id", id).Str("title", title).Bool("done", done).Msg("SERVICE: 更新任务")

	// 先检查任务是否存在
	task, err := s.repo.FindByID(id)
	if err != nil {
		s.log.Error().Uint("id", id).Err(err).Msg("任务不存在")
		return nil, err
	}

	// 更新任务
	task.Title = title
	task.Done = done
	if err := s.repo.Update(task); err != nil {
		s.log.Error().Uint("id", id).Err(err).Msg("更新任务失败")
		return nil, err
	}

	s.log.Debug().Uint("id", id).Msg("SERVICE: 任务更新成功")
	// 清除相关缓存
	s.log.Debug().Uint("id", id).Msg("CACHE INVALIDATE: 清除任务缓存")
	s.cache.InvalidateTask(ctx, id)

	return task, nil
}

// CountTasks 统计任务总数
func (s *TaskService) CountTasks() (int64, error) {
	s.log.Debug().Msg("SERVICE: 统计任务总数")
	return s.repo.Count()
}

// DeleteTask 删除任务（清除缓存）
func (s *TaskService) DeleteTask(id uint) error {
	ctx := context.Background()
	s.log.Debug().Uint("id", id).Msg("SERVICE: 删除任务")

	// 先检查是否存在
	if _, err := s.repo.FindByID(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		s.log.Error().Uint("id", id).Err(err).Msg("查询任务失败")
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		s.log.Error().Uint("id", id).Err(err).Msg("删除任务失败")
		return err
	}

	s.log.Debug().Uint("id", id).Msg("SERVICE: 任务删除成功")
	// 清除相关缓存
	s.log.Debug().Uint("id", id).Msg("CACHE INVALIDATE: 清除任务缓存")
	s.cache.InvalidateTask(ctx, id)

	return nil
}
