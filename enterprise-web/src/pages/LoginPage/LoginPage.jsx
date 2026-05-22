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
  const { login, loading } = useAuth();
  const navigate = useNavigate();
  const submittingRef = useRef(false);

  /** 提交登录 */
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (submittingRef.current) return;

    setError('');

    if (!username.trim() || !password.trim()) {
      setError('请输入用户名和密码');
      return;
    }

    submittingRef.current = true;
    const result = await login(username, password);
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
        <h2 className={styles.title}>用户登录</h2>

        {error && <div className={styles.error} role="alert">{error}</div>}

        <form onSubmit={handleSubmit} noValidate>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="login-username">用户名</label>
            <input
              id="login-username"
              className={styles.input}
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="请输入用户名"
              autoComplete="username"
              autoFocus
              aria-required="true"
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label} htmlFor="login-password">密码</label>
            <div className={styles.passwordWrapper}>
              <input
                id="login-password"
                className={styles.input}
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                autoComplete="current-password"
                aria-required="true"
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
