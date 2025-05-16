// frontend/static/js/statistics.js - Fixed version

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

    // Initialize charts with server data
    initializeChartsWithServerData();

    // Function to initialize charts with real data from server
    function initializeChartsWithServerData() {
        const now = new Date();
        const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
        const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0);

        // Fetch initial data for charts
        fetchFilteredStatistics({
            dateRangeKey: 'this-month',
            startDate: startOfMonth.toISOString(),
            endDate: endOfMonth.toISOString(),
            projectId: '',
            employeeIds: []
        });
    }
});