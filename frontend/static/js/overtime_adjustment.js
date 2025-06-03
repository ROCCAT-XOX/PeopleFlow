// Globale Variable für aktuelle Anpassung
let currentAdjustmentId = null;

// Event-Listener beim Laden der Seite erweitern
document.addEventListener('DOMContentLoaded', function() {
    // Bestehende Event-Listener...

    // Neue Event-Listener für Anpassungen
    const addAdjustmentForm = document.getElementById('addAdjustmentForm');
    if (addAdjustmentForm) {
        addAdjustmentForm.addEventListener('submit', submitAdjustment);
    }

    // Ausstehende Anpassungen laden (nur für Admin/Manager)
    const userRole = '{{.userRole}}';
    if (userRole === 'admin' || userRole === 'manager') {
        loadPendingAdjustments();
    }
});

// Neue Anpassung hinzufügen
function addOvertimeAdjustment(employeeId) {
    document.getElementById('adjustmentEmployeeId').value = employeeId;

    // Formular zurücksetzen
    document.getElementById('addAdjustmentForm').reset();
    document.getElementById('adjustmentEmployeeId').value = employeeId;

    openModal('addAdjustmentModal');
}

// Anpassung einreichen
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

    // API-Aufruf
    fetch(`/api/overtime/employee/${employeeId}/adjustment`, {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                closeModal('addAdjustmentModal');
                showNotification('Überstunden-Anpassung wurde eingereicht und wartet auf Genehmigung.', 'success');

                // Seite neu laden oder Daten aktualisieren
                setTimeout(() => {
                    window.location.reload();
                }, 1500);
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

// Ausstehende Anpassungen laden
function loadPendingAdjustments() {
    fetch('/api/overtime/adjustments/pending')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                displayPendingAdjustments(data.data);
            }
        })
        .catch(error => {
            console.error('Error loading pending adjustments:', error);
        });
}

// Ausstehende Anpassungen anzeigen
function displayPendingAdjustments(adjustments) {
    const container = document.getElementById('pendingAdjustments');
    const noDataDiv = document.getElementById('noPendingAdjustments');
    const countSpan = document.getElementById('pendingCount');

    if (!container) return;

    if (adjustments.length === 0) {
        container.innerHTML = '';
        noDataDiv.style.display = 'block';
        countSpan.textContent = '0 ausstehend';
        return;
    }

    noDataDiv.style.display = 'none';
    countSpan.textContent = `${adjustments.length} ausstehend`;

    const html = adjustments.map(adjustment => {
        const hoursClass = adjustment.hours >= 0 ? 'text-green-600' : 'text-red-600';
        const hoursText = adjustment.hours >= 0 ? `+${adjustment.hours.toFixed(1)}` : adjustment.hours.toFixed(1);

        return `
      <div class="px-4 py-4 border-b border-gray-200 last:border-b-0">
        <div class="flex items-center justify-between">
          <div class="flex-1">
            <div class="flex items-center space-x-4">
              <div class="flex-shrink-0">
                <div class="h-10 w-10 rounded-full bg-yellow-100 flex items-center justify-center">
                  <svg class="h-6 w-6 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                </div>
              </div>
              <div class="flex-1">
                <div class="flex items-center space-x-2">
                  <h4 class="text-sm font-medium text-gray-900">Mitarbeiter ID: ${adjustment.employeeId}</h4>
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    ${getAdjustmentTypeDisplay(adjustment.type)}
                  </span>
                  <span class="text-sm font-medium ${hoursClass}">${hoursText} Std</span>
                </div>
                <p class="text-sm text-gray-600 mt-1">${adjustment.reason}</p>
                ${adjustment.description ? `<p class="text-sm text-gray-500 mt-1">${adjustment.description}</p>` : ''}
                <div class="text-xs text-gray-500 mt-2">
                  Eingereicht von ${adjustment.adjusterName} am ${new Date(adjustment.createdAt).toLocaleDateString('de-DE')}
                </div>
              </div>
            </div>
          </div>
          <div class="flex space-x-2">
            <button onclick="showAdjustmentApproval('${adjustment.id}', '${adjustment.employeeId}', '${adjustment.type}', ${adjustment.hours}, '${adjustment.reason}', '${adjustment.description || ''}', '${adjustment.adjusterName}')" 
                    class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
              Prüfen
            </button>
          </div>
        </div>
      </div>
    `;
    }).join('');

    container.innerHTML = html;
}

// Anpassungstyp-Display-Namen
function getAdjustmentTypeDisplay(type) {
    const types = {
        'correction': 'Korrektur',
        'manual': 'Manuelle Anpassung',
        'bonus': 'Bonus/Ausgleich',
        'penalty': 'Abzug'
    };
    return types[type] || type;
}

// Anpassungs-Genehmigung anzeigen
function showAdjustmentApproval(adjustmentId, employeeId, type, hours, reason, description, adjusterName) {
    currentAdjustmentId = adjustmentId;

    const hoursClass = hours >= 0 ? 'text-green-600' : 'text-red-600';
    const hoursText = hours >= 0 ? `+${hours.toFixed(1)}` : hours.toFixed(1);

    const detailsHtml = `
    <div class="bg-gray-50 rounded-lg p-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <span class="text-sm font-medium text-gray-500">Mitarbeiter-ID:</span>
          <p class="text-sm text-gray-900">${employeeId}</p>
        </div>
        <div>
          <span class="text-sm font-medium text-gray-500">Art:</span>
          <p class="text-sm text-gray-900">${getAdjustmentTypeDisplay(type)}</p>
        </div>
        <div>
          <span class="text-sm font-medium text-gray-500">Stunden:</span>
          <p class="text-sm font-medium ${hoursClass}">${hoursText} Std</p>
        </div>
        <div>
          <span class="text-sm font-medium text-gray-500">Eingereicht von:</span>
          <p class="text-sm text-gray-900">${adjusterName}</p>
        </div>
      </div>
      <div class="mt-4">
        <span class="text-sm font-medium text-gray-500">Begründung:</span>
        <p class="text-sm text-gray-900 mt-1">${reason}</p>
      </div>
      ${description ? `
        <div class="mt-4">
          <span class="text-sm font-medium text-gray-500">Beschreibung:</span>
          <p class="text-sm text-gray-900 mt-1">${description}</p>
        </div>
      ` : ''}
    </div>
  `;

    document.getElementById('approvalDetails').innerHTML = detailsHtml;
    openModal('approveAdjustmentModal');
}

// Anpassung genehmigen/ablehnen
function approveAdjustment(action) {
    if (!currentAdjustmentId) return;

    const formData = new FormData();
    formData.append('action', action);

    fetch(`/api/overtime/adjustments/${currentAdjustmentId}/approve`, {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                closeModal('approveAdjustmentModal');
                const message = action === 'approve' ? 'Anpassung wurde genehmigt.' : 'Anpassung wurde abgelehnt.';
                showNotification(message, 'success');

                // Ausstehende Anpassungen neu laden
                loadPendingAdjustments();

                // Optional: Haupttabelle aktualisieren
                setTimeout(() => {
                    window.location.reload();
                }, 1500);
            } else {
                throw new Error(data.error || 'Fehler beim Verarbeiten der Anpassung');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showNotification('Fehler: ' + error.message, 'error');
        });
}