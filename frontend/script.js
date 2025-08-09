const API_URL = "http://localhost:8080/api";
const URL = "http://localhost:8080";

const noteFormTitle = document.getElementById('note-form-title');
const titleInput = document.getElementById('title');
const contentInput = document.getElementById('content');
const submitBtn = document.getElementById('submit-btn');
const cancelBtn = document.getElementById('cancel-btn');

// Добавляем глобальную переменную для хранения ID редактируемой заметки
let currentlyEditingNoteId = null;

async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    const response = await fetch(`${URL}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    const data = await response.json();
    localStorage.setItem('token', data.token);
    document.getElementById('login-form').style.display = 'none';
    loadNotes();
}

document.addEventListener('DOMContentLoaded', loadNotes);

async function loadNotes() {
    try {
        const response = await fetch(`${API_URL}/notes`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${errorText}`);
        }

        const notes = await response.json();
        renderNotes(notes);
    } catch (error) {
        console.error("Full error:", error);
    }
}

function renderNotes(notes) {
    const notesList = document.getElementById('notes-list');
    notesList.innerHTML = '';

    if (!notes || notes.length === 0) {
        notesList.innerHTML = '<p class="no-notes">No notes found</p>';
        return;
    }

    notes.forEach(note => {
        const noteElement = document.createElement('div');
        noteElement.className = 'note';
        noteElement.innerHTML = `
            <h3 class="note-title">${escapeHtml(note.title)}</h3>
            <p class="note-content">${escapeHtml(note.content)}</p>
            <div class="note-actions">
                <button onclick="deleteNote(${note.id})" class="delete-btn">Delete</button>
                <button onclick="openEditingWindow(${note.id})" class="edit-btn">Edit</button>
            </div>
        `;
        notesList.appendChild(noteElement);
    });
}

function escapeHtml(unsafe) {
    return unsafe
        .toString()
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

async function openEditingWindow(id) {
    try {
        // Получаем данные заметки
        const response = await fetch(`${API_URL}/notes/${id}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        if (!response.ok) throw new Error('Failed to fetch note');

        const note = await response.json();

        // Заполняем форму данными заметки
        currentlyEditingNoteId = id;
        noteFormTitle.textContent = 'Edit Note';
        titleInput.value = note.title;
        contentInput.value = note.content;
        submitBtn.textContent = 'Update';
        cancelBtn.style.display = 'inline-block';

        // Прокручиваем к форме
        document.querySelector('.note-form-container').scrollIntoView({ behavior: 'smooth' });
    } catch (error) {
        console.error("Error opening edit window:", error);
        alert('Failed to load note for editing');
    }
}

function newNoteMode() {
    currentlyEditingNoteId = null;
    noteFormTitle.textContent = 'Create New Note';
    titleInput.value = '';
    contentInput.value = '';
    submitBtn.textContent = 'Create';
    cancelBtn.style.display = 'none';
}

// Обновляем функцию saveNote для использования глобальных переменных
async function saveNote() {
    const title = titleInput.value;
    const content = contentInput.value;

    if (!title || !content) {
        alert('Please fill in both title and content');
        return;
    }

    try {
        if (currentlyEditingNoteId) {
            await editNote(currentlyEditingNoteId, title, content);
        } else {
            await createNote(title, content);
        }

        newNoteMode();
        loadNotes();
    } catch (error) {
        console.error("Error saving note:", error);
        alert('Failed to save note');
    }
}

async function createNote(title, content) {
    await fetch(`${API_URL}/notes`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ title, content })
    });
}

async function deleteNote(id) {
    if (!confirm('Are you sure you want to delete this note?')) return;

    try {
        await fetch(`${API_URL}/notes/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });
        loadNotes();
    } catch (error) {
        console.error("Error deleting note:", error);
    }
}

async function editNote(id, title, content) {
    const response = await fetch(`${API_URL}/notes/${id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ title, content })
    });

    if (!response.ok) throw new Error('Failed to update note');
}