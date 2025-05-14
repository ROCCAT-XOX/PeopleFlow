// Timebutler Integration
document.addEventListener('DOMContentLoaded', function() {
    // Status der Integrationen abrufen
    fetchIntegrationStatus();

    // Debug: Prüfen, ob das Element existiert
    const syncButtons = document.getElementById('timebutlerSyncButtons');
    console.log("Sync Buttons Element:", syncButtons);
});

function fetchIntegrationStatus() {
    fetch('/api/integrations/status')
        .then(response => response.json())
        .then(data => {
            console.log("Integration Status:", data); // Debug-Log
            updateTimebutlerStatus(data.timebutler.connected, data.timebutler.hasApiKey);
        })
        .catch(error => {
            console.error('Error fetching integration status:', error);
            updateTimebutlerStatus(false, false);
        });
}

function updateTimebutlerStatus(connected, hasApiKey) {
    console.log("Updating Status - Connected:", connected, "Has API Key:", hasApiKey); // Debug-Log

    const statusElement = document.getElementById('timebutlerStatus');
    const removeButton = document.getElementById('removeTimebutlerBtn');
    const apiKeyInput = document.getElementById('timebutler-api');
    const syncButtons = document.getElementById('timebutlerSyncButtons');

    console.log("Elements found - Status:", statusElement, "Sync Buttons:", syncButtons); // Debug-Log

    if (connected) {
        if (statusElement) {
            statusElement.className = 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800';
            statusElement.innerHTML = 'Verbunden';
        }

        if (removeButton) {
            removeButton.disabled = false;
        }

        // Synchronisierungsbuttons anzeigen
        if (syncButtons) {
            console.log("Showing sync buttons");
            syncButtons.style.display = 'flex';
        } else {
            console.error("Sync buttons element not found!");
        }

        // Wenn ein API-Schlüssel vorhanden ist, Sternchen anzeigen
        if (hasApiKey && apiKeyInput) {
            apiKeyInput.value = '••••••••••••••••••••••••••••••••';
            apiKeyInput.placeholder = 'API-Schlüssel ist gespeichert';
        }
    } else {
        if (statusElement) {
            statusElement.className = 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800';
            statusElement.innerHTML = 'Nicht verbunden';
        }

        if (removeButton) {
            removeButton.disabled = true;
        }

        // Synchronisierungsbuttons ausblenden
        if (syncButtons) {
            console.log("Hiding sync buttons");
            syncButtons.style.display = 'none';
        }

        // Input-Feld leeren
        if (apiKeyInput) {
            apiKeyInput.value = '';
            apiKeyInput.placeholder = 'Timebutler API-Schlüssel eingeben';
        }
    }
}

function saveTimebutlerApiKey() {
    const apiKeyInput = document.getElementById('timebutler-api');
    if (!apiKeyInput) {
        console.error("API Key input element not found!");
        return;
    }

    const apiKey = apiKeyInput.value;

    // Wenn der Input-Feld Sternchen enthält und der Schlüssel bereits gespeichert ist, nichts tun
    if (apiKey === '••••••••••••••••••••••••••••••••') {
        alert('API-Schlüssel ist bereits gespeichert');
        return;
    }

    if (!apiKey) {
        alert('Bitte geben Sie einen API-Schlüssel ein');
        return;
    }

    const form = document.getElementById('timebutlerForm');
    if (!form) {
        console.error("Timebutler form not found!");
        return;
    }

    const formData = new FormData(form);

    fetch('/api/integrations/timebutler/save', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Timebutler-Integration erfolgreich konfiguriert');
                fetchIntegrationStatus();
            } else {
                alert('Fehler: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Fehler bei der Verbindung mit dem Server');
        });
}

function syncTimebutlerUsers() {
    showSyncSpinner('Synchronisiere Benutzerdaten...');

    fetch('/api/integrations/timebutler/sync/users', {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            hideSyncSpinner();
            if (data.success) {
                alert(data.message);
            } else {
                alert('Fehler: ' + data.message);
            }
        })
        .catch(error => {
            hideSyncSpinner();
            console.error('Error:', error);
            alert('Fehler bei der Synchronisierung');
        });
}

function syncTimebutlerAbsences() {
    const year = new Date().getFullYear();
    showSyncSpinner('Synchronisiere Abwesenheiten...');

    fetch(`/api/integrations/timebutler/sync/absences?year=${year}`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            hideSyncSpinner();
            if (data.success) {
                alert(data.message);
            } else {
                alert('Fehler: ' + data.message);
            }
        })
        .catch(error => {
            hideSyncSpinner();
            console.error('Error:', error);
            alert('Fehler bei der Synchronisierung');
        });
}

// Hilfsfunktionen für Spinner
function showSyncSpinner(message) {
    // Einen Spinner entfernen, falls einer existiert
    hideSyncSpinner();

    // Füge einen Spinner zum Body hinzu
    const spinner = document.createElement('div');
    spinner.id = 'syncSpinner';
    spinner.className = 'fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50';
    spinner.innerHTML = `
        <div class="bg-white p-4 rounded-lg shadow-lg flex flex-col items-center">
            <svg class="animate-spin h-8 w-8 text-green-600 mb-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span class="text-gray-700">${message}</span>
        </div>
    `;
    document.body.appendChild(spinner);
}

function hideSyncSpinner() {
    const spinner = document.getElementById('syncSpinner');
    if (spinner) {
        spinner.remove();
    }
}