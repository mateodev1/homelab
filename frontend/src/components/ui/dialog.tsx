import * as React from 'react';
import { useEffect, useRef } from 'react';
import { cn } from '@/lib/utils';

export interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])';

/**
 * Minimal dependency-free modal (no Radix Dialog in this project's deps).
 * Handles overlay click-to-close, body scroll lock, and basic focus
 * management (focus moves into the dialog on open and restores to the
 * previously focused element on close). Escape is handled globally by
 * useKeyboardShortcuts to keep a single source of truth.
 */
function Dialog({ open, onOpenChange, children }: DialogProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const previouslyFocusedRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return;

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [open]);

  useEffect(() => {
    if (!open) return;

    previouslyFocusedRef.current = document.activeElement as HTMLElement | null;
    const container = containerRef.current;
    const focusable = container?.querySelector<HTMLElement>(FOCUSABLE_SELECTOR);
    focusable?.focus();

    return () => {
      previouslyFocusedRef.current?.focus();
    };
  }, [open]);

  if (!open) return null;

  return (
    <div
      ref={containerRef}
      className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto pt-[10vh]"
    >
      <button
        type="button"
        aria-label="Close dialog"
        className="fixed inset-0 bg-black/50"
        onClick={() => onOpenChange(false)}
      />
      {children}
    </div>
  );
}

export interface DialogContentProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Accessible name for the dialog, exposed via aria-label. */
  label?: string;
}

const DialogContent = React.forwardRef<HTMLDivElement, DialogContentProps>(
  ({ className, label, ...props }, ref) => (
    <div
      ref={ref}
      role="dialog"
      aria-modal="true"
      aria-label={label}
      className={cn('relative z-10 w-full max-w-lg px-4', className)}
      {...props}
    />
  ),
);
DialogContent.displayName = 'DialogContent';

export { Dialog, DialogContent };
