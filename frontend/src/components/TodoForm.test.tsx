import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TodoForm } from './TodoForm';

describe('TodoForm', () => {
  it('disables submit button when input is empty', () => {
    render(<TodoForm onAdd={vi.fn()} />);

    expect(screen.getByRole('button', { name: /add/i })).toBeDisabled();
  });

  it('acts as a controlled input', () => {
    render(<TodoForm onAdd={vi.fn()} />);

    const input = screen.getByRole('textbox', { name: /todo title/i });
    fireEvent.change(input, { target: { value: 'Buy milk' } });

    expect(input).toHaveValue('Buy milk');
    expect(screen.getByRole('button', { name: /add/i })).toBeEnabled();
  });

  it('calls onAdd with title and clears input on submit', () => {
    const onAdd = vi.fn();
    render(<TodoForm onAdd={onAdd} />);

    const input = screen.getByRole('textbox', { name: /todo title/i });
    fireEvent.change(input, { target: { value: 'Ship feature' } });
    fireEvent.submit(screen.getByRole('form', { name: /add todo form/i }));

    expect(onAdd).toHaveBeenCalledWith('Ship feature');
    expect(input).toHaveValue('');
    expect(screen.getByRole('button', { name: /add/i })).toBeDisabled();
  });
});
