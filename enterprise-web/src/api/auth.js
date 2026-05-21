import client from './client';

/**
 * 用户认证 API
 *
 * @module api/auth
 */

/**
 * 用户登录
 *
 * @param {string} username - 用户名
 * @param {string} password - 密码
 * @returns {Promise<object>} { data: { token, user } }
 */
export function login(username, password) {
  return client.post('/auth/login', { username, password });
}

/**
 * 用户注册
 *
 * @param {string} username - 用户名
 * @param {string} password - 密码
 * @returns {Promise<object>} { data: { token, user } }
 */
export function register(username, password) {
  return client.post('/auth/register', { username, password });
}
