package fsm

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type KeyAction struct {
	Action    string `yaml:"action"`
	Layer     string `yaml:"layer"`
	TimeoutMs int    `yaml:"timeout_ms"`
}

type StateDef struct {
	Hint   string               `yaml:"hint"`
	Sticky bool                 `yaml:"sticky"` // If true, don't reset to NAV after action
	Keys   map[string]KeyAction `yaml:"keys"`
}

type Keymap struct {
	Initial string              `yaml:"initial"`
	States  map[string]StateDef `yaml:"states"`
}

// Validate 验证 keymap 配置的正确性
func (km *Keymap) Validate() error {
	for name, st := range km.States {
		for key, act := range st.Keys {
			if act.Layer != "" {
				if _, ok := km.States[act.Layer]; !ok {
					return fmt.Errorf("state %s references missing layer %s for key %s", name, act.Layer, key)
				}
			}
		}
	}
	return nil
}

func LoadKeymap(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var km Keymap
	if err := yaml.Unmarshal(b, &km); err != nil {
		return err
	}

	// 验证配置
	if err := km.Validate(); err != nil {
		return fmt.Errorf("invalid keymap: %w", err)
	}

	KM = km
	return nil
}

var (
	KM Keymap
)
