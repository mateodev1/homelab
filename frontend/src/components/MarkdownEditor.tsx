import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Textarea } from './ui/textarea';

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
}

export function MarkdownEditor({ value, onChange, onBlur }: MarkdownEditorProps) {
  return (
    <div className="grid gap-3 md:grid-cols-2">
      <Textarea
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onBlur={onBlur}
        aria-label="Task body (markdown)"
        placeholder="Write markdown..."
        className="min-h-80 resize-none font-mono text-sm"
      />

      <div
        className="typeset typeset-docs min-h-80 rounded-md border border-border bg-card p-4"
        data-testid="markdown-preview"
      >
        {value.trim() ? (
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{value}</ReactMarkdown>
        ) : (
          <p className="text-muted-foreground">Nothing to preview yet.</p>
        )}
      </div>
    </div>
  );
}
