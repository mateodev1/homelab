import { useAuth0 } from '@auth0/auth0-react';

export function LoginButton() {
  const { loginWithRedirect } = useAuth0();
  return (
    <button
      type="button"
      onClick={() => loginWithRedirect()}
      className="app-header__theme-toggle"
      aria-label="Iniciar sesión"
      title="Iniciar sesión"
    >
      🔑
    </button>
  );
}
