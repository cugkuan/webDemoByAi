import { useState, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/useAuth';
import {
  USERNAME_MIN_LENGTH,
  USERNAME_MAX_LENGTH,
  PASSWORD_MIN_LENGTH,
  PASSWORD_MAX_LENGTH,
} from '../config';
import styles from './RegisterPage.module.css';

/**
 * 用户注册页面
 *
 * 提供用户名/密码/确认密码表单，含前端校验和防重复提交。
 */
export default function RegisterPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [error, setError] = useState('');
  const { register, loading } = useAuth();
  const navigate = useNavigate();
  const submittingRef = useRef(false);

  /** 表单校验 */
  const validate = () => {
    if (!username.trim() || !password.trim()) {
      return '请填写所有字段';
    }
    if (username.length < USERNAME_MIN_LENGTH || username.length > USERNAME_MAX_LENGTH) {
      return `用户名长度必须在 ${USERNAME_MIN_LENGTH}-${USERNAME_MAX_LENGTH} 个字符之间`;
    }
    if (password.length < PASSWORD_MIN_LENGTH || password.length > PASSWORD_MAX_LENGTH) {
      return `密码长度必须在 ${PASSWORD_MIN_LENGTH}-${PASSWORD_MAX_LENGTH} 个字符之间`;
    }
    if (password !== confirmPassword) {
      return '两次输入的密码不一致';
    }
    return null;
  };

  /** 提交注册 */
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (submittingRef.current) return;

    setError('');

    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }

    submittingRef.current = true;
    const result = await register(username, password);
    submittingRef.current = false;

    if (result.success) {
      navigate('/tasks');
    } else {
      setError(result.message);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2 className={styles.title}>用户注册</h2>

        {error && <div className={styles.error} role="alert">{error}</div>}

        <form onSubmit={handleSubmit} noValidate>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="reg-username">
              用户名（{USERNAME_MIN_LENGTH}-{USERNAME_MAX_LENGTH} 个字符）
            </label>
            <input
              id="reg-username"
              className={styles.input}
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder={`${USERNAME_MIN_LENGTH}-${USERNAME_MAX_LENGTH} 个字符`}
              autoComplete="username"
              autoFocus
              aria-required="true"
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label} htmlFor="reg-password">
              密码（{PASSWORD_MIN_LENGTH}-{PASSWORD_MAX_LENGTH} 个字符）
            </label>
            <div className={styles.passwordWrapper}>
              <input
                id="reg-password"
                className={styles.input}
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={`${PASSWORD_MIN_LENGTH}-${PASSWORD_MAX_LENGTH} 个字符`}
                autoComplete="new-password"
                aria-required="true"
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? '隐藏密码' : '显示密码'}
                tabIndex={-1}
              >
                {showPassword ? '🙈' : '👁️'}
              </button>
            </div>
          </div>

          <div className={styles.field}>
            <label className={styles.label} htmlFor="reg-confirm">确认密码</label>
            <div className={styles.passwordWrapper}>
              <input
                id="reg-confirm"
                className={styles.input}
                type={showConfirm ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="再次输入密码"
                autoComplete="new-password"
                aria-required="true"
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowConfirm((v) => !v)}
                aria-label={showConfirm ? '隐藏密码' : '显示密码'}
                tabIndex={-1}
              >
                {showConfirm ? '🙈' : '👁️'}
              </button>
            </div>
          </div>

          <button
            type="submit"
            className={`${styles.button}${loading ? ` ${styles.buttonDisabled}` : ''}`}
            disabled={loading}
          >
            {loading ? '注册中...' : '注册'}
          </button>
        </form>

        <p className={styles.footer}>
          已有账号？<Link to="/login" className={styles.link}>立即登录</Link>
        </p>
      </div>
    </div>
  );
}
