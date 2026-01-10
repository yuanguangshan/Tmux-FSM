# tmux-fsm

A flexible, configuration-driven FSM (Finite State Machine) based keybinding system for tmux, designed for efficient terminal navigation and pane management.

## ✨ Features

### 🏗️ **Modular Architecture**
- **FSM Engine**: Core state machine logic with layer and timeout support
- **Configurable Keymap**: YAML-based configuration for all key bindings
- **UI Abstraction**: Status line integration for state display
- **Neovim Integration**: Bidirectional mode synchronization

### 🎛️ **Configuration-Driven**
- **YAML Keymap**: Externalized key bindings for easy customization
- **State Management**: Multiple FSM states with hints and transitions
- **Layer Support**: Temporary sub-modes with timeout capabilities
- **Validation**: Built-in configuration validation

### ⌨️ **Advanced Key Handling**
- **Prefix Keys**: Support for chorded key sequences (e.g., `g` + `h` for goto-left)
- **Timeout Management**: Automatic state reset after timeout
- **Action Mapping**: Semantic actions mapped to key sequences

### 🔄 **Neovim Integration**
- **Mode Synchronization**: Automatic exit from FSM when Neovim enters insert mode
- **Bidirectional Communication**: FSM and Neovim can notify each other of mode changes

## 🚀 Installation

### Prerequisites
- Go 1.24+
- tmux 3.3+

### Installation Steps

1. Clone the repository:
```bash
git clone https://github.com/tmux-plugins/tmux-fsm.git ~/.tmux/plugins/tmux-fsm
```

2. Add to your `~/.tmux.conf`:
```tmux
set -g @plugin '~/.tmux/plugins/tmux-fsm'
```

3. Install TPM (Tmux Plugin Manager) if not already installed:
```bash
git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
```

4. Press `Prefix + I` to install plugins

## ⚙️ Configuration

### Keymap Configuration

The keymap is defined in `keymap.yaml` using a YAML format:

```yaml
# NOTE:
# layer + action should not exist simultaneously
# layer transition does not trigger action
states:
  NAV:
    hint: "h/j/k/l move · 0/$ line · g goto · : cmd · q quit"
    keys:
      h: { action: "move_left" }
      j: { action: "move_down" }
      k: { action: "move_up" }
      l: { action: "move_right" }
      "0": { action: "goto_line_start" }
      "$": { action: "goto_line_end" }
      g: { layer: "GOTO", timeout_ms: 800 }
      ":": { action: "prompt" }
      q: { action: "exit" }
      Escape: { action: "exit" }

  GOTO:
    hint: "h far-left · l far-right · g top · G bottom"
    keys:
      h: { action: "far_left" }
      l: { action: "far_right" }
      g: { action: "goto_top" }
      G: { action: "goto_bottom" }
      q: { action: "exit" }
      Escape: { action: "exit" }
```

### Keymap Structure

- **states**: Define different FSM states
- **hint**: Display text shown in status line
- **keys**: Key-to-action mappings
  - `action`: Direct action to execute
  - `layer`: Switch to another FSM state
  - `timeout_ms`: Timeout for layer transitions

## 🎮 Usage

### Basic Commands

- `Prefix + f`: Enter FSM mode (typically bound in tmux config)
- `Escape` or `q`: Exit FSM mode
- `C-c`: Exit FSM mode (alternative)

### Key Bindings

In FSM mode, the following keys are available based on your configuration:

- `h/j/k/l`: Move between panes
- `0/$`: Move to line start/end
- `g` + `h/l/g/G`: GOTO layer for extended navigation
- `:`: Command prompt
- `q` or `Escape`: Exit FSM

### Layer System

The FSM supports a layer system for temporary modes:
- Press `g` to enter GOTO layer
- Within GOTO layer, `h/l/g/G` have different meanings
- After 800ms timeout, returns to NAV state automatically

## 🔧 Commands

The `tmux-fsm` binary supports the following commands:

- `-enter`: Enter FSM mode
- `-exit`: Exit FSM mode
- `-key <key>`: Dispatch key to FSM
- `-reload`: Reload keymap configuration
- `-server`: Run as daemon server
- `-config <path>`: Path to keymap configuration file
- `-debug`: Enable debug logging
- `-help`: Show help information

Additional functionality is accessible through the server protocol:
- `__SHUTDOWN__`: Stop the running daemon
- `__PING__`: Check server status
- `__CLEAR_STATE__`: Reset FSM state

## 🏗️ Architecture

### Core Components

1. **Engine**: Manages FSM state, transitions, and key dispatch (`fsm/engine.go`)
2. **Keymap**: Handles YAML configuration loading and validation (`config.go`)
3. **Kernel**: Central processing unit coordinating components (`kernel/`)
4. **Weaver**: System composition and fact resolution (`weaver/`)
5. **Backend**: Tmux command execution layer (`backend/`)
6. **UI**: Status line integration for state display (`fsm/ui_stub.go`)

### Design Principles

- **Configuration-Driven**: Behavior defined in external YAML files
- **State Isolation**: Each FSM state is independent
- **UI Decoupling**: UI and logic are separated
- **Extensibility**: Easy to add new actions and states
- **Modularity**: Components are loosely coupled with clear interfaces

## 🧪 Testing

Run the full test suite:
```bash
go test ./...
```

Or run specific tests:
```bash
bash test_fsm.sh
```

The test suite covers:
- Build process
- Keymap validation
- Server mode
- FSM lifecycle
- Component integration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

This project builds upon the concepts of finite state machines applied to terminal navigation, with inspiration from modal editors like Vim.

## 🚀 Key Optimizations

### 1. **Engine Lifecycle Management**
- **Single Engine Instance**: Only one Engine instance exists throughout the application lifecycle
- **Explicit Initialization**: Engine is initialized with explicit Keymap injection
- **State Isolation**: Engine state is no longer tied to global variables

### 2. **Configuration Validation**
- **Runtime Validation**: Keymap configurations are validated at load time
- **Error Detection**: Invalid layer references are caught immediately
- **Fail-Fast**: Invalid configurations prevent system startup

### 3. **Dependency Injection**
- **Explicit Dependencies**: Keymap is injected into Engine rather than accessed globally
- **Testability**: Components can be instantiated with different configurations
- **Decoupling**: Reduced coupling between components

### 4. **UI/FSM Decoupling**
- **Status Line Integration**: Uses tmux status variables to display FSM state
- **Clean Interfaces**: UI only displays when FSM state is valid
- **Separation of Concerns**: UI updates are triggered by FSM state changes

### 5. **Layer and Timeout Management**
- **Proper State Transitions**: Layer transitions are handled correctly
- **Timeout Handling**: Goroutines properly capture Engine instance
- **Automatic Reset**: States automatically reset after timeout

### 6. **Neovim Integration**
- **Mode Synchronization**: Automatic exit from FSM when Neovim enters insert mode
- **Clean Communication**: Proper handling of mode changes
- **Non-Interference**: Avoids sending unwanted keystrokes to Neovim