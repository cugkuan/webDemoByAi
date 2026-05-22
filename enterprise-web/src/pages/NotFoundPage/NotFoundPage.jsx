import { Link } from 'react-router-dom';
import { usePageTitle } from '../../hooks/usePageTitle';
import styles from './NotFoundPage.module.css';

/**
 * 404 页面
 *
 * 当用户访问不存在的路由时显示。
 */
export default function NotFoundPage() {
  usePageTitle('页面未找到');

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1 className={styles.code}>404</h1>
        <h2 className={styles.title}>页面未找到</h2>
        <p className={styles.message}>
          抱歉，您访问的页面不存在或已被移除。
        </p>
        <Link to="/" className={styles.homeBtn}>
          返回首页
        </Link>
      </div>
    </div>
  );
}
