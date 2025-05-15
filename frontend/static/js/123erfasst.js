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