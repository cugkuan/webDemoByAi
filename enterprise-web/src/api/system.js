import client from './client';

/**
 * 系统管理 API
 *
 * @module api/system
 */

/**
 * 获取系统信息（版本、服务名、缓存配置等）
 *
 * @returns {Promise<object>} { data: { version, service, cache } }
 */
export function getSysInfo() {
  return client.get('/system/info');
}

/**
 * 获取系统运行状态（运行时间、缓存/数据库状态）
 *
 * @returns {Promise<object>} { data: { uptime, cache, database } }
 */
export function getSysStats() {
  return client.get('/system/stats');
}

/**
 * 健康检查
 *
 * @returns {Promise<object>} { data: { status, db, redis } }
 */
export function healthCheck() {
  return client.get('/health');
}
