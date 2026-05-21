import { createContext } from 'react';

/**
 * 认证上下文
 *
 * 管理用户登录状态、token 持久化，提供 login/register/logout 方法。
 */
export const AuthContext = createContext(null);
