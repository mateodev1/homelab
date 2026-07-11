import type { ReactNode } from 'react';
import { TodosBoardProvider, useTodosBoard } from '../context/TodosBoardContext';
import { Sidebar } from './Sidebar';
import { TaskForm } from './TaskForm';
import { Dialog, DialogContent } from './ui/dialog';

interface TodosLayoutProps {
  children: ReactNode;
}

/**
 * Shared shell for everything under `/todos`: the nav Sidebar and the
 * create/edit Dialog are identical on the board and the detail page — only
 * `children` (the board vs. the detail view) changes between them.
 */
export function TodosLayout({ children }: TodosLayoutProps) {
  return (
    <TodosBoardProvider>
      <TodosLayoutContent>{children}</TodosLayoutContent>
    </TodosBoardProvider>
  );
}

function TodosLayoutContent({ children }: TodosLayoutProps) {
  const {
    activeView,
    onSelectView,
    counts,
    openCreate,
    projects,
    activeProjectId,
    onSelectProject,
    projectCounts,
    addProject,
    removeProject,
    dialogOpen,
    editingTodo,
    closeDialog,
    addTodo,
    editTodo,
  } = useTodosBoard();

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar
        activeView={activeView}
        onSelectView={onSelectView}
        counts={counts}
        onCreateTask={openCreate}
        projects={projects}
        activeProjectId={activeProjectId}
        onSelectProject={onSelectProject}
        projectCounts={projectCounts}
        onCreateProject={addProject}
        onDeleteProject={removeProject}
      />

      {children}

      <Dialog open={dialogOpen} onOpenChange={(open) => !open && closeDialog()}>
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
