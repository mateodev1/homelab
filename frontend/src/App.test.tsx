import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import App from './App';

describe('App', () => {
  it('renders without crashing and shows HomeLab heading', () => {
    render(<App />);
    const heading = screen.getByRole('heading', { name: /homelab/i });
    expect(heading).toBeInTheDocument();
    expect(heading.tagName).toBe('H1');
  });

  it('renders the TodoList placeholder section', () => {
    render(<App />);
    const placeholder = screen.getByTestId('todo-list-placeholder');
    expect(placeholder).toBeInTheDocument();
  });
});
