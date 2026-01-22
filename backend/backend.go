package backend

import (
	"os/exec"
	"strings"
)

// EventType represents tmux event types for subscription
type EventType string

const (
	// ClientSessionChanged fires when client changes session
	ClientSessionChanged EventType = "client-session-changed"
	// PaneFocusIn fires when pane gains focus
	PaneFocusIn EventType = "pane-focus-in"
	// ClientKeyTableChanged fires when client key table changes
	ClientKeyTableChanged EventType = "client-key-table-changed"
)

// Event represents a tmux event notification
type Event struct {
	Type      EventType
	Target    string
	Data      string
	Timestamp int64
}

// Backend interface defines operations that interact with tmux
type Backend interface {
	SetUserOption(option, value string) error
	UnsetUserOption(option string) error
	GetUserOption(option string) (string, error)
	GetCommandOutput(cmd string) (string, error)
	SwitchClientTable(clientName, tableName string) error
	RefreshClient(clientName string) error
	GetActivePane(clientName string) (string, error)
	ExecRaw(cmd string) error

	// Phase 5.1: Event subscription support
	Subscribe(events ...EventType) (<-chan Event, error)
	Unsubscribe(ch <-chan Event) error
}

// GlobalBackend is the active backend instance
// Phase 5.1: Uses TmuxBackend by default, ControlModeBackend when configured
var GlobalBackend Backend = &TmuxBackend{}

// SetControlMode sets the backend to use control mode for persistent connections
// Phase 5.1: Allows switching between exec.Command and tmux control mode
func SetControlMode(useControlMode bool) {
	if useControlMode {
		GlobalBackend = &ControlModeBackend{}
	} else {
		GlobalBackend = &TmuxBackend{}
	}
}

// IsUsingControlMode checks if control mode is currently active
func IsUsingControlMode() bool {
	_, isControlMode := GlobalBackend.(*ControlModeBackend)
	return isControlMode
}

// TmuxBackend implements Backend interface using tmux commands
type TmuxBackend struct{}

// SetUserOption sets a tmux user option
func (b *TmuxBackend) SetUserOption(option, value string) error {
	cmd := exec.Command("tmux", "set", "-g", option, value)
	return cmd.Run()
}

// SwitchClientTable switches the client to a specific key table
func (b *TmuxBackend) SwitchClientTable(clientName, tableName string) error {
	args := []string{"switch-client", "-T", tableName}
	if clientName != "" && clientName != "default" {
		args = append(args, "-t", clientName)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// RefreshClient refreshes the client display
func (b *TmuxBackend) RefreshClient(clientName string) error {
	args := []string{"refresh-client", "-S"}
	if clientName != "" && clientName != "default" {
		args = append(args, "-t", clientName)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// GetActivePane gets the active pane ID
func (b *TmuxBackend) GetActivePane(clientName string) (string, error) {
	var cmd *exec.Cmd
	if clientName != "" && clientName != "default" {
		cmd = exec.Command("tmux", "display-message", "-p", "-t", clientName, "#{pane_id}")
	} else {
		cmd = exec.Command("tmux", "display-message", "-p", "#{pane_id}")
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// UnsetUserOption unsets a tmux user option
func (b *TmuxBackend) UnsetUserOption(option string) error {
	cmd := exec.Command("tmux", "set", "-u", "-g", option)
	return cmd.Run()
}

// GetUserOption gets a tmux user option value
func (b *TmuxBackend) GetUserOption(option string) (string, error) {
	cmd := exec.Command("tmux", "show-option", "-gv", option)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetCommandOutput executes a tmux command and returns its output
func (b *TmuxBackend) GetCommandOutput(cmd string) (string, error) {
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return "", nil
	}
	execCmd := exec.Command("tmux", parts...)
	output, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ExecRaw executes a raw tmux command string
func (b *TmuxBackend) ExecRaw(cmd string) error {
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return nil
	}
	execCmd := exec.Command("tmux", parts...)
	return execCmd.Run()
}

// Subscribe creates event channel subscriptions for specified tmux events
// Phase 5.1: Persistent event listening without exec.Command
func (b *TmuxBackend) Subscribe(events ...EventType) (<-chan Event, error) {
	eventCh := make(chan Event, 100)
	return eventCh, nil
}

// Unsubscribe signals backend to stop sending events
// Phase 5.1: Caller owns channel, backend just stops sending
func (b *TmuxBackend) Unsubscribe(ch <-chan Event) error {
	return nil
}
