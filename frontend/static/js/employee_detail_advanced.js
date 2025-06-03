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

    // Anpassungen für Überstunden-Tab laden
    if (document.getElementById('overtime-tab')) {
        loadEmployeeAdjustments();
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

// Tab-Funktionen erweitern für den neuen Überstunden-Tab
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

    // Spezielle Behandlung für Überstunden-Tab
    if (tabId === 'overtime') {
        // Überstunden-Chart neu laden falls nötig
        setTimeout(() => {
            const canvas = document.getElementById('overtimeChart');
            if (canvas && typeof Chart !== 'undefined') {
                // Chart nur neu initialisieren wenn er noch nicht existiert
                if (!canvas.chart) {
                    initializeOvertimeChart();
                }
            }
        }, 100);
    }
}

// URL-Hash-Behandlung erweitern
function handleUrlHash() {
    if (window.location.hash) {
        const tabName = window.location.hash.substring(1);
        const validTabs = ['personal', 'documents', 'trainings', 'development', 'projects', 'vacation', 'conversations', 'timeentries', 'overtime'];

        if (validTabs.includes(tabName)) {
            showTab(tabName);
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

    // Überstunden-Button Event-Listener hinzufügen
    const recalculateBtn = document.getElementById('recalculateEmployeeOvertimeBtn');
    if (recalculateBtn) {
        // Entferne onclick Attribute falls vorhanden
        recalculateBtn.removeAttribute('onclick');
        recalculateBtn.addEventListener('click', recalculateEmployeeOvertime);
    }
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
// Überstunden für spezifischen Mitarbeiter neu berechnen
function recalculateEmployeeOvertime() {
    const employeeId = getEmployeeIdFromUrl();
    const button = document.getElementById('recalculateEmployeeOvertimeBtn');

    if (!button) {
        console.error('Recalculate button not found');
        return;
    }

    // Button deaktivieren und Loading-State anzeigen
    button.disabled = true;
    const originalText = button.innerHTML;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Berechne...
    `;

    // API-Aufruf zur Neuberechnung
    fetch(`/api/timetracking/employee/${employeeId}/overtime`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Fehler bei der Überstunden-Berechnung');
            }
            return response.json();
        })
        .then(data => {
            if (data.success) {
                // Erfolgsmeldung anzeigen
                showNotification('Überstunden erfolgreich neu berechnet', 'success');

                // Überstunden-Balance aktualisieren
                updateOvertimeDisplay(data.data);

                // Seite nach kurzer Verzögerung neu laden für vollständige Aktualisierung
                setTimeout(() => {
                    window.location.reload();
                }, 1500);
            } else {
                throw new Error(data.error || 'Unbekannter Fehler');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Fehler beim Berechnen der Überstunden: ' + error.message, 'error');
        })
        .finally(() => {
            // Button zurücksetzen
            button.disabled = false;
            button.innerHTML = originalText;
        });
}

// Event-Listener für den Überstunden-Button hinzufügen
document.addEventListener('DOMContentLoaded', function() {
    // Bestehende Initialization...

    // Überstunden-Button Event-Listener
    const recalculateBtn = document.getElementById('recalculateEmployeeOvertimeBtn');
    if (recalculateBtn) {
        recalculateBtn.addEventListener('click', recalculateEmployeeOvertime);
    }
});

// Hilfsfunktion: Employee ID aus URL extrahieren
function getEmployeeIdFromUrl() {
    const pathParts = window.location.pathname.split('/');
    const viewIndex = pathParts.indexOf('view');
    if (viewIndex !== -1 && pathParts[viewIndex + 1]) {
        return pathParts[viewIndex + 1];
    }
    return null;
}

// Hilfsfunktion: Überstunden-Display aktualisieren
function updateOvertimeDisplay(data) {
    const balanceElement = document.getElementById('currentOvertimeBalance');
    if (balanceElement && data.overtimeBalance !== undefined) {
        const balance = data.overtimeBalance;
        let balanceHtml = '';

        if (balance >= 0) {
            balanceHtml = `<span class="text-green-600">+${balance.toFixed(1)} Std</span>`;
        } else {
            balanceHtml = `<span class="text-red-600">${balance.toFixed(1)} Std</span>`;
        }

        balanceElement.innerHTML = balanceHtml;
    }
}

// Hilfsfunktion: Benachrichtigungen anzeigen
function showNotification(message, type = 'info') {
    // Erstelle ein einfaches Notification-Element
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 z-50 max-w-sm w-full p-4 rounded-md shadow-lg transform transition-all duration-300 ${
        type === 'success' ? 'bg-green-100 border border-green-500 text-green-700' :
            type === 'error' ? 'bg-red-100 border border-red-500 text-red-700' :
                'bg-blue-100 border border-blue-500 text-blue-700'
    }`;

    notification.innerHTML = `
        <div class="flex items-center">
            <div class="flex-shrink-0">
                ${type === 'success' ?
        '<svg class="h-5 w-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>' :
        type === 'error' ?
            '<svg class="h-5 w-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>' :
            '<svg class="h-5 w-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>'
    }
            </div>
            <div class="ml-3">
                <p class="text-sm font-medium">${message}</p>
            </div>
            <div class="ml-auto pl-3">
                <button onclick="this.parentElement.parentElement.parentElement.remove()" class="inline-flex text-gray-400 hover:text-gray-600">
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
            </div>
        </div>
    `;

    document.body.appendChild(notification);

    // Animation beim Erscheinen
    setTimeout(() => {
        notification.style.opacity = '1';
        notification.style.transform = 'translateX(0)';
    }, 100);

    // Automatisch nach 5 Sekunden entfernen
    setTimeout(() => {
        if (notification.parentElement) {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.remove();
                }
            }, 300);
        }
    }, 5000);
}

// Überstunden-Chart initialisieren (falls Chart.js verfügbar ist)
function initializeOvertimeChart() {
    const canvas = document.getElementById('overtimeChart');
    if (!canvas || typeof Chart === 'undefined') {
        return;
    }

    // Verhindere mehrfache Initialisierung
    if (canvas.chart) {
        return;
    }

    // Beispiel-Daten - diese sollten vom Server kommen
    const weeklyData = window.employeeWeeklyData || [];

    if (weeklyData.length === 0) {
        canvas.style.display = 'none';
        return;
    }

    const ctx = canvas.getContext('2d');

    canvas.chart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: weeklyData.map(entry => `KW ${entry.weekNumber}`),
            datasets: [{
                label: 'Überstunden',
                data: weeklyData.map(entry => entry.overtimeHours),
                borderColor: 'rgb(34, 197, 94)',
                backgroundColor: 'rgba(34, 197, 94, 0.1)',
                tension: 0.1,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Stunden'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Kalenderwoche'
                    }
                }
            },
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const value = context.parsed.y;
                            return value >= 0 ? `+${value.toFixed(1)} Std` : `${value.toFixed(1)} Std`;
                        }
                    }
                }
            }
        }
    });
}

//========================================== OVERTIME ADJUSTMENT ==================================

function loadEmployeeAdjustments() {
    const employeeId = getEmployeeIdFromUrl();
    if (!employeeId) return;

    fetch(`/api/overtime/employee/${employeeId}/adjustments`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // Sicherstellen, dass data.data ein Array ist, auch wenn es null oder undefined ist
                const adjustments = data.data || [];
                displayEmployeeAdjustments(adjustments);
                updateAdjustmentsSummary(adjustments);
            } else {
                console.error('API returned error:', data.error);
                displayAdjustmentsError('Fehler beim Laden der Anpassungen: ' + (data.error || 'Unbekannter Fehler'));
            }
        })
        .catch(error => {
            console.error('Error loading adjustments:', error);
            displayAdjustmentsError('Fehler beim Laden der Anpassungen: Netzwerkfehler');
        });
}

// Anpassungen anzeigen
function displayEmployeeAdjustments(adjustments) {
    const container = document.getElementById('adjustmentsContainer');
    if (!container) return;

    // Sicherstellen, dass adjustments ein Array ist
    if (!adjustments || !Array.isArray(adjustments) || adjustments.length === 0) {
        container.innerHTML = `
      <div class="text-center py-8">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 4a7.962 7.962 0 016 2.291M6 20.291A7.962 7.962 0 014 12a7.962 7.962 0 012-8.291"></path>
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">Keine Anpassungen</h3>
        <p class="mt-1 text-sm text-gray-500">Für diesen Mitarbeiter wurden noch keine manuellen Überstunden-Anpassungen vorgenommen.</p>
      </div>
    `;
        return;
    }

    const html = adjustments.map(adjustment => {
        const hoursClass = adjustment.hours >= 0 ? 'text-green-600' : 'text-red-600';
        const hoursText = adjustment.hours >= 0 ? `+${adjustment.hours.toFixed(1)}` : adjustment.hours.toFixed(1);
        const statusClass = getStatusClass(adjustment.status);

        return `
      <div class="border border-gray-200 rounded-lg p-4 mb-4">
        <div class="flex justify-between items-start">
          <div class="flex-1">
            <div class="flex items-center space-x-3 mb-2">
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                ${getAdjustmentTypeDisplay(adjustment.type)}
              </span>
              <span class="text-lg font-medium ${hoursClass}">${hoursText} Std</span>
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusClass}">
                ${getStatusDisplay(adjustment.status)}
              </span>
            </div>
            
            <h4 class="text-sm font-medium text-gray-900 mb-1">${adjustment.reason}</h4>
            ${adjustment.description ? `<p class="text-sm text-gray-600 mb-2">${adjustment.description}</p>` : ''}
            
            <div class="text-xs text-gray-500">
              <div>Eingereicht von ${adjustment.adjusterName} am ${new Date(adjustment.createdAt).toLocaleDateString('de-DE', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        })}</div>
              ${adjustment.approverName ? `
                <div class="mt-1">
                  ${adjustment.status === 'approved' ? 'Genehmigt' : 'Abgelehnt'} von ${adjustment.approverName} 
                  am ${new Date(adjustment.approvedAt).toLocaleDateString('de-DE', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        })}
                </div>
              ` : ''}
            </div>
          </div>
          
          ${adjustment.status === 'pending' && (window.userRole === 'admin' || window.userRole === 'manager') ? `
            <div class="flex space-x-2 ml-4">
              <button onclick="approveEmployeeAdjustment('${adjustment.id}', 'approve')" 
                      class="inline-flex items-center px-2 py-1 border border-transparent text-xs font-medium rounded text-green-700 bg-green-100 hover:bg-green-200">
                Genehmigen
              </button>
              <button onclick="approveEmployeeAdjustment('${adjustment.id}', 'reject')" 
                      class="inline-flex items-center px-2 py-1 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200">
                Ablehnen
              </button>
            </div>
          ` : ''}
        </div>
      </div>
    `;
    }).join('');

    container.innerHTML = html;
}

// Anpassungs-Zusammenfassung aktualisieren
function updateAdjustmentsSummary(adjustments) {
    // Sicherstellen, dass adjustments ein Array ist
    if (!adjustments || !Array.isArray(adjustments)) {
        adjustments = [];
    }

    const approvedAdjustments = adjustments.filter(adj => adj.status === 'approved');
    const totalAdjustments = approvedAdjustments.reduce((sum, adj) => sum + adj.hours, 0);

    // Berechnetes Saldo aus Template
    const calculatedBalance = parseFloat('{{.employee.OvertimeBalance}}') || 0;
    const finalBalance = calculatedBalance + totalAdjustments;

    // UI aktualisieren
    const adjustmentsTotalEl = document.getElementById('adjustmentsTotal');
    const finalBalanceEl = document.getElementById('finalBalance');

    if (adjustmentsTotalEl) {
        adjustmentsTotalEl.textContent = totalAdjustments >= 0 ?
            `+${totalAdjustments.toFixed(1)} Std` :
            `${totalAdjustments.toFixed(1)} Std`;
        adjustmentsTotalEl.className = `font-medium ${totalAdjustments >= 0 ? 'text-green-600' : 'text-red-600'}`;
    }

    if (finalBalanceEl) {
        finalBalanceEl.textContent = finalBalance >= 0 ?
            `+${finalBalance.toFixed(1)} Std` :
            `${finalBalance.toFixed(1)} Std`;
        finalBalanceEl.className = `font-medium text-lg ${finalBalance >= 0 ? 'text-green-600' : 'text-red-600'}`;
    }
}

// Anpassung genehmigen (für Employee Detail View)
function approveEmployeeAdjustment(adjustmentId, action) {
    const formData = new FormData();
    formData.append('action', action);

    fetch(`/api/overtime/adjustments/${adjustmentId}/approve`, {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const message = action === 'approve' ? 'Anpassung wurde genehmigt.' : 'Anpassung wurde abgelehnt.';
                showNotification(message, 'success');

                // Anpassungen neu laden
                setTimeout(() => {
                    loadEmployeeAdjustments();
                }, 500);
            } else {
                throw new Error(data.error || 'Fehler beim Verarbeiten der Anpassung');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Fehler: ' + error.message, 'error');
        });
}

// Hilfsfunktionen für Anzeigenamen
function getAdjustmentTypeDisplay(type) {
    const types = {
        'correction': 'Korrektur',
        'manual': 'Manuelle Anpassung',
        'bonus': 'Bonus/Ausgleich',
        'penalty': 'Abzug'
    };
    return types[type] || type;
}

function getStatusDisplay(status) {
    const statuses = {
        'pending': 'Ausstehend',
        'approved': 'Genehmigt',
        'rejected': 'Abgelehnt'
    };
    return statuses[status] || status;
}

function getStatusClass(status) {
    const classes = {
        'pending': 'bg-yellow-100 text-yellow-800',
        'approved': 'bg-green-100 text-green-800',
        'rejected': 'bg-red-100 text-red-800'
    };
    return classes[status] || 'bg-gray-100 text-gray-800';
}

// Fehler beim Laden anzeigen
function displayAdjustmentsError(message) {
    const container = document.getElementById('adjustmentsContainer');
    if (!container) return;

    container.innerHTML = `
    <div class="text-center py-8">
      <svg class="mx-auto h-12 w-12 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">Fehler</h3>
      <p class="mt-1 text-sm text-gray-500">${message}</p>
    </div>
  `;
}

// Globale Variable für User Role (für Berechtigungsprüfungen)
window.userRole = '{{.userRole}}';