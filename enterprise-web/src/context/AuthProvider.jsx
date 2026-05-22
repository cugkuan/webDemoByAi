import { useState, useCallback, useEffect } from 'react';
import { AuthContext } from './AuthContext';
import { login as apiLogin, register as apiRegister } from '../api/auth';
import { clearAuth, STORAGE_KEYS } from '../api/client';

/**
 * 认证 Provider
 *
 * 提供用户认证状态和操作方法。
 */
export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEYS.USER);
      return saved ? JSON.parse(saved) : null;
    } catch {
      return null;
    }
  });

  const [token, setToken] = useState(() => localStorage.getItem(STORAGE_KEYS.TOKEN));
  const [loading, setLoading] = useState(false);

  const isAuthenticated = !!token;

  // token 变化时同步到 localStorage
  useEffect(() => {
    if (token) {
      localStorage.setItem(STORAGE_KEYS.TOKEN, token);
    } else {
      localStorage.removeItem(STORAGE_KEYS.TOKEN);
    }
  }, [token]);

  // user 变化时同步到 localStorage
  useEffect(() => {
    if (user) {
      localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
    } else {
      localStorage.removeItem(STORAGE_KEYS.USER);
    }
  }, [user]);

  /** 登录 */
  const login = useCallback(async (username, password) => {
    setLoading(true);
    try {
      const res = await apiLogin(username, password);
      const { token: newToken, user: userData } = res.data.data;
      setToken(newToken);
      setUser(userData);
      return { success: true };
    } catch (err) {
      return { success: false, message: err.message };
    } finally {
      setLoading(false);
    }
  }, []);

  /** 注册 */
  const register = useCallback(async (username, password) => {
    setLoading(true);
    try {
      const res = await apiRegister(username, password);
      const { token: newToken, user: userData } = res.data.data;
      setToken(newToken);
      setUser(userData);
      return { success: true };
    } catch (err) {
      return { success: false, message: err.message };
    } finally {
      setLoading(false);
    }
  }, []);

  /** 退出登录 */
  const logout = useCallback(async () => {
    clearAuth();
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, loading, isAuthenticated, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
