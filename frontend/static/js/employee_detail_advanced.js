// Add this script to your employee detail page or include it in a separate file

document.addEventListener('DOMContentLoaded', function() {
    // Only initialize if the vacation tab exists
    if (document.getElementById('vacation-tab')) {
        // Initialize vacation filters
        initVacationFilters();
    }
});

// Initialize vacation filters
function initVacationFilters() {
    const yearFilter = document.getElementById('vacation-year-filter');
    const typeFilter = document.getElementById('vacation-type-filter');
    const statusFilter = document.getElementById('vacation-status-filter');

    // Apply filters when any filter changes
    yearFilter.addEventListener('change', applyVacationFilters);
    typeFilter.addEventListener('change', applyVacationFilters);
    statusFilter.addEventListener('change', applyVacationFilters);

    // Initial filter application
    applyVacationFilters();
}

// Apply all vacation filters
function applyVacationFilters() {
    const yearFilter = document.getElementById('vacation-year-filter').value;
    const typeFilter = document.getElementById('vacation-type-filter').value;
    const statusFilter = document.getElementById('vacation-status-filter').value;

    const vacationItems = document.querySelectorAll('.vacation-item');

    vacationItems.forEach(item => {
        // Get data attributes
        const itemYear = item.getAttribute('data-year');
        const itemType = item.getAttribute('data-type');
        const itemStatus = item.getAttribute('data-status');

        // Check if the item passes all filters
        const yearMatch = yearFilter === 'all' || yearFilter === itemYear;
        const typeMatch = typeFilter === 'all' || typeFilter === itemType;
        const statusMatch = statusFilter === 'all' || statusFilter === itemStatus;

        // Show or hide based on filter matches
        if (yearMatch && typeMatch && statusMatch) {
            item.style.display = '';
        } else {
            item.style.display = 'none';
        }
    });

    // Update "no results" message
    const visibleItems = Array.from(vacationItems).filter(item => item.style.display !== 'none');
    const noResultsMessage = document.getElementById('no-vacation-results');

    if (visibleItems.length === 0 && !noResultsMessage) {
        const listContainer = document.getElementById('vacation-list');
        if (listContainer) {
            const message = document.createElement('p');
            message.id = 'no-vacation-results';
            message.className = 'text-sm text-gray-500 py-4';
            message.textContent = 'Keine Einträge gefunden, die den Filterkriterien entsprechen.';
            listContainer.appendChild(message);
        }
    } else if (visibleItems.length > 0 && noResultsMessage) {
        noResultsMessage.remove();
    }
}