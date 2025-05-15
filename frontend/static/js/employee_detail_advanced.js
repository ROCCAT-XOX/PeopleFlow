// Add this script to your employee detail page or include it in a separate file

document.addEventListener('DOMContentLoaded', function() {
    // Only initialize if the vacation tab exists
    if (document.getElementById('vacation-tab')) {
        // Initialize vacation filters
        initVacationFilters();
    }
});

// Initialize vacation filters
function initVacationFilters() {
    const yearFilter = document.getElementById('vacation-year-filter');
    const typeFilter = document.getElementById('vacation-type-filter');
    const statusFilter = document.getElementById('vacation-status-filter');

    // Apply filters when any filter changes
    yearFilter.addEventListener('change', applyVacationFilters);
    typeFilter.addEventListener('change', applyVacationFilters);
    statusFilter.addEventListener('change', applyVacationFilters);

    // Initial filter application
    applyVacationFilters();
}

// Apply all vacation filters
function applyVacationFilters() {
    const yearFilter = document.getElementById('vacation-year-filter').value;
    const typeFilter = document.getElementById('vacation-type-filter').value;
    const statusFilter = document.getElementById('vacation-status-filter').value;

    const vacationItems = document.querySelectorAll('.vacation-item');

    vacationItems.forEach(item => {
        // Get data attributes
        const itemYear = item.getAttribute('data-year');
        const itemType = item.getAttribute('data-type');
        const itemStatus = item.getAttribute('data-status');

        // Check if the item passes all filters
        const yearMatch = yearFilter === 'all' || yearFilter === itemYear;
        const typeMatch = typeFilter === 'all' || typeFilter === itemType;
        const statusMatch = statusFilter === 'all' || statusFilter === itemStatus;

        // Show or hide based on filter matches
        if (yearMatch && typeMatch && statusMatch) {
            item.style.display = '';
        } else {
            item.style.display = 'none';
        }
    });

    // Update "no results" message
    const visibleItems = Array.from(vacationItems).filter(item => item.style.display !== 'none');
    const noResultsMessage = document.getElementById('no-vacation-results');

    if (visibleItems.length === 0 && !noResultsMessage) {
        const listContainer = document.getElementById('vacation-list');
        if (listContainer) {
            const message = document.createElement('p');
            message.id = 'no-vacation-results';
            message.className = 'text-sm text-gray-500 py-4';
            message.textContent = 'Keine Einträge gefunden, die den Filterkriterien entsprechen.';
            listContainer.appendChild(message);
        }
    } else if (visibleItems.length > 0 && noResultsMessage) {
        noResultsMessage.remove();
    }
}

// Fügen Sie diese Funktion dem bestehenden JavaScript in employee_detail_advanced.html hinzu
function previewImage(input) {
    const preview = document.getElementById('profileImagePreview');

    if (input.files && input.files[0]) {
        const reader = new FileReader();

        reader.onload = function(e) {
            if (preview.tagName.toLowerCase() === 'img') {
                preview.src = e.target.result;
            } else {
                // Wenn es ein DIV ist (Standard-Avatar), ersetzen wir es durch ein Image
                const img = document.createElement('img');
                img.src = e.target.result;
                img.className = 'h-32 w-32 rounded-full mx-auto object-cover';
                img.id = 'profileImagePreview';
                preview.parentNode.replaceChild(img, preview);
            }
        }

        reader.readAsDataURL(input.files[0]);
    }
}

// In employee_detail_advanced.html
document.addEventListener('DOMContentLoaded', function() {
    const uploadProfileImageForm = document.getElementById('uploadProfileImageForm');
    if (uploadProfileImageForm) {
        uploadProfileImageForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            fetch(this.action, {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    if (!response.ok) {
                        return response.json().then(data => {
                            throw new Error(data.error || 'Fehler beim Hochladen des Profilbilds');
                        });
                    }
                    return response.json();
                })
                .then(data => {
                    closeModal('uploadProfileImageModal');
                    // Display success message
                    alert('Profilbild erfolgreich hochgeladen');
                    // Reload the page to show the updated profile image
                    window.location.reload();
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ein Fehler ist aufgetreten: ' + error.message);
                });
        });
    }
});

// Modal functions
function openModal(id) {
    document.getElementById(id).classList.remove('hidden');
    document.body.classList.add('overflow-hidden');
}

function closeModal(id) {
    document.getElementById(id).classList.add('hidden');
    document.body.classList.remove('overflow-hidden');
}

// Document upload for training
function openTrainingDocumentUpload(trainingId) {
    const docCategory = document.getElementById('docCategory');
    const docRelatedId = document.getElementById('docRelatedId');

    // Set the category and related ID
    if (docCategory) docCategory.value = 'training';
    if (docRelatedId) docRelatedId.value = trainingId;

    // Open the modal
    openModal('uploadDocumentModal');
}

// Document upload for evaluation
function openEvaluationDocumentUpload(evaluationId) {
    const docCategory = document.getElementById('docCategory');
    const docRelatedId = document.getElementById('docRelatedId');

    // Set the category and related ID
    if (docCategory) docCategory.value = 'evaluation';
    if (docRelatedId) docRelatedId.value = evaluationId;

    // Open the modal
    openModal('uploadDocumentModal');
}

// Document upload for absence
function openAbsenceDocumentUpload(absenceId) {
    const docCategory = document.getElementById('docCategory');
    const docRelatedId = document.getElementById('docRelatedId');

    // Set the category and related ID
    if (docCategory) docCategory.value = 'absence';
    if (docRelatedId) docRelatedId.value = absenceId;

    // Open the modal
    openModal('uploadDocumentModal');
}

// Document deletion confirmation
function confirmDeleteDocument(employeeId, documentId, category, relatedId = '') {
    const title = document.getElementById('confirmationTitle');
    const message = document.getElementById('confirmationMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');

    if (title) title.textContent = 'Dokument löschen';
    if (message) message.textContent = 'Sind Sie sicher, dass Sie dieses Dokument löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.';

    if (confirmBtn) {
        confirmBtn.onclick = function() {
            deleteDocument(employeeId, documentId, category, relatedId);
        };
    }

    openModal('confirmationModal');
}

// Training deletion confirmation
function confirmDeleteTraining(employeeId, trainingId) {
    const title = document.getElementById('confirmationTitle');
    const message = document.getElementById('confirmationMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');

    if (title) title.textContent = 'Weiterbildung löschen';
    if (message) message.textContent = 'Sind Sie sicher, dass Sie diese Weiterbildung löschen möchten? Alle zugehörigen Dokumente werden ebenfalls gelöscht. Diese Aktion kann nicht rückgängig gemacht werden.';

    if (confirmBtn) {
        confirmBtn.onclick = function() {
            deleteTraining(employeeId, trainingId);
        };
    }

    openModal('confirmationModal');
}

// Evaluation deletion confirmation
function confirmDeleteEvaluation(employeeId, evaluationId) {
    const title = document.getElementById('confirmationTitle');
    const message = document.getElementById('confirmationMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');

    if (title) title.textContent = 'Leistungsbeurteilung löschen';
    if (message) message.textContent = 'Sind Sie sicher, dass Sie diese Leistungsbeurteilung löschen möchten? Alle zugehörigen Dokumente werden ebenfalls gelöscht. Diese Aktion kann nicht rückgängig gemacht werden.';

    if (confirmBtn) {
        confirmBtn.onclick = function() {
            deleteEvaluation(employeeId, evaluationId);
        };
    }

    openModal('confirmationModal');
}

// Absence deletion confirmation
function confirmDeleteAbsence(employeeId, absenceId) {
    const title = document.getElementById('confirmationTitle');
    const message = document.getElementById('confirmationMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');

    if (title) title.textContent = 'Abwesenheit löschen';
    if (message) message.textContent = 'Sind Sie sicher, dass Sie diese Abwesenheit löschen möchten? Alle zugehörigen Dokumente werden ebenfalls gelöscht. Diese Aktion kann nicht rückgängig gemacht werden.';

    if (confirmBtn) {
        confirmBtn.onclick = function() {
            deleteAbsence(employeeId, absenceId);
        };
    }

    openModal('confirmationModal');
}

// Development item deletion confirmation
function confirmDeleteDevelopmentItem(employeeId, itemId) {
    const title = document.getElementById('confirmationTitle');
    const message = document.getElementById('confirmationMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');

    if (title) title.textContent = 'Entwicklungsziel löschen';
    if (message) message.textContent = 'Sind Sie sicher, dass Sie dieses Entwicklungsziel löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.';

    if (confirmBtn) {
        confirmBtn.onclick = function() {
            deleteDevelopmentItem(employeeId, itemId);
        };
    }

    openModal('confirmationModal');
}

// AJAX functions for deleting items
function deleteDocument(employeeId, documentId, category, relatedId = '') {
    let url = `/employees/${employeeId}/documents/${documentId}?category=${category}`;
    if (relatedId) {
        url += `&relatedId=${relatedId}`;
    }

    fetch(url, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Löschen des Dokuments');
            }
            return response.json();
        })
        .then(data => {
            closeModal('confirmationModal');
            // Reload the page to show the updated data
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

function deleteTraining(employeeId, trainingId) {
    fetch(`/employees/${employeeId}/trainings/${trainingId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Löschen der Weiterbildung');
            }
            return response.json();
        })
        .then(data => {
            closeModal('confirmationModal');
            // Reload the page to show the updated data
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

function deleteEvaluation(employeeId, evaluationId) {
    fetch(`/employees/${employeeId}/evaluations/${evaluationId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Löschen der Leistungsbeurteilung');
            }
            return response.json();
        })
        .then(data => {
            closeModal('confirmationModal');
            // Reload the page to show the updated data
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

function deleteAbsence(employeeId, absenceId) {
    fetch(`/employees/${employeeId}/absences/${absenceId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Löschen der Abwesenheit');
            }
            return response.json();
        })
        .then(data => {
            closeModal('confirmationModal');
            // Reload the page to show the updated data
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

function deleteDevelopmentItem(employeeId, itemId) {
    fetch(`/employees/${employeeId}/development/${itemId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Löschen des Entwicklungsziels');
            }
            return response.json();
        })
        .then(data => {
            closeModal('confirmationModal');
            // Reload the page to show the updated data
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// Function to approve or reject absence
function approveAbsence(employeeId, absenceId, action) {
    const formData = new FormData();
    formData.append('action', action);

    fetch(`/employees/${employeeId}/absences/${absenceId}/approve`, {
        method: 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Fehler beim ${action === 'approve' ? 'Genehmigen' : 'Ablehnen'} der Abwesenheit`);
            }
            return response.json();
        })
        .then(data => {
            // Reload the page to show the updated data
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// Function to show absence documents in a modal
function openAbsenceDocumentsModal(absenceId) {
    // This would need a new modal to be implemented
    alert('Funktion noch nicht implementiert');
}

// Function to display a specific tab
function showTab(tabId) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.add('hidden');
    });

    // Remove active class from all tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('bg-green-100', 'text-green-700');
        btn.classList.add('text-gray-500', 'hover:text-gray-700', 'hover:bg-gray-100');
    });

    // Show the selected tab content
    document.getElementById(tabId + '-tab').classList.remove('hidden');

    // Add active class to the selected tab button
    const activeBtn = document.querySelector(`.tab-btn[data-tab="${tabId}"]`);
    activeBtn.classList.remove('text-gray-500', 'hover:text-gray-700', 'hover:bg-gray-100');
    activeBtn.classList.add('bg-green-100', 'text-green-700');
}

// Initialize page with the first tab active
document.addEventListener('DOMContentLoaded', function() {
    showTab('personal');

    // Form submission handlers
    const uploadDocumentForm = document.getElementById('uploadDocumentForm');
    if (uploadDocumentForm) {
        uploadDocumentForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            fetch(this.action, {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Fehler beim Hochladen des Dokuments');
                    }
                    return response.json();
                })
                .then(data => {
                    closeModal('uploadDocumentModal');
                    // Reload the page to show the updated data
                    window.location.reload();
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ein Fehler ist aufgetreten: ' + error.message);
                });
        });
    }

    const addTrainingForm = document.getElementById('addTrainingForm');
    if (addTrainingForm) {
        addTrainingForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            fetch(this.action, {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Fehler beim Hinzufügen der Weiterbildung');
                    }
                    return response.json();
                })
                .then(data => {
                    closeModal('addTrainingModal');
                    // Reload the page to show the updated data
                    window.location.reload();
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ein Fehler ist aufgetreten: ' + error.message);
                });
        });
    }

    const addEvaluationForm = document.getElementById('addEvaluationForm');
    if (addEvaluationForm) {
        addEvaluationForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            fetch(this.action, {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Fehler beim Hinzufügen der Leistungsbeurteilung');
                    }
                    return response.json();
                })
                .then(data => {
                    closeModal('addEvaluationModal');
                    // Reload the page to show the updated data
                    window.location.reload();
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ein Fehler ist aufgetreten: ' + error.message);
                });
        });
    }

    const addAbsenceForm = document.getElementById('addAbsenceForm');
    if (addAbsenceForm) {
        addAbsenceForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            fetch(this.action, {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Fehler beim Hinzufügen der Abwesenheit');
                    }
                    return response.json();
                })
                .then(data => {
                    closeModal('addAbsenceModal');
                    // Reload the page to show the updated data
                    window.location.reload();
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ein Fehler ist aufgetreten: ' + error.message);
                });
        });
    }

    const addDevelopmentItemForm = document.getElementById('addDevelopmentItemForm');
    if (addDevelopmentItemForm) {
        addDevelopmentItemForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            fetch(this.action, {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Fehler beim Hinzufügen des Entwicklungsziels');
                    }
                    return response.json();
                })
                .then(data => {
                    closeModal('addDevelopmentItemModal');
                    // Reload the page to show the updated data
                    window.location.reload();
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ein Fehler ist aufgetreten: ' + error.message);
                });
        });
    }
});

// Löschen-Bestätigung für ein Gespräch
function confirmDeleteConversation(employeeId, conversationId) {
    const title = document.getElementById('confirmationTitle');
    const message = document.getElementById('confirmationMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');

    if (title) title.textContent = 'Gespräch löschen';
    if (message) message.textContent = 'Sind Sie sicher, dass Sie dieses Gespräch löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.';

    if (confirmBtn) {
        confirmBtn.onclick = function() {
            deleteConversation(employeeId, conversationId);
        };
    }

    openModal('confirmationModal');
}

// Gespräch löschen
function deleteConversation(employeeId, conversationId) {
    fetch(`/employees/${employeeId}/conversations/${conversationId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Löschen des Gesprächs');
            }
            return response.json();
        })
        .then(data => {
            closeModal('confirmationModal');
            // Seite neu laden, um die Änderungen anzuzeigen
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// Gespräch als abgeschlossen markieren
// Ergänzung der bestehenden Funktion zum Markieren als abgeschlossen
function markConversationCompleted(employeeId, conversationId) {
    fetch(`/employees/${employeeId}/conversations/${conversationId}/complete`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Markieren des Gesprächs als abgeschlossen');
            }
            return response.json();
        })
        .then(data => {
            // Seite neu laden, um die Änderungen anzuzeigen
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// Formular-Handling für das Hinzufügen von Gesprächen
document.addEventListener('DOMContentLoaded', function() {
    const addConversationForm = document.getElementById('addConversationForm');
    if (addConversationForm) {
        addConversationForm.addEventListener('submit', addConversationHandler);
    }
});

// Funktion zum Öffnen des Bearbeitungsmodals
function openEditConversationModal(id, title, description, date, notes) {
    // Das bestehende Modal-Formular verwenden
    const form = document.getElementById('addConversationForm');

    // Umbennen der Form-Action
    form.action = `/employees/{{.employee.ID.Hex}}/conversations/${id}`;
    form.method = 'POST';

    // Formularfelder ausfüllen
    const titleField = document.getElementById('conversationTitle');
    const dateField = document.getElementById('conversationDate');
    const descriptionField = document.getElementById('conversationDescription');
    const notesField = document.getElementById('conversationNotes');

    if (titleField) titleField.value = title;
    if (dateField) dateField.value = date;
    if (descriptionField) descriptionField.value = description;
    if (notesField) notesField.value = notes || '';

    // Überschrift ändern
    const modalTitle = document.querySelector('#addConversationModal .text-lg.font-medium');
    if (modalTitle) modalTitle.textContent = 'Gespräch bearbeiten';

    // Button-Text ändern
    const submitButton = form.querySelector('button[type="submit"]');
    if (submitButton) submitButton.textContent = 'Speichern';

    // Hidden Input für die ID hinzufügen (falls nicht vorhanden)
    let idField = form.querySelector('input[name="id"]');
    if (!idField) {
        idField = document.createElement('input');
        idField.type = 'hidden';
        idField.name = 'id';
        form.appendChild(idField);
    }
    idField.value = id;

    // Event-Listener hinzufügen (vorhandene entfernen)
    form.removeEventListener('submit', addConversationHandler);
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        updateConversation(id);
    });

    // Modal öffnen
    openModal('addConversationModal');
}

// Funktion zum Aktualisieren eines Gesprächs
function updateConversation(conversationId) {
    const form = document.getElementById('addConversationForm');
    const formData = new FormData(form);

    // AJAX-Request für das Update
    fetch(`/employees/{{.employee.ID.Hex}}/conversations/${conversationId}`, {
        method: 'PUT',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Aktualisieren des Gesprächs');
            }
            return response.json();
        })
        .then(data => {
            // Formular zurücksetzen
            form.reset();

            // Modal-Titel und Button-Text zurücksetzen
            const modalTitle = document.querySelector('#addConversationModal .text-lg.font-medium');
            if (modalTitle) modalTitle.textContent = 'Gespräch hinzufügen';

            const submitButton = form.querySelector('button[type="submit"]');
            if (submitButton) submitButton.textContent = 'Hinzufügen';

            // Event-Listener für das Hinzufügen wiederherstellen
            form.removeEventListener('submit', updateConversation);
            form.addEventListener('submit', addConversationHandler);

            // Modal schließen
            closeModal('addConversationModal');

            // Seite neu laden
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}
function addConversationHandler(e) {
    e.preventDefault();
    const formData = new FormData(this);

    fetch(this.action, {
        method: 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler beim Hinzufügen des Gesprächs');
            }
            return response.json();
        })
        .then(data => {
            closeModal('addConversationModal');
            // Seite neu laden, um die Änderungen anzuzeigen
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

document.addEventListener('DOMContentLoaded', function() {
    // Prüfen, ob ein Hashtag in der URL ist
    if (window.location.hash) {
        // Extrahieren des Tab-Namens aus dem Hash
        const tabName = window.location.hash.substring(1); // Entfernt das # vom Anfang

        // Wenn der Hash dem Tab "conversations" entspricht, diesen Tab anzeigen
        if (tabName === 'conversations') {
            showTab('conversations');

            // Scrolle ein bisschen nach unten, damit der Tab in der Mitte der Seite ist
            setTimeout(function() {
                window.scrollBy({
                    top: 200,
                    behavior: 'smooth'
                });
            }, 100);
        }
    }
});

