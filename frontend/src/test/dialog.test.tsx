import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { Dialog, DialogContent } from '../components/ui/dialog';

function TriggerAndDialog({ open }: { open: boolean }) {
  return (
    <div>
      <button type="button">Open trigger</button>
      <Dialog open={open} onOpenChange={vi.fn()}>
        <DialogContent label="Test dialog">
          <button type="button">Inside dialog</button>
        </DialogContent>
      </Dialog>
    </div>
  );
}

describe('Dialog', () => {
  it('exposes an accessible name via aria-label', () => {
    render(<TriggerAndDialog open={true} />);

    expect(screen.getByRole('dialog', { name: 'Test dialog' })).toBeInTheDocument();
  });

  it('moves focus into the dialog when opened', () => {
    render(<TriggerAndDialog open={true} />);

    expect(screen.getByRole('button', { name: 'Close dialog' })).toHaveFocus();
  });

  it('restores focus to the previously focused element on close', () => {
    const trigger = document.createElement('button');
    trigger.textContent = 'External trigger';
    document.body.appendChild(trigger);
    trigger.focus();

    const { rerender } = render(
      <Dialog open={true} onOpenChange={vi.fn()}>
        <DialogContent label="Test dialog">
          <button type="button">Inside dialog</button>
        </DialogContent>
      </Dialog>,
    );

    rerender(
      <Dialog open={false} onOpenChange={vi.fn()}>
        <DialogContent label="Test dialog">
          <button type="button">Inside dialog</button>
        </DialogContent>
      </Dialog>,
    );

    expect(trigger).toHaveFocus();
    document.body.removeChild(trigger);
  });

  it('renders nothing when closed', () => {
    render(<TriggerAndDialog open={false} />);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });
});
