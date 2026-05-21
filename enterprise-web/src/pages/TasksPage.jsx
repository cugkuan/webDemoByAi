import { useState, useEffect, useCallback } from 'react';
import { getTasksPage, createTask, updateTask, deleteTask } from '../api/tasks';
import Pagination from '../components/Pagination';
import { TaskSkeleton } from '../components/Skeleton';
import TaskItem from '../components/TaskItem';
import { PAGE_SIZE } from '../config';
import styles from './TasksPage.module.css';

/**
 * 任务管理页面
 *
 * 支持任务的创建、编辑、完成/未完成切换、删除和分页浏览。
 */
export default function TasksPage() {
  const [tasks, setTasks] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [newTitle, setNewTitle] = useState('');
  const [editingId, setEditingId] = useState(null);
  const [editTitle, setEditTitle] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  /** 获取任务列表 */
  const fetchTasks = useCallback(async (p) => {
    setLoading(true);
    try {
      const res = await getTasksPage(p, PAGE_SIZE);
      setTasks(res.data.items || []);
      setTotal(res.data.total || 0);
    } catch (err) {
      setError(err.message || '获取任务列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 页码变化时重新加载
  useEffect(() => {
    let cancelled = false;
    const loadData = async () => {
      setLoading(true);
      try {
        const res = await getTasksPage(page, PAGE_SIZE);
        if (!cancelled) {
          setTasks(res.data.items || []);
          setTotal(res.data.total || 0);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err.message || '获取任务列表失败');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };
    loadData();
    return () => { cancelled = true; };
  }, [page]);

  /** 创建新任务 */
  const handleCreate = async (e) => {
    e.preventDefault();
    if (!newTitle.trim()) return;

    try {
      await createTask(newTitle.trim());
      setNewTitle('');
      setPage(1);
      await fetchTasks(1);
    } catch (err) {
      setError(err.message || '创建任务失败');
    }
  };

  /** 切换任务完成状态 */
  const handleToggleDone = async (task) => {
    try {
      await updateTask(task.id, task.title, !task.done);
      await fetchTasks(page);
    } catch (err) {
      setError(err.message || '更新任务失败');
    }
  };

  /** 开始编辑任务 */
  const handleStartEdit = (task) => {
    setEditingId(task.id);
    setEditTitle(task.title);
  };

  /** 编辑标题变化 */
  const handleEditTitleChange = (value) => {
    setEditTitle(value);
  };

  /** 保存编辑 */
  const handleSaveEdit = async (id) => {
    if (!editTitle.trim()) return;
    try {
      await updateTask(id, editTitle.trim(), false);
      setEditingId(null);
      setEditTitle('');
      await fetchTasks(page);
    } catch (err) {
      setError(err.message || '更新任务失败');
    }
  };

  /** 取消编辑 */
  const handleCancelEdit = () => {
    setEditingId(null);
    setEditTitle('');
  };

  /** 删除任务 */
  const handleDelete = async (id) => {
    if (!window.confirm('确定要删除这个任务吗？')) return;
    try {
      await deleteTask(id);
      // 如果当前页删完了且不是第一页，回到上一页
      const newTotal = total - 1;
      const newTotalPages = Math.max(1, Math.ceil(newTotal / PAGE_SIZE));
      const targetPage = page > newTotalPages ? newTotalPages : page;
      setPage(targetPage);
      await fetchTasks(targetPage);
    } catch (err) {
      setError(err.message || '删除任务失败');
    }
  };

  // 首次加载显示骨架屏
  if (loading && tasks.length === 0) {
    return (
      <div className={styles.container}>
        <TaskSkeleton />
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        <div className={styles.header}>
          <h2 className={styles.title}>任务管理</h2>
          <span className={styles.badge}>共 {total} 个任务</span>
        </div>

        {error && (
          <div className={styles.error}>
            {error}
            <button onClick={() => setError('')} className={styles.closeBtn}>✕</button>
          </div>
        )}

        {/* 创建任务表单 */}
        <form onSubmit={handleCreate} className={styles.createForm}>
          <input
            className={styles.createInput}
            type="text"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            placeholder="输入新任务标题..."
            aria-label="新任务标题"
          />
          <button
            type="submit"
            className={styles.createBtn}
            disabled={!newTitle.trim()}
          >
            ➕ 添加任务
          </button>
        </form>

        {/* 任务列表 */}
        {tasks.length === 0 ? (
          <div className={styles.empty}>
            <p>暂无任务，快来创建第一个任务吧！</p>
          </div>
        ) : (
          <ul className={styles.taskList} aria-label="任务列表">
            {tasks.map((task) => (
              <TaskItem
                key={task.id}
                task={task}
                isEditing={editingId === task.id}
                editTitle={editTitle}
                onToggleDone={handleToggleDone}
                onStartEdit={handleStartEdit}
                onEditTitleChange={handleEditTitleChange}
                onSaveEdit={handleSaveEdit}
                onCancelEdit={handleCancelEdit}
                onDelete={handleDelete}
              />
            ))}
          </ul>
        )}

        {/* 分页 */}
        <Pagination
          page={page}
          totalPages={totalPages}
          total={total}
          onPageChange={setPage}
        />
      </div>
    </div>
  );
}
