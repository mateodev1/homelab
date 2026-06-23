# Plan — Dark mode default with toggle

## Tasks

- [ ] Crear `src/styles/global.css` con tokens, reset, base y estilos de componentes (reemplaza `index.css`)
- [ ] Actualizar import en `main.tsx`: `./index.css` → `./styles/global.css`
- [ ] Crear `src/context/ThemeContext.tsx` con lógica de toggle + localStorage
- [ ] Actualizar `main.tsx` para envolver con `<ThemeProvider>`
- [ ] Agregar botón toggle (🌙/☀️) en el header de `App.tsx`

## Detail

### Estrategia general

- Usar `data-theme` attribute en `<html>` (vía `document.documentElement`)
- Default: `"dark"` (hardcodeado como fallback, pero persiste en `localStorage`)
- Dos bloques de tokens: `:root` = light, `[data-theme="dark"]` = dark
- El context inicializa desde `localStorage` → fallback `"dark"`

---

### 1. `global.css` — estructura y tokens

Crear `frontend/src/styles/global.css` con secciones bien delimitadas. El archivo reemplaza `index.css` (eliminar el original una vez creado).

**Estructura de secciones** (en este orden):

```
/* ===== 1. Theme tokens    ===== */
/* ===== 2. Reset & Base    ===== */
/* ===== 3. Layout          ===== */
/* ===== 4. Header          ===== */
/* ===== 5. NoteForm        ===== */
/* ===== 6. Color Picker    ===== */
/* ===== 7. NoteGrid        ===== */
/* ===== 8. NoteCard        ===== */
/* ===== 9. Spinner         ===== */
```

Sección 1 — tokens (agregar al **tope del archivo**):

```css
/* ===== Theme tokens ===== */
:root {
  --color-bg:           #f8f9fa;
  --color-surface:      #ffffff;
  --color-surface-alt:  #f1f3f4;
  --color-border:       #e0e0e0;
  --color-text-primary: #202124;
  --color-text-muted:   #5f6368;
  --color-text-subtle:  #80868b;
  --color-accent:       #1a73e8;
  --color-error:        #d93025;
  --color-spinner-track:#e0e0e0;
  --color-hover-overlay:rgba(0, 0, 0, 0.06);
  --color-hover-icon:   rgba(0, 0, 0, 0.08);
}

[data-theme="dark"] {
  --color-bg:           #202124;
  --color-surface:      #2d2e30;
  --color-surface-alt:  #303134;
  --color-border:       #5f6368;
  --color-text-primary: #e8eaed;
  --color-text-muted:   #9aa0a6;
  --color-text-subtle:  #9aa0a6;
  --color-accent:       #8ab4f8;
  --color-error:        #f28b82;
  --color-spinner-track:#5f6368;
  --color-hover-overlay:rgba(255, 255, 255, 0.06);
  --color-hover-icon:   rgba(255, 255, 255, 0.08);
}
```

El resto del contenido de `index.css` se migra íntegro, reorganizado en las secciones indicadas, y con todos los hex reemplazados por variables.

Mapa de reemplazo completo:

| Valor original        | Variable                  |
|-----------------------|---------------------------|
| `#f8f9fa`             | `var(--color-bg)`         |
| `#202124`             | `var(--color-text-primary)` |
| `#fff` / `#ffffff`    | `var(--color-surface)`    |
| `#f1f3f4`             | `var(--color-surface-alt)`|
| `#e0e0e0`             | `var(--color-border)`     |
| `#5f6368`             | `var(--color-text-muted)` |
| `#80868b`             | `var(--color-text-subtle)`|
| `#1a73e8`             | `var(--color-accent)`     |
| `#d93025`             | `var(--color-error)`      |
| `rgba(0,0,0,0.06)`    | `var(--color-hover-overlay)` |
| `rgba(0,0,0,0.08)`    | `var(--color-hover-icon)` |

Las sombras (`box-shadow`) con `rgba(0,0,0,...)` pueden quedarse igual — funcionan bien en ambos themes por ser semi-transparentes.

La nota de color del card (`.note-card` background que viene del inline style de `colors.ts`) no cambia — es intencional que las notas conserven su color elegido por el usuario.

---

---

### 2. ThemeContext — `frontend/src/context/ThemeContext.tsx`

```tsx
import { createContext, useContext, useEffect, useState } from 'react';

type Theme = 'dark' | 'light';

interface ThemeContextValue {
  theme: Theme;
  toggle: () => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<Theme>(() => {
    return (localStorage.getItem('theme') as Theme) ?? 'dark';
  });

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem('theme', theme);
  }, [theme]);

  // Apply theme immediately on mount (before first paint)
  // This avoids flash of wrong theme
  useEffect(() => {
    document.documentElement.dataset.theme =
      (localStorage.getItem('theme') as Theme) ?? 'dark';
  }, []);

  const toggle = () => setTheme((t) => (t === 'dark' ? 'light' : 'dark'));

  return (
    <ThemeContext.Provider value={{ theme, toggle }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider');
  return ctx;
}
```

**Anti-flash**: al inicializar el estado desde `localStorage` directamente en el `useState` initializer, el primer render ya aplica el theme correcto. El segundo `useEffect` es para el caso edge de SSR-like (no aplica aquí, pero es buena práctica).

---

### 3. main.tsx — cambiar import y envolver con ThemeProvider

```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { ThemeProvider } from './context/ThemeContext';
import App from './App';
import './styles/global.css'; // antes era ./index.css

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider>
      <App />
    </ThemeProvider>
  </StrictMode>
);
```

---

### 4. Toggle button en App.tsx

Importar `useTheme` y agregar el botón en el header, a la derecha del search:

```tsx
const { theme, toggle } = useTheme();

// En el header, después del div de search:
<button
  type="button"
  className="app-header__theme-toggle"
  onClick={toggle}
  aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
>
  {theme === 'dark' ? '☀️' : '🌙'}
</button>
```

CSS para el botón (agregar en la sección `/* === Header === */` de `global.css`):

```css
.app-header__theme-toggle {
  font-size: 20px;
  padding: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  transition: background 0.15s;
}

.app-header__theme-toggle:hover {
  background: var(--color-hover-icon);
}
```

---

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `frontend/src/styles/global.css` | Nuevo — reemplaza `index.css`. Tokens + todo el CSS reorganizado por secciones |
| `frontend/src/index.css` | Eliminar una vez migrado el contenido |
| `frontend/src/context/ThemeContext.tsx` | Nuevo archivo |
| `frontend/src/main.tsx` | Import → `./styles/global.css`, envolver con ThemeProvider |
| `frontend/src/App.tsx` | Importar useTheme, agregar toggle button |

### Verificación

Correr `pnpm --filter frontend dev` (o equivalente) y verificar:
1. La app abre en dark mode
2. El toggle cambia a light
3. Al recargar la página se mantiene la preferencia
