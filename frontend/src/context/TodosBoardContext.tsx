import { useNavigate } from '@tanstack/react-router';
import { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { KIND_VIEWS, type TaskView, VIEWS } from '../components/Sidebar';
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts';
import { useProjects } from '../hooks/useProjects';
import { useTodos } from '../hooks/useTodos';
import type { Project } from '../types/project';
import type { Todo } from '../types/todo';

interface TodosBoardContextValue {
  todos: Todo[];
  loading: boolean;
  error: string | null;
  addTodo: ReturnType<typeof useTodos>['addTodo'];
  editTodo: ReturnType<typeof useTodos>['editTodo'];
  removeTodo: ReturnType<typeof useTodos>['removeTodo'];

  projects: Project[];
  addProject: (name: string) => void;
  removeProject: (id: number) => void;

  query: string;
  setQuery: (query: string) => void;
  activeView: TaskView;
  activeProjectId: number | null;
  onSelectView: (view: TaskView) => void;
  onSelectProject: (id: number | null) => void;

  filteredTasks: Todo[];
  activeIndex: number;
  counts: Record<TaskView, number>;
  projectCounts: Record<number, number>;

  dialogOpen: boolean;
  editingTodo: Todo | null;
  openCreate: () => void;
  openEditForActive: () => void;
  openEditForTask: (id: number) => void;
  closeDialog: () => void;

  searchInputRef: React.RefObject<HTMLInputElement>;
}

const TodosBoardContext = createContext<TodosBoardContextValue | null>(null);

function matchesQuery(query: string) {
  const normalized = query.trim().toLowerCase();

  return (todo: { title: string; body: string }) => {
    if (!normalized) {
      return true;
    }

    return (
      todo.title.toLowerCase().includes(normalized) || todo.body.toLowerCase().includes(normalized)
    );
  };
}

export function TodosBoardProvider({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const { todos, groupedTodos, loading, error, addTodo, editTodo, removeTodo } = useTodos();
  const { projects, addProject, removeProject } = useProjects();

  const [query, setQuery] = useState('');
  const [activeView, setActiveView] = useState<TaskView>('all');
  const [activeProjectId, setActiveProjectId] = useState<number | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTodoID, setEditingTodoID] = useState<number | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  const viewTasks = useMemo(() => {
    if (activeView === 'all') {
      return [
        ...groupedTodos.todo,
        ...groupedTodos.in_progress,
        ...groupedTodos.done,
        ...groupedTodos.cancelled,
      ];
    }
    if (activeView === 'note' || activeView === 'issue') {
      return todos.filter((todo) => todo.kind === activeView);
    }
    return groupedTodos[activeView];
  }, [activeView, groupedTodos, todos]);

  const projectFilteredTasks = useMemo(() => {
    if (activeProjectId === null) return viewTasks;
    return viewTasks.filter((todo) => todo.project_id === activeProjectId);
  }, [viewTasks, activeProjectId]);

  const filteredTasks = useMemo(
    () => projectFilteredTasks.filter(matchesQuery(query)),
    [projectFilteredTasks, query],
  );

  const projectCounts = useMemo(() => {
    const counts: Record<number, number> = {};
    for (const todo of todos) {
      if (todo.project_id != null) {
        counts[todo.project_id] = (counts[todo.project_id] ?? 0) + 1;
      }
    }
    return counts;
  }, [todos]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: reset the active row when the view, project or search query changes
  useEffect(() => {
    setActiveIndex(0);
  }, [activeView, activeProjectId, query]);

  const counts: Record<TaskView, number> = {
    all: todos.length,
    todo: groupedTodos.todo.length,
    in_progress: groupedTodos.in_progress.length,
    done: groupedTodos.done.length,
    cancelled: groupedTodos.cancelled.length,
    note: todos.filter((todo) => todo.kind === 'note').length,
    issue: todos.filter((todo) => todo.kind === 'issue').length,
  };

  const editingTodo =
    editingTodoID == null ? null : (todos.find((todo) => todo.id === editingTodoID) ?? null);

  const openCreate = () => {
    setEditingTodoID(null);
    setDialogOpen(true);
  };

  const openEditForActive = () => {
    const task = filteredTasks[activeIndex];
    if (!task) return;
    setEditingTodoID(task.id);
    setDialogOpen(true);
  };

  const openEditForTask = (id: number) => {
    setEditingTodoID(id);
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setEditingTodoID(null);
  };

  const moveActive = (delta: number) => {
    setActiveIndex((current) => {
      if (filteredTasks.length === 0) return 0;
      return Math.max(0, Math.min(filteredTasks.length - 1, current + delta));
    });
  };

  const onSelectView = (view: TaskView) => {
    setActiveView(view);
    void navigate({ to: '/todos' });
  };

  const onSelectProject = (id: number | null) => {
    setActiveProjectId(id);
    void navigate({ to: '/todos' });
  };

  useKeyboardShortcuts({
    onCreate: () => !dialogOpen && openCreate(),
    onSearch: () => !dialogOpen && searchInputRef.current?.focus(),
    onMoveDown: () => !dialogOpen && moveActive(1),
    onMoveUp: () => !dialogOpen && moveActive(-1),
    onOpen: () => !dialogOpen && openEditForActive(),
    onClose: () => dialogOpen && closeDialog(),
  });

  const value: TodosBoardContextValue = {
    todos,
    loading,
    error,
    addTodo,
    editTodo,
    removeTodo,

    projects,
    addProject: (name) => void addProject(name),
    removeProject: (id) => {
      if (activeProjectId === id) setActiveProjectId(null);
      void removeProject(id);
    },

    query,
    setQuery,
    activeView,
    activeProjectId,
    onSelectView,
    onSelectProject,

    filteredTasks,
    activeIndex,
    counts,
    projectCounts,

    dialogOpen,
    editingTodo,
    openCreate,
    openEditForActive,
    openEditForTask,
    closeDialog,

    searchInputRef,
  };

  return <TodosBoardContext.Provider value={value}>{children}</TodosBoardContext.Provider>;
}

export function useTodosBoard(): TodosBoardContextValue {
  const ctx = useContext(TodosBoardContext);

  if (!ctx) {
    throw new Error('useTodosBoard must be used within TodosBoardProvider');
  }

  return ctx;
}

export { VIEWS, KIND_VIEWS };
export type { TaskView };
