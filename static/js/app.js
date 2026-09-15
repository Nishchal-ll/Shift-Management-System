// ==========================================================================
// ShiftPro - Client-side Interactive Logic & Real-time MQTT Service
// ==========================================================================

// 1. Edit Modal Logic
function openEditModal(id, empName, shiftName, startDate, endDate) {
    var modal = document.getElementById('editModal');
    if (modal) {
        modal.style.display = "flex"; 

        // Fill Form Data
        document.getElementById('modal-id').value = id;
        document.getElementById('modal-employee').value = empName;
        document.getElementById('modal-shift').value = shiftName;
        
        // Format dates to YYYY-MM-DD for the input field
        if(startDate) document.getElementById('modal-start').value = startDate.split('T')[0];
        if(endDate) document.getElementById('modal-end').value = endDate.split('T')[0];
    }
}

// Close modal on outside click
window.onclick = function(event) {
    var modal = document.getElementById('editModal');
    if (modal && event.target == modal) {
        modal.style.display = "none";
    }
};

// ==========================================================================
// 2. Notification System (Toasts & Bell Dropdown)
// ==========================================================================

function updateNotificationBadge(increment = 0) {
    const badge = document.getElementById('notifCountBadge');
    if (!badge) return;
    if (increment > 0) {
        let cur = parseInt(badge.innerText) || 0;
        let next = cur + increment;
        badge.innerText = next > 9 ? '9+' : next;
        badge.style.display = 'flex';
    }
}

function prependNotificationItem(payload) {
    const listEl = document.getElementById('notifDropdownList');
    if (!listEl) return;

    const emptyState = document.getElementById('notifEmptyState');
    if (emptyState) {
        emptyState.remove();
    }

    const timeStr = 'Just now';
    const notifItem = document.createElement('div');
    notifItem.className = 'notif-dropdown-item unread';
    notifItem.innerHTML = `
        <div class="notif-item-icon">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
        </div>
        <div class="notif-item-body">
            <div class="notif-item-title">${escapeHtml(payload.title || 'Notification')}</div>
            <div class="notif-item-msg">${escapeHtml(payload.message || '')}</div>
            <div class="notif-item-time">${timeStr}</div>
        </div>
    `;
    listEl.prepend(notifItem);
    updateNotificationBadge(1);
}

function clearAllNotifications() {
    const listEl = document.getElementById('notifDropdownList');
    const badge = document.getElementById('notifCountBadge');
    
    if (listEl) {
        listEl.innerHTML = '<div class="notif-dropdown-empty" id="notifEmptyState">No new notifications</div>';
    }
    if (badge) {
        badge.innerText = '0';
        badge.style.display = 'none';
    }

    fetch('/api/notifications/clear', { method: 'POST' })
        .catch(err => console.error('Error clearing notifications:', err));
}

function showToastNotification(payload) {
    const container = document.getElementById('toastContainer');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = 'toast-item toast-info';
    
    if (payload.type === 'shift_assigned' || payload.type === 'swap_approved') {
        toast.className = 'toast-item toast-success';
    } else if (payload.type === 'swap_requested') {
        toast.className = 'toast-item toast-warning';
    }

    toast.innerHTML = `
        <div class="toast-icon">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
            </svg>
        </div>
        <div class="toast-body">
            <div class="toast-title">${escapeHtml(payload.title || 'System Notification')}</div>
            <div class="toast-message">${escapeHtml(payload.message || '')}</div>
        </div>
        <button class="toast-close" aria-label="Close Notification">&times;</button>
    `;

    // Close button handler
    toast.querySelector('.toast-close').addEventListener('click', () => {
        removeToast(toast);
    });

    container.appendChild(toast);

    // Auto dismiss after 6 seconds
    setTimeout(() => {
        removeToast(toast);
    }, 6000);
}

function removeToast(toast) {
    if (!toast) return;
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    setTimeout(() => {
        if (toast.parentNode) {
            toast.parentNode.removeChild(toast);
        }
    }, 300);
}

function escapeHtml(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

// ==========================================================================
// 3. Real-Time MQTT Listener Connection
// ==========================================================================

function initRealtimeMQTT() {
    if (typeof mqtt === 'undefined') {
        console.warn('MQTT.js library not loaded. Retrying in 1s...');
        setTimeout(initRealtimeMQTT, 1000);
        return;
    }

    const currentUser = document.body.getAttribute('data-current-user');
    const currentRole = document.body.getAttribute('data-current-role');

    const host = window.location.hostname || 'localhost';
    const wsUrl = `ws://${host}:9001`;

    console.log(`[ShiftPro MQTT] Connecting to broker at ${wsUrl}...`);

    try {
        const client = mqtt.connect(wsUrl, {
            clientId: `web_${currentUser || 'anon'}_${Math.random().toString(16).substr(2, 8)}`,
            clean: true,
            connectTimeout: 4000,
            reconnectPeriod: 3000,
        });

        client.on('connect', () => {
            console.log('[ShiftPro MQTT] Connected to broker successfully.');

            // Subscribe to broad shifts channel
            client.subscribe('shiftpro/shifts', (err) => {
                if (!err) console.log('[ShiftPro MQTT] Subscribed to topic: shiftpro/shifts');
            });

            // If user is logged in, subscribe to user-specific channel
            if (currentUser) {
                const userTopic = `shiftpro/notifications/${currentUser}`;
                client.subscribe(userTopic, (err) => {
                    if (!err) console.log(`[ShiftPro MQTT] Subscribed to personal topic: ${userTopic}`);
                });
            }
        });

        client.on('message', (topic, message) => {
            try {
                const payload = JSON.parse(message.toString());
                console.log('[ShiftPro MQTT] Received message on topic:', topic, payload);

                // If user-specific notification or relevant to current user
                if (topic.startsWith('shiftpro/notifications/') || (currentUser && payload.employee === currentUser)) {
                    prependNotificationItem(payload);
                    showToastNotification(payload);
                } else if (!currentUser && payload.type) {
                    // For admin or general listener
                    showToastNotification(payload);
                }
            } catch (err) {
                console.error('[ShiftPro MQTT] Error parsing message:', err);
            }
        });

        client.on('error', (err) => {
            console.warn('[ShiftPro MQTT] Connection warning:', err);
        });

    } catch (err) {
        console.error('[ShiftPro MQTT] Initialization error:', err);
    }
}

// ==========================================================================
// 4. Cron Auto-Scheduler Control Functions
// ==========================================================================

function toggleCronScheduler(enabled) {
    const statusPill = document.getElementById('cronStatusPill');
    const statusText = document.getElementById('cronStatusText');
    const toggleLabel = document.getElementById('cronToggleLabel');

    if (toggleLabel) {
        toggleLabel.innerText = enabled ? 'Cron Enabled' : 'Cron Paused';
    }

    if (statusPill && statusText) {
        if (enabled) {
            statusPill.className = 'cron-status-pill status-active';
            statusText.innerText = 'Active (Running every 1m)';
        } else {
            statusPill.className = 'cron-status-pill status-paused';
            statusText.innerText = 'Paused';
        }
    }

    const formData = new FormData();
    formData.append('enabled', enabled ? 'true' : 'false');

    fetch('/admin/cron/toggle', {
        method: 'POST',
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'Accept': 'application/json'
        },
        body: formData
    })
    .then(res => res.json())
    .then(data => {
        if (data.success) {
            showToastNotification({
                type: enabled ? 'shift_assigned' : 'swap_requested',
                title: 'Auto-Scheduler Updated',
                message: enabled ? 'Shift Auto-Scheduler is now ACTIVE and running in background.' : 'Shift Auto-Scheduler is now PAUSED.'
            });
        }
    })
    .catch(err => console.error('Error toggling cron scheduler:', err));
}

function runCronNow() {
    const btn = document.getElementById('btnRunCronNow');
    if (btn) {
        btn.classList.add('loading');
        btn.innerHTML = '<span>Running Auto-Assign...</span>';
    }

    fetch('/admin/cron/run-now', {
        method: 'POST',
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'Accept': 'application/json'
        }
    })
    .then(res => res.json())
    .then(data => {
        if (btn) {
            btn.classList.remove('loading');
            btn.innerHTML = `
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polygon points="5 3 19 12 5 21 5 3"></polygon>
                </svg>
                <span>Run Auto-Assign Now</span>
            `;
        }

        if (data.success) {
            const count = data.count || 0;
            const lastRunEl = document.getElementById('cronLastRun');
            const shiftsAssignedEl = document.getElementById('cronShiftsAssigned');
            if (lastRunEl) lastRunEl.innerText = 'Just now';
            if (shiftsAssignedEl && data.status) shiftsAssignedEl.innerText = data.status.shifts_assigned;

            showToastNotification({
                type: 'shift_assigned',
                title: 'Auto-Allocation Complete',
                message: `Successfully generated and allocated ${count} shifts across staff with live MQTT notifications.`
            });

            // Reload page after 1.2s to show newly generated roster table
            setTimeout(() => {
                window.location.reload();
            }, 1200);
        } else {
            showToastNotification({
                type: 'swap_requested',
                title: 'Auto-Allocation Notice',
                message: data.error || 'No new shifts needed or all slots filled.'
            });
        }
    })
    .catch(err => {
        if (btn) {
            btn.classList.remove('loading');
            btn.innerHTML = `
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polygon points="5 3 19 12 5 21 5 3"></polygon>
                </svg>
                <span>Run Auto-Assign Now</span>
            `;
        }
        console.error('Error running cron cycle:', err);
    });
}

// ==========================================================================
// 5. Global Page Load Logic (Calendar, Banners, Search)
// ==========================================================================

document.addEventListener('DOMContentLoaded', function() {
    initRealtimeMQTT();

    // --- Error Banner Handling ---
    const urlParams = new URLSearchParams(window.location.search);
    const errorMsg = urlParams.get('error');

    if (errorMsg) {
        const banner = document.getElementById('error-banner');
        const textSpan = document.getElementById('error-text');
        if (banner && textSpan) {
            textSpan.innerText = decodeURIComponent(errorMsg);
            banner.style.display = 'flex';
        }
    }

    // --- Global Search Keyboard Shortcut (⌘K / Ctrl+K) ---
    const searchInput = document.getElementById('globalSearchInput');
    if (searchInput) {
        document.addEventListener('keydown', function(e) {
            if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
                e.preventDefault();
                searchInput.focus();
            }
        });
    }

    // --- FullCalendar Initialization ---
    var calendarEl = document.getElementById('calendar');
    if (calendarEl && typeof FullCalendar !== 'undefined') {
        function getDynamicColor(text) {
            if (text === 'Morning') return '#d97706';   // Amber 600
            if (text === 'Afternoon') return '#0284c7'; // Sky 600
            if (text === 'Night') return '#7c3aed';     // Violet 600
            
            const colorPalette = ['#4f46e5', '#059669', '#d97706', '#dc2626', '#7c3aed', '#0284c7', '#ea580c'];
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
                    headerToolbar: {
                        left: 'prev,next today',
                        center: 'title',
                        right: 'dayGridMonth,timeGridWeek'
                    },
                    height: 650
                });
                calendar.render();
            })
            .catch(err => console.error("Error loading calendar allocations:", err));
    }
});