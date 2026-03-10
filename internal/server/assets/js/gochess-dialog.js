// ============================================
// Custom Dialog System
// ============================================
// Provides custom alert and confirm dialogs
// to replace standard browser dialogs

/**
 * Show a custom alert dialog
 * @param {string} message - The message to display
 * @param {string} type - The type of alert: 'info', 'success', 'warning', 'error'
 * @param {function} callback - Optional callback function to execute after dialog closes
 */
function showDialog(message, type = 'info', callback = null) {
    const dialog = document.getElementById('customDialog');
    const dialogMessage = document.getElementById('dialogMessage');
    const dialogIcon = document.getElementById('dialogIcon');
    const dialogOkBtn = document.getElementById('dialogOkBtn');
    const dialogCancelBtn = document.getElementById('dialogCancelBtn');
    
    // Set message
    dialogMessage.textContent = message;
    
    // Set icon based on type (no background color changes)
    const icons = {
        'info': '💬',
        'success': '✅',
        'warning': '⚠️',
        'error': '❌',
        'game-over': '🏁',
        'timeout': '⏰'
    };
    
    dialogIcon.textContent = icons[type] || icons['info'];
    
    // Show only OK button for alerts
    dialogOkBtn.style.display = 'inline-block';
    dialogCancelBtn.style.display = 'none';
    
    // Show dialog
    dialog.style.display = 'flex';
    
    // Focus on OK button
    setTimeout(() => dialogOkBtn.focus(), 100);
    
    // Handle OK button click
    const handleOk = () => {
        dialog.style.display = 'none';
        dialogOkBtn.removeEventListener('click', handleOk);
        if (callback) callback();
    };
    
    dialogOkBtn.addEventListener('click', handleOk);
    
    // Handle Enter key
    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            handleOk();
            document.removeEventListener('keydown', handleKeyPress);
        }
    };
    document.addEventListener('keydown', handleKeyPress);
}

/**
 * Show a custom confirm dialog
 * @param {string} message - The message to display
 * @param {function} onConfirm - Callback function to execute if user confirms
 * @param {function} onCancel - Optional callback function to execute if user cancels
 */
function showConfirm(message, onConfirm, onCancel = null) {
    const dialog = document.getElementById('customDialog');
    const dialogMessage = document.getElementById('dialogMessage');
    const dialogIcon = document.getElementById('dialogIcon');
    const dialogOkBtn = document.getElementById('dialogOkBtn');
    const dialogCancelBtn = document.getElementById('dialogCancelBtn');
    
    // Set message
    dialogMessage.textContent = message;
    
    // Set icon for confirmation
    dialogIcon.textContent = '❓';
    
    // Show both OK and Cancel buttons
    dialogOkBtn.style.display = 'inline-block';
    dialogCancelBtn.style.display = 'inline-block';
    dialogOkBtn.textContent = 'Yes';
    
    // Show dialog
    dialog.style.display = 'flex';
    
    // Focus on Cancel button (safer default)
    setTimeout(() => dialogCancelBtn.focus(), 100);
    
    // Handle OK button click
    const handleOk = () => {
        dialog.style.display = 'none';
        cleanup();
        if (onConfirm) onConfirm();
    };
    
    // Handle Cancel button click
    const handleCancel = () => {
        dialog.style.display = 'none';
        cleanup();
        if (onCancel) onCancel();
    };
    
    // Handle Enter key (confirm) and Escape key (cancel)
    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            handleOk();
        } else if (e.key === 'Escape') {
            handleCancel();
        }
    };
    
    const cleanup = () => {
        dialogOkBtn.removeEventListener('click', handleOk);
        dialogCancelBtn.removeEventListener('click', handleCancel);
        document.removeEventListener('keydown', handleKeyPress);
        dialogOkBtn.textContent = 'OK';
    };
    
    dialogOkBtn.addEventListener('click', handleOk);
    dialogCancelBtn.addEventListener('click', handleCancel);
    document.addEventListener('keydown', handleKeyPress);
}

/**
 * Show a game over dialog with special styling
 * @param {string} message - The game over message
 * @param {function} callback - Optional callback function
 */
function showGameOver(message, callback = null) {
    showDialog(message, 'game-over', callback);
}

/**
 * Show a timeout dialog with special styling
 * @param {string} message - The timeout message
 * @param {function} callback - Optional callback function
 */
function showTimeout(message, callback = null) {
    showDialog(message, 'timeout', callback);
}
