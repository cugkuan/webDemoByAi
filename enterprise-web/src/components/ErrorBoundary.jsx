import { Component } from 'react';
import styles from './ErrorBoundary.module.css';

export default class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('[ErrorBoundary]', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className={styles.container}>
          <div className={styles.card}>
            <h2 className={styles.title}>⚠️ 页面出错了</h2>
            <p className={styles.message}>
              {this.state.error?.message || '发生了未知错误'}
            </p>
            <button onClick={this.handleRetry} className={styles.retryBtn}>
              重试
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
