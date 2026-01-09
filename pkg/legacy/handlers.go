package legacy

import (
	"fmt"
	"strings"
	"time"
	"tmux-fsm/pkg/protocol"
)

// FSMState represents the state of the FSM
type FSMState struct {
	Mode                 string                 `json:"mode"`
	Operator             string                 `json:"operator"`
	Count                int                    `json:"count"`
	PendingKeys          string                 `json:"pending_keys"`
	Register             string                 `json:"register"`
	LastRepeatableAction map[string]interface{} `json:"last_repeatable_action"`
	UndoStack            []Transaction          `json:"undo_stack"`
	RedoStack            []Transaction          `json:"redo_stack"`
	LastUndoFailure      string                 `json:"last_undo_failure,omitempty"`
	LastUndoSafetyLevel  string                 `json:"last_undo_safety_level,omitempty"`
	AllowPartial         bool                   `json:"allow_partial"` // Phase 7: Explicit permission for fuzzy resolution
}

// Transaction represents a single transaction
type Transaction struct {
	ID               TransactionID  `json:"id"`
	Records          []ActionRecord `json:"records"`
	CreatedAt        time.Time      `json:"created_at"`
	Applied          bool           `json:"applied"`
	Skipped          bool           `json:"skipped"`
	SafetyLevel      string         `json:"safety_level,omitempty"`       // exact, fuzzy
	PreSnapshotHash  string         `json:"pre_snapshot_hash,omitempty"`  // Phase 8: World state before transaction
	PostSnapshotHash string         `json:"post_snapshot_hash,omitempty"` // Phase 8: World state after transaction
}

type TransactionID uint64

type ActionRecord = protocol.ActionRecord

// ProcessKey handles key presses that are not handled by the FSM.
// It updates the FSM state and returns the action string to be executed.
func ProcessKey(state *FSMState, key string) string {
	if key == "Escape" || key == "C-c" {
		// Reset FSM state on escape/cancel
		state.Count = 0
		state.Operator = ""
		state.PendingKeys = ""
		// fsm.Reset() // Note: fsm package dependency would need to be imported
		return ""
	}

	// Check for count prefix
	if count, ok := isDigit(key); ok {
		if state.Count == 0 { // If no previous count, start accumulating
			state.Count = count
		} else { // Append digit to existing count
			state.Count = state.Count*10 + count
		}
		state.PendingKeys = fmt.Sprintf("%d", state.Count)
		return "" // Key handled as count, wait for next key
	}

	// If we have a count and received a motion
	if state.Count > 0 {
		// If the key is a motion
		if isMotion(key) {
			// Store motion for operator
			state.Operator = key // This is a simplification. Operator + Motion logic is complex.
			state.PendingKeys = fmt.Sprintf("%d%s", state.Count, key)
			// We need to capture this operator+motion for repeat
			state.LastRepeatableAction = map[string]interface{}{
				"action": state.Operator + "_" + state.Operator, // Placeholder, need proper motion mapping
				"count":  state.Count,
			}
			state.Count = 0 // Reset count after operator+motion
			return ""       // Key handled as count, wait for next key
		} else {
			// If it's not a motion, reset count and process key normally
			// e.g. 3j then 'd' is correct, but 3j then 'i' is wrong.
			// For simplicity, we reset count and let the key be processed as usual.
			// A more robust FSM would handle operator pending state better.

			// Rethink: if count is pending, and key is not a motion,
			// maybe it's an operator for the count? e.g. 3i<char>
			// For now, simpler reset.
			action := state.Operator + "_" + key
			state.Count = 0
			state.Operator = ""
			state.PendingKeys = ""
			return action
		}
	}

	// If we have an operator pending (e.g. 'd', 'c')
	if state.Operator != "" {
		// Check if key is a motion
		if isMotion(key) {
			action := state.Operator + "_" + key
			state.PendingKeys = fmt.Sprintf("%s%s", state.Operator, key)
			state.LastRepeatableAction = map[string]interface{}{
				"action": action,
				"count":  state.Count,
			}
			state.Count = 0 // Reset count after operator+motion
			state.Operator = ""
			return action
		} else {
			// Operator pending, but key is not a motion. Reset.
			// e.g., 'd' then 'a' (delete around word). This is wrong.
			// If it's another operator, e.g., 'd' then 'd' -> dd
			if key == state.Operator { // e.g., 'd' then 'd'
				action := state.Operator + "_" + state.Operator
				state.LastRepeatableAction = map[string]interface{}{
					"action": action,
					"count":  state.Count,
				}
				state.Count = 0
				state.Operator = ""
				return action
			}
			// Reset operator and pending keys, process key normally
			state.Count = 0
			state.Operator = ""
			state.PendingKeys = ""
			// Fallthrough to process key normally
		}
	}

	// If key is a known operator (d, c, y, etc.)
	if isOperator(key) {
		state.Operator = key
		state.PendingKeys = key
		state.Count = 0 // Reset count when a new operator is pressed
		return ""
	}

	// If key is insert mode related
	if strings.HasPrefix(key, "insert") || strings.HasPrefix(key, "replace") || strings.HasPrefix(key, "toggle") || strings.HasPrefix(key, "paste") {
		state.PendingKeys = ""
		state.Operator = ""
		state.Count = 0
		return key
	}

	// If key is a motion
	if isMotion(key) {
		// If no operator is pending, just move
		state.PendingKeys = key
		return "move_" + key
	}

	// Clear pending keys if not recognized and not part of an operator/motion sequence
	if state.PendingKeys != "" && !strings.HasPrefix(key, "move_") { // Allow move_ actions to be appended
		state.PendingKeys = ""
		state.Operator = ""
		state.Count = 0
	}

	// Handle special keys like Esc or Ctrl+C
	if key == "Escape" || key == "C-c" {
		state.Count = 0
		state.Operator = ""
		state.PendingKeys = ""
		// fsm.Reset() // Reset FSM state
		return ""
	}

	// For any other key, return it as is (or handle specific ones like search)
	// Add explicit handling for search keys if not caught by FSM
	if strings.HasPrefix(key, "search_") {
		state.PendingKeys = key
		return key
	}

	// If key is unknown, clear state
	state.Count = 0
	state.Operator = ""
	state.PendingKeys = ""

	return ""
}

func isOperator(key string) bool {
	switch key {
	case "d", "c", "y":
		return true
	default:
		return false
	}
}

func isMotion(key string) bool {
	switch key {
	case "h", "j", "k", "l", "w", "b", "e", "0", "$", "gg", "G", // basic motions
		"up", "down", "left", "right", "word_forward", "word_backward", "end_of_word", // mapped motions
		"start_of_line", "end_of_line", "start_of_file", "end_of_file":
		return true
	default:
		return false
	}
}

func isDigit(s string) (int, bool) {
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return int(s[0] - '0'), true
	}
	return 0, false
}

// ProcessKeyLegacy processes a key using legacy logic
func ProcessKeyLegacy(key string) string {
	// This is a placeholder that should be implemented based on the actual legacy logic
	// For now, we'll return an empty string to avoid compilation errors
	// This function should be implemented with the actual legacy processing logic
	// For now, we'll return an empty string to avoid compilation errors
	// This function should be implemented with the actual legacy processing logic
	// We'll implement it based on the original processKey function from main.go
	state := &FSMState{} // Create a temporary state for processing
	return ProcessKey(state, key)
}
