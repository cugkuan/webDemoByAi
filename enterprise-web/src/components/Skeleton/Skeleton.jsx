import styles from './Skeleton.module.css';

export function TaskSkeleton() {
  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.headerTitle} />
        <div className={styles.headerBadge} />
      </div>
      <div className={styles.form}>
        <div className={styles.formInput} />
        <div className={styles.formBtn} />
      </div>
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className={styles.item} />
      ))}
    </div>
  );
}

export function ProfileSkeleton() {
  return (
    <div className={styles.profileContainer}>
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className={styles.profileCard} />
      ))}
    </div>
  );
}
