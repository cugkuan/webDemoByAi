import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/useAuth';
import styles from './Navbar.module.css';

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
          📋 任务管理系统
        </Link>

        <div className={styles.links}>
          {isAuthenticated ? (
            <>
              <Link to="/tasks" className={styles.link}>任务列表</Link>
              <Link to="/profile" className={styles.link}>
                👤 {user?.username || '用户'}
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
