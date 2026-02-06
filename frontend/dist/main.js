/**
 * JIRA Work Calendar – Frontend
 *
 * Uses Wails bindings (window.go.main.App) to call backend methods:
 * - GetConfig, SaveConfig: Jira configuration
 * - GetUserInfo: current user (avatar, timezone, etc.)
 * - GetWorkLogs(startDate, endDate): work logs for date range (YYYY-MM-DD)
 * - AddWorklog(issueKey, comment, started, timeSpentSeconds): create work log
 *
 * Screens: config → date (home with calendar widget) → calendar (table) + add-worklog modal.
 * State: calendarState (date picker), currentCalendarDateRange (for refresh), worklog form and date picker.
 */
let app;

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
    
    // Setup Add Worklog button
    document.getElementById('add-worklog-btn').addEventListener('click', () => {
        showAddWorklogModal();
    });
    
    // Setup modal close buttons
    document.getElementById('close-worklog-modal').addEventListener('click', () => {
        hideAddWorklogModal();
    });
    
    document.getElementById('cancel-worklog-btn').addEventListener('click', () => {
        hideAddWorklogModal();
    });
    
    // Setup worklog form
    setupWorklogForm();
}

// Worklog date picker state
let worklogDatePickerState = {
    currentDate: new Date(),
    selectedDate: null
};

function showAddWorklogModal() {
    const modal = document.getElementById('add-worklog-modal');
    modal.classList.remove('hidden');
    
    // Clear any previous errors
    hideWorklogError();
    
    // Set default date to today
    const today = new Date();
    worklogDatePickerState.currentDate = new Date(today);
    worklogDatePickerState.selectedDate = new Date(today);
    
    // Update date picker display
    updateWorklogDateDisplay();
    initWorklogDatePicker();
    
    // Set default time to 9:30 AM
    document.getElementById('worklog-time').value = '09:30';
    
    // Clear form
    document.getElementById('worklog-issue-key').value = '';
    document.getElementById('worklog-duration').value = '';
    document.getElementById('worklog-comment').value = '';
}

function showWorklogError(message) {
    const errorDiv = document.getElementById('worklog-error');
    if (errorDiv) {
        errorDiv.textContent = message;
        errorDiv.classList.remove('hidden');
        // Scroll to error
        errorDiv.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
}

function hideWorklogError() {
    const errorDiv = document.getElementById('worklog-error');
    if (errorDiv) {
        errorDiv.textContent = '';
        errorDiv.classList.add('hidden');
    }
}

let worklogDatePickerCloseHandler = null;
let worklogDatePickerInitialized = false;

function initWorklogDatePicker() {
    // Only initialize once
    if (worklogDatePickerInitialized) {
        updateWorklogDatePicker();
        return;
    }
    
    // Setup date picker button
    const pickerBtn = document.getElementById('worklog-date-picker-btn');
    const pickerPopup = document.getElementById('worklog-date-picker');
    const displayInput = document.getElementById('worklog-date-display');
    
    const togglePicker = () => {
        const isHidden = pickerPopup.classList.contains('hidden');
        pickerPopup.classList.toggle('hidden');
        if (isHidden) {
            // Position the popup relative to the button
            const btnRect = pickerBtn.getBoundingClientRect();
            const wrapperRect = pickerBtn.closest('.date-picker-wrapper').getBoundingClientRect();
            pickerPopup.style.top = `${btnRect.bottom + window.scrollY + 8}px`;
            pickerPopup.style.left = `${wrapperRect.left + window.scrollX}px`;
            pickerPopup.style.width = `${Math.max(300, wrapperRect.width)}px`;
            
            updateWorklogDatePicker();
            // Remove old handler if exists
            if (worklogDatePickerCloseHandler) {
                document.removeEventListener('click', worklogDatePickerCloseHandler);
            }
            // Add click outside handler
            worklogDatePickerCloseHandler = (e) => {
                if (!pickerPopup.contains(e.target) && e.target !== pickerBtn && e.target !== displayInput) {
                    pickerPopup.classList.add('hidden');
                    document.removeEventListener('click', worklogDatePickerCloseHandler);
                    worklogDatePickerCloseHandler = null;
                }
            };
            setTimeout(() => {
                document.addEventListener('click', worklogDatePickerCloseHandler);
            }, 0);
        } else {
            // Remove handler when closing
            if (worklogDatePickerCloseHandler) {
                document.removeEventListener('click', worklogDatePickerCloseHandler);
                worklogDatePickerCloseHandler = null;
            }
        }
    };
    
    pickerBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        togglePicker();
    });
    
    displayInput.addEventListener('click', () => {
        if (pickerPopup.classList.contains('hidden')) {
            togglePicker();
        }
    });
    
    // Setup navigation
    document.getElementById('worklog-date-prev').addEventListener('click', (e) => {
        e.stopPropagation();
        worklogDatePickerState.currentDate.setMonth(worklogDatePickerState.currentDate.getMonth() - 1);
        updateWorklogDatePicker();
    });
    
    document.getElementById('worklog-date-next').addEventListener('click', (e) => {
        e.stopPropagation();
        worklogDatePickerState.currentDate.setMonth(worklogDatePickerState.currentDate.getMonth() + 1);
        updateWorklogDatePicker();
    });
    
    worklogDatePickerInitialized = true;
    updateWorklogDatePicker();
}

function updateWorklogDatePicker() {
    const grid = document.getElementById('worklog-date-grid');
    grid.innerHTML = '';
    
    const year = worklogDatePickerState.currentDate.getFullYear();
    const month = worklogDatePickerState.currentDate.getMonth();
    
    // Update month/year display
    const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
        'July', 'August', 'September', 'October', 'November', 'December'];
    document.getElementById('worklog-date-month-year').textContent = 
        `${monthNames[month]} ${year}`;
    
    // Get first day of month and number of days
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startingDayOfWeek = firstDay.getDay();
    
    // Add empty cells for days before month starts
    for (let i = 0; i < startingDayOfWeek; i++) {
        const emptyCell = document.createElement('div');
        emptyCell.className = 'date-picker-day empty';
        grid.appendChild(emptyCell);
    }
    
    // Add cells for each day of the month
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    for (let day = 1; day <= daysInMonth; day++) {
        const date = new Date(year, month, day);
        const cell = document.createElement('div');
        cell.className = 'date-picker-day';
        cell.textContent = day;
        
        // Check if today
        if (date.getTime() === today.getTime()) {
            cell.classList.add('today');
        }
        
        // Check if selected
        if (worklogDatePickerState.selectedDate) {
            const selected = new Date(worklogDatePickerState.selectedDate);
            selected.setHours(0, 0, 0, 0);
            if (date.getTime() === selected.getTime()) {
                cell.classList.add('selected');
            }
        }
        
        // Check if weekend
        const dayOfWeek = date.getDay();
        if (dayOfWeek === 0 || dayOfWeek === 6) {
            cell.classList.add('weekend');
        }
        
        // Add click handler
        cell.addEventListener('click', (e) => {
            e.stopPropagation();
            worklogDatePickerState.selectedDate = new Date(date);
            updateWorklogDateDisplay();
            const pickerPopup = document.getElementById('worklog-date-picker');
            pickerPopup.classList.add('hidden');
            // Remove click outside handler
            if (worklogDatePickerCloseHandler) {
                document.removeEventListener('click', worklogDatePickerCloseHandler);
                worklogDatePickerCloseHandler = null;
            }
        });
        
        grid.appendChild(cell);
    }
    
    // Add empty cells at the end to always have 42 cells (6 rows)
    const cellsAdded = startingDayOfWeek + daysInMonth;
    const remainingCells = 42 - cellsAdded;
    for (let i = 0; i < remainingCells; i++) {
        const emptyCell = document.createElement('div');
        emptyCell.className = 'date-picker-day empty';
        grid.appendChild(emptyCell);
    }
}

function updateWorklogDateDisplay() {
    if (worklogDatePickerState.selectedDate) {
        const date = worklogDatePickerState.selectedDate;
        document.getElementById('worklog-date-display').value = formatDateDisplay(date);
        document.getElementById('worklog-date').value = formatDate(date);
    } else {
        // If no date selected, clear the display
        document.getElementById('worklog-date-display').value = '';
        document.getElementById('worklog-date').value = '';
    }
}

function hideAddWorklogModal() {
    const modal = document.getElementById('add-worklog-modal');
    modal.classList.add('hidden');
}

function setupWorklogForm() {
    const form = document.getElementById('add-worklog-form');
    
    // Clear errors when user starts typing in any field
    const formInputs = form.querySelectorAll('input, textarea');
    formInputs.forEach(input => {
        input.addEventListener('input', () => {
            hideWorklogError();
        });
    });
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const issueKey = document.getElementById('worklog-issue-key').value.trim().toUpperCase();
        const date = document.getElementById('worklog-date').value;
        const time = document.getElementById('worklog-time').value;
        const duration = document.getElementById('worklog-duration').value.trim();
        const comment = document.getElementById('worklog-comment').value.trim();
        
        // Parse duration to seconds
        const timeSpentSeconds = parseDuration(duration);
        if (timeSpentSeconds === 0) {
            alert('Invalid duration format. Please use formats like: 2h 30m, 150m, 1.5h');
            return;
        }
        
        // Format date/time for Jira API (ISO 8601 with timezone)
        const dateTime = new Date(`${date}T${time}`);
        const started = formatDateTimeForJira(dateTime);
        
        // Clear any previous errors
        hideWorklogError();
        
        try {
            await app.AddWorklog(issueKey, comment, started, timeSpentSeconds);
            
            // Close modal immediately - do this synchronously
            const modal = document.getElementById('add-worklog-modal');
            if (modal) {
                modal.classList.add('hidden');
            }
            
            // Clear form
            form.reset();
            
            // Refresh the calendar to show the new worklog (async, don't wait)
            refreshCalendar().catch(err => {
                console.error('Calendar refresh failed:', err);
            });
            
            // Show success message after a short delay to allow UI to update
            setTimeout(() => {
                alert('Worklog added successfully!');
            }, 50);
        } catch (err) {
            // Show error in the form
            showWorklogError(err.message || err.toString());
        }
    });
}

function parseDuration(duration) {
    if (!duration) return 0;
    
    duration = duration.toLowerCase().trim();
    let totalSeconds = 0;
    
    // Parse hours and minutes (e.g., "2h 30m", "2h30m", "2.5h")
    const hourMatch = duration.match(/(\d+(?:\.\d+)?)\s*h/);
    if (hourMatch) {
        totalSeconds += parseInt(parseFloat(hourMatch[1]) * 3600);
    }
    
    const minuteMatch = duration.match(/(\d+)\s*m/);
    if (minuteMatch) {
        totalSeconds += parseInt(minuteMatch[1]) * 60;
    }
    
    // If no units found, try parsing as just a number (assume minutes)
    if (totalSeconds === 0) {
        const numMatch = duration.match(/^(\d+(?:\.\d+)?)$/);
        if (numMatch) {
            // If it's a decimal, assume hours; otherwise assume minutes
            if (duration.includes('.')) {
                totalSeconds = parseInt(parseFloat(numMatch[1]) * 3600);
            } else {
                totalSeconds = parseInt(numMatch[1]) * 60;
            }
        }
    }
    
    return totalSeconds;
}

function formatDateTimeForJira(date) {
    // Format as ISO 8601: "2024-01-15T10:00:00.000+0000"
    // For simplicity, we'll use UTC offset 0000
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    
    return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}.000+0000`;
}

function setupSettingsButtons() {
    const settingsBtnDate = document.getElementById('settings-btn-date');
    const refreshBtnCalendar = document.getElementById('refresh-btn-calendar');
    
    if (settingsBtnDate) {
        settingsBtnDate.addEventListener('click', () => {
            showConfigScreen();
        });
    }
    
    if (refreshBtnCalendar) {
        refreshBtnCalendar.addEventListener('click', async () => {
            await refreshCalendar();
        });
    }
}

async function refreshCalendar() {
    // Only refresh if we're on the calendar screen and have a date range
    const calendarScreen = document.getElementById('calendar-screen');
    if (!calendarScreen || calendarScreen.classList.contains('hidden')) {
        return;
    }
    
    if (!currentCalendarDateRange.startDate || !currentCalendarDateRange.endDate) {
        console.log('Cannot refresh: date range not set');
        return;
    }
    
    // Show loading animation
    const refreshBtn = document.getElementById('refresh-btn-calendar');
    if (refreshBtn) {
        refreshBtn.classList.add('refreshing');
    }
    
    try {
        const workLogs = await app.GetWorkLogs(currentCalendarDateRange.startDate, currentCalendarDateRange.endDate);
        showCalendarScreen(workLogs, currentCalendarDateRange.startDate, currentCalendarDateRange.endDate);
    } catch (err) {
        console.error('Failed to refresh calendar:', err);
        alert('Failed to refresh calendar: ' + err);
    } finally {
        // Remove loading animation
        if (refreshBtn) {
            refreshBtn.classList.remove('refreshing');
        }
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

// Store current date range for refresh functionality
let currentCalendarDateRange = {
    startDate: null,
    endDate: null
};

function showCalendarScreen(workLogs, startDate, endDate) {
    document.getElementById('config-screen').classList.add('hidden');
    document.getElementById('date-screen').classList.add('hidden');
    document.getElementById('calendar-screen').classList.remove('hidden');
    
    // Store current date range for refresh
    currentCalendarDateRange.startDate = startDate;
    currentCalendarDateRange.endDate = endDate;
    
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
