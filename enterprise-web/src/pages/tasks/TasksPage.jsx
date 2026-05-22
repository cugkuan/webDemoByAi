import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { getTasksPage, createTask, updateTask, deleteTask } from '../../api/tasks';
import Pagination from '../../components/Pagination/Pagination';
import { TaskSkeleton } from '../../components/Skeleton/Skeleton';
import TaskItem from '../../components/TaskItem/TaskItem';
import ConfirmDialog from '../../components/ConfirmDialog/ConfirmDialog';
import { usePageTitle } from '../../hooks/usePageTitle';
import { PAGE_SIZE } from '../../config';
import styles from './TasksPage.module.css';

/**
 * 任务管理页面
 *
 * 支持任务的创建、编辑、完成/未完成切换、删除和分页浏览。
 */
export default function TasksPage() {
  usePageTitle('任务列表');
  const navigate = useNavigate();
  const [tasks, setTasks] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [newTitle, setNewTitle] = useState('');
  const [editingId, setEditingId] = useState(null);
  const [editTitle, setEditTitle] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [deleteTarget, setDeleteTarget] = useState(null);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  /** 获取任务列表（统一的数据加载函数） */
  const fetchTasks = useCallback(async (targetPage) => {
    setLoading(true);
    setError('');
    try {
      const res = await getTasksPage(targetPage, PAGE_SIZE);
      setTasks(res.data.data.items || []);
      setTotal(res.data.data.total || 0);
    } catch (err) {
      setError(err.message || '获取任务列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 首次加载 + 页码变化时重新获取数据
  useEffect(() => {
    let cancelled = false;
    const loadData = async () => {
      setLoading(true);
      setError('');
      try {
        const res = await getTasksPage(page, PAGE_SIZE);
        if (!cancelled) {
          setTasks(res.data.data.items || []);
          setTotal(res.data.data.total || 0);
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
    } catch (err) {
      setError(err.message || '创建任务失败');
    }
  };

  /** 切换任务完成状态 */
  const handleToggleDone = async (task) => {
    try {
      await updateTask(task.id, task.title, !task.done);
      // 乐观更新：直接修改本地状态，避免重新请求
      setTasks((prev) =>
        prev.map((t) => (t.id === task.id ? { ...t, done: !t.done } : t))
      );
    } catch (err) {
      setError(err.message || '更新任务失败');
      // 失败时回退：重新获取当前页数据
      fetchTasks(page);
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
      // 乐观更新
      setTasks((prev) =>
        prev.map((t) => (t.id === id ? { ...t, title: editTitle.trim() } : t))
      );
    } catch (err) {
      setError(err.message || '更新任务失败');
      fetchTasks(page);
    }
  };

  /** 取消编辑 */
  const handleCancelEdit = () => {
    setEditingId(null);
    setEditTitle('');
  };

  /** 点击删除按钮，弹出确认对话框 */
  const handleDeleteClick = (id) => {
    setDeleteTarget(id);
  };

  /** 确认删除 */
  const handleDeleteConfirm = async () => {
    if (deleteTarget === null) return;
    try {
      await deleteTask(deleteTarget);
      setDeleteTarget(null);
      // 如果当前页删完了且不是第一页，回到上一页
      const newTotal = total - 1;
      const newTotalPages = Math.max(1, Math.ceil(newTotal / PAGE_SIZE));
      const targetPage = page > newTotalPages ? newTotalPages : page;
      setPage(targetPage);
    } catch (err) {
      setError(err.message || '删除任务失败');
      setDeleteTarget(null);
    }
  };

  /** 查看任务详情 */
  const handleViewDetail = (id) => {
    navigate(`/tasks/${id}`);
  };

  /** 取消删除 */
  const handleDeleteCancel = () => {
    setDeleteTarget(null);
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
                onDelete={handleDeleteClick}
                onViewDetail={handleViewDetail}
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

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={deleteTarget !== null}
        title="删除任务"
        message="确定要删除这个任务吗？此操作不可恢复。"
        confirmText="删除"
        cancelText="取消"
        variant="danger"
        onConfirm={handleDeleteConfirm}
        onCancel={handleDeleteCancel}
      />
    </div>
  );
}
