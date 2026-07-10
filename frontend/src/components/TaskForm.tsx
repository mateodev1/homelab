import { type FormEvent, Suspense, lazy, useEffect, useState } from 'react';
import type { Project } from '../types/project';
import type { IssueType, Todo, TodoKind, TodoStatus } from '../types/todo';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select } from './ui/select';
import { Textarea } from './ui/textarea';

const MDEditor = lazy(async () => {
  const mod = await import('@uiw/react-md-editor');
  return { default: mod.default };
});

interface TaskFormProps {
  todo: Todo | null;
  projects: Project[];
  onCreate: (
    title: string,
    body?: string,
    priority?: 0 | 1 | 2 | 3,
    dueDate?: string | null,
    kind?: TodoKind,
    issueType?: IssueType | null,
    projectId?: number | null,
  ) => Promise<void>;
  onUpdate: (
    id: number,
    changes: Partial<
      Pick<
        Todo,
        'title' | 'body' | 'status' | 'priority' | 'due_date' | 'kind' | 'issue_type' | 'project_id'
      >
    >,
  ) => Promise<void>;
  onCancelEdit: () => void;
}

export function TaskForm({ todo, projects, onCreate, onUpdate, onCancelEdit }: TaskFormProps) {
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [status, setStatus] = useState<TodoStatus>('todo');
  const [priority, setPriority] = useState<0 | 1 | 2 | 3>(0);
  const [dueDate, setDueDate] = useState<string>('');
  const [kind, setKind] = useState<TodoKind>('note');
  const [issueType, setIssueType] = useState<IssueType | ''>('');
  const [projectId, setProjectId] = useState<number | ''>('');

  const isEditing = Boolean(todo);

  useEffect(() => {
    if (!todo) {
      setTitle('');
      setBody('');
      setStatus('todo');
      setPriority(0);
      setDueDate('');
      setKind('note');
      setIssueType('');
      setProjectId('');
      return;
    }

    setTitle(todo.title);
    setBody(todo.body);
    setStatus(todo.status);
    setPriority(todo.priority);
    setDueDate(todo.due_date ?? '');
    setKind(todo.kind);
    setIssueType(todo.issue_type ?? '');
    setProjectId(todo.project_id ?? '');
  }, [todo]);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();

    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      return;
    }

    const resolvedIssueType = kind === 'issue' ? issueType || null : null;
    const resolvedProjectId = projectId === '' ? null : projectId;

    if (todo) {
      await onUpdate(todo.id, {
        title: trimmedTitle,
        body,
        status,
        priority,
        due_date: dueDate || null,
        kind,
        issue_type: resolvedIssueType,
        project_id: resolvedProjectId,
      });
      return;
    }

    await onCreate(
      trimmedTitle,
      body,
      priority,
      dueDate || null,
      kind,
      resolvedIssueType,
      resolvedProjectId,
    );
    setTitle('');
    setBody('');
    setPriority(0);
    setDueDate('');
    setKind('note');
    setIssueType('');
    setProjectId('');
  };

  return (
    <Card>
      <form onSubmit={handleSubmit}>
        <CardHeader>
          <CardTitle className="text-base">{isEditing ? 'Edit task' : 'Create task'}</CardTitle>
        </CardHeader>

        <CardContent>
          <Input
            type="text"
            placeholder="Task title"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
          />

          <div className="grid gap-3 sm:grid-cols-3">
            <Label className="flex-col items-stretch gap-1.5">
              Kind
              <Select
                value={kind}
                onChange={(event) => {
                  const nextKind = event.target.value as TodoKind;
                  setKind(nextKind);
                  if (nextKind !== 'issue') {
                    setIssueType('');
                  }
                }}
              >
                <option value="note">Note</option>
                <option value="issue">Issue</option>
              </Select>
            </Label>

            {kind === 'issue' ? (
              <Label className="flex-col items-stretch gap-1.5">
                Issue type
                <Select
                  value={issueType}
                  onChange={(event) => setIssueType(event.target.value as IssueType | '')}
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
                value={priority}
                onChange={(event) => setPriority(Number(event.target.value) as 0 | 1 | 2 | 3)}
              >
                <option value={0}>None</option>
                <option value={1}>Low</option>
                <option value={2}>Medium</option>
                <option value={3}>High</option>
              </Select>
            </Label>

            <Label className="flex-col items-stretch gap-1.5">
              Due date
              <Input
                type="date"
                value={dueDate}
                onChange={(event) => setDueDate(event.target.value)}
              />
            </Label>

            <Label className="flex-col items-stretch gap-1.5">
              Project
              <Select
                value={projectId}
                onChange={(event) =>
                  setProjectId(event.target.value === '' ? '' : Number(event.target.value))
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

            <Label className="flex-col items-stretch gap-1.5">
              Status
              <Select
                value={status}
                onChange={(event) => setStatus(event.target.value as TodoStatus)}
              >
                <option value="todo">Todo</option>
                <option value="in_progress">In progress</option>
                <option value="done">Done</option>
                <option value="cancelled">Cancelled</option>
              </Select>
            </Label>
          </div>

          <Suspense
            fallback={
              <Textarea value={body} onChange={(event) => setBody(event.target.value)} rows={8} />
            }
          >
            <div data-color-mode="light">
              <MDEditor
                value={body}
                onChange={(value) => setBody(value ?? '')}
                preview="edit"
                height={240}
                textareaProps={{ placeholder: 'Write markdown...' }}
              />
            </div>
          </Suspense>
        </CardContent>

        <div className="flex justify-end gap-2 p-4 pt-0">
          {isEditing ? (
            <Button type="button" variant="outline" onClick={onCancelEdit}>
              Cancel
            </Button>
          ) : null}
          <Button type="submit">{isEditing ? 'Update task' : 'Add task'}</Button>
        </div>
      </form>
    </Card>
  );
}
