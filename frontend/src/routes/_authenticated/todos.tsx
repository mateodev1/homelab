import { Outlet, createFileRoute } from '@tanstack/react-router';
import { TodosLayout } from '../../components/TodosLayout';

// Layout: `/todos` and `/todos/$id` are siblings that both need this route
// matched (it's their parent). The shared nav Sidebar and the create/edit
// Dialog live in `TodosLayout` so they're identical on the board and the
// detail page — only the `Outlet` content (board vs. detail) changes.
export const Route = createFileRoute('/_authenticated/todos')({
  component: () => (
    <TodosLayout>
      <Outlet />
    </TodosLayout>
  ),
});
