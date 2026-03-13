import { FormEvent, useEffect, useState } from "react";

import "./styles.css";

type Todo = {
  id: string;
  title: string;
  completed: boolean;
};

type AppProps = {
  apiBaseUrl?: string;
};

export function App({ apiBaseUrl = resolveApiBaseUrl() }: AppProps) {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [newTodoTitle, setNewTodoTitle] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    let isMounted = true;

    async function loadTodos() {
      try {
        const response = await fetch(`${apiBaseUrl}/api/todos`);
        if (!response.ok) {
          throw new Error("Todo一覧の取得に失敗しました");
        }

        const nextTodos = (await response.json()) as Todo[];
        if (isMounted) {
          setTodos(nextTodos);
        }
      } catch (error) {
        if (isMounted) {
          setErrorMessage(resolveErrorMessage(error));
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadTodos();

    return () => {
      isMounted = false;
    };
  }, [apiBaseUrl]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const title = newTodoTitle.trim();
    if (title === "") {
      setErrorMessage("Todoタイトルを入力してください");
      return;
    }

    setErrorMessage("");

    try {
      const response = await fetch(`${apiBaseUrl}/api/todos`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          title
        })
      });

      if (!response.ok) {
        throw new Error("Todoの追加に失敗しました");
      }

      const createdTodo = (await response.json()) as Todo;
      setTodos((currentTodos) => [...currentTodos, createdTodo]);
      setNewTodoTitle("");
    } catch (error) {
      setErrorMessage(resolveErrorMessage(error));
    }
  }

  return (
    <main className="app-shell">
      <section className="panel">
        <p className="eyebrow">Local k8s Todo</p>
        <h1>Todoアプリ</h1>
        <p className="description">
          Frontend は React、Backend は Go で構成し、ローカルの k8s 上で動かす前提の最小構成です。
        </p>

        <form className="todo-form" onSubmit={handleSubmit}>
          <label className="field">
            <span>Todoタイトル</span>
            <input
              aria-label="Todoタイトル"
              value={newTodoTitle}
              onChange={(event) => setNewTodoTitle(event.target.value)}
              placeholder="例: Deployment で起動確認する"
            />
          </label>
          <button type="submit">追加する</button>
        </form>

        {errorMessage !== "" ? <p role="alert">{errorMessage}</p> : null}

        {isLoading ? (
          <p>読み込み中です...</p>
        ) : (
          <ul className="todo-list" aria-label="Todo一覧">
            {todos.map((todo) => (
              <li key={todo.id} className="todo-card">
                <span>{todo.title}</span>
                <span className="status">{todo.completed ? "完了" : "未完了"}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </main>
  );
}

function resolveApiBaseUrl() {
  // k8s では Frontend コンテナのビルド成果物を Nginx で配信しつつ、
  // Ingress から `/api` を Backend Service に流す構成が分かりやすい。
  // そのため既定値は同一オリジンに寄せ、開発時だけ環境変数で差し替えられるようにする。
  return import.meta.env.VITE_API_BASE_URL ?? "";
}

function resolveErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }

  return "不明なエラーが発生しました";
}
