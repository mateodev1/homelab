import { Button } from '@/components/ui/button';
import { useAuth0 } from '@auth0/auth0-react';
import { LogOut } from 'lucide-react';

export function LogoutButton() {
  const { logout } = useAuth0();
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}
      aria-label="Cerrar sesión"
      title="Cerrar sesión"
    >
      <LogOut aria-hidden="true" />
    </Button>
  );
}
