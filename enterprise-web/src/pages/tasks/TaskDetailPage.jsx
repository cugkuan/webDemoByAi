import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { getTaskById, updateTask, deleteTask } from '../../api/tasks';
import ConfirmDialog from '../../components/ConfirmDialog/ConfirmDialog';
import { usePageTitle } from '../../hooks/usePageTitle';
import styles from './TaskDetailPage.module.css';

/**
 * 任务详情页面
 *
 * 展示单个任务的详细信息，支持编辑标题、切换完成状态和删除。
 */
export default function TaskDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  usePageTitle('任务详情');

  const [task, setTask] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  /** 加载任务详情 */
  useEffect(() => {
    let cancelled = false;
    const loadTask = async () => {
      setLoading(true);
      setError('');
      try {
        const res = await getTaskById(id);
        if (!cancelled) {
          setTask(res.data.data);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err.message || '获取任务详情失败');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };
    loadTask();
    return () => { cancelled = true; };
  }, [id]);

  /** 切换完成状态 */
  const handleToggleDone = async () => {
    if (!task) return;
    try {
      const res = await updateTask(task.id, task.title, !task.done);
      setTask(res.data.data);
    } catch (err) {
      setError(err.message || '更新任务失败');
    }
  };

  /** 开始编辑 */
  const handleStartEdit = () => {
    setEditTitle(task.title);
    setEditing(true);
  };

  /** 保存编辑 */
  const handleSaveEdit = async () => {
    if (!editTitle.trim() || !task) return;
    try {
      const res = await updateTask(task.id, editTitle.trim(), task.done);
      setTask(res.data.data);
      setEditing(false);
    } catch (err) {
      setError(err.message || '更新任务失败');
    }
  };

  /** 取消编辑 */
  const handleCancelEdit = () => {
    setEditing(false);
    setEditTitle('');
  };

  /** 确认删除 */
  const handleDeleteConfirm = async () => {
    if (!task) return;
    try {
      await deleteTask(task.id);
      navigate('/tasks', { replace: true });
    } catch (err) {
      setError(err.message || '删除任务失败');
      setShowDeleteConfirm(false);
    }
  };

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.skeleton}>
          <div className={styles.skeletonLine} style={{ width: '60%' }} />
          <div className={styles.skeletonLine} style={{ width: '40%' }} />
          <div className={styles.skeletonLine} style={{ width: '80%' }} />
        </div>
      </div>
    );
  }

  if (error && !task) {
    return (
      <div className={styles.container}>
        <div className={styles.errorCard}>
          <p>{error}</p>
          <Link to="/tasks" className={styles.backLink}>← 返回任务列表</Link>
        </div>
      </div>
    );
  }

  if (!task) {
    return (
      <div className={styles.container}>
        <div className={styles.errorCard}>
          <p>任务不存在</p>
          <Link to="/tasks" className={styles.backLink}>← 返回任务列表</Link>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        {/* 顶部导航 */}
        <div className={styles.nav}>
          <Link to="/tasks" className={styles.backLink}>← 返回任务列表</Link>
          <span className={styles.taskId}>ID: {task.id}</span>
        </div>

        {/* 错误提示 */}
        {error && (
          <div className={styles.error}>
            {error}
            <button onClick={() => setError('')} className={styles.closeBtn}>✕</button>
          </div>
        )}

        {/* 任务标题区域 */}
        <div className={styles.titleSection}>
          <input
            type="checkbox"
            checked={task.done}
            onChange={handleToggleDone}
            className={styles.checkbox}
            aria-label={`标记任务为${task.done ? '未完成' : '已完成'}`}
          />

          {editing ? (
            <div className={styles.editGroup}>
              <input
                className={styles.editInput}
                type="text"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                autoFocus
                aria-label="编辑任务标题"
              />
              <button onClick={handleSaveEdit} className={styles.saveBtn}>保存</button>
              <button onClick={handleCancelEdit} className={styles.cancelBtn}>取消</button>
            </div>
          ) : (
            <h1 className={`${styles.title}${task.done ? ` ${styles.titleDone}` : ''}`}>
              {task.title}
            </h1>
          )}
        </div>

        {/* 任务状态 */}
        <div className={styles.meta}>
          <div className={styles.metaItem}>
            <span className={styles.metaLabel}>状态</span>
            <span className={`${styles.status} ${task.done ? styles.statusDone : styles.statusPending}`}>
              {task.done ? '✅ 已完成' : '⏳ 待完成'}
            </span>
          </div>
        </div>

        {/* 操作按钮 */}
        <div className={styles.actions}>
          {!editing && (
            <button onClick={handleStartEdit} className={styles.editBtn} disabled={task.done}>
              ✏️ 编辑标题
            </button>
          )}
          <button onClick={() => setShowDeleteConfirm(true)} className={styles.deleteBtn}>
            🗑️ 删除任务
          </button>
        </div>
      </div>

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={showDeleteConfirm}
        title="删除任务"
        message={`确定要删除任务"${task.title}"吗？此操作不可恢复。`}
        confirmText="删除"
        cancelText="取消"
        variant="danger"
        onConfirm={handleDeleteConfirm}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </div>
  );
}
