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

// Function to save Timebutler API key
function saveTimebutlerApiKey() {
    const apiKey = document.getElementById('timebutler-api').value;

    if (!apiKey) {
        showNotification('Fehler', 'Bitte geben Sie einen API-Schlüssel ein.', 'error');
        return;
    }

    // Show loading state
    const button = event.currentTarget;
    const originalText = button.innerHTML;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Speichern...
    `;
    button.disabled = true;

    // Create form data
    const formData = new FormData();
    formData.append('timebutler-api', apiKey);

    // Save API key
    fetch('/api/integrations/timebutler/save', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            // Restore button state
            button.innerHTML = originalText;
            button.disabled = false;

            if (data.success) {
                // Show success notification
                showNotification(
                    'API-Schlüssel gespeichert',
                    'Der Timebutler API-Schlüssel wurde erfolgreich gespeichert.',
                    'success'
                );

                // Update status and enable sync buttons
                document.getElementById('timebutlerStatus').innerHTML = 'Verbunden';
                document.getElementById('timebutlerStatus').className = 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800';
                document.getElementById('removeTimebutlerBtn').disabled = false;
                document.getElementById('timebutlerSyncButtons').style.display = 'flex';

                // Check if Timebutler integration is connected
                loadIntegrationStatus();
            } else {
                // Show error notification
                showNotification(
                    'Fehler beim Speichern',
                    data.message || 'Der API-Schlüssel konnte nicht gespeichert werden.',
                    'error'
                );
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification(
                'Fehler beim Speichern',
                'Beim Speichern des API-Schlüssels ist ein Fehler aufgetreten.',
                'error'
            );
            console.error('Error:', error);
        });
}

// Function to synchronize users from Timebutler
function syncTimebutlerUsers() {
    // Show loading state
    const button = event.currentTarget;
    const originalText = button.innerHTML;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Synchronisiere...
    `;
    button.disabled = true;

    // Call API to sync users
    fetch('/api/integrations/timebutler/sync/users', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => response.json())
        .then(data => {
            // Restore button state
            button.innerHTML = originalText;
            button.disabled = false;

            if (data.success) {
                // Show success notification
                showNotification(
                    'Synchronisierung erfolgreich',
                    `Es wurden ${data.updatedCount} Mitarbeiter aktualisiert.`,
                    'success'
                );
            } else {
                // Show error notification
                showNotification(
                    'Fehler bei der Synchronisierung',
                    data.message || 'Die Mitarbeiterdaten konnten nicht synchronisiert werden.',
                    'error'
                );
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification(
                'Fehler bei der Synchronisierung',
                'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten.',
                'error'
            );
            console.error('Error:', error);
        });
}

// Function to synchronize absences from Timebutler
function syncTimebutlerAbsences() {
    // Show loading state
    const button = event.currentTarget;
    const originalText = button.innerHTML;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Synchronisiere...
    `;
    button.disabled = true;

    // Get current year
    const currentYear = new Date().getFullYear();

    // Call API to sync absences
    fetch(`/api/integrations/timebutler/sync/absences?year=${currentYear}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => response.json())
        .then(data => {
            // Restore button state
            button.innerHTML = originalText;
            button.disabled = false;

            if (data.success) {
                // Show success notification
                showNotification(
                    'Synchronisierung erfolgreich',
                    `Es wurden ${data.updatedCount} Mitarbeiter mit Abwesenheitsdaten aktualisiert.`,
                    'success'
                );
            } else {
                // Show error notification
                showNotification(
                    'Fehler bei der Synchronisierung',
                    data.message || 'Die Abwesenheitsdaten konnten nicht synchronisiert werden.',
                    'error'
                );
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification(
                'Fehler bei der Synchronisierung',
                'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten.',
                'error'
            );
            console.error('Error:', error);
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



// Function to synchronize holiday entitlements from Timebutler
function syncTimebutlerHolidayEntitlements() {
    // Show loading state
    const button = event.currentTarget;
    const originalText = button.innerHTML;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Synchronisiere...
    `;
    button.disabled = true;

    // Get current year
    const currentYear = new Date().getFullYear();

    // Call API to sync holiday entitlements
    fetch(`/api/integrations/timebutler/sync/holidayentitlements?year=${currentYear}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => response.json())
        .then(data => {
            // Restore button state
            button.innerHTML = originalText;
            button.disabled = false;

            if (data.success) {
                // Show success notification
                showNotification(
                    'Synchronisierung erfolgreich',
                    `Es wurden ${data.updatedCount} Mitarbeiter mit Urlaubsansprüchen aktualisiert.`,
                    'success'
                );
            } else {
                // Show error notification
                showNotification(
                    'Fehler bei der Synchronisierung',
                    data.message || 'Die Urlaubsansprüche konnten nicht synchronisiert werden.',
                    'error'
                );
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification(
                'Fehler bei der Synchronisierung',
                'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten.',
                'error'
            );
            console.error('Error:', error);
        });
}