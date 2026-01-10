package kernel

import (
	"fmt"
	"log"
	"tmux-fsm/backend"
)

// Execute a decision made by the kernel.
func (k *Kernel) Execute(decision *Decision) {
	if decision == nil {
		log.Println("kernel.Execute called with nil decision")
		return
	}

	if k.Exec == nil {
		log.Println("kernel.Execute called with nil executor")
		return
	}

	switch decision.Kind {
	case DecisionNone, DecisionLegacy:
		return // Do nothing intentionally.

	case DecisionIntent:
		// This is a full-fledged intent from the grammar.
		// Process it via the standard execution path.
		if decision.Intent == nil {
			log.Println("DecisionIntent without an intent")
			return
		}
		_ = k.Exec.Process(decision.Intent)

	case DecisionFSM:
		// This is a simple FSM action that should be executed.
		// Instead of calling FSM's RunAction directly (which violates architecture),
		// we execute the action through the proper backend.
		if decision.Action != "" {
			executeFSMAction(decision.Action)
		}

	default:
		log.Printf("Unknown or unhandled decision kind: %v", decision.Kind)
	}
}

// executeFSMAction 执行 FSM 动作，通过适当的后端
func executeFSMAction(action string) {
	// 根据动作类型执行相应的 tmux 命令
	cmd := getTmuxCommandForAction(action)
	if cmd != "" {
		err := backend.GlobalBackend.ExecRaw(cmd)
		if err != nil {
			log.Printf("Error executing tmux command '%s': %v", cmd, err)
		}
	} else {
		log.Printf("Unknown FSM action: %s", action)
	}
}

// getTmuxCommandForAction 将 FSM 动作映射到相应的 tmux 命令
func getTmuxCommandForAction(action string) string {
	switch action {
	case "pane_left":
		return "select-pane -L"
	case "pane_right":
		return "select-pane -R"
	case "pane_up":
		return "select-pane -U"
	case "pane_down":
		return "select-pane -D"
	case "next_pane":
		return "select-pane -t :.+"
	case "prev_pane":
		return "select-pane -t :.-"
	case "far_left":
		return "select-pane -t :.0"
	case "far_right":
		return "select-pane -t :.$"
	case "goto_top":
		return "select-pane -t :.0"
	case "goto_bottom":
		return "select-pane -t :.$"
	case "goto_line_start":
		return "send-keys -t . Home"
	case "goto_line_end":
		return "send-keys -t . End"
	case "move_left":
		return "send-keys -t . Left"
	case "move_right":
		return "send-keys -t . Right"
	case "move_up":
		return "send-keys -t . Up"
	case "move_down":
		return "send-keys -t . Down"
	case "exit":
		// 特殊处理：退出 FSM
		go func() {
			// 延迟执行，避免在执行过程中修改状态
			fmt.Println("Exiting FSM...")
		}()
		return ""
	case "prompt":
		return "command-prompt"
	default:
		return ""
	}
}
