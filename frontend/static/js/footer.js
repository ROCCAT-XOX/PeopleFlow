<!-- Notification JavaScript -->
// Initialen Zustand beim Laden der Seite setzen
document.addEventListener('DOMContentLoaded', function() {
    const isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
    if (isCollapsed) {
        document.body.classList.add('sidebar-collapsed');
    } else {
        document.body.classList.remove('sidebar-collapsed');
    }
});
// JavaScript für die Navigation
document.addEventListener('DOMContentLoaded', function() {
    // Mobile Navigation
    const mobileMenuButton = document.getElementById('mobile-sidebar-button');
    const closeButton = document.getElementById('close-sidebar-button');
    const mobileMenu = document.getElementById('mobile-menu');

    if (mobileMenuButton) {
        mobileMenuButton.addEventListener('click', function() {
            mobileMenu.style.display = 'block';
        });
    }

    if (closeButton) {
        closeButton.addEventListener('click', function() {
            mobileMenu.style.display = 'none';
        });
    }

    // Sidebar Toggle (einklappen/ausklappen)
    const sidebar = document.getElementById('sidebar');
    const mainContent = document.getElementById('main-content');
    const toggleButton = document.getElementById('toggle-sidebar');
    const expandButton = document.getElementById('expand-sidebar-button');
    const sidebarLogoFull = document.getElementById('sidebar-logo-full');
    const sidebarLogoIcon = document.getElementById('sidebar-logo-icon');
    const menuTexts = document.querySelectorAll('.menu-text');

    // Initialer Zustand der Sidebar
    if (sidebar) {
        sidebar.style.display = 'flex';
    }

    function collapseSidebar() {
        if (sidebar && mainContent) {
            sidebar.style.width = '5rem';
            mainContent.style.paddingLeft = '5rem';
            document.body.classList.add('sidebar-collapsed'); // Neue Zeile

            // Logos umschalten
            if (sidebarLogoFull) sidebarLogoFull.classList.add('hidden');
            if (sidebarLogoIcon) sidebarLogoIcon.classList.remove('hidden');

            // Text in Menüpunkten ausblenden
            menuTexts.forEach(text => {
                text.style.display = 'none';
            });

            // Toggle-Button-Icons umschalten
            document.getElementById('collapse-icon').classList.add('hidden');
            document.getElementById('expand-icon').classList.remove('hidden');

            localStorage.setItem('sidebarCollapsed', 'true');
        }
    }

// In der expandSidebar-Funktion hinzufügen:
    function expandSidebar() {
        if (sidebar && mainContent) {
            sidebar.style.width = '16rem';
            mainContent.style.paddingLeft = '16rem';
            document.body.classList.remove('sidebar-collapsed'); // Neue Zeile

            // Logos umschalten
            if (sidebarLogoIcon) sidebarLogoIcon.classList.add('hidden');
            if (sidebarLogoFull) sidebarLogoFull.classList.remove('hidden');

            // Text in Menüpunkten anzeigen
            menuTexts.forEach(text => {
                text.style.display = 'inline';
            });

            // Toggle-Button-Icons umschalten
            document.getElementById('expand-icon').classList.add('hidden');
            document.getElementById('collapse-icon').classList.remove('hidden');

            localStorage.setItem('sidebarCollapsed', 'false');
        }
    }

    // Speichern des Sidebar-Status im localStorage
    const savedSidebarState = localStorage.getItem('sidebarCollapsed');
    const isCollapsed = savedSidebarState === 'true';

    // Initialen Zustand setzen
    if (isCollapsed) {
        collapseSidebar();
    } else {
        expandSidebar();
    }

    // Event-Listener für Toggle-Buttons
    if (toggleButton) {
        toggleButton.addEventListener('click', function() {
            const currentWidth = sidebar.style.width;
            if (currentWidth === '16rem') {
                collapseSidebar();
            } else {
                expandSidebar();
            }
        });
    }

    if (expandButton) {
        expandButton.addEventListener('click', function() {
            expandSidebar();
        });
    }

    // Tab-Wechsel Funktionalität (falls vorhanden)
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    if (tabBtns.length > 0 && tabContents.length > 0) {
        tabBtns.forEach(btn => {
            btn.addEventListener('click', function() {
                const tab = this.getAttribute('data-tab');

                // Aktiven Button-Zustand ändern
                tabBtns.forEach(button => {
                    button.classList.remove('border-green-500', 'text-green-600');
                    button.classList.add('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
                });

                this.classList.remove('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
                this.classList.add('border-green-500', 'text-green-600');

                // Tab-Inhalte ein-/ausblenden
                tabContents.forEach(content => {
                    content.classList.add('hidden');
                });

                document.getElementById(tab + '-tab').classList.remove('hidden');
            });
        });
    }

    // Suchfunktion für Tabellen (falls vorhanden)
    const searchInput = document.getElementById('searchInput');
    const userItems = document.querySelectorAll('.user-item');

    if (searchInput && userItems.length > 0) {
        searchInput.addEventListener('input', function() {
            const searchTerm = this.value.toLowerCase();

            userItems.forEach(item => {
                const text = item.textContent.toLowerCase();
                if (text.includes(searchTerm)) {
                    item.style.display = '';
                } else {
                    item.style.display = 'none';
                }
            });
        });
    }

    // Modal-Funktionen (falls vorhanden)
    window.openModal = function(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.remove('hidden');
            document.body.classList.add('overflow-hidden');
        }
    }

    window.closeModal = function(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.add('hidden');
            document.body.classList.remove('overflow-hidden');
        }
    }

    // Benutzer bearbeiten Funktion (falls vorhanden)
    window.openEditUserModal = function(id, firstName, lastName, email, role, status) {
        // Formular-Aktion aktualisieren
        const form = document.getElementById('editUserForm');
        if (form) {
            form.action = '/users/edit/' + id;

            // Formularfelder vorausfüllen
            document.getElementById('edit-user-id').value = id;
            document.getElementById('edit-firstName').value = firstName;
            document.getElementById('edit-lastName').value = lastName;
            document.getElementById('edit-email').value = email;
            document.getElementById('edit-role').value = role;
            document.getElementById('edit-status').value = status;

            // Modal öffnen
            openModal('editUserModal');
        }
    }

    // Benutzer löschen Funktion (falls vorhanden)
    window.confirmDeleteUser = function(id, name) {
        const messageElem = document.getElementById('delete-user-message');
        if (messageElem) {
            // Lösch-Message personalisieren
            messageElem.textContent = `Sind Sie sicher, dass Sie den Benutzer "${name}" löschen möchten? Diese Aktion kann nicht rückgängig gemacht werden.`;
        }

        // Bestätigungs-Button konfigurieren
        const confirmBtn = document.getElementById('confirmDeleteBtn');
        if (confirmBtn) {
            confirmBtn.onclick = function() {
                deleteUser(id);
            };
        }

        // Modal öffnen
        openModal('deleteUserModal');
    }

    // Benutzer löschen AJAX-Aufruf (falls verwendet)
    window.deleteUser = function(id) {
        fetch('/users/delete/' + id, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            }
        })
            .then(response => response.json())
            .then(data => {
                closeModal('deleteUserModal');
                // URL zur Einstellungsseite aktualisieren
                window.location.href = '/settings?success=deleted';
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Ein Fehler ist aufgetreten. Bitte versuchen Sie es erneut.');
            });
    }

    // Event-Listener für Close-Buttons in Modals
    const closeModalButtons = document.querySelectorAll('[data-close-modal]');
    if (closeModalButtons.length > 0) {
        closeModalButtons.forEach(button => {
            button.addEventListener('click', function() {
                const modalId = this.getAttribute('data-close-modal');
                closeModal(modalId);
            });
        });
    }

    // Passwort-Validierung (falls verwendet)
    const passwordForm = document.querySelector('form[action="/users/change-password"]');
    if (passwordForm) {
        passwordForm.addEventListener('submit', function(e) {
            const newPassword = document.getElementById('newPassword').value;
            const confirmPassword = document.getElementById('confirmPassword').value;

            if (newPassword !== confirmPassword) {
                e.preventDefault();
                alert('Die Passwörter stimmen nicht überein.');
                return false;
            }

            if (newPassword.length < 6) {
                e.preventDefault();
                alert('Das Passwort muss mindestens 6 Zeichen lang sein.');
                return false;
            }
        });
    }
});
function showNotification(title, message, type = 'info', duration = 5000) {
    const container = document.getElementById('notification-container');

    // Create notification element
    const notification = document.createElement('div');
    notification.className = 'rounded-lg shadow-lg overflow-hidden transform transition-all duration-300 opacity-0 translate-x-full';

    // Set notification color based on type
    let bgColor, iconColor, iconSvg;
    switch(type) {
        case 'success':
            bgColor = 'bg-green-50 border-l-4 border-green-500';
            iconColor = 'text-green-500';
            iconSvg = '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>';
            break;
        case 'error':
            bgColor = 'bg-red-50 border-l-4 border-red-500';
            iconColor = 'text-red-500';
            iconSvg = '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>';
            break;
        case 'warning':
            bgColor = 'bg-yellow-50 border-l-4 border-yellow-500';
            iconColor = 'text-yellow-500';
            iconSvg = '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>';
            break;
        default: // info
            bgColor = 'bg-blue-50 border-l-4 border-blue-500';
            iconColor = 'text-blue-500';
            iconSvg = '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>';
    }

    // Set notification content
    notification.innerHTML = `
      <div class="${bgColor} p-4 flex">
        <div class="flex-shrink-0">
          <div class="${iconColor}">
            ${iconSvg}
          </div>
        </div>
        <div class="ml-3 w-0 flex-1">
          <p class="text-sm font-medium text-gray-900">${title}</p>
          <p class="mt-1 text-sm text-gray-500">${message}</p>
        </div>
        <div class="ml-4 flex-shrink-0 flex">
          <button class="inline-flex text-gray-400 hover:text-gray-500 focus:outline-none">
            <span class="sr-only">Close</span>
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    `;

    // Add notification to container
    container.appendChild(notification);

    // Add click event to close button
    notification.querySelector('button').addEventListener('click', () => {
        notification.classList.add('opacity-0', 'translate-x-full');
        setTimeout(() => {
            container.removeChild(notification);
        }, 300);
    });

    // Animate notification in
    setTimeout(() => {
        notification.classList.remove('opacity-0', 'translate-x-full');
    }, 10);

    // Auto-remove notification after duration
    setTimeout(() => {
        if (notification.parentNode === container) {
            notification.classList.add('opacity-0', 'translate-x-full');
            setTimeout(() => {
                if (notification.parentNode === container) {
                    container.removeChild(notification);
                }
            }, 300);
        }
    }, duration);
}
