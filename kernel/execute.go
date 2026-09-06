package kernel

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"tmux-fsm/backend"
	"tmux-fsm/fsm"
)

// Execute a decision made by the kernel.
// hctx 携带 RequestID/ActorID——ActorID 用于在边界注入 PaneID
// （Grammar 永不产生 PaneID），此前该值未传到这里，
// 导致 weaver 路径的 intent.PaneID 恒为空。
func (k *Kernel) Execute(hctx HandleContext, decision *Decision) {
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
		// 优先走上下文路径：执行器能拿到 RequestID/ActorID，
		// 并在边界完成 PaneID 注入；否则回退无上下文 Process。

		// PaneID 边界注入（Kernel 职责，与 legacy ProcessIntent 对齐）：
		// Grammar 永不产生 PaneID；ActorID 格式 "paneID|clientName"。
		if decision.Intent != nil && decision.Intent.PaneID == "" && hctx.ActorID != "" {
			if parts := strings.SplitN(hctx.ActorID, "|", 2); len(parts) > 0 {
				decision.Intent.PaneID = parts[0]
			}
		}

		if ctxExec, ok := k.Exec.(ContextualIntentExecutor); ok {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = ctxExec.ProcessWithContext(ctx, hctx, decision.Intent)
			// M1.2：意图已完整消化，清零 FSM 侧计数（单一来源在 Grammar 侧）
			if k.FSM != nil {
				k.FSM.ResetCount()
			}
			return
		}
		_ = k.Exec.Process(decision.Intent)

	case DecisionPassthrough:
		// M4-core：字面透传——把可打印字符原样送进目标 pane 的 shell
		// （用户在 FSM 导航模式下自由输入，未绑定键不再被吞）。
		// -H 十六进制形式对引号/分号等特殊字符也安全。
		paneID := ""
		if hctx.ActorID != "" {
			if parts := strings.SplitN(hctx.ActorID, "|", 2); len(parts) > 0 {
				paneID = parts[0]
			}
		}
		_ = backend.GlobalBackend.ExecRaw(fmt.Sprintf(
			"send-keys -t %s -H %02X", paneID, decision.PassthroughKey[0]))

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
	// M1.1 修复（P0-1）：exit 是生命周期动作，必须走 fsm.ExitFSM()
	// （解除 key-table + 持久化状态）。此前 getTmuxCommandForAction("exit")
	// 返回空串、stdout 又被守护进程丢弃，q 键形同虚设。
	if action == "exit" {
		fsm.ExitFSM()
		return
	}

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
	case "paste_from_buffer":
		// M4.3：y 系列复制进 tmux buffer 后，p 粘贴回 shell 行
		// （-p = bracketed paste，光标落在粘贴内容尾部）
		return "paste-buffer -p"
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
