// 1. Edit Modal Logic
function openEditModal(id, empName, shiftName, startDate, endDate) {
    var modal = document.getElementById('editModal');
    modal.style.display = "flex"; 

    // Fill Form Data
    document.getElementById('modal-id').value = id;
    document.getElementById('modal-employee').value = empName;
    document.getElementById('modal-shift').value = shiftName;
    
    // Format dates to YYYY-MM-DD for the input field
    if(startDate) document.getElementById('modal-start').value = startDate.split('T')[0];
    if(endDate) document.getElementById('modal-end').value = endDate.split('T')[0];
}

// Close modal on outside click
window.onclick = function(event) {
    var modal = document.getElementById('editModal');
    if (event.target == modal) {
        modal.style.display = "none";
    }
}

// 2. Global Page Load Logic (Error Banners & Calendar)
document.addEventListener('DOMContentLoaded', function() {
    
    // --- Error Banner Handling ---
    const urlParams = new URLSearchParams(window.location.search);
    const errorMsg = urlParams.get('error');

    if (errorMsg) {
        const banner = document.getElementById('error-banner');
        const textSpan = document.getElementById('error-text');
        textSpan.innerText = decodeURIComponent(errorMsg);
        banner.style.display = 'block';
        // Clean URL so refresh doesn't show error again
        window.history.replaceState({}, document.title, window.location.pathname + "?view=schedule");
    }

    // --- Calendar Logic ---
    var calendarEl = document.getElementById('calendar');
    
    // Only run if the calendar element exists on this page
    if(calendarEl) {
        const colorPalette = ['#0d6efd', '#6610f2', '#6f42c1', '#d63384', '#dc3545', '#fd7e14', '#198754', '#20c997', '#0dcaf0', '#343a40'];

        function getDynamicColor(text) {
            if (text === 'Morning') return '#17a2b8';
            if (text === 'Afternoon') return '#fd7e14';
            if (text === 'Night') return '#343a40';
            
            let hash = 0;
            for (let i = 0; i < text.length; i++) {
                hash = text.charCodeAt(i) + ((hash << 5) - hash);
            }
            return colorPalette[Math.abs(hash) % colorPalette.length];
        }

        fetch('/api/allocations')
            .then(response => response.json())
            .then(data => {
                if(!data) return;
                var calendarEvents = data.map(function(item) {
                    var color = getDynamicColor(item.ShiftName);
                    return {
                        title: item.EmployeeName + ' (' + item.ShiftName + ')',
                        start: item.StartDate,
                        end: item.EndDate,
                        backgroundColor: color,
                        borderColor: color,
                        textColor: '#ffffff'
                    };
                });

                var calendar = new FullCalendar.Calendar(calendarEl, {
                    initialView: 'dayGridMonth',
                    displayEventTime: false,
                    events: calendarEvents,
                    height: 650
                });
                calendar.render();
            });
    }
});