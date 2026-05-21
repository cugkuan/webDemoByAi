import { useContext } from 'react';
import { AuthContext } from './AuthContext';

/**
 * 认证 Hook
 *
 * 提供用户认证状态和操作方法。
 * 必须在 AuthProvider 内部使用。
 *
 * @returns {{ user, token, loading, isAuthenticated, login, register, logout }}
 *
 * @example
 * const { user, isAuthenticated, login, logout } = useAuth();
 */
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
