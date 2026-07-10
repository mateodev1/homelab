import { X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { KIND_VIEWS, Sidebar, type TaskView, VIEWS } from './components/Sidebar';
import { TaskForm } from './components/TaskForm';
import { TaskList } from './components/TaskList';
import { Button } from './components/ui/button';
import { Dialog, DialogContent } from './components/ui/dialog';
import { Input } from './components/ui/input';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useProjects } from './hooks/useProjects';
import { useTodos } from './hooks/useTodos';

function App() {
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

  const viewTitle =
    VIEWS.find((view) => view.key === activeView)?.label ??
    KIND_VIEWS.find((view) => view.key === activeView)?.label ??
    'All';

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

  useKeyboardShortcuts({
    onCreate: () => !dialogOpen && openCreate(),
    onSearch: () => !dialogOpen && searchInputRef.current?.focus(),
    onMoveDown: () => !dialogOpen && moveActive(1),
    onMoveUp: () => !dialogOpen && moveActive(-1),
    onOpen: () => !dialogOpen && openEditForActive(),
    onClose: () => dialogOpen && closeDialog(),
  });

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar
        activeView={activeView}
        onSelectView={setActiveView}
        counts={counts}
        onCreateTask={openCreate}
        projects={projects}
        activeProjectId={activeProjectId}
        onSelectProject={setActiveProjectId}
        projectCounts={projectCounts}
        onCreateProject={(name) => void addProject(name)}
        onDeleteProject={(id) => {
          if (activeProjectId === id) setActiveProjectId(null);
          void removeProject(id);
        }}
      />

      <main className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-12 shrink-0 items-center gap-4 border-b border-border px-4">
          <h1 className="shrink-0 text-sm font-semibold text-foreground">
            {viewTitle}
            <span className="ml-1.5 font-normal text-muted-foreground">{filteredTasks.length}</span>
          </h1>

          <div className="relative max-w-xs flex-1">
            <Input
              ref={searchInputRef}
              type="search"
              placeholder="Search tasks"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              aria-label="Search tasks"
              className="h-7 pl-3 pr-8 text-sm"
            />
            {query && (
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setQuery('')}
                aria-label="Clear search"
                className="absolute right-0.5 top-0.5 h-6 w-6"
              >
                <X aria-hidden="true" className="size-3.5" />
              </Button>
            )}
          </div>
        </header>

        <div className="flex-1 overflow-y-auto">
          <TaskList
            tasks={filteredTasks}
            projects={projects}
            loading={loading}
            error={error}
            activeIndex={activeIndex}
            onSelectTask={(id) => {
              setEditingTodoID(id);
              setDialogOpen(true);
            }}
            onDeleteTask={removeTodo}
          />
        </div>
      </main>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent label={editingTodo ? 'Edit task' : 'Create task'}>
          <TaskForm
            todo={editingTodo}
            projects={projects}
            onCreate={async (title, body, priority, dueDate, kind, issueType, projectId) => {
              await addTodo(title, body, priority, dueDate, kind, issueType, projectId);
              closeDialog();
            }}
            onUpdate={async (id, changes) => {
              await editTodo(id, changes);
              closeDialog();
            }}
            onCancelEdit={closeDialog}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

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

export default App;
