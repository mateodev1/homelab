import { useEffect } from 'react';

interface KeyboardShortcutHandlers {
  onCreate: () => void;
  onSearch: () => void;
  onMoveDown: () => void;
  onMoveUp: () => void;
  onOpen: () => void;
  onClose: () => void;
}

function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false;
  }

  const tag = target.tagName;
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
}

/**
 * Linear-style single-letter shortcuts, scoped to the whole document.
 * Disabled while the user is typing in a form field (except Escape).
 */
export function useKeyboardShortcuts(
  { onCreate, onSearch, onMoveDown, onMoveUp, onOpen, onClose }: KeyboardShortcutHandlers,
  enabled = true,
) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
        return;
      }

      if (isTypingTarget(event.target) || event.metaKey || event.ctrlKey || event.altKey) {
        return;
      }

      switch (event.key) {
        case 'c':
          event.preventDefault();
          onCreate();
          break;
        case '/':
          event.preventDefault();
          onSearch();
          break;
        case 'j':
        case 'ArrowDown':
          event.preventDefault();
          onMoveDown();
          break;
        case 'k':
        case 'ArrowUp':
          event.preventDefault();
          onMoveUp();
          break;
        case 'Enter':
          event.preventDefault();
          onOpen();
          break;
        default:
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [enabled, onCreate, onSearch, onMoveDown, onMoveUp, onOpen, onClose]);
}
