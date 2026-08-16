import { createFileRoute } from '@tanstack/react-router';
import { SecretsPage } from '../../components/SecretsPage';

export const Route = createFileRoute('/_authenticated/secrets')({
  component: SecretsPage,
});
