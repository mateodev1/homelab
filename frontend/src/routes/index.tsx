import { useAuth0 } from '@auth0/auth0-react';
import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';

export const Route = createFileRoute('/')({
  component: IndexPage,
});

// Wait for Auth0 to finish initializing before redirecting.
// Using beforeLoad here would race with the Auth0 callback processing.
function IndexPage() {
  const { isLoading } = useAuth0();
  const navigate = useNavigate();

  useEffect(() => {
    if (!isLoading) {
      navigate({ to: '/todos', replace: true });
    }
  }, [isLoading, navigate]);

  return null;
}
