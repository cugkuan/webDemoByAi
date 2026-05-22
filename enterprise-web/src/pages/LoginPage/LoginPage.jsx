import { useState, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/useAuth';
import { usePageTitle } from '../../hooks/usePageTitle';
import { EyeIcon, EyeOffIcon } from '../../components/Icons/EyeIcon';
import styles from './LoginPage.module.css';

/**
 * 用户登录页面
 *
 * 提供用户名/密码表单，含防重复提交和错误提示。
 */
export default function LoginPage() {
  usePageTitle('用户登录');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState({});
  const { login, loading } = useAuth();
  const navigate = useNavigate();
  const submittingRef = useRef(false);
  const usernameRef = useRef(null);

  /** 处理用户名输入变化 */
  const handleUsernameChange = (e) => {
    setUsername(e.target.value);
    if (fieldErrors.username) {
      setFieldErrors((prev) => ({ ...prev, username: '' }));
    }
  };

  /** 处理密码输入变化 */
  const handlePasswordChange = (e) => {
    setPassword(e.target.value);
    if (fieldErrors.password) {
      setFieldErrors((prev) => ({ ...prev, password: '' }));
    }
  };

  /** 提交登录 */
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (submittingRef.current) return;

    setError('');
    setFieldErrors({});

    // 字段级校验
    const errors = {};
    if (!username.trim()) {
      errors.username = '请输入用户名';
    }
    if (!password.trim()) {
      errors.password = '请输入密码';
    }
    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }

    submittingRef.current = true;
    const result = await login(username, password);
    submittingRef.current = false;

    if (result.success) {
      navigate('/tasks');
    } else {
      setError(result.message);
      // 密码错误时聚焦到用户名输入框
      if (usernameRef.current) {
        usernameRef.current.focus();
        usernameRef.current.select();
      }
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2 className={styles.title}>用户登录</h2>

        {/* 顶部错误提示 - 用于服务端返回的错误（如密码错误） */}
        {error && (
          <div className={styles.error} role="alert">
            <span className={styles.errorIcon}>⚠️</span>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} noValidate>
          <div className={`${styles.field}${fieldErrors.username ? ` ${styles.fieldError}` : ''}`}>
            <label className={styles.label} htmlFor="login-username">用户名</label>
            <input
              ref={usernameRef}
              id="login-username"
              className={`${styles.input}${fieldErrors.username ? ` ${styles.inputError}` : ''}`}
              type="text"
              value={username}
              onChange={handleUsernameChange}
              placeholder="请输入用户名"
              autoComplete="username"
              autoFocus
              aria-required="true"
              aria-invalid={!!fieldErrors.username}
              aria-describedby={fieldErrors.username ? 'username-error' : undefined}
            />
            {fieldErrors.username && (
              <p id="username-error" className={styles.fieldErrorText} role="alert">
                {fieldErrors.username}
              </p>
            )}
          </div>

          <div className={`${styles.field}${fieldErrors.password ? ` ${styles.fieldError}` : ''}`}>
            <label className={styles.label} htmlFor="login-password">密码</label>
            <div className={styles.passwordWrapper}>
              <input
                id="login-password"
                className={`${styles.input}${fieldErrors.password ? ` ${styles.inputError}` : ''}`}
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={handlePasswordChange}
                placeholder="请输入密码"
                autoComplete="current-password"
                aria-required="true"
                aria-invalid={!!fieldErrors.password}
                aria-describedby={fieldErrors.password ? 'password-error' : undefined}
              />
              <button
                type="button"
                className={styles.eyeBtn}
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? '隐藏密码' : '显示密码'}
                tabIndex={-1}
              >
                {showPassword ? <EyeOffIcon /> : <EyeIcon />}
              </button>
            </div>
            {fieldErrors.password && (
              <p id="password-error" className={styles.fieldErrorText} role="alert">
                {fieldErrors.password}
              </p>
            )}
          </div>

          <button
            type="submit"
            className={`${styles.button}${loading ? ` ${styles.buttonDisabled}` : ''}`}
            disabled={loading}
          >
            {loading ? '登录中...' : '登录'}
          </button>
        </form>

        <p className={styles.footer}>
          还没有账号？<Link to="/register" className={styles.link}>立即注册</Link>
        </p>
      </div>
    </div>
  );
}
