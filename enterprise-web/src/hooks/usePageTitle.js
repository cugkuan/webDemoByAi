import { useEffect } from 'react';

/**
 * 动态设置页面标题的 Hook
 *
 * @param {string} title - 页面标题
 * @param {string} [suffix='企业级任务管理系统'] - 标题后缀
 *
 * @example
 * usePageTitle('任务列表');
 * // 页面标题变为 "任务列表 - 企业级任务管理系统"
 */
export function usePageTitle(title, suffix = '企业级任务管理系统') {
  useEffect(() => {
    document.title = title ? `${title} - ${suffix}` : suffix;
    return () => {
      document.title = suffix;
    };
  }, [title, suffix]);
}
