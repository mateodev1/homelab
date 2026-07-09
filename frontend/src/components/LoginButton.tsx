import { Button } from '@/components/ui/button';
import { useAuth0 } from '@auth0/auth0-react';
import { LogIn } from 'lucide-react';

export function LoginButton() {
  const { loginWithRedirect } = useAuth0();
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onClick={() => loginWithRedirect()}
      aria-label="Iniciar sesión"
      title="Iniciar sesión"
    >
      <LogIn aria-hidden="true" />
    </Button>
  );
}
