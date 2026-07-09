import { renderHook } from '@testing-library/react';
import { fireEvent } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts';

function makeHandlers() {
  return {
    onCreate: vi.fn(),
    onSearch: vi.fn(),
    onMoveDown: vi.fn(),
    onMoveUp: vi.fn(),
    onOpen: vi.fn(),
    onClose: vi.fn(),
  };
}

describe('useKeyboardShortcuts', () => {
  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('invokes the matching handler for each shortcut key', () => {
    const handlers = makeHandlers();
    renderHook(() => useKeyboardShortcuts(handlers));

    fireEvent.keyDown(document, { key: 'c' });
    fireEvent.keyDown(document, { key: '/' });
    fireEvent.keyDown(document, { key: 'j' });
    fireEvent.keyDown(document, { key: 'ArrowDown' });
    fireEvent.keyDown(document, { key: 'k' });
    fireEvent.keyDown(document, { key: 'ArrowUp' });
    fireEvent.keyDown(document, { key: 'Enter' });
    fireEvent.keyDown(document, { key: 'Escape' });

    expect(handlers.onCreate).toHaveBeenCalledTimes(1);
    expect(handlers.onSearch).toHaveBeenCalledTimes(1);
    expect(handlers.onMoveDown).toHaveBeenCalledTimes(2);
    expect(handlers.onMoveUp).toHaveBeenCalledTimes(2);
    expect(handlers.onOpen).toHaveBeenCalledTimes(1);
    expect(handlers.onClose).toHaveBeenCalledTimes(1);
  });

  it('ignores letter shortcuts while typing in a form field, except Escape', () => {
    const handlers = makeHandlers();
    renderHook(() => useKeyboardShortcuts(handlers));

    const input = document.createElement('input');
    document.body.appendChild(input);

    fireEvent.keyDown(input, { key: 'c' });
    fireEvent.keyDown(input, { key: 'Escape' });

    expect(handlers.onCreate).not.toHaveBeenCalled();
    expect(handlers.onClose).toHaveBeenCalledTimes(1);
  });

  it('ignores shortcuts combined with modifier keys', () => {
    const handlers = makeHandlers();
    renderHook(() => useKeyboardShortcuts(handlers));

    fireEvent.keyDown(document, { key: 'c', metaKey: true });
    fireEvent.keyDown(document, { key: 'j', ctrlKey: true });

    expect(handlers.onCreate).not.toHaveBeenCalled();
    expect(handlers.onMoveDown).not.toHaveBeenCalled();
  });

  it('does not attach listeners when disabled', () => {
    const handlers = makeHandlers();
    renderHook(() => useKeyboardShortcuts(handlers, false));

    fireEvent.keyDown(document, { key: 'c' });

    expect(handlers.onCreate).not.toHaveBeenCalled();
  });

  it('removes the listener on unmount', () => {
    const handlers = makeHandlers();
    const { unmount } = renderHook(() => useKeyboardShortcuts(handlers));

    unmount();
    fireEvent.keyDown(document, { key: 'c' });

    expect(handlers.onCreate).not.toHaveBeenCalled();
  });
});
