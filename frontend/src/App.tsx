import { TodoForm } from './components/TodoForm';
import { TodoList } from './components/TodoList';
import { useTodos } from './hooks/useTodos';

function App() {
  const { todos, loading, error, addTodo, toggleTodo, removeTodo } = useTodos();

  return (
    <main>
      <h1>Todo App</h1>
      <TodoForm onAdd={addTodo} />
      <TodoList
        todos={todos}
        loading={loading}
        error={error}
        onToggle={toggleTodo}
        onDelete={removeTodo}
      />
    </main>
  );
}

export default App;
