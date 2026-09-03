import { NotesService } from '../bindings/changeme'

let notes = []
let currentNote = null

// Load notes on startup
async function loadNotes() {
    notes = await NotesService.GetAll()
    renderNotesList()
}

// Render notes list
function renderNotesList() {
    const notesList = document.getElementById('notes-list')

    if (notes.length === 0) {
        notesList.innerHTML = '<div style="padding: 20px; text-align: center; color: #999;">No notes yet</div>'
        return
    }

    notesList.innerHTML = notes.map(note => `
        <div class="note-item ${currentNote?.id === note.id ? 'active' : ''}" data-id="${note.id}">
            <h3>${note.title || 'Untitled'}</h3>
            <p>${note.content || 'No content'}</p>
        </div>
    `).join('')

    // Add click handlers
    document.querySelectorAll('.note-item').forEach(item => {
        item.addEventListener('click', () => {
            const id = item.dataset.id
            selectNote(id)
        })
    })
}

// Select a note
function selectNote(id) {
    currentNote = notes.find(n => n.id === id)
    if (currentNote) {
        document.getElementById('empty-state').style.display = 'none'
        document.getElementById('note-editor').style.display = 'flex'
        document.getElementById('note-title').value = currentNote.title
        document.getElementById('note-content').value = currentNote.content
        document.getElementById('last-updated').textContent =
            `Last updated: ${new Date(currentNote.updatedAt).toLocaleString()}`
        renderNotesList()
    }
}

// Create new note
document.getElementById('new-note-btn').addEventListener('click', async () => {
    const note = await NotesService.Create('Untitled', '')
    notes.push(note)
    selectNote(note.id)
    // Focus the title input and select all text so user can immediately type
    const titleInput = document.getElementById('note-title')
    titleInput.focus()
    titleInput.select()
})

// Update note on input
let updateTimeout
function scheduleUpdate() {
    clearTimeout(updateTimeout)
    updateTimeout = setTimeout(async () => {
        if (currentNote) {
            const title = document.getElementById('note-title').value
            const content = document.getElementById('note-content').value

            await NotesService.Update(currentNote.id, title, content)

            // Update local copy
            const note = notes.find(n => n.id === currentNote.id)
            if (note) {
                note.title = title
                note.content = content
                note.updatedAt = new Date().toISOString()
            }

            renderNotesList()
            document.getElementById('last-updated').textContent =
                `Last updated: ${new Date().toLocaleString()}`
        }
    }, 500)
}

document.getElementById('note-title').addEventListener('input', scheduleUpdate)
document.getElementById('note-content').addEventListener('input', scheduleUpdate)

// Delete note
document.getElementById('delete-btn').addEventListener('click', async () => {
    if (!currentNote) return

    try {
        await NotesService.Delete(currentNote.id)
        notes = notes.filter(n => n.id !== currentNote.id)
        currentNote = null
        document.getElementById('empty-state').style.display = 'flex'
        document.getElementById('note-editor').style.display = 'none'
        renderNotesList()
    } catch (error) {
        console.error('Delete failed:', error)
    }
})

// Save to file
document.getElementById('save-btn').addEventListener('click', async () => {
    try {
        await NotesService.SaveToFile()
    } catch (error) {
        if (error) console.error('Save failed:', error)
    }
})

// Load from file
document.getElementById('load-btn').addEventListener('click', async () => {
    try {
        await NotesService.LoadFromFile()
        notes = await NotesService.GetAll()
        currentNote = null
        document.getElementById('empty-state').style.display = 'flex'
        document.getElementById('note-editor').style.display = 'none'
        renderNotesList()
    } catch (error) {
        if (error) console.error('Load failed:', error)
    }
})

// Initialize
loadNotes()