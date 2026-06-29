import { createRootRoute, Outlet, useRouter } from '@tanstack/react-router';
import { Auth0Provider } from '@auth0/auth0-react';
import { ThemeProvider } from '../context/ThemeContext';

export const Route = createRootRoute({
  component: RootLayout,
});

function RootLayout() {
  const router = useRouter();

  return (
    <ThemeProvider>
      <Auth0Provider
        domain={import.meta.env.VITE_AUTH0_DOMAIN ?? ''}
        clientId={import.meta.env.VITE_AUTH0_CLIENT_ID ?? ''}
        authorizationParams={{ redirect_uri: window.location.origin }}
        onRedirectCallback={(appState) => {
          // Navigate AFTER Auth0 finishes processing the callback —
          // this avoids the race condition where isAuthenticated is still false
          router.navigate({
            to: (appState?.returnTo as string) ?? '/todos',
            replace: true,
          });
        }}
      >
        <Outlet />
      </Auth0Provider>
    </ThemeProvider>
  );
}
