import { useEffect, useState } from 'react';
import { useAuth } from '../../context/useAuth';
import { getSysInfo, getSysStats, healthCheck } from '../../api/system';
import { ProfileSkeleton } from '../../components/Skeleton/Skeleton';
import { usePageTitle } from '../../hooks/usePageTitle';
import styles from './ProfilePage.module.css';

export default function ProfilePage() {
  usePageTitle('个人中心');
  const { user } = useAuth();
  const [sysInfo, setSysInfo] = useState(null);
  const [sysStats, setSysStats] = useState(null);
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    const fetchData = async () => {
      setLoading(true);
      setError('');
      try {
        const [infoRes, statsRes, healthRes] = await Promise.all([
          getSysInfo(),
          getSysStats(),
          healthCheck(),
        ]);
        if (!cancelled) {
          setSysInfo(infoRes.data);
          setSysStats(statsRes.data);
          setHealth(healthRes.data);
        }
      } catch (err) {
        if (!cancelled) {
          const msg = err?.message || '获取系统信息失败';
          console.warn('[Profile] 获取系统信息失败:', msg);
          setError(msg);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };
    fetchData();
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return (
      <div className={styles.container}>
        <ProfileSkeleton />
      </div>
    );
  }

  const statusClass = (status) => {
    return status === 'ok' ? styles.statusOk : styles.statusError;
  };

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        {error && (
          <div className={styles.error}>
            {error}
            <button onClick={() => setError('')} className={styles.closeBtn}>✕</button>
          </div>
        )}

        {/* 用户信息卡片 */}
        <div className={styles.card}>
          <h3 className={styles.cardTitle}>👤 用户信息</h3>
          {user ? (
            <div className={styles.infoGrid}>
              <div className={styles.infoItem}>
                <span className={styles.label}>用户 ID</span>
                <span className={styles.value}>{user.id}</span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>用户名</span>
                <span className={styles.value}>{user.username}</span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>创建时间</span>
                <span className={styles.value}>
                  {user.created_at ? new Date(user.created_at).toLocaleString('zh-CN') : '-'}
                </span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>更新时间</span>
                <span className={styles.value}>
                  {user.updated_at ? new Date(user.updated_at).toLocaleString('zh-CN') : '-'}
                </span>
              </div>
            </div>
          ) : (
            <p className={styles.noData}>未获取到用户信息</p>
          )}
        </div>

        {/* 系统信息卡片 */}
        <div className={styles.card}>
          <h3 className={styles.cardTitle}>⚙️ 系统信息</h3>
          {sysInfo ? (
            <div className={styles.infoGrid}>
              <div className={styles.infoItem}>
                <span className={styles.label}>服务版本</span>
                <span className={styles.value}>{sysInfo.version}</span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>服务名称</span>
                <span className={styles.value}>{sysInfo.service}</span>
              </div>
              {sysInfo.cache && (
                <>
                  <div className={styles.infoItem}>
                    <span className={styles.label}>L1 缓存 TTL</span>
                    <span className={styles.value}>{sysInfo.cache.L1_TTL}</span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.label}>L2 缓存 TTL</span>
                    <span className={styles.value}>{sysInfo.cache.L2_TTL}</span>
                  </div>
                </>
              )}
            </div>
          ) : (
            <p className={styles.noData}>未获取到系统信息</p>
          )}
        </div>

        {/* 系统状态卡片 */}
        <div className={styles.card}>
          <h3 className={styles.cardTitle}>📊 系统状态</h3>
          {sysStats ? (
            <div className={styles.infoGrid}>
              <div className={styles.infoItem}>
                <span className={styles.label}>运行状态</span>
                <span className={styles.value}>{sysStats.uptime}</span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>缓存状态</span>
                <span className={styles.value}>{sysStats.cache?.status || '-'}</span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>数据库状态</span>
                <span className={styles.value}>{sysStats.database?.status || '-'}</span>
              </div>
            </div>
          ) : (
            <p className={styles.noData}>未获取到系统状态</p>
          )}
        </div>

        {/* 健康检查卡片 */}
        <div className={styles.card}>
          <h3 className={styles.cardTitle}>❤️ 健康检查</h3>
          {health ? (
            <div className={styles.infoGrid}>
              <div className={styles.infoItem}>
                <span className={styles.label}>整体状态</span>
                <span className={`${styles.statusBadge} ${statusClass(health.status)}`}>
                  {health.status === 'ok' ? '正常' : '异常'}
                </span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>数据库</span>
                <span className={`${styles.statusBadge} ${statusClass(health.db)}`}>
                  {health.db === 'ok' ? '正常' : health.db}
                </span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>Redis</span>
                <span className={`${styles.statusBadge} ${health.redis === 'ok' ? styles.statusOk : styles.statusWarn}`}>
                  {health.redis === 'ok' ? '正常' : health.redis}
                </span>
              </div>
            </div>
          ) : (
            <p className={styles.noData}>未获取到健康检查信息</p>
          )}
        </div>
      </div>
    </div>
  );
}
