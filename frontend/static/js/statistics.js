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

    // Mitarbeiter-Dropdown Funktionalität
    const employeeDropdownToggle = document.getElementById('employee-dropdown-toggle');
    const employeeDropdown = document.getElementById('employee-dropdown');
    const selectAllEmployees = document.getElementById('select-all-employees');
    const employeeSelectionDisplay = document.getElementById('employee-selection-display');
    let selectedEmployees = []; // Array für ausgewählte Mitarbeiter-IDs

    employeeDropdownToggle.addEventListener('click', function() {
        employeeDropdown.classList.toggle('hidden');
    });

    // Klick außerhalb des Dropdowns schließt es
    document.addEventListener('click', function(event) {
        if (!employeeDropdownToggle.contains(event.target) && !employeeDropdown.contains(event.target)) {
            employeeDropdown.classList.add('hidden');
        }
    });

    // "Alle auswählen" Checkbox Funktionalität
    selectAllEmployees.addEventListener('change', function() {
        const checkboxes = document.querySelectorAll('.employee-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.checked = this.checked;
        });

        updateEmployeeSelection();
    });

    // Einzelne Mitarbeiter-Checkboxen
    document.querySelectorAll('.employee-checkbox').forEach(checkbox => {
        checkbox.addEventListener('change', function() {
            updateEmployeeSelection();

            // Überprüfen, ob alle Checkboxen ausgewählt sind
            const allCheckboxes = document.querySelectorAll('.employee-checkbox');
            const allChecked = Array.from(allCheckboxes).every(cb => cb.checked);
            selectAllEmployees.checked = allChecked;
        });
    });

    // Klick auf Labels oder Zeilen schaltet die Checkboxen um
    document.querySelectorAll('.employee-option').forEach(option => {
        option.addEventListener('click', function(e) {
            // Verhindern, dass das Klicken auf die Checkbox selbst doppelt registriert wird
            if (e.target.type !== 'checkbox') {
                const checkbox = this.querySelector('input[type="checkbox"]');
                checkbox.checked = !checkbox.checked;

                // Manuell das Change-Event auslösen
                const event = new Event('change');
                checkbox.dispatchEvent(event);
            }
        });
    });

    // Aktualisiert die Anzeige der ausgewählten Mitarbeiter
    function updateEmployeeSelection() {
        selectedEmployees = [];
        const checkboxes = document.querySelectorAll('.employee-checkbox:checked');

        checkboxes.forEach(checkbox => {
            selectedEmployees.push(checkbox.value);
        });

        if (selectedEmployees.length === 0) {
            employeeSelectionDisplay.textContent = 'Alle Mitarbeiter';
        } else if (selectedEmployees.length === 1) {
            const employeeName = document.querySelector(`label[for="employee-${selectedEmployees[0]}"]`).textContent.trim();
            employeeSelectionDisplay.textContent = employeeName;
        } else {
            employeeSelectionDisplay.textContent = `${selectedEmployees.length} Mitarbeiter ausgewählt`;
        }
    }

    // Filter anwenden
    const applyFilterBtn = document.getElementById('apply-filter');
    const projectFilter = document.getElementById('project-filter');

    applyFilterBtn.addEventListener('click', function() {
        // Datumsfilterung
        const dateRangeValue = dateRange.value;
        let startDate = null;
        let endDate = null;

        const now = new Date();
        const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

        // Datumsbereich basierend auf Auswahl setzen
        if (dateRangeValue === 'custom') {
            const startDateInput = document.getElementById('start-date');
            const endDateInput = document.getElementById('end-date');

            if (startDateInput.value) {
                startDate = new Date(startDateInput.value);
            }

            if (endDateInput.value) {
                endDate = new Date(endDateInput.value);
                // Setzen auf Ende des Tages für korrekten Vergleich
                endDate.setHours(23, 59, 59, 999);
            }
        } else if (dateRangeValue === 'this-month') {
            // Dieser Monat
            startDate = new Date(now.getFullYear(), now.getMonth(), 1);
            endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0);
            endDate.setHours(23, 59, 59, 999);
        } else if (dateRangeValue === 'last-month') {
            // Letzter Monat
            startDate = new Date(now.getFullYear(), now.getMonth() - 1, 1);
            endDate = new Date(now.getFullYear(), now.getMonth(), 0);
            endDate.setHours(23, 59, 59, 999);
        } else if (dateRangeValue === 'this-quarter') {
            // Dieses Quartal
            const quarter = Math.floor(now.getMonth() / 3);
            startDate = new Date(now.getFullYear(), quarter * 3, 1);
            endDate = new Date(now.getFullYear(), quarter * 3 + 3, 0);
            endDate.setHours(23, 59, 59, 999);
        } else if (dateRangeValue === 'last-quarter') {
            // Letztes Quartal
            const quarter = Math.floor(now.getMonth() / 3) - 1;
            const year = quarter < 0 ? now.getFullYear() - 1 : now.getFullYear();
            const adjustedQuarter = quarter < 0 ? 3 : quarter;
            startDate = new Date(year, adjustedQuarter * 3, 1);
            endDate = new Date(year, adjustedQuarter * 3 + 3, 0);
            endDate.setHours(23, 59, 59, 999);
        } else if (dateRangeValue === 'this-year') {
            // Dieses Jahr
            startDate = new Date(now.getFullYear(), 0, 1);
            endDate = new Date(now.getFullYear(), 11, 31);
            endDate.setHours(23, 59, 59, 999);
        }

        console.log('Filter - Date Range:', { startDate, endDate });

        // Ausgewählte Mitarbeiter
        console.log('Filter - Employees:', selectedEmployees);

        // Ausgewähltes Projekt
        const selectedProject = projectFilter.value;
        console.log('Filter - Project:', selectedProject);

        // AJAX-Anfrage zum Laden der gefilterten Daten
        // Im echten System würde hier ein API-Call erfolgen
        // Für dieses Beispiel simulieren wir die Antwort

        // Anzeige wird aktualisiert, um zu zeigen, dass die Filter angewendet wurden
        showNotification(
            'Filter angewendet',
            'Die Statistikdaten wurden entsprechend Ihrer Filterkriterien aktualisiert.',
            'success'
        );

        // In einer echten Anwendung würden die Charts basierend auf den API-Antworten aktualisiert
        updateChartsWithFilteredData();
    });

    // Simulierte Funktion zur Aktualisierung der Charts mit gefilterten Daten
    function updateChartsWithFilteredData() {
        // Diese Funktion würde in einer echten Anwendung die Charts aktualisieren
        // basierend auf den Antworten der API auf die Filteranfrage

        // Beispiel für Aktualisierung von Statistik-Karten
        const totalHoursCard = document.getElementById('total-hours-card');
        const productivityRateCard = document.getElementById('productivity-rate-card');
        const absenceDaysCard = document.getElementById('absence-days-card');

        // Simuliere neue Werte für die Statistik-Karten
        if (totalHoursCard) totalHoursCard.textContent = (Math.random() * 1000).toFixed(1) + ' Std';
        if (productivityRateCard) productivityRateCard.textContent = (75 + Math.random() * 20).toFixed(1) + '%';
        if (absenceDaysCard) absenceDaysCard.textContent = (Math.random() * 100).toFixed(0) + ' Tage';

        // In einer echten Anwendung würden hier die Chart.js-Objekte aktualisiert
        // und neue Daten zugewiesen
    }

    // Chartdaten exportieren
    window.exportChartData = function(chartId, format) {
        // Diese Funktion würde in einer echten Anwendung die Chartdaten exportieren
        const chart = Chart.getChart(chartId);
        if (!chart) {
            showNotification('Fehler', 'Das angegebene Chart wurde nicht gefunden.', 'error');
            return;
        }

        const chartName = chartId.replace('Chart', '');
        showNotification(
            'Export gestartet',
            `Die Daten für "${chartName}" werden als ${format.toUpperCase()} exportiert...`,
            'info'
        );

        // Simuliere eine Verzögerung für die Demonstration
        setTimeout(() => {
            showNotification(
                'Export erfolgreich',
                `Die Daten für "${chartName}" wurden erfolgreich als ${format.toUpperCase()} exportiert.`,
                'success'
            );
        }, 1500);
    };

    // Funktion zum Wechseln des Chart-Typs
    window.changeChartType = function(chartId, newType) {
        const chart = Chart.getChart(chartId);
        if (!chart) {
            showNotification('Fehler', 'Das angegebene Chart wurde nicht gefunden.', 'error');
            return;
        }

        // Speichere die aktuellen Daten und Optionen
        const data = chart.data;
        const options = chart.options;

        // Zerstöre das alte Chart
        chart.destroy();

        // Erstelle ein neues Chart mit dem gewünschten Typ
        const ctx = document.getElementById(chartId).getContext('2d');
        new Chart(ctx, {
            type: newType,
            data: data,
            options: options
        });

        showNotification(
            'Chart-Typ geändert',
            `Der Chart-Typ wurde zu ${newType} geändert.`,
            'success'
        );
    };

    // Farbpalette definieren
    const chartColors = [
        '#15803D', // Dunkelgrün
        '#22C55E', // Grün
        '#86EFAC', // Hellgrün
        '#C3E657', // Gelbgrün
        '#EFE176'  // Gelb
    ];

    // Initialisiere alle Charts - hier könnte in einer echten Anwendung
    // eine Funktion sein, die die Charts mit tatsächlichen Daten initialisiert

    // Arbeitszeit nach Wochentag Chart
    const weekdayHoursCtx = document.getElementById('weekdayHoursChart')?.getContext('2d');
    if (weekdayHoursCtx) {
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

    // Monatliche Abwesenheitstage Chart
    const absenceCtx = document.getElementById('absenceChart')?.getContext('2d');
    if (absenceCtx) {
        new Chart(absenceCtx, {
            type: 'line',
            data: {
                labels: ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez'],
                datasets: [
                    {
                        label: 'Urlaub',
                        data: [5, 7, 12, 10, 8, 15, 22, 18, 10, 8, 6, 12],
                        borderColor: chartColors[0],
                        backgroundColor: 'transparent',
                        tension: 0.4
                    },
                    {
                        label: 'Krankheit',
                        data: [8, 10, 6, 5, 4, 3, 2, 5, 7, 9, 12, 10],
                        borderColor: chartColors[3],
                        backgroundColor: 'transparent',
                        tension: 0.4
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Tage'
                        }
                    }
                }
            }
        });
    }

    // Produktivität nach Mitarbeiter Chart
    const employeeProductivityCtx = document.getElementById('employeeProductivityChart')?.getContext('2d');
    if (employeeProductivityCtx) {
        new Chart(employeeProductivityCtx, {
            type: 'bar',
            data: {
                labels: ['Max Mustermann', 'Anna Schmidt', 'Thomas Weber', 'Sarah Becker', 'Michael Schulz'],
                datasets: [{
                    label: 'Produktivität (%)',
                    data: [92, 87, 85, 83, 80],
                    backgroundColor: chartColors[0],
                    borderRadius: 6
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                indexAxis: 'y',
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    x: {
                        beginAtZero: true,
                        max: 100,
                        title: {
                            display: true,
                            text: 'Produktivität (%)'
                        }
                    }
                }
            }
        });
    }

    // Produktivität im Zeitverlauf Chart
    const productivityTimeCtx = document.getElementById('productivityTimeChart')?.getContext('2d');
    if (productivityTimeCtx) {
        new Chart(productivityTimeCtx, {
            type: 'line',
            data: {
                labels: ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez'],
                datasets: [{
                    label: 'Produktivitätsrate (%)',
                    data: [75, 78, 77, 80, 82, 85, 86, 85, 87, 88, 87, 89],
                    borderColor: chartColors[1],
                    backgroundColor: `${chartColors[1]}20`,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: false,
                        min: 70,
                        max: 100,
                        title: {
                            display: true,
                            text: 'Produktivität (%)'
                        }
                    }
                }
            }
        });
    }

    // Weitere Charts initialisieren...

    // Initialisiere alle anderen Charts entsprechend
    initializeRemainingCharts();

    function initializeRemainingCharts() {
        // Hier würden alle weiteren Charts initialisiert werden
        // Dies würde die Funktion sehr lang machen, daher ist dies nur ein Platzhalter

        // In einer echten Anwendung könnten die Chart-Daten von der API geladen werden
        // und die Charts dann dynamisch initialisiert werden
    }
});