import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";

import { App } from "./App";

describe("App", () => {
  test("Todo一覧を表示して追加できる", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [
          {
            id: "todo-1",
            title: "既存のTodo",
            completed: false
          }
        ]
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: "todo-2",
          title: "新しいTodo",
          completed: false
        })
      });

    vi.stubGlobal("fetch", fetchMock);

    render(<App apiBaseUrl="http://backend.example" />);

    expect(await screen.findByText("既存のTodo")).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText("Todoタイトル"), "新しいTodo");
    await userEvent.click(screen.getByRole("button", { name: "追加する" }));

    await waitFor(() => {
      expect(screen.getByText("新しいTodo")).toBeInTheDocument();
    });

    expect(fetchMock).toHaveBeenNthCalledWith(1, "http://backend.example/api/todos");
    expect(fetchMock).toHaveBeenNthCalledWith(2, "http://backend.example/api/todos", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        title: "新しいTodo"
      })
    });
  });
});
