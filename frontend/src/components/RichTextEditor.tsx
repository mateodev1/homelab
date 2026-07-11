import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { Markdown, type MarkdownStorage } from 'tiptap-markdown';

// `tiptap-markdown`'s type declarations don't augment Tiptap's `Storage`
// interface, so `editor.storage.markdown` is untyped without this.
declare module '@tiptap/core' {
  interface Storage {
    markdown: MarkdownStorage;
  }
}

interface RichTextEditorProps {
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
}

export function RichTextEditor({ value, onChange, onBlur }: RichTextEditorProps) {
  const editor = useEditor({
    extensions: [StarterKit, Markdown.configure({ html: false })],
    content: value,
    onUpdate: ({ editor: currentEditor }) => {
      onChange(currentEditor.storage.markdown.getMarkdown());
    },
    onBlur: () => onBlur?.(),
    editorProps: {
      attributes: {
        'aria-label': 'Task body',
        class: 'typeset typeset-docs min-h-80 focus:outline-none',
      },
    },
  });

  return <EditorContent editor={editor} />;
}
