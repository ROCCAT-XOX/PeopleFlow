// frontend/static/js/enhanced-statistics.js

document.addEventListener('DOMContentLoaded', function() {
    // Tab functionality
    const tabButtons = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    tabButtons.forEach(btn => {
        btn.addEventListener('click', function() {
            const tabId = this.getAttribute('data-tab');

            // Change active button state
            tabButtons.forEach(button => {
                button.classList.remove('bg-green-100', 'text-green-700', 'border-green-500');
                button.classList.add('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
            });

            this.classList.remove('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
            this.classList.add('bg-green-100', 'text-green-700', 'border-green-500');

            // Show/hide tab contents
            tabContents.forEach(content => {
                content.classList.add('hidden');
            });

            document.getElementById(tabId + '-tab').classList.remove('hidden');
        });
    });

    // Filter functionality
    const dateRange = document.getElementById('date-range');
    const customDateRange = document.getElementById('custom-date-range');

    dateRange.addEventListener('change', function() {
        if (this.value === 'custom') {
            customDateRange.classList.remove('hidden');
        } else {
            customDateRange.classList.add('hidden');
        }
    });

    // Apply filter button
    const applyFilterBtn = document.getElementById('apply-filter');
    const projectFilter = document.getElementById('project-filter');
    const employeeFilter = document.getElementById('employee-filter');

    applyFilterBtn.addEventListener('click', function() {
        // Get date range values
        const dateRangeValue = dateRange.value;
        let startDate = null;
        let endDate = null;

        const now = new Date();

        // Set date range based on selection
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
            // This month
            startDate = new Date(now.getFullYear(), now.getMonth(), 1);
            endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        } else if (dateRangeValue === 'last-month') {
            // Last month
            startDate = new Date(now.getFullYear(), now.getMonth() - 1, 1);
            endDate = new Date(now.getFullYear(), now.getMonth(), 0);
        } else if (dateRangeValue === 'this-quarter') {
            // This quarter
            const quarter = Math.floor(now.getMonth() / 3);
            startDate = new Date(now.getFullYear(), quarter * 3, 1);
            endDate = new Date(now.getFullYear(), quarter * 3 + 3, 0);
        } else if (dateRangeValue === 'last-quarter') {
            // Last quarter
            const quarter = Math.floor(now.getMonth() / 3) - 1;
            const year = quarter < 0 ? now.getFullYear() - 1 : now.getFullYear();
            const adjustedQuarter = quarter < 0 ? 3 : quarter;
            startDate = new Date(year, adjustedQuarter * 3, 1);
            endDate = new Date(year, adjustedQuarter * 3 + 3, 0);
        } else if (dateRangeValue === 'this-year') {
            // This year
            startDate = new Date(now.getFullYear(), 0, 1);
            endDate = new Date(now.getFullYear(), 11, 31);
        }

        // Get selected project and employee
        const selectedProject = projectFilter.value;
        const selectedEmployee = employeeFilter.value;

        // Prepare filter data
        const filterData = {
            startDate: startDate ? startDate.toISOString() : null,
            endDate: endDate ? endDate.toISOString() : null,
            projectId: selectedProject,
            employeeIds: selectedEmployee ? [selectedEmployee] : [],
            dateRangeKey: dateRangeValue
        };

        // Show loading state
        toggleLoadingState(true);

        // Fetch data with the filter
        fetchFilteredStatistics(filterData);
    });

    // Toggle loading state
    function toggleLoadingState(isLoading) {
        const applyFilterBtn = document.getElementById('apply-filter');

        if (isLoading) {
            applyFilterBtn.disabled = true;
            applyFilterBtn.innerHTML = `
                <svg class="animate-spin -ml-1 mr-2 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Lädt...
            `;
        } else {
            applyFilterBtn.disabled = false;
            applyFilterBtn.innerHTML = `
                <svg class="mr-2 -ml-1 h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fill-rule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z" clip-rule="evenodd" />
                </svg>
                Filter anwenden
            `;
        }
    }

    // Fetch filtered statistics data
    function fetchFilteredStatistics(filterData) {
        fetch('/api/statistics/extended', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(filterData)
        })
            .then(response => response.json())
            .then(data => {
                // Hide loading state
                toggleLoadingState(false);

                if (data.success) {
                    // Update the statistics display with the new data
                    updateStatisticsDisplay(data.data);

                    // Show success notification
                    showNotification(
                        'Filter angewendet',
                        'Die Statistikdaten wurden entsprechend Ihrer Filterkriterien aktualisiert.',
                        'success'
                    );
                } else {
                    // Show error notification
                    showNotification(
                        'Fehler',
                        data.error || 'Die Statistikdaten konnten nicht aktualisiert werden.',
                        'error'
                    );
                }
            })
            .catch(error => {
                // Hide loading state
                toggleLoadingState(false);

                // Show error notification
                showNotification(
                    'Fehler',
                    'Bei der Aktualisierung der Statistikdaten ist ein Fehler aufgetreten.',
                    'error'
                );
                console.error('Error:', error);

                // Update with dummy data in case of error
                updateStatisticsDisplay(generateDummyData());
            });
    }

    // Update statistics display with the new data
    function updateStatisticsDisplay(data) {
        // Update card values
        updateCardValues(data);

        // Update charts
        updateAllCharts(data);

        // Update tables
        updateTables(data);
    }

    // Update card values
    function updateCardValues(data) {
        document.getElementById('total-hours-card').textContent = data.totalHours.toFixed(1) + " Std";
        document.getElementById('productivity-rate-card').textContent = data.productivityRate.toFixed(1) + "%";
        document.getElementById('absence-days-card').textContent = Math.round(data.totalAbsenceDays) + " Tage";
        document.getElementById('active-projects-card').textContent = data.activeProjects;
    }

    // Update tables
    function updateTables(data) {
        // Update productivity ranking table
        const productivityTableBody = document.getElementById('productivity-table-body');
        if (productivityTableBody && data.productivityRanking) {
            let tableHTML = '';

            data.productivityRanking.forEach((employee, index) => {
                tableHTML += `
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${index + 1}</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="flex items-center">
                            <div class="flex-shrink-0 h-10 w-10">
                                ${employee.hasProfileImage
                    ? `<img class="h-10 w-10 rounded-full" src="/employees/${employee.id}/profile-image" alt="">`
                    : `<div class="h-10 w-10 rounded-full bg-green-100 flex items-center justify-center text-green-800 font-medium">
                                        ${getInitials(employee.name)}
                                       </div>`
                }
                            </div>
                            <div class="ml-4">
                                <div class="text-sm font-medium text-gray-900">${employee.name}</div>
                            </div>
                        </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${employee.department}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${employee.hours.toFixed(1)} Std</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="flex items-center">
                            <div class="w-full bg-gray-200 rounded-full h-2.5">
                                <div class="bg-green-600 h-2.5 rounded-full" style="width: ${employee.productivityRate.toFixed(1)}%"></div>
                            </div>
                            <span class="ml-2 text-sm text-gray-900">${employee.productivityRate.toFixed(1)}%</span>
                        </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm">
                        ${employee.isTrendPositive
                    ? `<span class="text-green-600">${employee.trendFormatted}</span>`
                    : employee.isTrendNegative
                        ? `<span class="text-red-600">${employee.trendFormatted}</span>`
                        : `<span class="text-gray-500">0.0%</span>`
                }
                    </td>
                </tr>
                `;
            });

            productivityTableBody.innerHTML = tableHTML;
        }

        // Update projects table
        const projectsTableBody = document.getElementById('projects-table-body');
        if (projectsTableBody && data.projectDetails) {
            let tableHTML = '';

            data.projectDetails.forEach(project => {
                tableHTML += `
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="text-sm font-medium text-gray-900">${project.name}</div>
                        <div class="text-sm text-gray-500">ID: ${project.id}</div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full
                            ${project.status === 'Abgeschlossen' ? 'bg-green-100 text-green-800' :
                    project.status === 'In Arbeit' ? 'bg-blue-100 text-blue-800' :
                        project.status === 'Kritisch' ? 'bg-red-100 text-red-800' :
                            'bg-yellow-100 text-yellow-800'}">
                            ${project.status}
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${project.teamSize} Mitarbeiter
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${project.hoursFormatted}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="flex items-center">
                            <div class="w-full bg-gray-200 rounded-full h-2.5">
                                <div class="${project.efficiencyClass} h-2.5 rounded-full" style="width: ${project.efficiencyFormatted}"></div>
                            </div>
                            <span class="ml-2 text-sm text-gray-900">${project.efficiencyFormatted}</span>
                        </div>
                    </td>
                </tr>
                `;
            });

            projectsTableBody.innerHTML = tableHTML;
        }

        // Update absence table
        const absenceTableBody = document.getElementById('absence-table-body');
        if (absenceTableBody && data.currentAbsences) {
            let tableHTML = '';

            data.currentAbsences.forEach(absence => {
                tableHTML += `
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="flex items-center">
                            <div class="flex-shrink-0 h-10 w-10">
                                ${absence.hasProfileImage
                    ? `<img class="h-10 w-10 rounded-full" src="/employees/${absence.employeeId}/profile-image" alt="">`
                    : `<div class="h-10 w-10 rounded-full bg-green-100 flex items-center justify-center text-green-800 font-medium">
                                        ${getInitials(absence.employeeName)}
                                       </div>`
                }
                            </div>
                            <div class="ml-4">
                                <div class="text-sm font-medium text-gray-900">${absence.employeeName}</div>
                            </div>
                        </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full
                            ${absence.type === 'vacation' ? 'bg-green-100 text-green-800' :
                    absence.type === 'sick' ? 'bg-red-100 text-red-800' :
                        'bg-blue-100 text-blue-800'}">
                            ${absence.type === 'vacation' ? 'Urlaub' :
                    absence.type === 'sick' ? 'Krank' :
                        'Sonderurlaub'}
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${formatDate(absence.startDate)}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${formatDate(absence.endDate)}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${absence.days.toFixed(1)} Tage</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full
                            ${absence.status === 'approved' ? 'bg-green-100 text-green-800' :
                    absence.status === 'rejected' ? 'bg-red-100 text-red-800' :
                        absence.status === 'requested' ? 'bg-yellow-100 text-yellow-800' :
                            'bg-gray-100 text-gray-800'}">
                            ${absence.status === 'approved' ? 'Genehmigt' :
                    absence.status === 'rejected' ? 'Abgelehnt' :
                        absence.status === 'requested' ? 'Beantragt' :
                            'Storniert'}
                        </span>
                    </td>
                </tr>
                `;
            });

            absenceTableBody.innerHTML = tableHTML;
        }
    }

    // Helper function to format dates
    function formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('de-DE', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric'
        });
    }

    // Helper function to get initials
    function getInitials(name) {
        if (!name) return '?';

        const parts = name.split(' ');
        if (parts.length === 1) {
            return parts[0].charAt(0).toUpperCase();
        }

        return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
    }

    // Chart instances
    let charts = {};

    // Color palette
    const chartColors = [
        '#15803D', // Dark green
        '#22C55E', // Green
        '#4ADE80', // Light green
        '#86EFAC', // Very light green
        '#BBF7D0', // Pastel green
        '#2563EB', // Blue
        '#3B82F6', // Light blue
        '#F59E0B', // Yellow/Orange
        '#FBBF24'  // Yellow
    ];

    // Update all charts
    function updateAllCharts(data) {
        // Weekday hours chart
        updateChart('weekdayHoursChart', 'bar', {
            labels: Object.keys(data.weekdayHours),
            datasets: [{
                label: 'Durchschnittliche Stunden',
                data: Object.values(data.weekdayHours),
                backgroundColor: chartColors[1],
                borderRadius: 6
            }]
        }, {
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
        });

        // Project hours chart
        updateChart('projectHoursChart', 'doughnut', {
            labels: data.projectHours.map(project => project.name),
            datasets: [{
                data: data.projectHours.map(project => project.hours),
                backgroundColor: chartColors.slice(0, data.projectHours.length),
                borderWidth: 1
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                }
            }
        });

        // Productivity timeline chart
        updateChart('productivityTimelineChart', 'line', {
            labels: data.productivityTimeline.map(item => item.month),
            datasets: [{
                label: 'Produktivitätsrate',
                data: data.productivityTimeline.map(item => item.rate),
                borderColor: chartColors[0],
                backgroundColor: 'rgba(34, 197, 94, 0.1)',
                fill: true,
                tension: 0.4
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    min: 50,
                    max: 100,
                    title: {
                        display: true,
                        text: 'Produktivität (%)'
                    }
                }
            }
        });

        // Absence types chart
        updateChart('absenceTypesChart', 'pie', {
            labels: Object.keys(data.absenceTypes),
            datasets: [{
                data: Object.values(data.absenceTypes),
                backgroundColor: [chartColors[1], chartColors[5], chartColors[7]],
                borderWidth: 1
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                }
            }
        });

        // Project productivity chart
        updateChart('projectProductivityChart', 'bar', {
            labels: data.projectProductivity.map(project => project.name),
            datasets: [{
                label: 'Produktivitätsrate (%)',
                data: data.projectProductivity.map(project => project.rate),
                backgroundColor: chartColors[1],
                borderRadius: 6
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    min: 0,
                    max: 100,
                    title: {
                        display: true,
                        text: 'Produktivität (%)'
                    }
                }
            }
        });

        // Employee productivity chart
        updateChart('employeeProductivityChart', 'bar', {
            labels: data.employeeProductivity.map(employee => employee.name),
            datasets: [{
                label: 'Produktivitätsrate (%)',
                data: data.employeeProductivity.map(employee => employee.rate),
                backgroundColor: chartColors[0],
                borderRadius: 6
            }]
        }, {
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                x: {
                    min: 0,
                    max: 100,
                    title: {
                        display: true,
                        text: 'Produktivität (%)'
                    }
                }
            }
        });

        // Project progress chart
        updateChart('projectProgressChart', 'bar', {
            labels: data.projectProgress.map(project => project.name),
            datasets: [{
                label: 'Fortschritt (%)',
                data: data.projectProgress.map(project => project.progress),
                backgroundColor: chartColors[1],
                borderRadius: 6
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    min: 0,
                    max: 100,
                    title: {
                        display: true,
                        text: 'Fortschritt (%)'
                    }
                }
            }
        });

        // Resource allocation chart
        updateChart('resourceAllocationChart', 'doughnut', {
            labels: data.resourceAllocation.map(project => project.name),
            datasets: [{
                data: data.resourceAllocation.map(project => project.teamSize),
                backgroundColor: chartColors.slice(0, data.resourceAllocation.length),
                borderWidth: 1
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `${context.label}: ${context.raw} Mitarbeiter`;
                        }
                    }
                }
            }
        });

        // Absence type detail chart
        updateChart('absenceTypeDetailChart', 'pie', {
            labels: Object.keys(data.absenceTypeDetail),
            datasets: [{
                data: Object.values(data.absenceTypeDetail),
                backgroundColor: [
                    chartColors[1], // Urlaub
                    chartColors[5], // Krankheit
                    chartColors[7], // Sonderurlaub
                    chartColors[3], // Elternzeit
                    chartColors[8]  // Fortbildung
                ],
                borderWidth: 1
            }]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                }
            }
        });

        // Absence timeline chart
        updateChart('absenceTimelineChart', 'bar', {
            labels: data.absenceTimeline.map(item => item.month),
            datasets: [
                {
                    label: 'Urlaub',
                    data: data.absenceTimeline.map(item => item.vacation),
                    backgroundColor: 'rgba(34, 197, 94, 0.5)',
                    borderColor: chartColors[1],
                    borderWidth: 1,
                    stack: 'Stack 0'
                },
                {
                    label: 'Krankheit',
                    data: data.absenceTimeline.map(item => item.sick),
                    backgroundColor: 'rgba(37, 99, 235, 0.5)',
                    borderColor: chartColors[5],
                    borderWidth: 1,
                    stack: 'Stack 0'
                },
                {
                    label: 'Sonstige',
                    data: data.absenceTimeline.map(item => item.other),
                    backgroundColor: 'rgba(245, 158, 11, 0.5)',
                    borderColor: chartColors[7],
                    borderWidth: 1,
                    stack: 'Stack 0'
                }
            ]
        }, {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                tooltip: {
                    mode: 'index',
                    intersect: false
                }
            },
            scales: {
                x: {
                    stacked: true,
                },
                y: {
                    stacked: true,
                    title: {
                        display: true,
                        text: 'Tage'
                    }
                }
            }
        });
    }

    // Create or update chart helper function
    function updateChart(chartId, type, data, options) {
        const ctx = document.getElementById(chartId)?.getContext('2d');
        if (!ctx) return;

        // Destroy existing chart if it exists
        if (charts[chartId]) {
            charts[chartId].destroy();
        }

        // Create new chart
        charts[chartId] = new Chart(ctx, {
            type: type,
            data: data,
            options: options
        });
    }

    // Notification function
    function showNotification(title, message, type = 'info', duration = 3000) {
        const container = document.getElementById('notification-container');
        if (!container) return;

        // Create notification element
        const notification = document.createElement('div');
        notification.className = 'rounded-lg shadow-lg overflow-hidden transform transition-all duration-300 opacity-0 translate-x-full';

        // Set color based on type
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
                <span class="sr-only">Schließen</span>
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

    // Generate dummy data for testing
    function generateDummyData() {
        return {
            totalHours: 1425.5,
            productivityRate: 87.2,
            totalAbsenceDays: 45,
            activeProjects: 5,

            weekdayHours: {
                'Montag': 8.2,
                'Dienstag': 8.5,
                'Mittwoch': 7.9,
                'Donnerstag': 8.3,
                'Freitag': 7.1,
                'Samstag': 1.4,
                'Sonntag': 0.3
            },

            projectHours: [
                { id: 'proj-1', name: 'Website Redesign', hours: 320, share: 22.4 },
                { id: 'proj-2', name: 'Mobile App', hours: 780, share: 54.7 },
                { id: 'proj-3', name: 'Datenmigration', hours: 150, share: 10.5 },
                { id: 'proj-4', name: 'Security Audit', hours: 95, share: 6.7 },
                { id: 'proj-5', name: 'CRM Implementation', hours: 80, share: 5.7 }
            ],

            productivityTimeline: [
                { month: 'Jan', rate: 82 },
                { month: 'Feb', rate: 83 },
                { month: 'Mär', rate: 80 },
                { month: 'Apr', rate: 84 },
                { month: 'Mai', rate: 87 },
                { month: 'Jun', rate: 85 },
                { month: 'Jul', rate: 88 },
                { month: 'Aug', rate: 89 },
                { month: 'Sep', rate: 86 },
                { month: 'Okt', rate: 85 },
                { month: 'Nov', rate: 87 },
                { month: 'Dez', rate: 90 }
            ],

            absenceTypes: {
                'Urlaub': 120,
                'Krankheit': 45,
                'Sonderurlaub': 15
            },

            projectProductivity: [
                { id: 'proj-1', name: 'Website Redesign', rate: 87 },
                { id: 'proj-2', name: 'Mobile App', rate: 92 },
                { id: 'proj-3', name: 'Datenmigration', rate: 89 },
                { id: 'proj-4', name: 'Security Audit', rate: 65 },
                { id: 'proj-5', name: 'CRM Implementation', rate: 90 }
            ],

            employeeProductivity: [
                { id: 'emp-1', name: 'Max Mustermann', rate: 95 },
                { id: 'emp-2', name: 'Anna Schmidt', rate: 89 },
                { id: 'emp-3', name: 'Timo Becker', rate: 87 },
                { id: 'emp-4', name: 'Lisa Meier', rate: 92 },
                { id: 'emp-5', name: 'Jan Weber', rate: 84 }
            ],

            productivityRanking: [
                {
                    id: 'emp-1',
                    name: 'Max Mustermann',
                    department: 'Entwicklung',
                    hours: 160.5,
                    productivityRate: 95,
                    trend: 3,
                    trendFormatted: '+3.0%',
                    isTrendPositive: true,
                    isTrendNegative: false,
                    hasProfileImage: false
                },
                {
                    id: 'emp-4',
                    name: 'Lisa Meier',
                    department: 'Design',
                    hours: 145.2,
                    productivityRate: 92,
                    trend: 1.5,
                    trendFormatted: '+1.5%',
                    isTrendPositive: true,
                    isTrendNegative: false,
                    hasProfileImage: false
                },
                {
                    id: 'emp-2',
                    name: 'Anna Schmidt',
                    department: 'Marketing',
                    hours: 138.7,
                    productivityRate: 89,
                    trend: -1.2,
                    trendFormatted: '-1.2%',
                    isTrendPositive: false,
                    isTrendNegative: true,
                    hasProfileImage: false
                },
                {
                    id: 'emp-3',
                    name: 'Timo Becker',
                    department: 'Entwicklung',
                    hours: 152.8,
                    productivityRate: 87,
                    trend: 2.3,
                    trendFormatted: '+2.3%',
                    isTrendPositive: true,
                    isTrendNegative: false,
                    hasProfileImage: false
                },
                {
                    id: 'emp-5',
                    name: 'Jan Weber',
                    department: 'Vertrieb',
                    hours: 130.5,
                    productivityRate: 84,
                    trend: -2.5,
                    trendFormatted: '-2.5%',
                    isTrendPositive: false,
                    isTrendNegative: true,
                    hasProfileImage: false
                }
            ],

            projectProgress: [
                { id: 'proj-1', name: 'Website Redesign', progress: 65 },
                { id: 'proj-2', name: 'Mobile App', progress: 40 },
                { id: 'proj-3', name: 'Datenmigration', progress: 100 },
                { id: 'proj-4', name: 'Security Audit', progress: 30 },
                { id: 'proj-5', name: 'CRM Implementation', progress: 10 }
            ],

            resourceAllocation: [
                { id: 'proj-1', name: 'Website Redesign', teamSize: 5 },
                { id: 'proj-2', name: 'Mobile App', teamSize: 8 },
                { id: 'proj-3', name: 'Datenmigration', teamSize: 3 },
                { id: 'proj-4', name: 'Security Audit', teamSize: 2 },
                { id: 'proj-5', name: 'CRM Implementation', teamSize: 6 }
            ],

            projectDetails: [
                {
                    id: 'proj-1',
                    name: 'Website Redesign',
                    status: 'In Arbeit',
                    teamSize: 5,
                    hours: 320,
                    hoursFormatted: '320.0 Std',
                    efficiency: 87.2,
                    efficiencyFormatted: '87.2%',
                    efficiencyClass: 'bg-green-600'
                },
                {
                    id: 'proj-2',
                    name: 'Mobile App',
                    status: 'In Arbeit',
                    teamSize: 8,
                    hours: 780,
                    hoursFormatted: '780.0 Std',
                    efficiency: 92.5,
                    efficiencyFormatted: '92.5%',
                    efficiencyClass: 'bg-green-600'
                },
                {
                    id: 'proj-3',
                    name: 'Datenmigration',
                    status: 'Abgeschlossen',
                    teamSize: 3,
                    hours: 150,
                    hoursFormatted: '150.0 Std',
                    efficiency: 89.0,
                    efficiencyFormatted: '89.0%',
                    efficiencyClass: 'bg-green-600'
                },
                {
                    id: 'proj-4',
                    name: 'Security Audit',
                    status: 'Kritisch',
                    teamSize: 2,
                    hours: 95,
                    hoursFormatted: '95.0 Std',
                    efficiency: 65.0,
                    efficiencyFormatted: '65.0%',
                    efficiencyClass: 'bg-yellow-600'
                },
                {
                    id: 'proj-5',
                    name: 'CRM Implementation',
                    status: 'Geplant',
                    teamSize: 6,
                    hours: 80,
                    hoursFormatted: '80.0 Std',
                    efficiency: 90.0,
                    efficiencyFormatted: '90.0%',
                    efficiencyClass: 'bg-green-600'
                }
            ],

            absenceTypeDetail: {
                'Urlaub': 150,
                'Krankheit': 75,
                'Sonderurlaub': 25,
                'Elternzeit': 30,
                'Fortbildung': 40
            },

            absenceTimeline: [
                { month: 'Jan', vacation: 5, sick: 8, other: 2 },
                { month: 'Feb', vacation: 7, sick: 10, other: 3 },
                { month: 'Mär', vacation: 10, sick: 6, other: 4 },
                { month: 'Apr', vacation: 8, sick: 5, other: 5 },
                { month: 'Mai', vacation: 12, sick: 4, other: 3 },
                { month: 'Jun', vacation: 25, sick: 3, other: 2 },
                { month: 'Jul', vacation: 30, sick: 4, other: 3 },
                { month: 'Aug', vacation: 35, sick: 5, other: 4 },
                { month: 'Sep', vacation: 15, sick: 7, other: 2 },
                { month: 'Okt', vacation: 8, sick: 9, other: 3 },
                { month: 'Nov', vacation: 7, sick: 12, other: 4 },
                { month: 'Dez', vacation: 20, sick: 10, other: 3 }
            ],

            currentAbsences: [
                {
                    id: 'abs-1',
                    employeeId: 'emp-1',
                    employeeName: 'Max Mustermann',
                    type: 'vacation',
                    startDate: '2023-06-15T00:00:00Z',
                    endDate: '2023-06-30T00:00:00Z',
                    days: 12,
                    status: 'approved',
                    hasProfileImage: false,
                    affectedProjects: ['Website Redesign', 'Mobile App']
                },
                {
                    id: 'abs-2',
                    employeeId: 'emp-3',
                    employeeName: 'Timo Becker',
                    type: 'sick',
                    startDate: '2023-06-10T00:00:00Z',
                    endDate: '2023-06-12T00:00:00Z',
                    days: 3,
                    status: 'approved',
                    hasProfileImage: false,
                    affectedProjects: ['Security Audit']
                },
                {
                    id: 'abs-3',
                    employeeId: 'emp-2',
                    employeeName: 'Anna Schmidt',
                    type: 'vacation',
                    startDate: '2023-07-05T00:00:00Z',
                    endDate: '2023-07-15T00:00:00Z',
                    days: 9,
                    status: 'requested',
                    hasProfileImage: false,
                    affectedProjects: ['Mobile App', 'CRM Implementation']
                }
            ]
        };
    }

    // Initialize the page with dummy data
    updateStatisticsDisplay(generateDummyData());
});