import { X } from 'lucide-react';
import { KIND_VIEWS, VIEWS } from './components/Sidebar';
import { TaskBoard } from './components/TaskBoard';
import { Button } from './components/ui/button';
import { Input } from './components/ui/input';
import { useTodosBoard } from './context/TodosBoardContext';

function App() {
  const {
    projects,
    loading,
    error,
    editTodo,
    removeTodo,
    query,
    setQuery,
    activeView,
    filteredTasks,
    activeIndex,
    searchInputRef,
    openEditForTask,
  } = useTodosBoard();

  const viewTitle =
    VIEWS.find((view) => view.key === activeView)?.label ??
    KIND_VIEWS.find((view) => view.key === activeView)?.label ??
    'All';

  return (
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
        <TaskBoard
          tasks={filteredTasks}
          projects={projects}
          loading={loading}
          error={error}
          activeIndex={activeIndex}
          onSelectTask={openEditForTask}
          onDeleteTask={removeTodo}
          onStatusChange={(id, status) => void editTodo(id, { status })}
        />
      </div>
    </main>
  );
}

export default App;
