// Overtime Adjustment JavaScript für overtime.html

document.addEventListener('DOMContentLoaded', function() {
    // Ausstehende Anpassungen laden wenn der User Admin oder Manager ist
    const userRole = getUserRole();
    if (userRole === 'admin' || userRole === 'manager') {
        loadPendingAdjustments();
    }

    // Event-Listener für Anpassungsformular
    const addAdjustmentForm = document.getElementById('addAdjustmentForm');
    if (addAdjustmentForm) {
        addAdjustmentForm.addEventListener('submit', submitAdjustment);
    }
});

// Ausstehende Anpassungen laden
function loadPendingAdjustments() {
    fetch('/api/overtime/adjustments/pending')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                displayPendingAdjustments(data.data || []);
                updatePendingCount(data.data ? data.data.length : 0);
            } else {
                console.error('Error loading pending adjustments:', data.error);
                displayPendingAdjustmentsError('Fehler beim Laden der ausstehenden Anpassungen');
            }
        })
        .catch(error => {
            console.error('Network error loading pending adjustments:', error);
            displayPendingAdjustmentsError('Netzwerkfehler beim Laden der ausstehenden Anpassungen');
        });
}

// Ausstehende Anpassungen anzeigen
function displayPendingAdjustments(adjustments) {
    const container = document.getElementById('pendingAdjustments');
    const noDataElement = document.getElementById('noPendingAdjustments');

    if (!container) return;

    if (!adjustments || adjustments.length === 0) {
        container.innerHTML = '';
        if (noDataElement) {
            noDataElement.classList.remove('hidden');
        }
        return;
    }

    if (noDataElement) {
        noDataElement.classList.add('hidden');
    }

    const html = adjustments.map(adjustment => {
        const hoursClass = adjustment.hours >= 0 ? 'text-green-600' : 'text-red-600';
        const hoursText = adjustment.hours >= 0 ? `+${adjustment.hours.toFixed(1)}` : adjustment.hours.toFixed(1);

        return `
            <div class="px-4 py-5 sm:px-6">
                <div class="flex justify-between items-start">
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center space-x-3 mb-2">
                            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                ${getAdjustmentTypeDisplay(adjustment.type)}
                            </span>
                            <span class="text-lg font-semibold ${hoursClass}">${hoursText} Std</span>
                        </div>
                        
                        <h4 class="text-sm font-medium text-gray-900 mb-1">${adjustment.reason}</h4>
                        ${adjustment.description ? `<p class="text-sm text-gray-600 mb-2">${adjustment.description}</p>` : ''}
                        
                        <div class="flex items-center text-xs text-gray-500 space-x-4">
                            <div>
                                <span class="font-medium">Mitarbeiter:</span> 
                                <span id="employee-${adjustment.employeeId}">Lädt...</span>
                            </div>
                            <div>
                                <span class="font-medium">Eingereicht von:</span> ${adjustment.adjusterName}
                            </div>
                            <div>
                                <span class="font-medium">Datum:</span> 
                                ${new Date(adjustment.createdAt).toLocaleDateString('de-DE', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        })}
                            </div>
                        </div>
                    </div>
                    
                    <div class="flex space-x-2 ml-4">
                        <button onclick="showApprovalModal('${adjustment.id}', '${adjustment.reason}', ${adjustment.hours}, '${adjustment.adjusterName}')" 
                                class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            Details & Genehmigung
                        </button>
                        <button onclick="approveAdjustmentQuick('${adjustment.id}', 'approve')" 
                                class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-green-700 bg-green-100 hover:bg-green-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500">
                            <svg class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                            </svg>
                            Genehmigen
                        </button>
                        <button onclick="approveAdjustmentQuick('${adjustment.id}', 'reject')" 
                                class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500">
                            <svg class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                            </svg>
                            Ablehnen
                        </button>
                    </div>
                </div>
            </div>
        `;
    }).join('');

    container.innerHTML = html;

    // Mitarbeiternamen laden
    adjustments.forEach(adjustment => {
        loadEmployeeName(adjustment.employeeId);
    });
}

// Mitarbeitername laden und anzeigen
function loadEmployeeName(employeeId) {
    fetch(`/api/employees/${employeeId}/name`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const element = document.getElementById(`employee-${employeeId}`);
                if (element) {
                    element.textContent = data.name;
                }
            }
        })
        .catch(error => {
            console.error('Error loading employee name:', error);
            const element = document.getElementById(`employee-${employeeId}`);
            if (element) {
                element.textContent = 'Unbekannt';
            }
        });
}

// Anzahl ausstehender Anpassungen aktualisieren
function updatePendingCount(count) {
    const countElement = document.getElementById('pendingCount');
    if (countElement) {
        countElement.textContent = `${count} ausstehend`;
        countElement.className = count > 0
            ? 'inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800'
            : 'inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800';
    }
}

// Fehler beim Laden anzeigen
function displayPendingAdjustmentsError(message) {
    const container = document.getElementById('pendingAdjustments');
    if (container) {
        container.innerHTML = `
            <div class="px-4 py-8 text-center text-red-500">
                <svg class="mx-auto h-12 w-12 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <h3 class="mt-2 text-sm font-medium text-gray-900">Fehler</h3>
                <p class="mt-1 text-sm text-gray-500">${message}</p>
            </div>
        `;
    }
}

// Schnelle Genehmigung/Ablehnung
function approveAdjustmentQuick(adjustmentId, action) {
    const confirmText = action === 'approve'
        ? 'Möchten Sie diese Anpassung wirklich genehmigen?'
        : 'Möchten Sie diese Anpassung wirklich ablehnen?';

    if (!confirm(confirmText)) {
        return;
    }

    processAdjustmentApproval(adjustmentId, action);
}

// Genehmigungsmodal anzeigen
function showApprovalModal(adjustmentId, reason, hours, adjusterName) {
    const modal = document.getElementById('approveAdjustmentModal');
    const detailsContainer = document.getElementById('approvalDetails');

    if (!modal || !detailsContainer) return;

    const hoursClass = hours >= 0 ? 'text-green-600' : 'text-red-600';
    const hoursText = hours >= 0 ? `+${hours.toFixed(1)}` : hours.toFixed(1);

    detailsContainer.innerHTML = `
        <div class="space-y-3">
            <div>
                <label class="block text-sm font-medium text-gray-700">Anpassung</label>
                <div class="mt-1 text-lg font-semibold ${hoursClass}">${hoursText} Stunden</div>
            </div>
            <div>
                <label class="block text-sm font-medium text-gray-700">Begründung</label>
                <div class="mt-1 text-sm text-gray-900">${reason}</div>
            </div>
            <div>
                <label class="block text-sm font-medium text-gray-700">Eingereicht von</label>
                <div class="mt-1 text-sm text-gray-900">${adjusterName}</div>
            </div>
        </div>
    `;

    // Aktuelle Adjustment-ID speichern
    modal.setAttribute('data-adjustment-id', adjustmentId);

    openModal('approveAdjustmentModal');
}

// Anpassung genehmigen/ablehnen (vom Modal aus)
function approveAdjustment(action) {
    const modal = document.getElementById('approveAdjustmentModal');
    const adjustmentId = modal.getAttribute('data-adjustment-id');

    if (!adjustmentId) {
        alert('Fehler: Keine Anpassungs-ID gefunden');
        return;
    }

    closeModal('approveAdjustmentModal');
    processAdjustmentApproval(adjustmentId, action);
}

// Genehmigungsprozess durchführen
function processAdjustmentApproval(adjustmentId, action) {
    const formData = new FormData();
    formData.append('action', action);

    fetch(`/api/overtime/adjustments/${adjustmentId}/approve`, {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const message = action === 'approve'
                    ? 'Anpassung wurde erfolgreich genehmigt.'
                    : 'Anpassung wurde erfolgreich abgelehnt.';

                showNotification(message, 'success');

                // Ausstehende Anpassungen neu laden
                setTimeout(() => {
                    loadPendingAdjustments();
                }, 500);
            } else {
                throw new Error(data.error || 'Fehler beim Verarbeiten der Anpassung');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Fehler beim Verarbeiten der Anpassung: ' + error.message, 'error');
        });
}

// Neue Anpassung hinzufügen
function addAdjustment(employeeId) {
    document.getElementById('adjustmentEmployeeId').value = employeeId;

    // Formular zurücksetzen
    document.getElementById('addAdjustmentForm').reset();
    document.getElementById('adjustmentEmployeeId').value = employeeId;

    openModal('addAdjustmentModal');
}

// Anpassungsformular absenden
function submitAdjustment(event) {
    event.preventDefault();

    const formData = new FormData(event.target);
    const employeeId = formData.get('employeeId');

    // Validierung
    const hours = parseFloat(formData.get('hours'));
    if (isNaN(hours)) {
        alert('Bitte geben Sie eine gültige Stundenanzahl ein.');
        return;
    }

    const reason = formData.get('reason').trim();
    if (!reason) {
        alert('Bitte geben Sie eine Begründung an.');
        return;
    }

    // Button deaktivieren
    const submitBtn = event.target.querySelector('button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.disabled = true;
    submitBtn.innerHTML = 'Wird eingereicht...';

    fetch(`/api/overtime/employee/${employeeId}/adjustment`, {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                closeModal('addAdjustmentModal');
                showNotification('Überstunden-Anpassung wurde eingereicht und wartet auf Genehmigung.', 'success');

                // Ausstehende Anpassungen neu laden
                setTimeout(() => {
                    loadPendingAdjustments();
                }, 500);
            } else {
                throw new Error(data.error || 'Fehler beim Einreichen der Anpassung');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Fehler beim Einreichen der Anpassung: ' + error.message, 'error');
        })
        .finally(() => {
            submitBtn.disabled = false;
            submitBtn.innerHTML = originalText;
        });
}

// Hilfsfunktionen
function getAdjustmentTypeDisplay(type) {
    const types = {
        'correction': 'Korrektur',
        'manual': 'Manuelle Anpassung',
        'bonus': 'Bonus/Ausgleich',
        'penalty': 'Abzug'
    };
    return types[type] || type;
}

function getUserRole() {
    // Versuche die Rolle aus verschiedenen Quellen zu ermitteln
    if (window.userRole) {
        return window.userRole;
    }

    // Fallback: Aus Template-Variable (falls verfügbar)
    const roleElement = document.querySelector('[data-user-role]');
    if (roleElement) {
        return roleElement.getAttribute('data-user-role');
    }

    return null;
}

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