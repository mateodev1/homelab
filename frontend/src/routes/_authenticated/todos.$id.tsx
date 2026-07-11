import { createFileRoute } from '@tanstack/react-router';
import { TodoDetail } from '../../components/TodoDetail';
import { useTodosBoard } from '../../context/TodosBoardContext';

export const Route = createFileRoute('/_authenticated/todos/$id')({
  component: TodoDetailRoute,
});

function TodoDetailRoute() {
  const { id } = Route.useParams();
  const { projects, removeTodo } = useTodosBoard();

  return <TodoDetail id={Number(id)} projects={projects} removeTodo={removeTodo} />;
}
