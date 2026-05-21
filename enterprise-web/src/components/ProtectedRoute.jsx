import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/useAuth';

/**
 * 路由守卫组件
 *
 * 如果用户未登录，自动跳转到登录页。
 * 如果已登录，渲染子组件。
 *
 * @param {object} props
 * @param {React.ReactNode} props.children - 需要保护的路由内容
 */
export default function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return null;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
}
