import { X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { Sidebar, type TaskView, VIEWS } from './components/Sidebar';
import { TaskForm } from './components/TaskForm';
import { TaskList } from './components/TaskList';
import { Button } from './components/ui/button';
import { Dialog, DialogContent } from './components/ui/dialog';
import { Input } from './components/ui/input';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useTodos } from './hooks/useTodos';

function App() {
  const { todos, groupedTodos, loading, error, addTodo, editTodo, removeTodo } = useTodos();
  const [query, setQuery] = useState('');
  const [activeView, setActiveView] = useState<TaskView>('all');
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
    return groupedTodos[activeView];
  }, [activeView, groupedTodos]);

  const filteredTasks = useMemo(() => viewTasks.filter(matchesQuery(query)), [viewTasks, query]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: reset the active row when the view or search query changes
  useEffect(() => {
    setActiveIndex(0);
  }, [activeView, query]);

  const counts: Record<TaskView, number> = {
    all: todos.length,
    todo: groupedTodos.todo.length,
    in_progress: groupedTodos.in_progress.length,
    done: groupedTodos.done.length,
    cancelled: groupedTodos.cancelled.length,
  };

  const editingTodo =
    editingTodoID == null ? null : (todos.find((todo) => todo.id === editingTodoID) ?? null);

  const viewTitle = VIEWS.find((view) => view.key === activeView)?.label ?? 'All';

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
            onCreate={async (title, body, priority, dueDate) => {
              await addTodo(title, body, priority, dueDate);
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
