package main

import (
	"errors"
	"sync"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoStats struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Active    int `json:"active"`
}

type TodoService struct {
	todos  []Todo
	nextID int
	mu     sync.Mutex
}

func NewTodoService() *TodoService {
	return &TodoService{
		todos:  []Todo{},
		nextID: 1,
	}
}

func (t *TodoService) GetStats() TodoStats {
	t.mu.Lock()
	defer t.mu.Unlock()

	stats := TodoStats{
		Total: len(t.todos),
	}

	for _, todo := range t.todos {
		if todo.Completed {
			stats.Completed++
		} else {
			stats.Active++
		}
	}
	return stats
}

func (t *TodoService) GetAll() []Todo {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.todos
}

func (t *TodoService) Add(title string) (*Todo, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	todo := Todo{
		ID:        t.nextID,
		Title:     title,
		Completed: false,
	}

	t.todos = append(t.todos, todo)
	t.nextID++

	return &todo, nil
}

func (t *TodoService) Toggle(id int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, todo := range t.todos {
		if todo.ID == id {
			t.todos[i].Completed = !todo.Completed
			return nil
		}
	}
	return errors.New("todo not found")
}

func (t *TodoService) Delete(id int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, todo := range t.todos {
		if todo.ID == id {
			t.todos = append(t.todos[:i], t.todos[i+1:]...)
			return nil
		}
	}
	return errors.New("todo not found")
}

func (t *TodoService) ClearCompleted() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	removed := 0
	newTodos := []Todo{}

	for _, todo := range t.todos {
		if !todo.Completed {
			newTodos = append(newTodos, todo)
		} else {
			removed++
		}
	}

	t.todos = newTodos
	return removed
}
