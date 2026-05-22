import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { usePageTitle } from '../hooks/usePageTitle';

describe('usePageTitle', () => {
  beforeEach(() => {
    document.title = '企业级任务管理系统';
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should set document title with suffix', () => {
    renderHook(() => usePageTitle('任务列表'));
    expect(document.title).toBe('任务列表 - 企业级任务管理系统');
  });

  it('should set document title with custom suffix', () => {
    renderHook(() => usePageTitle('登录', 'My App'));
    expect(document.title).toBe('登录 - My App');
  });

  it('should set document title to suffix when title is empty', () => {
    renderHook(() => usePageTitle(''));
    expect(document.title).toBe('企业级任务管理系统');
  });

  it('should restore title on unmount', () => {
    const { unmount } = renderHook(() => usePageTitle('任务列表'));
    expect(document.title).toBe('任务列表 - 企业级任务管理系统');
    unmount();
    expect(document.title).toBe('企业级任务管理系统');
  });
});
