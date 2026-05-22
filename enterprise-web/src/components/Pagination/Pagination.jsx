import styles from './Pagination.module.css';

export default function Pagination({ page, totalPages, total, onPageChange }) {
  if (totalPages <= 1) return null;

  const goToPage = (p) => {
    if (p >= 1 && p <= totalPages) {
      onPageChange(p);
    }
  };

  const pages = [];
  const maxVisible = 5;
  let start = Math.max(1, page - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages, start + maxVisible - 1);
  if (end - start + 1 < maxVisible) {
    start = Math.max(1, end - maxVisible + 1);
  }

  const btnClass = (isActive, isDisabled) => {
    let cls = styles.btn;
    if (isActive) cls += ` ${styles.btnActive}`;
    if (isDisabled) cls += ` ${styles.btnDisabled}`;
    return cls;
  };

  pages.push(
    <button
      key="prev"
      onClick={() => goToPage(page - 1)}
      disabled={page === 1}
      className={btnClass(false, page === 1)}
      aria-label="上一页"
    >
      ‹ 上一页
    </button>
  );

  if (start > 1) {
    pages.push(
      <button key={1} onClick={() => goToPage(1)} className={styles.btn} aria-label="第 1 页">1</button>
    );
    if (start > 2) {
      pages.push(<span key="dots1" className={styles.dots}>...</span>);
    }
  }

  for (let i = start; i <= end; i++) {
    pages.push(
      <button
        key={i}
        onClick={() => goToPage(i)}
        className={btnClass(i === page)}
        aria-label={`第 ${i} 页`}
        aria-current={i === page ? 'page' : undefined}
      >
        {i}
      </button>
    );
  }

  if (end < totalPages) {
    if (end < totalPages - 1) {
      pages.push(<span key="dots2" className={styles.dots}>...</span>);
    }
    pages.push(
      <button key={totalPages} onClick={() => goToPage(totalPages)} className={styles.btn} aria-label={`第 ${totalPages} 页`}>
        {totalPages}
      </button>
    );
  }

  pages.push(
    <button
      key="next"
      onClick={() => goToPage(page + 1)}
      disabled={page === totalPages}
      className={btnClass(false, page === totalPages)}
      aria-label="下一页"
    >
      下一页 ›
    </button>
  );

  return (
    <div className={styles.container}>
      {pages}
      <span className={styles.total}>共 {total} 条</span>
    </div>
  );
}
