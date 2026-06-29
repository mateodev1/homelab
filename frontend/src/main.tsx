import { createRoot } from 'react-dom/client';
import { RouterProvider, createRouter } from '@tanstack/react-router';
import { routeTree } from './routeTree.gen';
import './styles/global.css';

const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById('root');

if (!rootElement) {
  throw new Error('Root element not found');
}

// StrictMode removed: Auth0 authorization codes are single-use.
// In dev, StrictMode runs effects twice which causes the second
// code exchange attempt to fail with access_denied.
createRoot(rootElement).render(
  <RouterProvider router={router} />,
);
