import client from './client';

/**
 * 任务管理 API
 *
 * @module api/tasks
 */

/**
 * 获取任务分页列表
 *
 * @param {number} page - 页码（从 1 开始）
 * @param {number} pageSize - 每页条数
 * @returns {Promise<object>} { data: { items: Task[], total: number } }
 */
export function getTasksPage(page, pageSize) {
  return client.get('/tasks', { params: { page, page_size: pageSize } });
}

/**
 * 创建新任务
 *
 * @param {string} title - 任务标题
 * @returns {Promise<object>} { data: Task }
 */
export function createTask(title) {
  return client.post('/tasks', { title });
}

/**
 * 更新任务
 *
 * @param {number} id - 任务 ID
 * @param {string} title - 任务标题
 * @param {boolean} done - 是否完成
 * @returns {Promise<object>} { data: Task }
 */
export function updateTask(id, title, done) {
  return client.put(`/tasks/${id}`, { title, done });
}

/**
 * 删除任务
 *
 * @param {number} id - 任务 ID
 * @returns {Promise<object>}
 */
export function deleteTask(id) {
  return client.delete(`/tasks/${id}`);
}
