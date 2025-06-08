// Status-Aktualisierung
function updateErfasst123Status() {
    fetch('/api/integrations/status')
        .then(response => response.json())
        .then(data => {
            const statusElem = document.getElementById('erfasst123Status');
            const syncButtons = document.getElementById('erfasst123SyncButtons');
            const removeBtn = document.getElementById('removeErfasst123Btn');
            const emailInput = document.getElementById('erfasst123-email');
            const passwordInput = document.getElementById('erfasst123-password');
            const configForm = document.getElementById('erfasst123ConfigForm');
            const syncSettings = document.getElementById('erfasst123SyncSettings');

            if (data['123erfasst']) {
                const erfasst123 = data['123erfasst'];

                if (erfasst123.connected) {
                    // Status auf "Verbunden" setzen
                    statusElem.textContent = 'Verbunden';
                    statusElem.classList.remove('bg-gray-100', 'text-gray-800', 'bg-red-100', 'text-red-800');
                    statusElem.classList.add('bg-green-100', 'text-green-800');

                    // Synchronisierungsbuttons anzeigen
                    if (syncButtons) {
                        syncButtons.style.display = 'flex';
                    }

                    // Konfigurationsformular ausblenden
                    if (configForm) {
                        configForm.style.display = 'none';
                    }

                    // Synchronisierungseinstellungen anzeigen
                    if (syncSettings) {
                        syncSettings.style.display = 'block';
                        loadErfasst123SyncSettings();
                    }

                    // Remove-Button aktivieren
                    if (removeBtn) {
                        removeBtn.disabled = false;
                    }

                    // Anmeldedaten mit Sternchen maskieren
                    if (emailInput && erfasst123.hasApiKey) {
                        emailInput.value = '********';
                        emailInput.setAttribute('placeholder', 'E-Mail (gespeichert)');
                    }

                    if (passwordInput && erfasst123.hasApiKey) {
                        passwordInput.value = '********';
                        passwordInput.setAttribute('placeholder', 'Passwort (gespeichert)');
                    }
                } else if (erfasst123.hasApiKey) {
                    // Status auf "Verbindungsfehler" setzen
                    statusElem.textContent = 'Verbindungsfehler';
                    statusElem.classList.remove('bg-gray-100', 'text-gray-800', 'bg-green-100', 'text-green-800');
                    statusElem.classList.add('bg-red-100', 'text-red-800');

                    // Synchronisierungsbuttons verstecken
                    if (syncButtons) {
                        syncButtons.style.display = 'none';
                    }

                    // Konfigurationsformular anzeigen
                    if (configForm) {
                        configForm.style.display = 'block';
                    }

                    // Synchronisierungseinstellungen ausblenden
                    if (syncSettings) {
                        syncSettings.style.display = 'none';
                    }

                    // Remove-Button aktivieren
                    if (removeBtn) {
                        removeBtn.disabled = false;
                    }

                    // Anmeldedaten mit Sternchen maskieren, da sie noch gespeichert sind
                    if (emailInput) {
                        emailInput.value = '********';
                        emailInput.setAttribute('placeholder', 'E-Mail (gespeichert)');
                    }

                    if (passwordInput) {
                        passwordInput.value = '********';
                        passwordInput.setAttribute('placeholder', 'Passwort (gespeichert)');
                    }
                } else {
                    // Status auf "Nicht verbunden" setzen
                    statusElem.textContent = 'Nicht verbunden';
                    statusElem.classList.remove('bg-green-100', 'text-green-800', 'bg-red-100', 'text-red-800');
                    statusElem.classList.add('bg-gray-100', 'text-gray-800');

                    // Synchronisierungsbuttons verstecken
                    if (syncButtons) {
                        syncButtons.style.display = 'none';
                    }

                    // Konfigurationsformular anzeigen
                    if (configForm) {
                        configForm.style.display = 'block';
                    }

                    // Synchronisierungseinstellungen ausblenden
                    if (syncSettings) {
                        syncSettings.style.display = 'none';
                    }

                    // Remove-Button deaktivieren
                    if (removeBtn) {
                        removeBtn.disabled = true;
                    }

                    // Leere Eingabefelder
                    if (emailInput) {
                        emailInput.value = '';
                        emailInput.setAttribute('placeholder', 'E-Mail eingeben');
                    }

                    if (passwordInput) {
                        passwordInput.value = '';
                        passwordInput.setAttribute('placeholder', 'Passwort eingeben');
                    }
                }
            }
        })
        .catch(error => {
            console.error('Fehler beim Abrufen des Integrationsstatus:', error);
        });
}

// Beim Laden der Seite Status aktualisieren
document.addEventListener('DOMContentLoaded', function() {
    updateErfasst123Status();

    // Set default start date to beginning of current year
    const startDateInput = document.getElementById('erfasst123-sync-start-date');
    if (startDateInput) {
        const yearStart = new Date(new Date().getFullYear(), 0, 1);
        startDateInput.value = yearStart.toISOString().split('T')[0];
    }

    // Event-Listener für Remove-Button
    const removeBtn = document.getElementById('removeErfasst123Btn');
    if (removeBtn) {
        removeBtn.addEventListener('click', removeErfasst123Integration);
    }

    // Event-Listener für Fokus auf Input-Feldern
    const emailInput = document.getElementById('erfasst123-email');
    const passwordInput = document.getElementById('erfasst123-password');

    // Wenn auf maskierte Felder geklickt wird, diese leeren für neue Eingabe
    if (emailInput) {
        emailInput.addEventListener('focus', function() {
            if (this.value === '********') {
                this.value = '';
            }
        });
    }

    if (passwordInput) {
        passwordInput.addEventListener('focus', function() {
            if (this.value === '********') {
                this.value = '';
            }
        });
    }
});

// Function to save 123erfasst credentials
function saveErfasst123Credentials() {
    const email = document.getElementById('erfasst123-email').value;
    const password = document.getElementById('erfasst123-password').value;
    const syncStartDate = document.getElementById('erfasst123-sync-start-date').value;

    if (!email || !password) {
        showNotification('error', 'Bitte geben Sie E-Mail und Passwort ein.');
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
        Integration einrichten...
    `;
    button.disabled = true;

    // Create form data
    const formData = new FormData();
    formData.append('erfasst123-email', email);
    formData.append('erfasst123-password', password);
    if (syncStartDate) {
        formData.append('erfasst123-sync-start-date', syncStartDate);
    }

    // Save credentials
    fetch('/api/integrations/123erfasst/save', {
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
                showNotification('success', 'Die 123erfasst Integration wurde erfolgreich eingerichtet. Die erste Synchronisierung wurde gestartet.');

                // Update status and enable sync buttons
                updateErfasst123Status();

                // Trigger initial full sync
                setTimeout(() => {
                    triggerErfasst123FullSync();
                }, 1000);
            } else {
                // Show error notification
                showNotification('error', data.message || 'Die Anmeldedaten konnten nicht gespeichert werden.');
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification('error', 'Beim Speichern der Anmeldedaten ist ein Fehler aufgetreten.');
            console.error('Error:', error);
        });
}

// Integration entfernen
function removeErfasst123Integration() {
    if (confirm('Möchten Sie die 123erfasst-Integration wirklich entfernen?')) {
        fetch('/api/integrations/123erfasst/remove', {
            method: 'POST'
        })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showNotification('success', '123erfasst-Integration erfolgreich entfernt!');
                    // Felder zurücksetzen
                    document.getElementById('erfasst123-email').value = '';
                    document.getElementById('erfasst123-password').value = '';
                    // Status aktualisieren
                    updateErfasst123Status();
                } else {
                    showNotification('error', data.message || 'Fehler beim Entfernen der Integration');
                }
            })
            .catch(error => {
                console.error('Fehler beim Entfernen der Integration:', error);
                showNotification('error', 'Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.');
            });
    }
}

// Function to synchronize 123erfasst projects
function syncErfasst123Projects() {
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

    // Get date range for the current month
    const now = new Date();
    const startDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0];
    const endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().split('T')[0];

    // Call API to sync projects
    fetch(`/api/integrations/123erfasst/sync/projects?startDate=${startDate}&endDate=${endDate}`, {
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
                showNotification('success', `Es wurden ${data.updatedCount} Mitarbeiter mit Projektdaten aktualisiert.`);

                // Refresh the last sync time
                loadErfasst123SyncSettings();
            } else {
                // Show error notification
                showNotification('error', data.message || 'Die Projektdaten konnten nicht synchronisiert werden.');
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification('error', 'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten.');
            console.error('Error:', error);
        });
}

// Function to synchronize employees from 123erfasst
function syncErfasst123Employees() {
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

    // Call API to sync employees
    fetch('/api/integrations/123erfasst/sync/employees', {
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
                showNotification('success', `Es wurden ${data.updatedCount} Mitarbeiter aktualisiert.`);

                // Refresh the last sync time
                loadErfasst123SyncSettings();
            } else {
                // Show error notification
                showNotification('error', data.message || 'Die Mitarbeiterdaten konnten nicht synchronisiert werden.');
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification('error', 'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten.');
            console.error('Error:', error);
        });
}

// Function to synchronize time entries from 123erfasst
function syncErfasst123TimeEntries() {
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

    // Get date range for the current month
    const now = new Date();
    const startDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0];
    const endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().split('T')[0];

    // Call API to sync time entries
    fetch(`/api/integrations/123erfasst/sync/times?startDate=${startDate}&endDate=${endDate}`, {
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
                // Show success notification with the number of updated records
                showNotification('success', `Es wurden ${data.updatedCount} Mitarbeiter mit Zeiterfassungsdaten aktualisiert.`);

                // Refresh the last sync time
                loadErfasst123SyncSettings();
            } else {
                // Show error notification
                showNotification('error', data.message || 'Die Zeiterfassungsdaten konnten nicht synchronisiert werden.');
            }
        })
        .catch(error => {
            // Restore button state and show error
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification('error', 'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten.');
            console.error('Error:', error);
        });
}

// Load sync settings for 123erfasst
function loadErfasst123SyncSettings() {
    fetch('/api/integrations/123erfasst/sync-status')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const settings = data.data;

                // Update auto-sync checkbox
                const autoSyncCheckbox = document.getElementById('erfasst123-auto-sync');
                if (autoSyncCheckbox) {
                    autoSyncCheckbox.checked = settings.autoSync;
                }

                // Update start date
                const startDateInput = document.getElementById('erfasst123-start-date');
                if (startDateInput && settings.syncStartDate) {
                    startDateInput.value = settings.syncStartDate;
                }

                // Update last sync display
                const lastSyncElem = document.getElementById('erfasst123-last-sync');
                if (lastSyncElem) {
                    lastSyncElem.textContent = settings.lastSync;
                }
            }
        })
        .catch(error => {
            console.error('Error loading sync settings:', error);
        });
}

// Update 123erfasst sync settings
function updateErfasst123SyncSettings() {
    const autoSync = document.getElementById('erfasst123-auto-sync').checked;
    const startDate = document.getElementById('erfasst123-start-date').value;

    // Update auto-sync setting
    const formData = new FormData();
    formData.append('enabled', autoSync.toString());

    fetch('/api/integrations/123erfasst/set-auto-sync', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (!data.success) {
                showNotification('error', data.message || 'Fehler beim Aktualisieren der Auto-Sync-Einstellung');
            }
        })
        .catch(error => {
            console.error('Error updating auto-sync:', error);
            showNotification('error', 'Fehler beim Aktualisieren der Auto-Sync-Einstellung');
        });

    // Update start date if provided
    if (startDate) {
        const startDateFormData = new FormData();
        startDateFormData.append('startDate', startDate);

        fetch('/api/integrations/123erfasst/set-sync-start-date', {
            method: 'POST',
            body: startDateFormData
        })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showNotification('success', 'Synchronisierungseinstellungen wurden aktualisiert');
                } else {
                    showNotification('error', data.message || 'Fehler beim Aktualisieren des Startdatums');
                }
            })
            .catch(error => {
                console.error('Error updating start date:', error);
                showNotification('error', 'Fehler beim Aktualisieren des Startdatums');
            });
    } else {
        showNotification('success', 'Auto-Sync-Einstellung wurde aktualisiert');
    }
}

// Trigger a full sync for 123erfasst
function triggerErfasst123FullSync() {
    // Show notification
    showNotification('info', 'Vollständige Synchronisierung wurde gestartet. Dies kann einige Minuten dauern...');

    // Call API to trigger full sync
    fetch('/api/integrations/123erfasst/full-sync', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // Show success notification
                showNotification('success', data.message || 'Synchronisierung erfolgreich abgeschlossen');

                // Refresh last sync time
                loadErfasst123SyncSettings();
            } else {
                // Show error notification
                showNotification('error', data.message || 'Fehler bei der Synchronisierung');
            }
        })
        .catch(error => {
            console.error('Error during full sync:', error);
            showNotification('error', 'Bei der Synchronisierung ist ein unerwarteter Fehler aufgetreten');
        });
}

// Helper function to show notifications
function showNotification(type, message) {
    const notificationContainer = document.getElementById('notification-container');
    if (!notificationContainer) {
        console.warn('Notification container not found');
        return;
    }

    // Create notification element
    const notification = document.createElement('div');
    notification.classList.add('mb-2', 'p-4', 'rounded-md', 'shadow-md', 'transform', 'transition-all', 'duration-300');

    // Set notification style based on type
    switch(type) {
        case 'success':
            notification.classList.add('bg-green-50', 'text-green-800', 'border', 'border-green-200');
            break;
        case 'error':
            notification.classList.add('bg-red-50', 'text-red-800', 'border', 'border-red-200');
            break;
        case 'info':
            notification.classList.add('bg-blue-50', 'text-blue-800', 'border', 'border-blue-200');
            break;
        default:
            notification.classList.add('bg-gray-50', 'text-gray-800', 'border', 'border-gray-200');
    }

    // Set notification content
    notification.textContent = message;

    // Add notification to container
    notificationContainer.appendChild(notification);

    // Animate in
    setTimeout(() => {
        notification.classList.add('opacity-100');
    }, 10);

    // Auto-remove after 5 seconds
    setTimeout(() => {
        notification.classList.add('opacity-0', 'translate-y-2');
        setTimeout(() => {
            notification.remove();
        }, 300);
    }, 5000);
}

function cleanup123ErfasstDuplicates() {
    if (!confirm('Möchten Sie wirklich alle doppelten Zeiteinträge bereinigen? Dies kann einige Zeit dauern.')) {
        return;
    }

    const button = event.currentTarget;
    const originalText = button.innerHTML;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Bereinige Duplikate...
    `;
    button.disabled = true;

    fetch('/api/integrations/123erfasst/cleanup-duplicates', {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            button.innerHTML = originalText;
            button.disabled = false;

            if (data.success) {
                showNotification('success', data.message);
            } else {
                showNotification('error', data.message || 'Fehler beim Bereinigen der Duplikate');
            }
        })
        .catch(error => {
            button.innerHTML = originalText;
            button.disabled = false;
            showNotification('error', 'Ein Fehler ist aufgetreten: ' + error);
            console.error('Error:', error);
        });
}