import {TodoService} from "../bindings/changeme";

let currentFilter = 'all';

async function loadTodos() {
    const [todos, stats] = await Promise.all([
        
        TodoService.GetFiltered(currentFilter),
        TodoService.GetStats()
    ]);

    document.getElementById('stats')!.textContent =
        `${stats.active} active, ${stats.completed} completed`;

    const list = document.getElementById('todo-list')! as HTMLDivElement;
    list.innerHTML = (todos ?? []).map(todo => `
        <div class="todo ${todo.completed ? 'completed' : ''}" data-todo-id="${todo.id}">
            <input type="checkbox"
                ${todo.completed ? 'checked' : ''}
                onchange="toggleTodo(${todo.id})">
            <span ondblclick="editTodo(${todo.id})">${todo.title}</span>
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
        setFilter: (f: string) => Promise<void>;
        editTodo: (id: number) => void;
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

window.setFilter = async (f: string) => {
    currentFilter = f;
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.classList.toggle('is-active', (btn as HTMLElement).dataset.filter === f);
    });
    await loadTodos();
};

window.editTodo = (id: number) => {
    const span = document.querySelector(`[data-todo-id="${id}"] span`) as HTMLSpanElement | null;
    if (!span) return;

    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'edit-input';
    input.value = span.textContent ?? '';

    span.replaceWith(input);
    input.focus();
    input.select();

    let done = false;
    const commit = async () => {
        if (done) return;
        done = true;
        const title = input.value.trim();
        if (title && title !== span.textContent) {
            try {
                await TodoService.Update(id, title);
            } catch (err) {
                console.error('Failed to update todo:', err);
            }
        }
        await loadTodos();
    };

    input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') input.blur();          // blur triggers commit
        if (e.key === 'Escape') { done = true; loadTodos(); }  // cancel: re-render restores the span
    });
    input.addEventListener('blur', commit);
};

// Load todos on startup
loadTodos();