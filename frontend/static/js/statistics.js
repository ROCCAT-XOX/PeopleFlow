document.addEventListener('DOMContentLoaded', function() {
    // Tab-Funktionalität
    const tabButtons = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    tabButtons.forEach(btn => {
        btn.addEventListener('click', function() {
            const tabId = this.getAttribute('data-tab');

            // Aktiven Button-Zustand ändern
            tabButtons.forEach(button => {
                button.classList.remove('bg-green-100', 'text-green-700', 'border-green-500');
                button.classList.add('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
            });

            this.classList.remove('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
            this.classList.add('bg-green-100', 'text-green-700', 'border-green-500');

            // Tab-Inhalte ein-/ausblenden
            tabContents.forEach(content => {
                content.classList.add('hidden');
            });

            document.getElementById(tabId + '-tab').classList.remove('hidden');
        });
    });

    // Filter-Funktionalität
    const dateRange = document.getElementById('date-range');
    const customDateRange = document.getElementById('custom-date-range');

    dateRange.addEventListener('change', function() {
        if (this.value === 'custom') {
            customDateRange.classList.remove('hidden');
        } else {
            customDateRange.classList.add('hidden');
        }
    });

    // Filter anwenden
    const applyFilterBtn = document.getElementById('apply-filter');
    const projectFilter = document.getElementById('project-filter');

    applyFilterBtn.addEventListener('click', function() {
        // Datumsfilterung
        const dateRangeValue = dateRange.value;
        let startDate = null;
        let endDate = null;

        const now = new Date();

        // Datumsbereich basierend auf Auswahl setzen
        if (dateRangeValue === 'custom') {
            const startDateInput = document.getElementById('start-date');
            const endDateInput = document.getElementById('end-date');

            if (startDateInput.value) {
                startDate = new Date(startDateInput.value);
            }

            if (endDateInput.value) {
                endDate = new Date(endDateInput.value);
            }
        } else if (dateRangeValue === 'this-month') {
            // Dieser Monat
            startDate = new Date(now.getFullYear(), now.getMonth(), 1);
            endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        } else if (dateRangeValue === 'last-month') {
            // Letzter Monat
            startDate = new Date(now.getFullYear(), now.getMonth() - 1, 1);
            endDate = new Date(now.getFullYear(), now.getMonth(), 0);
        } else if (dateRangeValue === 'this-quarter') {
            // Dieses Quartal
            const quarter = Math.floor(now.getMonth() / 3);
            startDate = new Date(now.getFullYear(), quarter * 3, 1);
            endDate = new Date(now.getFullYear(), quarter * 3 + 3, 0);
        } else if (dateRangeValue === 'last-quarter') {
            // Letztes Quartal
            const quarter = Math.floor(now.getMonth() / 3) - 1;
            const year = quarter < 0 ? now.getFullYear() - 1 : now.getFullYear();
            const adjustedQuarter = quarter < 0 ? 3 : quarter;
            startDate = new Date(year, adjustedQuarter * 3, 1);
            endDate = new Date(year, adjustedQuarter * 3 + 3, 0);
        } else if (dateRangeValue === 'this-year') {
            // Dieses Jahr
            startDate = new Date(now.getFullYear(), 0, 1);
            endDate = new Date(now.getFullYear(), 11, 31);
        }

        // Ausgewähltes Projekt
        const selectedProject = projectFilter.value;

        // Anzeige wird aktualisiert, um zu zeigen, dass die Filter angewendet wurden
        showNotification(
            'Filter angewendet',
            'Die Statistikdaten wurden entsprechend Ihrer Filterkriterien aktualisiert.',
            'success'
        );

        // Charts aktualisieren
        initializeCharts();
    });

    // Hilfsfunktion zum Anzeigen von Benachrichtigungen
    function showNotification(title, message, type = 'info', duration = 3000) {
        const container = document.getElementById('notification-container');
        if (!container) return;

        // Benachrichtigung erstellen
        const notification = document.createElement('div');
        notification.className = 'rounded-lg shadow-lg overflow-hidden transform transition-all duration-300 opacity-0 translate-x-full';

        // Farbe je nach Typ festlegen
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
            default: // info
                bgColor = 'bg-blue-50 border-l-4 border-blue-500';
                iconColor = 'text-blue-500';
                iconSvg = '<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>';
        }

        // Inhalt der Benachrichtigung
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
                <span class="sr-only">Schließen</span>
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        `;

        // Benachrichtigung zum Container hinzufügen
        container.appendChild(notification);

        // Event-Listener für Schließen-Button
        notification.querySelector('button').addEventListener('click', () => {
            notification.classList.add('opacity-0', 'translate-x-full');
            setTimeout(() => {
                container.removeChild(notification);
            }, 300);
        });

        // Animation einblenden
        setTimeout(() => {
            notification.classList.remove('opacity-0', 'translate-x-full');
        }, 10);

        // Automatisch ausblenden nach der angegebenen Zeit
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

    // Farbpalette
    const chartColors = [
        '#15803D', // Dunkelgrün
        '#22C55E', // Grün
        '#86EFAC', // Hellgrün
        '#C3E657', // Gelbgrün
        '#EFE176'  // Gelb
    ];

    // Charts initialisieren
    function initializeCharts() {
        // Arbeitszeit nach Wochentag Chart
        const weekdayHoursCtx = document.getElementById('weekdayHoursChart')?.getContext('2d');
        if (weekdayHoursCtx) {
            try {
                const existingChart = Chart.getChart('weekdayHoursChart');
                if (existingChart) {
                    existingChart.destroy();
                }
            } catch (e) {
                console.log("No existing chart found, creating new one");
            }

            new Chart(weekdayHoursCtx, {
                type: 'bar',
                data: {
                    labels: ['Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag', 'Sonntag'],
                    datasets: [{
                        label: 'Durchschnittliche Stunden',
                        data: [8.2, 8.5, 7.9, 8.3, 7.1, 1.4, 0.3],
                        backgroundColor: chartColors[1],
                        borderRadius: 6
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Stunden'
                            }
                        }
                    }
                }
            });
        }

        // Arbeitszeit nach Projekt Chart
        const projectHoursCtx = document.getElementById('projectHoursChart')?.getContext('2d');
        if (projectHoursCtx) {
            try {
                const existingChart = Chart.getChart('projectHoursChart');
                if (existingChart) {
                    existingChart.destroy();
                }
            } catch (e) {
                console.log("No existing chart found, creating new one");
            }

            new Chart(projectHoursCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Projekt A', 'Projekt B', 'Projekt C', 'Projekt D', 'Andere'],
                    datasets: [{
                        data: [35, 25, 15, 10, 15],
                        backgroundColor: chartColors,
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'right'
                        }
                    }
                }
            });
        }
    }

    // Initial Charts erstellen
    initializeCharts();
});