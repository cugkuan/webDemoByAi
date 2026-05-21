import axios from 'axios';

/**
 * Axios 实例 - 统一 HTTP 客户端
 *
 * - baseURL: 从环境变量读取，开发环境默认 http://localhost:8080/api
 * - timeout: 10 秒超时
 * - 请求拦截器: 自动注入 JWT Bearer Token
 * - 响应拦截器: 统一错误处理，401 自动跳转登录
 */

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || 'http://localhost:8080/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
});

// 请求拦截器：自动注入 token
client.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器：统一错误处理
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // 避免在登录页循环跳转
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }
    const message = error.response?.data?.error || error.message || '请求失败';
    return Promise.reject(new Error(message));
  }
);

/**
 * 创建一个可取消的请求
 *
 * @param {string} method - HTTP 方法 (get/post/put/delete)
 * @param {string} url - 请求路径
 * @param {object} [data] - 请求体数据
 * @returns {{ promise: Promise, cancel: () => void }}
 *
 * @example
 * const { promise, cancel } = cancellableRequest('get', '/tasks');
 * promise.then(res => ...).catch(err => { if (!axios.isCancel(err)) ... });
 * // 组件卸载时取消
 * useEffect(() => () => cancel(), []);
 */
export function cancellableRequest(method, url, data) {
  const controller = new AbortController();
  const promise = client({
    method,
    url,
    data,
    signal: controller.signal,
  });
  return { promise, cancel: () => controller.abort() };
}

export default client;
