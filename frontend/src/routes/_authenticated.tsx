import { createFileRoute, Outlet } from '@tanstack/react-router';
import { useAuth0 } from '@auth0/auth0-react';

export const Route = createFileRoute('/_authenticated')({
  component: AuthenticatedLayout,
});

function AuthenticatedLayout() {
  const { isAuthenticated, isLoading, loginWithRedirect } = useAuth0();

  if (isLoading) {
    return (
      <div className="app">
        <div className="app-main" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '60vh' }}>
          <p style={{ color: 'var(--color-text-muted)' }}>Cargando…</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="app">
        <div className="app-main" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: '1rem', minHeight: '60vh' }}>
          <p style={{ color: 'var(--color-text-primary)', fontWeight: 600 }}>Iniciá sesión para ver tus notas</p>
          <button
            type="button"
            onClick={() => loginWithRedirect()}
            className="app-header__theme-toggle"
            style={{ padding: '0.5rem 1.5rem', borderRadius: '999px', background: 'var(--color-accent)', color: '#fff', border: 'none', cursor: 'pointer', fontWeight: 600 }}
          >
            Iniciar sesión
          </button>
        </div>
      </div>
    );
  }

  return <Outlet />;
}
