# tmux-fsm

A flexible, configuration-driven FSM (Finite State Machine) based keybinding system for tmux, designed for efficient terminal navigation and pane management.

## ✨ Features

### 🏗️ **Modular Architecture**
- **FSM Engine**: Core state machine logic with layer and timeout support
- **Configurable Keymap**: YAML-based configuration for all key bindings
- **UI Abstraction**: Pluggable UI backends (popup, status, etc.)
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
- tmux 3.3+ (for popup UI support)

### Installation Steps

1. Clone the repository:
```bash
git clone https://github.com/your-username/tmux-fsm.git ~/.tmux/plugins/tmux-fsm
```

2. Add to your `~/.tmux.conf`:
```tmux
set -g @plugin 'your-username/tmux-fsm'
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
    hint: "h/j/k/l move · g goto · : cmd · q quit"
    keys:
      h: { action: pane_left }
      j: { action: pane_down }
      k: { action: pane_up }
      l: { action: pane_right }
      g: { layer: GOTO, timeout_ms: 800 }
      ":": { action: prompt }
      q: { action: exit }
      Escape: { action: exit }

  GOTO:
    hint: "h far-left · l far-right · g top · G bottom"
    keys:
      h: { action: far_left }
      l: { action: far_right }
      g: { action: goto_top }
      G: { action: goto_bottom }
      q: { action: exit }
      Escape: { action: exit }
```

### Keymap Structure

- **states**: Define different FSM states
- **hint**: Display text shown in UI
- **keys**: Key-to-action mappings
  - `action`: Direct action to execute
  - `layer`: Switch to another FSM state
  - `timeout_ms`: Timeout for layer transitions

## 🎮 Usage

### Basic Commands

- `Prefix + f`: Enter FSM mode
- `Escape` or `q`: Exit FSM mode
- `C-c`: Exit FSM mode (alternative)

### Key Bindings

In FSM mode, the following keys are available based on your configuration:

- `h/j/k/l`: Move between panes
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
- `-nvim-mode <mode>`: Handle Neovim mode change
- `-reload`: Reload keymap configuration
- `-server`: Run as daemon server
- `-stop`: Stop the running daemon
- `-ui-show`: Show UI
- `-ui-hide`: Hide UI
- `-config <path>`: Path to keymap configuration file

## 🏗️ Architecture

### Core Components

1. **Engine**: Manages FSM state, transitions, and key dispatch
2. **Keymap**: Handles YAML configuration loading and validation
3. **UI**: Abstract interface for different UI backends
4. **Neovim**: Integration for bidirectional mode synchronization

### Design Principles

- **Configuration-Driven**: Behavior defined in external YAML files
- **State Isolation**: Each FSM state is independent
- **UI Decoupling**: UI and logic are completely separated
- **Extensibility**: Easy to add new actions and states

## 🧪 Testing

Run the full test suite:
```bash
bash test_fsm.sh
```

The test suite covers:
- Build process
- Keymap validation
- Server mode
- FSM lifecycle
- UI functionality

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
- **No Fallback Logic**: UI components no longer guess FSM state
- **Clean Interfaces**: UI only displays when FSM state is valid
- **State Provider**: Abstract interface for UI to access FSM state

### 5. **Layer and Timeout Management**
- **Proper State Transitions**: Layer transitions are handled correctly
- **Timeout Handling**: Goroutines properly capture Engine instance
- **Automatic Reset**: States automatically reset after timeout

### 6. **Neovim Integration**
- **Mode Synchronization**: Automatic exit from FSM when Neovim enters insert mode
- **Clean Communication**: Proper handling of mode changes
- **Non-Interference**: Avoids sending unwanted keystrokes to Neovim