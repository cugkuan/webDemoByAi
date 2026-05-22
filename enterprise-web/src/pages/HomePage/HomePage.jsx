import { Link } from 'react-router-dom';
import { useAuth } from '../../context/useAuth';
import { usePageTitle } from '../../hooks/usePageTitle';
import styles from './HomePage.module.css';

export default function HomePage() {
  usePageTitle('');
  const { isAuthenticated } = useAuth();

  return (
    <div className={styles.container}>
      <div className={styles.hero}>
        <h1 className={styles.title}>📋 企业级任务管理系统</h1>
        <p className={styles.subtitle}>
          基于 Go + Gin + GORM + Redis 构建的高性能任务管理平台
        </p>

        <div className={styles.features}>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>🔐</span>
            <h3>用户认证</h3>
            <p>JWT + Redis 白名单，安全可靠</p>
          </div>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>⚡</span>
            <h3>二级缓存</h3>
            <p>L1 内存缓存 + L2 Redis 缓存，极致性能</p>
          </div>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>📊</span>
            <h3>任务管理</h3>
            <p>创建、编辑、完成、删除，完整 CRUD</p>
          </div>
          <div className={styles.feature}>
            <span className={styles.featureIcon}>🛡️</span>
            <h3>限流保护</h3>
            <p>内置限流中间件，防止恶意请求</p>
          </div>
        </div>

        <div className={styles.actions}>
          {isAuthenticated ? (
            <Link to="/tasks" className={styles.primaryBtn}>
              进入任务管理 →
            </Link>
          ) : (
            <>
              <Link to="/login" className={styles.primaryBtn}>
                立即登录
              </Link>
              <Link to="/register" className={styles.secondaryBtn}>
                注册账号
              </Link>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
