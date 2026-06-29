import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App';
import { ThemeProvider } from './context/ThemeContext';
import { useTodos } from './hooks/useTodos';

const renderApp = () =>
  render(
    <ThemeProvider>
      <App />
    </ThemeProvider>,
  );

vi.mock('./hooks/useTodos', () => ({
  useTodos: vi.fn(),
}));

const mockedUseTodos = vi.mocked(useTodos);

const mockHookBase = {
  loading: false,
  error: null,
  addTodo: vi.fn(),
  editTodo: vi.fn(),
  toggleTodo: vi.fn(),
  togglePin: vi.fn(),
  removeTodo: vi.fn(),
};

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders NoteForm and NoteGrid using useTodos state', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [
        {
          id: 1,
          title: 'From hook',
          body: '',
          color: 'default',
          pinned: false,
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
    });

    renderApp();

    expect(screen.getByRole('searchbox', { name: /buscar notas/i })).toBeInTheDocument();
    expect(screen.getByText('From hook')).toBeInTheDocument();
  });

  it('renders loading spinner when loading state is true', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [],
      loading: true,
    });

    renderApp();

    expect(screen.getByLabelText(/cargando notas/i)).toBeInTheDocument();
  });

  it('search filtering — typing filters visible cards', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [
        {
          id: 1,
          title: 'Alpha',
          body: '',
          color: 'default',
          pinned: false,
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
        {
          id: 2,
          title: 'Beta',
          body: '',
          color: 'default',
          pinned: false,
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
    });

    renderApp();

    fireEvent.change(screen.getByRole('searchbox', { name: /buscar notas/i }), {
      target: { value: 'alp' },
    });

    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
  });

  it('clear search button — click restores all todos', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [
        {
          id: 1,
          title: 'Alpha',
          body: '',
          color: 'default',
          pinned: false,
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
        {
          id: 2,
          title: 'Beta',
          body: '',
          color: 'default',
          pinned: false,
          done: false,
          created_at: '2026-06-21T03:00:00Z',
          updated_at: '2026-06-21T03:00:00Z',
        },
      ],
    });

    renderApp();

    fireEvent.change(screen.getByRole('searchbox', { name: /buscar notas/i }), {
      target: { value: 'alp' },
    });

    expect(screen.queryByText('Beta')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /limpiar búsqueda/i }));

    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('error state — error message renders in NoteGrid', () => {
    mockedUseTodos.mockReturnValue({
      ...mockHookBase,
      todos: [],
      error: 'server error',
    });

    renderApp();

    expect(screen.getByText(/error al cargar notas: server error/i)).toBeInTheDocument();
  });
});
