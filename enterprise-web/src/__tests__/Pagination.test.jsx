import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import Pagination from '../components/Pagination/Pagination';

describe('Pagination', () => {
  it('should not render when totalPages <= 1', () => {
    const { container } = render(
      <Pagination page={1} totalPages={1} total={5} onPageChange={() => {}} />
    );
    expect(container.innerHTML).toBe('');
  });

  it('should render page buttons', () => {
    render(
      <Pagination page={1} totalPages={5} total={50} onPageChange={() => {}} />
    );
    expect(screen.getByLabelText('第 1 页')).toBeDefined();
    expect(screen.getByLabelText('第 5 页')).toBeDefined();
    expect(screen.getByLabelText('上一页')).toBeDefined();
    expect(screen.getByLabelText('下一页')).toBeDefined();
  });

  it('should highlight current page', () => {
    render(
      <Pagination page={3} totalPages={5} total={50} onPageChange={() => {}} />
    );
    const currentBtn = screen.getByLabelText('第 3 页');
    expect(currentBtn.getAttribute('aria-current')).toBe('page');
  });

  it('should call onPageChange when clicking a page', () => {
    const onPageChange = vi.fn();
    render(
      <Pagination page={1} totalPages={5} total={50} onPageChange={onPageChange} />
    );
    fireEvent.click(screen.getByLabelText('第 2 页'));
    expect(onPageChange).toHaveBeenCalledWith(2);
  });

  it('should disable prev button on first page', () => {
    render(
      <Pagination page={1} totalPages={5} total={50} onPageChange={() => {}} />
    );
    expect(screen.getByLabelText('上一页').hasAttribute('disabled')).toBe(true);
  });

  it('should disable next button on last page', () => {
    render(
      <Pagination page={5} totalPages={5} total={50} onPageChange={() => {}} />
    );
    expect(screen.getByLabelText('下一页').hasAttribute('disabled')).toBe(true);
  });

  it('should show total count', () => {
    render(
      <Pagination page={1} totalPages={5} total={50} onPageChange={() => {}} />
    );
    expect(screen.getByText('共 50 条')).toBeDefined();
  });

  it('should show ellipsis for many pages', () => {
    render(
      <Pagination page={10} totalPages={20} total={200} onPageChange={() => {}} />
    );
    const dots = screen.getAllByText('...');
    expect(dots.length).toBeGreaterThanOrEqual(1);
  });
});
