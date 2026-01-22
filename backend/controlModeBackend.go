package backend

// controlModeBackend provides stub implementation for Tmux Control Mode
// Phase 5.1: Full control mode implementation requires tmux control protocol
// This stub provides interface but uses exec.Command as a placeholder
// Future: Replace exec.Command with tmux control mode (-C socket) for persistent connections

import (
	"os/exec"
	"strings"
)

// ControlModeBackend implements Backend using tmux commands with control mode placeholder
type ControlModeBackend struct{}

// NewControlMode creates a control mode backend instance
func NewControlModeBackend() *ControlModeBackend {
	return &ControlModeBackend{}
}

// SetUserOption sets a tmux user option
func (b *ControlModeBackend) SetUserOption(option, value string) error {
	cmd := exec.Command("tmux", "set", "-g", option, value)
	return cmd.Run()
}

// UnsetUserOption unsets a tmux user option
func (b *ControlModeBackend) UnsetUserOption(option string) error {
	cmd := exec.Command("tmux", "set", "-u", "-g", option)
	return cmd.Run()
}

// GetUserOption gets a tmux user option value
func (b *ControlModeBackend) GetUserOption(option string) (string, error) {
	cmd := exec.Command("tmux", "show-option", "-gv", option)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// SwitchClientTable switches client to a specific key table
func (b *ControlModeBackend) SwitchClientTable(clientName, tableName string) error {
	args := []string{"switch-client", "-T", tableName}
	if clientName != "" && clientName != "default" {
		args = append(args, "-t", clientName)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// RefreshClient refreshes client display
func (b *ControlModeBackend) RefreshClient(clientName string) error {
	args := []string{"refresh-client", "-S"}
	if clientName != "" && clientName != "default" {
		args = append(args, "-t", clientName)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// GetActivePane gets the active pane ID
func (b *ControlModeBackend) GetActivePane(clientName string) (string, error) {
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

// GetCommandOutput executes a tmux command and returns its output
func (b *ControlModeBackend) GetCommandOutput(cmd string) (string, error) {
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
func (b *ControlModeBackend) ExecRaw(cmd string) error {
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return nil
	}
	execCmd := exec.Command("tmux", parts...)
	return execCmd.Run()
}

// Subscribe creates event channel subscriptions for specified tmux events
// Phase 5.2: Stub implementation - full control mode required for persistent connections
// Future: Replace exec.Command with tmux control mode (-C socket) for event-driven subscriptions
func (b *ControlModeBackend) Subscribe(events ...EventType) (<-chan Event, error) {
	eventCh := make(chan Event, 100)
	return eventCh, nil
}

// Unsubscribe signals backend to stop sending events
// Phase 5.1: Caller owns channel, backend just stops sending
func (b *ControlModeBackend) Unsubscribe(ch <-chan Event) error {
	return nil
}
