// Wails runtime will be available as window.go
let app;

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    // Wait for Wails runtime
    if (window.go && window.go.main && window.go.main.App) {
        app = window.go.main.App;
        initApp();
    } else {
        console.error('Wails runtime not available');
        // Fallback for development
        setTimeout(initApp, 1000);
    }
});

function initApp() {
    // Check if config exists
    checkConfig();
    
    // Setup form handlers
    setupConfigForm();
    setupDateForm();
    setupCalendarButtons();
    setupSettingsButtons();
}

async function checkConfig() {
    try {
        const config = await app.GetConfig();
        if (config && config.baseURL && config.username) {
            // Config exists, try to authenticate
            try {
                await app.SaveConfig(config);
                showDateScreen();
            } catch (err) {
                console.error('Auth failed:', err);
                showConfigScreen();
            }
        } else {
            showConfigScreen();
        }
    } catch (err) {
        console.error('Error loading config:', err);
        showConfigScreen();
    }
}

function setupConfigForm() {
    const form = document.getElementById('config-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const config = {
            baseURL: document.getElementById('baseURL').value,
            username: document.getElementById('username').value,
            apiToken: document.getElementById('apiToken').value
        };
        
        try {
            await app.SaveConfig(config);
            showDateScreen();
        } catch (err) {
            alert('Failed to save configuration: ' + err);
        }
    });
    
    // Setup close button
    const closeBtn = document.getElementById('close-config-btn');
    if (closeBtn) {
        closeBtn.addEventListener('click', () => {
            showDateScreen();
        });
    }
    
    // Setup API Token visibility toggle
    const toggleBtn = document.getElementById('toggle-api-token');
    const apiTokenInput = document.getElementById('apiToken');
    if (toggleBtn && apiTokenInput) {
        toggleBtn.addEventListener('click', () => {
            if (apiTokenInput.type === 'password') {
                apiTokenInput.type = 'text';
                toggleBtn.textContent = '🙈';
                toggleBtn.title = 'Hide API Token';
            } else {
                apiTokenInput.type = 'password';
                toggleBtn.textContent = '👁️';
                toggleBtn.title = 'Show API Token';
            }
        });
    }
}

// Calendar widget state
let calendarState = {
    currentDate: new Date(),
    selectedStartDate: null,
    selectedEndDate: null,
    selectingStart: true
};

function setupDateForm() {
    // Initialize calendar widget
    initCalendarWidget();
    
    // Set default dates (last 7 days)
    const endDate = new Date();
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - 7);
    
    calendarState.selectedStartDate = startDate;
    calendarState.selectedEndDate = endDate;
    calendarState.selectingStart = false;
    
    updateCalendarDisplay();
    updateSelectionDisplay();
    
    // Setup navigation buttons
    document.getElementById('prev-month').addEventListener('click', () => {
        calendarState.currentDate.setMonth(calendarState.currentDate.getMonth() - 1);
        updateCalendarDisplay();
    });
    
    document.getElementById('next-month').addEventListener('click', () => {
        calendarState.currentDate.setMonth(calendarState.currentDate.getMonth() + 1);
        updateCalendarDisplay();
    });
    
    // Setup load button
    document.getElementById('load-calendar-btn').addEventListener('click', async () => {
        if (!calendarState.selectedStartDate || !calendarState.selectedEndDate) {
            return;
        }
        
        const startDateStr = formatDate(calendarState.selectedStartDate);
        const endDateStr = formatDate(calendarState.selectedEndDate);
        
        try {
            const workLogs = await app.GetWorkLogs(startDateStr, endDateStr);
            showCalendarScreen(workLogs, startDateStr, endDateStr);
        } catch (err) {
            alert('Failed to load work logs: ' + err);
        }
    });
}

function initCalendarWidget() {
    updateCalendarDisplay();
}

function updateCalendarDisplay() {
    const grid = document.getElementById('calendar-grid');
    grid.innerHTML = '';
    
    const year = calendarState.currentDate.getFullYear();
    const month = calendarState.currentDate.getMonth();
    
    // Update month/year display
    const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
        'July', 'August', 'September', 'October', 'November', 'December'];
    document.getElementById('current-month-year').textContent = 
        `${monthNames[month]} ${year}`;
    
    // Get first day of month and number of days
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startingDayOfWeek = firstDay.getDay();
    
    // Calculate total cells needed (always 6 rows = 42 cells for consistent sizing)
    const totalCells = 42; // 6 weeks * 7 days
    
    // Add empty cells for days before month starts
    for (let i = 0; i < startingDayOfWeek; i++) {
        const emptyCell = document.createElement('div');
        emptyCell.className = 'calendar-day empty';
        grid.appendChild(emptyCell);
    }
    
    // Add cells for each day of the month
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    for (let day = 1; day <= daysInMonth; day++) {
        const date = new Date(year, month, day);
        const cell = document.createElement('div');
        cell.className = 'calendar-day';
        cell.textContent = day;
        cell.dataset.date = formatDate(date);
        
        // Check if today
        if (date.getTime() === today.getTime()) {
            cell.classList.add('today');
        }
        
        // Check if weekend
        const dayOfWeek = date.getDay();
        if (dayOfWeek === 0 || dayOfWeek === 6) {
            cell.classList.add('weekend');
        }
        
        // Check if in selected range
        const cellDate = new Date(date);
        cellDate.setHours(0, 0, 0, 0);
        const cellTime = cellDate.getTime();
        
        if (calendarState.selectedStartDate && calendarState.selectedEndDate) {
            const startDate = new Date(calendarState.selectedStartDate);
            startDate.setHours(0, 0, 0, 0);
            const endDate = new Date(calendarState.selectedEndDate);
            endDate.setHours(0, 0, 0, 0);
            
            const startTime = startDate.getTime();
            const endTime = endDate.getTime();
            
            if (cellTime >= startTime && cellTime <= endTime) {
                cell.classList.add('selected');
                if (cellTime === startTime) {
                    cell.classList.add('range-start');
                }
                if (cellTime === endTime) {
                    cell.classList.add('range-end');
                }
            }
        } else if (calendarState.selectedStartDate) {
            const startDate = new Date(calendarState.selectedStartDate);
            startDate.setHours(0, 0, 0, 0);
            const startTime = startDate.getTime();
            if (cellTime === startTime) {
                cell.classList.add('selected', 'range-start');
            }
        }
        
        // Add click handler
        cell.addEventListener('click', () => handleDateClick(date));
        
        grid.appendChild(cell);
    }
    
    // Add empty cells at the end to always have 42 cells (6 rows)
    const cellsAdded = startingDayOfWeek + daysInMonth;
    const remainingCells = totalCells - cellsAdded;
    for (let i = 0; i < remainingCells; i++) {
        const emptyCell = document.createElement('div');
        emptyCell.className = 'calendar-day empty';
        grid.appendChild(emptyCell);
    }
    
    // Add hover handlers for range preview (only when start date is selected but end date is not)
    if (calendarState.selectedStartDate && !calendarState.selectedEndDate) {
        const allCells = document.querySelectorAll('.calendar-day:not(.empty)');
        allCells.forEach(cell => {
            const cellDateStr = cell.dataset.date;
            if (cellDateStr) {
                cell.addEventListener('mouseenter', () => {
                    const cellDate = new Date(cellDateStr);
                    handleDateHover(cellDate);
                });
            }
        });
        
        // Clear preview when mouse leaves the grid
        grid.addEventListener('mouseleave', clearHoverPreview);
    }
}

function handleDateClick(date) {
    // Normalize dates to midnight for comparison
    const clickedDate = new Date(date);
    clickedDate.setHours(0, 0, 0, 0);
    
    if (calendarState.selectingStart || !calendarState.selectedStartDate) {
        // Start new selection - always clear end date when selecting a new start
        calendarState.selectedStartDate = new Date(clickedDate);
        calendarState.selectedEndDate = null;
        calendarState.selectingStart = false;
    } else {
        // We have a start date but no end date yet
        const startDate = new Date(calendarState.selectedStartDate);
        startDate.setHours(0, 0, 0, 0);
        
        if (clickedDate < startDate) {
            // If clicked date is before start, make it the new start and clear end
            calendarState.selectedStartDate = new Date(clickedDate);
            calendarState.selectedEndDate = null;
            calendarState.selectingStart = false;
        } else {
            // Set as end date
            calendarState.selectedEndDate = new Date(clickedDate);
            calendarState.selectingStart = true;
        }
    }
    
    updateCalendarDisplay();
    updateSelectionDisplay();
}

function handleDateHover(hoveredDate) {
    if (!calendarState.selectedStartDate || calendarState.selectedEndDate) {
        return;
    }
    
    const hovered = new Date(hoveredDate);
    hovered.setHours(0, 0, 0, 0);
    const hoveredTime = hovered.getTime();
    
    const startDate = new Date(calendarState.selectedStartDate);
    startDate.setHours(0, 0, 0, 0);
    const startTime = startDate.getTime();
    
    // Determine the range (start to hovered, or hovered to start if hovered is before start)
    const rangeStart = hoveredTime < startTime ? hoveredTime : startTime;
    const rangeEnd = hoveredTime < startTime ? startTime : hoveredTime;
    
    // Add preview class to all cells in the range
    const allCells = document.querySelectorAll('.calendar-day:not(.empty)');
    allCells.forEach(cell => {
        const cellDateStr = cell.dataset.date;
        if (cellDateStr) {
            const cellDate = new Date(cellDateStr);
            cellDate.setHours(0, 0, 0, 0);
            const cellTime = cellDate.getTime();
            
            if (cellTime >= rangeStart && cellTime <= rangeEnd) {
                cell.classList.add('preview-range');
                if (cellTime === rangeStart) {
                    cell.classList.add('preview-start');
                }
                if (cellTime === rangeEnd) {
                    cell.classList.add('preview-end');
                }
            }
        }
    });
}

function clearHoverPreview() {
    const allCells = document.querySelectorAll('.calendar-day');
    allCells.forEach(cell => {
        cell.classList.remove('preview-range', 'preview-start', 'preview-end');
    });
}

function updateSelectionDisplay() {
    const startDisplay = document.getElementById('selected-start-date');
    const endDisplay = document.getElementById('selected-end-date');
    const loadBtn = document.getElementById('load-calendar-btn');
    
    if (calendarState.selectedStartDate) {
        startDisplay.textContent = formatDateDisplay(calendarState.selectedStartDate);
    } else {
        startDisplay.textContent = 'Not selected';
    }
    
    if (calendarState.selectedEndDate) {
        endDisplay.textContent = formatDateDisplay(calendarState.selectedEndDate);
        loadBtn.disabled = false;
    } else {
        endDisplay.textContent = 'Not selected';
        loadBtn.disabled = true;
    }
}

function formatDateDisplay(date) {
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
        'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return `${monthNames[date.getMonth()]} ${date.getDate()}, ${date.getFullYear()}`;
}

function setupCalendarButtons() {
    document.getElementById('close-btn').addEventListener('click', () => {
        showDateScreen();
    });
}

function setupSettingsButtons() {
    const settingsBtnDate = document.getElementById('settings-btn-date');
    const settingsBtnCalendar = document.getElementById('settings-btn-calendar');
    
    if (settingsBtnDate) {
        settingsBtnDate.addEventListener('click', () => {
            showConfigScreen();
        });
    }
    
    if (settingsBtnCalendar) {
        settingsBtnCalendar.addEventListener('click', () => {
            showConfigScreen();
        });
    }
}

async function showConfigScreen() {
    document.getElementById('config-screen').classList.remove('hidden');
    document.getElementById('date-screen').classList.add('hidden');
    document.getElementById('calendar-screen').classList.add('hidden');
    
    // Load current config values into the form
    try {
        const config = await app.GetConfig();
        if (config) {
            document.getElementById('baseURL').value = config.baseURL || '';
            document.getElementById('username').value = config.username || '';
            document.getElementById('apiToken').value = config.apiToken || '';
        }
    } catch (err) {
        console.error('Error loading config:', err);
    }
}

function showDateScreen() {
    document.getElementById('config-screen').classList.add('hidden');
    document.getElementById('date-screen').classList.remove('hidden');
    document.getElementById('calendar-screen').classList.add('hidden');
    
    // Reset calendar to current month
    calendarState.currentDate = new Date();
    updateCalendarDisplay();
    
    // Load user info if available
    loadUserInfo();
}

async function loadUserInfo() {
    try {
        const userInfo = await app.GetUserInfo();
        const userInfoDiv = document.getElementById('user-info');
        
        let infoHTML = `
            <div class="user-card">
                <div class="user-header">
        `;
        
        if (userInfo.avatarURL) {
            infoHTML += `<img src="${userInfo.avatarURL}" alt="Avatar" class="user-avatar" />`;
        }
        
        infoHTML += `
                    <h3>${userInfo.displayName || 'N/A'}</h3>
                </div>
                <p><strong>Email:</strong> ${userInfo.emailAddress || 'N/A'}</p>
        `;
        
        if (userInfo.accountID) {
            infoHTML += `<p><strong>Account ID:</strong> ${userInfo.accountID}</p>`;
        }
        
        if (userInfo.timeZone) {
            infoHTML += `<p><strong>Timezone:</strong> ${userInfo.timeZone}</p>`;
        }
        
        if (userInfo.locale) {
            infoHTML += `<p><strong>Locale:</strong> ${userInfo.locale}</p>`;
        }
        
        if (userInfo.accountType) {
            infoHTML += `<p><strong>Account Type:</strong> ${userInfo.accountType}</p>`;
        }
        
        infoHTML += `<p><strong>Status:</strong> ${userInfo.active ? 'Active' : 'Inactive'}</p>`;
        infoHTML += `</div>`;
        
        userInfoDiv.innerHTML = infoHTML;
    } catch (err) {
        // Not authenticated yet
        document.getElementById('user-info').innerHTML = '';
    }
}

function showCalendarScreen(workLogs, startDate, endDate) {
    document.getElementById('config-screen').classList.add('hidden');
    document.getElementById('date-screen').classList.add('hidden');
    document.getElementById('calendar-screen').classList.remove('hidden');
    
    renderCalendar(workLogs, startDate, endDate);
}

function renderCalendar(workLogs, startDate, endDate) {
    const tableDiv = document.getElementById('calendar-table');
    
    if (!workLogs || workLogs.length === 0) {
        tableDiv.innerHTML = '<p>No work logs found for the selected date range.</p>';
        return;
    }
    
    // Parse dates
    const start = new Date(startDate);
    const end = new Date(endDate);
    
    // Generate all dates in range
    const dates = [];
    const currentDate = new Date(start);
    while (currentDate <= end) {
        dates.push(new Date(currentDate));
        currentDate.setDate(currentDate.getDate() + 1);
    }
    
    // Build data structure
    const ticketMap = new Map();
    const dailyTotals = new Map();
    const ticketEarliestDate = new Map();
    let grandTotal = 0;
    
    workLogs.forEach(wl => {
        const dateKey = formatDate(new Date(wl.started));
        const ticketKey = wl.issueKey;
        
        // Track earliest date per ticket
        const wlDate = new Date(wl.started);
        if (!ticketEarliestDate.has(ticketKey) || wlDate < ticketEarliestDate.get(ticketKey)) {
            ticketEarliestDate.set(ticketKey, wlDate);
        }
        
        // Initialize ticket if needed
        if (!ticketMap.has(ticketKey)) {
            ticketMap.set(ticketKey, {
                key: ticketKey,
                summary: wl.issueSummary,
                hours: new Map(),
                total: 0
            });
        }
        
        const ticket = ticketMap.get(ticketKey);
        const hours = wl.timeSpentSeconds / 3600;
        
        // Add hours for this date
        if (!ticket.hours.has(dateKey)) {
            ticket.hours.set(dateKey, 0);
        }
        ticket.hours.set(dateKey, ticket.hours.get(dateKey) + hours);
        ticket.total += hours;
        
        // Update daily totals
        if (!dailyTotals.has(dateKey)) {
            dailyTotals.set(dateKey, 0);
        }
        dailyTotals.set(dateKey, dailyTotals.get(dateKey) + hours);
        grandTotal += hours;
    });
    
    // Sort tickets by earliest date
    const sortedTickets = Array.from(ticketMap.values()).sort((a, b) => {
        const dateA = ticketEarliestDate.get(a.key);
        const dateB = ticketEarliestDate.get(b.key);
        return dateA - dateB;
    });
    
    // Create table
    const table = document.createElement('table');
    table.className = 'calendar-table';
    
    // Header row
    const headerRow = document.createElement('tr');
    
    // Ticket column header
    const ticketHeader = document.createElement('th');
    ticketHeader.className = 'sticky-col';
    ticketHeader.textContent = 'Ticket';
    headerRow.appendChild(ticketHeader);
    
    // Date column headers
    dates.forEach(date => {
        const th = document.createElement('th');
        const day = date.getDate();
        const dayName = date.toLocaleDateString('en-US', { weekday: 'short' });
        th.innerHTML = `${String(day).padStart(2, '0')}<br>${dayName}`;
        
        // Highlight weekends
        const dayOfWeek = date.getDay();
        if (dayOfWeek === 0 || dayOfWeek === 6) {
            th.classList.add('weekend');
        }
        headerRow.appendChild(th);
    });
    
    // Total column header
    const totalHeader = document.createElement('th');
    totalHeader.className = 'sticky-col-right';
    totalHeader.textContent = 'Total';
    headerRow.appendChild(totalHeader);
    
    table.appendChild(headerRow);
    
    // Ticket rows
    sortedTickets.forEach(ticket => {
        const row = document.createElement('tr');
        
        // Ticket column
        const ticketCell = document.createElement('td');
        ticketCell.className = 'sticky-col';
        ticketCell.textContent = `${ticket.key} - ${ticket.summary}`;
        row.appendChild(ticketCell);
        
        // Date columns
        dates.forEach(date => {
            const dateKey = formatDate(date);
            const cell = document.createElement('td');
            const hours = ticket.hours.get(dateKey) || 0;
            
            if (hours > 0) {
                cell.textContent = `${Math.round(hours)}h`;
            }
            
            // Highlight weekends
            const dayOfWeek = date.getDay();
            if (dayOfWeek === 0 || dayOfWeek === 6) {
                cell.classList.add('weekend');
            }
            
            row.appendChild(cell);
        });
        
        // Total column
        const totalCell = document.createElement('td');
        totalCell.className = 'sticky-col-right';
        totalCell.textContent = `${Math.round(ticket.total)}h`;
        totalCell.style.fontWeight = 'bold';
        row.appendChild(totalCell);
        
        table.appendChild(row);
    });
    
    // Total row
    const totalRow = document.createElement('tr');
    totalRow.style.fontWeight = 'bold';
    
    const totalLabelCell = document.createElement('td');
    totalLabelCell.className = 'sticky-col';
    totalLabelCell.textContent = 'Total Hours';
    totalRow.appendChild(totalLabelCell);
    
    dates.forEach(date => {
        const dateKey = formatDate(date);
        const cell = document.createElement('td');
        const total = dailyTotals.get(dateKey) || 0;
        
        if (total > 0) {
            cell.textContent = `${Math.round(total)}h`;
        }
        
        // Highlight weekends
        const dayOfWeek = date.getDay();
        if (dayOfWeek === 0 || dayOfWeek === 6) {
            cell.classList.add('weekend');
        }
        
        totalRow.appendChild(cell);
    });
    
    const grandTotalCell = document.createElement('td');
    grandTotalCell.className = 'sticky-col-right';
    grandTotalCell.textContent = `${Math.round(grandTotal)}h`;
    totalRow.appendChild(grandTotalCell);
    
    table.appendChild(totalRow);
    
    // Clear and add table
    tableDiv.innerHTML = '';
    tableDiv.appendChild(table);
}

function formatDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}
