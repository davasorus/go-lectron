import {TodoService} from "../bindings/changeme";

async function loadTodos() {
    const [todos, stats] = await Promise.all([
        TodoService.GetAll(),
        TodoService.GetStats()
    ]);

    document.getElementById('stats')!.textContent =
        `${stats.active} active, ${stats.completed} completed`;

    const list = document.getElementById('todo-list')! as HTMLDivElement;
    list.innerHTML = (todos ?? []).map(todo => `
        <div class="todo ${todo.completed ? 'completed' : ''}">
            <input type="checkbox"
                ${todo.completed ? 'checked' : ''}
                onchange="toggleTodo(${todo.id})">
            <span>${todo.title}</span>
            <button onclick="deleteTodo(${todo.id})">Delete</button>
        </div>
    `).join('');
}

declare global {
    interface Window {
        addTodo: () => Promise<void>;
        toggleTodo: (id: number) => Promise<void>;
        deleteTodo: (id: number) => Promise<void>;
        clearCompleted: () => Promise<void>;
    }
}

window.addTodo = async () => {
    const input = document.getElementById('todo-input')! as HTMLInputElement;
    const title = input.value.trim();
    if (title) {
        await TodoService.Add(title);
        input.value = '';
        await loadTodos();
    }
};

window.toggleTodo = async (id: number) => {
    await TodoService.Toggle(id);
    await loadTodos();
};

window.deleteTodo = async (id: number) => {
    await TodoService.Delete(id);
    await loadTodos();
};

window.clearCompleted = async () => {
    await TodoService.ClearCompleted();
    await loadTodos();
};

// Load todos on startup
loadTodos();