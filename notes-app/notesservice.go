package main

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NotesService struct {
	notes []Note
}

func NewNotesService() *NotesService {
	return &NotesService{
		notes: make([]Note, 0),
	}
}

func (n *NotesService) GetAll() []Note {
	return n.notes
}

func (n *NotesService) Create(title, content string) Note {
	note := Note{
		ID:        generateID(),
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	n.notes = append(n.notes, note)
	return note
}

func (n *NotesService) Update(id, title, content string) error {
	for i := range n.notes {
		if n.notes[i].ID == id {
			n.notes[i].Title = title
			n.notes[i].Content = content
			n.notes[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return errors.New("note not found")
}

func (n *NotesService) Delete(id string) error {
	for i := range n.notes {
		if n.notes[i].ID == id {
			n.notes = append(n.notes[:i], n.notes[i+1:]...)
			return nil
		}
	}
	return errors.New("note not found")
}

func (n *NotesService) SaveToFile() error {
	app := application.Get()

	path, err := app.Dialog.SaveFile().
		SetFilename("notes.json").
		AddFilter("JSON Files", "*.json").
		PromptForSingleSelection()
	if err != nil {
		return err
	}
	if path == "" {
		return nil // user cancelled
	}

	data, err := json.MarshalIndent(n.notes, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	app.Dialog.Info().SetMessage("Notes saved successfully").Show()
	return nil
}

func (n *NotesService) LoadFromFile() error {
	app := application.Get()

	path, err := app.Dialog.OpenFile().
		AddFilter("JSON Files", "*.json").
		PromptForSingleSelection()
	if err != nil {
		return err
	}
	if path == "" {
		return nil // user cancelled
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &n.notes); err != nil {
		return err
	}

	app.Dialog.Info().SetMessage("Notes loaded successfully").Show()
	return nil
}

func generateID() string {
	return time.Now().Format("20060102150405")
}
