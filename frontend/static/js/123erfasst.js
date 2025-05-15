// Datei: frontend/static/js/123erfasst.js

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

// Anmeldedaten speichern
function saveErfasst123Credentials() {
    const email = document.getElementById('erfasst123-email').value;
    const password = document.getElementById('erfasst123-password').value;

    // Überprüfen, ob es sich um die maskierten Werte handelt
    if ((email === '********' && password === '********') || (!email || !password)) {
        alert('Bitte geben Sie E-Mail und Passwort ein.');
        return;
    }

    const formData = new FormData();
    formData.append('erfasst123-email', email);
    formData.append('erfasst123-password', password);

    fetch('/api/integrations/123erfasst/save', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('123erfasst-Integration erfolgreich konfiguriert!');
                // Status aktualisieren
                updateErfasst123Status();
            } else {
                alert('Fehler: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Fehler beim Speichern der Anmeldedaten:', error);
            alert('Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.');
        });
}

// Mitarbeiterdaten synchronisieren
function syncErfasst123Employees() {
    fetch('/api/integrations/123erfasst/sync/employees', {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert(`Synchronisierung erfolgreich! ${data.updatedCount} Mitarbeiter wurden aktualisiert.`);
            } else {
                alert('Fehler: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Fehler bei der Synchronisierung:', error);
            alert('Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.');
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
                    alert('123erfasst-Integration erfolgreich entfernt!');
                    // Felder zurücksetzen
                    document.getElementById('erfasst123-email').value = '';
                    document.getElementById('erfasst123-password').value = '';
                    // Status aktualisieren
                    updateErfasst123Status();
                } else {
                    alert('Fehler: ' + data.message);
                }
            })
            .catch(error => {
                console.error('Fehler beim Entfernen der Integration:', error);
                alert('Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.');
            });
    }
}

// Function to synchronize 123erfasst projects
function syncErfasst123Projects() {
    // Show loading state
    const button = event.target;
    const originalText = button.innerHTML;
    button.disabled = true;
    button.innerHTML = `
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Synchronisiere...
    `;

    // Get current month date range
    const now = new Date();
    const startDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0];
    const endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().split('T')[0];

    // Make AJAX request
    fetch(`/api/integrations/123erfasst/sync/projects?startDate=${startDate}&endDate=${endDate}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => response.json())
        .then(data => {
            // Reset button
            button.disabled = false;
            button.innerHTML = originalText;

            // Show success or error message
            if (data.success) {
                showNotification('success', data.message);
            } else {
                showNotification('error', data.message);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            button.disabled = false;
            button.innerHTML = originalText;
            showNotification('error', 'Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.');
        });
}

// Helper function to show notifications (if not already defined)
function showNotification(type, message) {
    // Create notification element if it doesn't exist
    let notification = document.getElementById('notification');
    if (!notification) {
        notification = document.createElement('div');
        notification.id = 'notification';
        notification.className = 'fixed bottom-4 right-4 px-4 py-2 rounded-md shadow-lg transform transition-all duration-300 opacity-0 translate-y-2';
        document.body.appendChild(notification);
    }

    // Set notification type
    if (type === 'success') {
        notification.className = 'fixed bottom-4 right-4 px-4 py-2 bg-green-50 text-green-800 border border-green-200 rounded-md shadow-lg transform transition-all duration-300 opacity-0 translate-y-2';
    } else {
        notification.className = 'fixed bottom-4 right-4 px-4 py-2 bg-red-50 text-red-800 border border-red-200 rounded-md shadow-lg transform transition-all duration-300 opacity-0 translate-y-2';
    }

    // Set message
    notification.textContent = message;

    // Show notification
    setTimeout(() => {
        notification.style.opacity = '1';
        notification.style.transform = 'translateY(0)';
    }, 10);

    // Hide notification after 3 seconds
    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transform = 'translateY(2px)';
    }, 3000);
}