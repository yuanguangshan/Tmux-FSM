package fsm

// currentMode holds the current FSM mode (NAV, INSERT, etc.)
var currentMode string

// SetMode sets the current FSM mode
func SetMode(mode string) {
	currentMode = mode
}

// RefreshUI triggers UI refresh without manipulating tmux key table
// This function only triggers status bar refresh (actual UI logic remains in main/updateStatusBar)
func RefreshUI() {
	// UI update is delegated to main/updateStatusBar
	// This function exists as a hook point for future UI abstraction
}
