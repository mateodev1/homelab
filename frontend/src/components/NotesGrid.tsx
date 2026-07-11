import type { Project } from '../types/project';
import type { Todo } from '../types/todo';
import { NoteCard } from './NoteCard';

interface NotesGridProps {
  notes: Todo[];
  projects: Project[];
  activeId: number | null;
  onSelectTask: (id: number) => void;
  onDeleteTask: (id: number) => void;
}

export function NotesGrid({
  notes,
  projects,
  activeId,
  onSelectTask,
  onDeleteTask,
}: NotesGridProps) {
  return (
    <div className="columns-1 gap-3 sm:columns-2 md:columns-3 lg:columns-4">
      {notes.map((note) => (
        <div key={note.id} className="mb-3 break-inside-avoid">
          <NoteCard
            todo={note}
            project={projects.find((project) => project.id === note.project_id) ?? null}
            active={note.id === activeId}
            onSelect={onSelectTask}
            onDelete={onDeleteTask}
          />
        </div>
      ))}
    </div>
  );
}
