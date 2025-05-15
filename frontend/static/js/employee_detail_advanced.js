// employee_detail_advanced.js
document.addEventListener('DOMContentLoaded', function() {
    // Tab-Funktionalität initialisieren
    initTabs();

    // Überprüfen, ob ein Hash in der URL vorhanden ist
    handleUrlHash();

    // Wenn Vacation-Tab existiert, Vakation-Filter initialisieren
    if (document.getElementById('vacation-tab')) {
        initVacationFilters();
    }

    // Form-Handler initialisieren
    initFormHandlers();
});

// ========== TAB-FUNKTIONEN ==========
function initTabs() {
    showTab('personal'); // Standardmäßig den 'personal' Tab anzeigen

    // Tab-Button-Handler hinzufügen
    const tabButtons = document.querySelectorAll('.tab-btn');
    if (tabButtons.length > 0) {
        tabButtons.forEach(btn => {
            btn.addEventListener('click', function() {
                const tab = this.getAttribute('data-tab');
                showTab(tab);
            });
        });
    }
}

function showTab(tabId) {
    // Alle Tab-Inhalte ausblenden
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.add('hidden');
    });

    // Aktive Klasse von allen Tab-Buttons entfernen
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('bg-green-100', 'text-green-700');
        btn.classList.add('text-gray-500', 'hover:text-gray-700', 'hover:bg-gray-100');
    });

    // Gewählten Tab-Inhalt anzeigen
    const tabElement = document.getElementById(tabId + '-tab');
    if (tabElement) {
        tabElement.classList.remove('hidden');
    }

    // Aktive Klasse zum gewählten Tab-Button hinzufügen
    const activeBtn = document.querySelector(`.tab-btn[data-tab="${tabId}"]`);
    if (activeBtn) {
        activeBtn.classList.remove('text-gray-500', 'hover:text-gray-700', 'hover:bg-gray-100');
        activeBtn.classList.add('bg-green-100', 'text-green-700');
    }
}

function handleUrlHash() {
    if (window.location.hash) {
        const tabName = window.location.hash.substring(1);
        if (tabName === 'conversations') {
            showTab('conversations');
            setTimeout(function() {
                window.scrollBy({
                    top: 200,
                    behavior: 'smooth'
                });
            }, 100);
        }
    }
}

// ========== VACATION FILTER FUNCTIONS ==========
function initVacationFilters() {
    const yearFilter = document.getElementById('vacation-year-filter');
    const typeFilter = document.getElementById('vacation-type-filter');
    const statusFilter = document.getElementById('vacation-status-filter');

    if (yearFilter && typeFilter && statusFilter) {
        // Auf Änderungen bei den Filtern reagieren
        yearFilter.addEventListener('change', applyVacationFilters);
        typeFilter.addEventListener('change', applyVacationFilters);
        statusFilter.addEventListener('change', applyVacationFilters);

        // Filter initial anwenden
        applyVacationFilters();
    }
}

function applyVacationFilters() {
    const yearFilter = document.getElementById('vacation-year-filter').value;
    const typeFilter = document.getElementById('vacation-type-filter').value;
    const statusFilter = document.getElementById('vacation-status-filter').value;
    const vacationItems = document.querySelectorAll('.vacation-item');
    let visibleCount = 0;

    vacationItems.forEach(item => {
        // Datenattribute abrufen
        const itemYear = item.getAttribute('data-year');
        const itemType = item.getAttribute('data-type');
        const itemStatus = item.getAttribute('data-status');

        // Prüfen, ob das Element alle Filter erfüllt
        const yearMatch = yearFilter === 'all' || yearFilter === itemYear;
        const typeMatch = typeFilter === 'all' || typeFilter === itemType;
        const statusMatch = statusFilter === 'all' || statusFilter === itemStatus;

        // Element anzeigen oder ausblenden
        if (yearMatch && typeMatch && statusMatch) {
            item.style.display = '';
            visibleCount++;
        } else {
            item.style.display = 'none';
        }
    });

    // "Keine Ergebnisse" Meldung aktualisieren
    const noResultsMessage = document.getElementById('no-vacation-results');
    const listContainer = document.getElementById('vacation-list');

    if (visibleCount === 0 && !noResultsMessage && listContainer) {
        const message = document.createElement('p');
        message.id = 'no-vacation-results';
        message.className = 'text-sm text-gray-500 py-4';
        message.textContent = 'Keine Einträge gefunden, die den Filterkriterien entsprechen.';
        listContainer.appendChild(message);
    } else if (visibleCount > 0 && noResultsMessage) {
        noResultsMessage.remove();
    }
}

// ========== MODAL FUNCTIONS ==========
function openModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.remove('hidden');
        document.body.classList.add('overflow-hidden');
    }
}

function closeModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.add('hidden');
        document.body.classList.remove('overflow-hidden');
    }
}

// ========== IMAGE PREVIEW FUNCTION ==========
function previewImage(input) {
    const preview = document.getElementById('profileImagePreview');
    if (!preview || !input.files || !input.files[0]) return;

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
    };
    reader.readAsDataURL(input.files[0]);
}

// ========== CONFIRM DELETION FUNCTIONS ==========
// Gemeinsame Funktion für Löschbestätigungen aller Typen
function confirmDelete(title, message, callback) {
    const titleElement = document.getElementById('confirmationTitle');
    const messageElement = document.getElementById('confirmationMessage');
    const confirmButton = document.getElementById('confirmActionBtn');

    if (titleElement) titleElement.textContent = title;
    if (messageElement) messageElement.textContent = message;
    if (confirmButton) confirmButton.onclick = callback;

    openModal('confirmationModal');
}

// Spezifische Bestätigungsfunktionen
function confirmDeleteDocument(employeeId, documentId, category, relatedId = '') {
    confirmDelete(
        'Dokument löschen',
        'Sind Sie sicher, dass Sie dieses Dokument löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.',
        () => deleteDocument(employeeId, documentId, category, relatedId)
    );
}

function confirmDeleteTraining(employeeId, trainingId) {
    confirmDelete(
        'Weiterbildung löschen',
        'Sind Sie sicher, dass Sie diese Weiterbildung löschen möchten? Alle zugehörigen Dokumente werden ebenfalls gelöscht. Diese Aktion kann nicht rückgängig gemacht werden.',
        () => deleteTraining(employeeId, trainingId)
    );
}

function confirmDeleteEvaluation(employeeId, evaluationId) {
    confirmDelete(
        'Leistungsbeurteilung löschen',
        'Sind Sie sicher, dass Sie diese Leistungsbeurteilung löschen möchten? Alle zugehörigen Dokumente werden ebenfalls gelöscht. Diese Aktion kann nicht rückgängig gemacht werden.',
        () => deleteEvaluation(employeeId, evaluationId)
    );
}

function confirmDeleteAbsence(employeeId, absenceId) {
    confirmDelete(
        'Abwesenheit löschen',
        'Sind Sie sicher, dass Sie diese Abwesenheit löschen möchten? Alle zugehörigen Dokumente werden ebenfalls gelöscht. Diese Aktion kann nicht rückgängig gemacht werden.',
        () => deleteAbsence(employeeId, absenceId)
    );
}

function confirmDeleteDevelopmentItem(employeeId, itemId) {
    confirmDelete(
        'Entwicklungsziel löschen',
        'Sind Sie sicher, dass Sie dieses Entwicklungsziel löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.',
        () => deleteDevelopmentItem(employeeId, itemId)
    );
}

function confirmDeleteConversation(employeeId, conversationId) {
    confirmDelete(
        'Gespräch löschen',
        'Sind Sie sicher, dass Sie dieses Gespräch löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.',
        () => deleteConversation(employeeId, conversationId)
    );
}

// ========== DOCUMENT UPLOAD FUNCTIONS ==========
function openDocumentUpload(category, relatedId) {
    const docCategory = document.getElementById('docCategory');
    const docRelatedId = document.getElementById('docRelatedId');

    if (docCategory) docCategory.value = category;
    if (docRelatedId) docRelatedId.value = relatedId;

    openModal('uploadDocumentModal');
}

// Funktionskurzformen für verschiedene Dokumenttypen
function openTrainingDocumentUpload(trainingId) {
    openDocumentUpload('training', trainingId);
}

function openEvaluationDocumentUpload(evaluationId) {
    openDocumentUpload('evaluation', evaluationId);
}

function openAbsenceDocumentUpload(absenceId) {
    openDocumentUpload('absence', absenceId);
}

// ========== AJAX DELETE FUNCTIONS ==========
// Generische DELETE-Funktion für AJAX-Anfragen
function performDelete(url, errorMessage) {
    return fetch(url, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(errorMessage);
            }
            return response.json();
        })
        .then(() => {
            closeModal('confirmationModal');
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// Spezifische DELETE-Funktionen
function deleteDocument(employeeId, documentId, category, relatedId = '') {
    let url = `/employees/${employeeId}/documents/${documentId}?category=${category}`;
    if (relatedId) url += `&relatedId=${relatedId}`;

    performDelete(url, 'Fehler beim Löschen des Dokuments');
}

function deleteTraining(employeeId, trainingId) {
    performDelete(
        `/employees/${employeeId}/trainings/${trainingId}`,
        'Fehler beim Löschen der Weiterbildung'
    );
}

function deleteEvaluation(employeeId, evaluationId) {
    performDelete(
        `/employees/${employeeId}/evaluations/${evaluationId}`,
        'Fehler beim Löschen der Leistungsbeurteilung'
    );
}

function deleteAbsence(employeeId, absenceId) {
    performDelete(
        `/employees/${employeeId}/absences/${absenceId}`,
        'Fehler beim Löschen der Abwesenheit'
    );
}

function deleteDevelopmentItem(employeeId, itemId) {
    performDelete(
        `/employees/${employeeId}/development/${itemId}`,
        'Fehler beim Löschen des Entwicklungsziels'
    );
}

function deleteConversation(employeeId, conversationId) {
    performDelete(
        `/employees/${employeeId}/conversations/${conversationId}`,
        'Fehler beim Löschen des Gesprächs'
    );
}

// ========== FORM SUBMISSION HANDLERS ==========
function initFormHandlers() {
    // Array mit Form-IDs und ihren entsprechenden Aktionsbeschreibungen
    const forms = [
        { id: 'uploadProfileImageForm', error: 'Fehler beim Hochladen des Profilbilds', success: 'Profilbild erfolgreich hochgeladen' },
        { id: 'uploadDocumentForm', error: 'Fehler beim Hochladen des Dokuments' },
        { id: 'addTrainingForm', error: 'Fehler beim Hinzufügen der Weiterbildung' },
        { id: 'addEvaluationForm', error: 'Fehler beim Hinzufügen der Leistungsbeurteilung' },
        { id: 'addAbsenceForm', error: 'Fehler beim Hinzufügen der Abwesenheit' },
        { id: 'addDevelopmentItemForm', error: 'Fehler beim Hinzufügen des Entwicklungsziels' },
        { id: 'addConversationForm', error: 'Fehler beim Hinzufügen des Gesprächs', handler: addConversationHandler }
    ];

    forms.forEach(form => {
        const formElement = document.getElementById(form.id);
        if (formElement) {
            formElement.addEventListener('submit', form.handler || function(e) {
                handleFormSubmit(e, form.id, form.error, form.success);
            });
        }
    });
}

// Generischer Form-Submission-Handler
function handleFormSubmit(e, formId, errorMessage, successMessage) {
    e.preventDefault();
    const form = document.getElementById(formId);
    const formData = new FormData(form);

    fetch(form.action, {
        method: form.method || 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || errorMessage);
                });
            }
            return response.json();
        })
        .then(data => {
            if (formId.includes('Modal')) {
                closeModal(formId.replace('Form', 'Modal'));
            }
            if (successMessage) {
                alert(successMessage);
            }
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// Spezieller Handler für Conversations
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
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// ========== CONVERSATION FUNCTIONS ==========
function openEditConversationModal(id, title, description, date, notes) {
    // Form und Elemente abrufen
    const form = document.getElementById('addConversationForm');
    if (!form) return;

    // Form umkonfigurieren
    form.action = `/employees/{{.employee.ID.Hex}}/conversations/${id}`;
    form.method = 'POST';

    // Felder befüllen
    const fields = {
        'conversationTitle': title,
        'conversationDate': date,
        'conversationDescription': description,
        'conversationNotes': notes || ''
    };

    Object.entries(fields).forEach(([fieldId, value]) => {
        const field = document.getElementById(fieldId);
        if (field) field.value = value;
    });

    // Überschrift ändern
    const modalTitle = document.querySelector('#addConversationModal .text-lg.font-medium');
    if (modalTitle) modalTitle.textContent = 'Gespräch bearbeiten';

    // Button-Text ändern
    const submitButton = form.querySelector('button[type="submit"]');
    if (submitButton) submitButton.textContent = 'Speichern';

    // ID-Feld hinzufügen/aktualisieren
    let idField = form.querySelector('input[name="id"]');
    if (!idField) {
        idField = document.createElement('input');
        idField.type = 'hidden';
        idField.name = 'id';
        form.appendChild(idField);
    }
    idField.value = id;

    // Event-Listener austauschen
    form.removeEventListener('submit', addConversationHandler);
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        updateConversation(id);
    });

    // Modal öffnen
    openModal('addConversationModal');
}

function updateConversation(conversationId) {
    const form = document.getElementById('addConversationForm');
    const formData = new FormData(form);

    fetch(`/employees/{{.employee.ID.Hex}}/conversations/${conversationId}`, {
        method: 'PUT',
        body: formData
    })
        .then(response => {
            if (!response.ok) throw new Error('Fehler beim Aktualisieren des Gesprächs');
            return response.json();
        })
        .then(data => {
            // Form zurücksetzen
            form.reset();

            // Modal-Titel und Button-Text zurücksetzen
            const modalTitle = document.querySelector('#addConversationModal .text-lg.font-medium');
            if (modalTitle) modalTitle.textContent = 'Gespräch hinzufügen';

            const submitButton = form.querySelector('button[type="submit"]');
            if (submitButton) submitButton.textContent = 'Hinzufügen';

            // Event-Listener zurücksetzen
            form.removeEventListener('submit', updateConversation);
            form.addEventListener('submit', addConversationHandler);

            // Modal schließen und Seite neu laden
            closeModal('addConversationModal');
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

function markConversationCompleted(employeeId, conversationId) {
    fetch(`/employees/${employeeId}/conversations/${conversationId}/complete`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) throw new Error('Fehler beim Markieren des Gesprächs als abgeschlossen');
            return response.json();
        })
        .then(data => {
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

// ========== ABSENCE FUNCTIONS ==========
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
            window.location.reload();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Ein Fehler ist aufgetreten: ' + error.message);
        });
}

function openAbsenceDocumentsModal(absenceId) {
    // Stub-Funktion, noch nicht implementiert
    alert('Funktion noch nicht implementiert');
}


// ============ TIME TRACKING ==============
document.addEventListener('DOMContentLoaded', function() {
    // Check if the time entries tab is selected and initialize chart
    const timeEntriesTab = document.getElementById('timeentries-tab');
    const timeDistributionChart = document.getElementById('timeDistributionChart');

    if (timeEntriesTab && timeDistributionChart) {
        // Get project data from the rendered page
        const projectLabels = JSON.parse('{{.projectLabels | json}}' || '[]');
        const projectHours = JSON.parse('{{.projectHours | json}}' || '[]');

        // Initialize chart when the tab is clicked
        document.querySelector('[data-tab="timeentries"]').addEventListener('click', function() {
            if (!window.timeChart) {
                window.timeChart = new Chart(timeDistributionChart, {
                    type: 'pie',
                    data: {
                        labels: projectLabels,
                        datasets: [{
                            data: projectHours,
                            backgroundColor: [
                                '#10B981', // green-500
                                '#3B82F6', // blue-500
                                '#F59E0B', // amber-500
                                '#EF4444', // red-500
                                '#8B5CF6', // purple-500
                                '#EC4899', // pink-500
                                '#14B8A6', // teal-500
                                '#F97316', // orange-500
                                '#6366F1', // indigo-500
                                '#06B6D4'  // cyan-500
                            ],
                            borderWidth: 1
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        legend: {
                            position: 'right',
                        },
                        tooltips: {
                            callbacks: {
                                label: function(tooltipItem, data) {
                                    const dataset = data.datasets[tooltipItem.datasetIndex];
                                    const total = dataset.data.reduce((acc, val) => acc + val, 0);
                                    const currentValue = dataset.data[tooltipItem.index];
                                    const percentage = Math.round((currentValue / total) * 100);
                                    return `${data.labels[tooltipItem.index]}: ${currentValue.toFixed(2)} Std. (${percentage}%)`;
                                }
                            }
                        }
                    }
                });
            }
        });
    }
});