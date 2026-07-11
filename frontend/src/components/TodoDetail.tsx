import { cn } from '@/lib/utils';
import { Link, useNavigate } from '@tanstack/react-router';
import { ArrowLeft, Loader2, Trash2 } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useTodo } from '../hooks/useTodo';
import type { Project } from '../types/project';
import type { IssueType, Todo, TodoKind, TodoStatus } from '../types/todo';
import { RichTextEditor } from './RichTextEditor';
import { Button, buttonVariants } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select } from './ui/select';

const SAVE_DEBOUNCE_MS = 500;

interface TodoDetailProps {
  id: number;
  projects: Project[];
  removeTodo: (id: number) => Promise<void>;
}

export function TodoDetail({ id, projects, removeTodo }: TodoDetailProps) {
  const navigate = useNavigate();
  const { todo, loading, error, save } = useTodo(id);

  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const saveTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (todo) {
      setTitle(todo.title);
      setBody(todo.body);
    }
  }, [todo]);

  useEffect(() => {
    return () => {
      if (saveTimeout.current) clearTimeout(saveTimeout.current);
    };
  }, []);

  const saveNow = (changes: Partial<Todo>) => {
    if (saveTimeout.current) clearTimeout(saveTimeout.current);
    void save(changes);
  };

  const scheduleSave = (changes: Partial<Todo>) => {
    if (saveTimeout.current) clearTimeout(saveTimeout.current);
    saveTimeout.current = setTimeout(() => {
      void save(changes);
    }, SAVE_DEBOUNCE_MS);
  };

  const handleDelete = async () => {
    if (!todo) return;
    await removeTodo(todo.id);
    navigate({ to: '/todos' });
  };

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2
          role="status"
          aria-label="Loading task"
          className="size-8 animate-spin text-muted-foreground"
        />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center py-8">
        <p className="text-sm text-destructive">Failed to load task: {error}</p>
      </div>
    );
  }

  if (!todo) {
    return null;
  }

  return (
    <div className="flex flex-1 overflow-hidden">
      <div className="flex-1 overflow-y-auto">
        <div className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
          <div className="flex items-center justify-between">
            <Link to="/todos" className={cn(buttonVariants({ variant: 'ghost', size: 'sm' }))}>
              <ArrowLeft aria-hidden="true" className="size-4" />
              Back
            </Link>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => void handleDelete()}
              aria-label={`Delete ${todo.title}`}
            >
              <Trash2 aria-hidden="true" className="size-4" />
            </Button>
          </div>

          <input
            type="text"
            value={title}
            onChange={(event) => {
              setTitle(event.target.value);
              scheduleSave({ title: event.target.value });
            }}
            onBlur={() => saveNow({ title })}
            placeholder="Task title"
            aria-label="Task title"
            className="w-full border-none bg-transparent text-2xl font-semibold text-foreground outline-none"
          />

          <RichTextEditor
            value={body}
            onChange={(next) => {
              setBody(next);
              scheduleSave({ body: next });
            }}
            onBlur={() => saveNow({ body })}
          />
        </div>
      </div>

      <aside className="w-72 shrink-0 overflow-y-auto border-l border-border p-4">
        <h2 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Properties
        </h2>

        <div className="flex flex-col gap-3">
          <Label className="flex-col items-stretch gap-1.5">
            Kind
            <Select
              value={todo.kind}
              onChange={(event) => {
                const nextKind = event.target.value as TodoKind;
                saveNow({
                  kind: nextKind,
                  issue_type: nextKind === 'issue' ? todo.issue_type : null,
                });
              }}
            >
              <option value="note">Note</option>
              <option value="issue">Issue</option>
            </Select>
          </Label>

          {todo.kind === 'issue' ? (
            <Label className="flex-col items-stretch gap-1.5">
              Issue type
              <Select
                value={todo.issue_type ?? ''}
                onChange={(event) =>
                  saveNow({ issue_type: (event.target.value || null) as IssueType | null })
                }
              >
                <option value="">Unclassified</option>
                <option value="feature">Feature</option>
                <option value="bug">Bug</option>
                <option value="improvement">Improvement</option>
              </Select>
            </Label>
          ) : null}

          <Label className="flex-col items-stretch gap-1.5">
            Priority
            <Select
              value={todo.priority}
              onChange={(event) =>
                saveNow({ priority: Number(event.target.value) as 0 | 1 | 2 | 3 })
              }
            >
              <option value={0}>None</option>
              <option value={1}>Low</option>
              <option value={2}>Medium</option>
              <option value={3}>High</option>
            </Select>
          </Label>

          <Label className="flex-col items-stretch gap-1.5">
            Status
            <Select
              value={todo.status}
              onChange={(event) => saveNow({ status: event.target.value as TodoStatus })}
            >
              <option value="todo">Todo</option>
              <option value="in_progress">In progress</option>
              <option value="done">Done</option>
              <option value="cancelled">Cancelled</option>
            </Select>
          </Label>

          <Label className="flex-col items-stretch gap-1.5">
            Due date
            <Input
              type="date"
              value={todo.due_date ?? ''}
              onChange={(event) => saveNow({ due_date: event.target.value || null })}
            />
          </Label>

          <Label className="flex-col items-stretch gap-1.5">
            Project
            <Select
              value={todo.project_id ?? ''}
              onChange={(event) =>
                saveNow({
                  project_id: event.target.value === '' ? null : Number(event.target.value),
                })
              }
            >
              <option value="">No project</option>
              {projects.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name}
                </option>
              ))}
            </Select>
          </Label>
        </div>
      </aside>
    </div>
  );
}
