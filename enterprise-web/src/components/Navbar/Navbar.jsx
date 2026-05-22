import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/useAuth';
import styles from './Navbar.module.css';

/**
 * 应用顶部导航栏
 *
 * 根据登录状态显示不同的导航链接。
 */
export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <nav className={styles.nav}>
      <div className={styles.container}>
        <Link to="/" className={styles.brand}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginRight: 6, verticalAlign: 'middle' }}>
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <line x1="3" y1="9" x2="21" y2="9" />
            <line x1="9" y1="21" x2="9" y2="9" />
          </svg>
          任务管理系统
        </Link>

        <div className={styles.links}>
          {isAuthenticated ? (
            <>
              <Link to="/tasks" className={styles.link}>任务列表</Link>
              <Link to="/profile" className={styles.link}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginRight: 4, verticalAlign: 'middle' }}>
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                  <circle cx="12" cy="7" r="4" />
                </svg>
                {user?.username || '用户'}
              </Link>
              <button onClick={handleLogout} className={styles.logoutBtn}>
                退出登录
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className={styles.link}>登录</Link>
              <Link to="/register" className={styles.link}>注册</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
