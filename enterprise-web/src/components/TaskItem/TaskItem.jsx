import { memo } from 'react';
import styles from './TaskItem.module.css';

/**
 * 单个任务项组件
 *
 * @param {object} props
 * @param {object} props.task - 任务对象 { id, title, done }
 * @param {boolean} props.isEditing - 是否正在编辑
 * @param {string} props.editTitle - 编辑中的标题
 * @param {function} props.onToggleDone - 切换完成状态
 * @param {function} props.onStartEdit - 开始编辑
 * @param {function} props.onEditTitleChange - 编辑标题变化
 * @param {function} props.onSaveEdit - 保存编辑
 * @param {function} props.onCancelEdit - 取消编辑
 * @param {function} props.onDelete - 删除任务
 * @param {function} props.onViewDetail - 查看详情
 */
function TaskItem({
  task,
  isEditing,
  editTitle,
  onToggleDone,
  onStartEdit,
  onEditTitleChange,
  onSaveEdit,
  onCancelEdit,
  onDelete,
  onViewDetail,
}) {
  return (
    <li
      className={`${styles.item}${task.done ? ` ${styles.done}` : ''}`}
    >
      <input
        type="checkbox"
        checked={task.done}
        onChange={() => onToggleDone(task)}
        className={styles.checkbox}
        aria-label={`标记任务"${task.title}"为${task.done ? '未完成' : '已完成'}`}
      />

      {isEditing ? (
        <div className={styles.editGroup}>
          <input
            className={styles.editInput}
            type="text"
            value={editTitle}
            onChange={(e) => onEditTitleChange(e.target.value)}
            autoFocus
            aria-label="编辑任务标题"
          />
          <button
            onClick={() => onSaveEdit(task.id)}
            className={styles.saveBtn}
            aria-label="保存修改"
          >
            保存
          </button>
          <button
            onClick={onCancelEdit}
            className={styles.cancelBtn}
            aria-label="取消编辑"
          >
            取消
          </button>
        </div>
      ) : (
        <>
          <span
            className={`${styles.title}${task.done ? ` ${styles.titleDone}` : ''}`}
            onClick={() => onViewDetail?.(task.id)}
            title="点击查看详情"
          >
            {task.title}
          </span>

          <div className={styles.actions}>
            <button
              onClick={() => onStartEdit(task)}
              className={styles.iconBtn}
              disabled={task.done}
              title="编辑"
              aria-label={`编辑任务"${task.title}"`}
            >
              ✏️
            </button>
            <button
              onClick={() => onDelete(task.id)}
              className={styles.iconBtn}
              title="删除"
              aria-label={`删除任务"${task.title}"`}
            >
              🗑️
            </button>
          </div>
        </>
      )}
    </li>
  );
}

export default memo(TaskItem);
