import { useEffect, useRef } from 'react';
import styles from './ConfirmDialog.module.css';

/**
 * 确认对话框组件
 *
 * 用于删除等危险操作的二次确认，替代原生 window.confirm。
 *
 * @param {object} props
 * @param {boolean} props.open - 是否显示对话框
 * @param {string} props.title - 对话框标题
 * @param {string} props.message - 确认提示信息
 * @param {string} [props.confirmText='确定'] - 确认按钮文字
 * @param {string} [props.cancelText='取消'] - 取消按钮文字
 * @param {'danger'|'primary'} [props.variant='danger'] - 确认按钮样式
 * @param {() => void} props.onConfirm - 确认回调
 * @param {() => void} props.onCancel - 取消回调
 */
export default function ConfirmDialog({
  open,
  title = '确认操作',
  message = '确定要执行此操作吗？',
  confirmText = '确定',
  cancelText = '取消',
  variant = 'danger',
  onConfirm,
  onCancel,
}) {
  const confirmRef = useRef(null);

  // 打开时聚焦确认按钮，支持 Escape 关闭
  useEffect(() => {
    if (!open) return;

    const timer = setTimeout(() => confirmRef.current?.focus(), 50);

    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        onCancel?.();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      clearTimeout(timer);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [open, onCancel]);

  if (!open) return null;

  return (
    <div className={styles.overlay} onClick={onCancel} role="dialog" aria-modal="true">
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <h3 className={styles.title}>{title}</h3>
        <p className={styles.message}>{message}</p>
        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onCancel}>
            {cancelText}
          </button>
          <button
            ref={confirmRef}
            className={`${styles.confirmBtn} ${variant === 'danger' ? styles.danger : styles.primary}`}
            onClick={onConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
