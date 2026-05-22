import { describe, it, expect } from 'vitest';
import {
  PAGE_SIZE,
  MAX_VISIBLE_PAGES,
  USERNAME_MIN_LENGTH,
  USERNAME_MAX_LENGTH,
  PASSWORD_MIN_LENGTH,
  PASSWORD_MAX_LENGTH,
  API_TIMEOUT,
} from '../config';

describe('config', () => {
  it('PAGE_SIZE should be a positive integer', () => {
    expect(PAGE_SIZE).toBeGreaterThan(0);
    expect(Number.isInteger(PAGE_SIZE)).toBe(true);
  });

  it('MAX_VISIBLE_PAGES should be a positive integer', () => {
    expect(MAX_VISIBLE_PAGES).toBeGreaterThan(0);
    expect(Number.isInteger(MAX_VISIBLE_PAGES)).toBe(true);
  });

  it('USERNAME_MIN_LENGTH should be less than USERNAME_MAX_LENGTH', () => {
    expect(USERNAME_MIN_LENGTH).toBeLessThan(USERNAME_MAX_LENGTH);
  });

  it('PASSWORD_MIN_LENGTH should be less than PASSWORD_MAX_LENGTH', () => {
    expect(PASSWORD_MIN_LENGTH).toBeLessThan(PASSWORD_MAX_LENGTH);
  });

  it('API_TIMEOUT should be a positive integer', () => {
    expect(API_TIMEOUT).toBeGreaterThan(0);
    expect(Number.isInteger(API_TIMEOUT)).toBe(true);
  });
});
