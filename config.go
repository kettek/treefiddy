package main

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Actions   Actions
	Binds     Binds
	Modes     Modes
	Shortcuts Shortcuts
	UseMouse  bool    `yaml:"use_mouse"`
	Systems   Systems `yaml:"systems"`
}

type Systems struct {
	JavaScriptPlugins bool `yaml:"javascript_plugins"`
}

type Actions struct {
	Click           string
	Enter           string
	PostEdictEdicts map[string]string `yaml:"post_edict_edicts"`
}

type Shortcuts []Shortcut

type Shortcut struct {
	Edict     string
	Arguments []string
	Keyword   string
}

type Binds []Bind

type Bind struct {
	Edict     string   `yaml:"edict,omitempty"`
	Arguments []string `yaml:"arguments,omitempty"`
	Key       int      `yaml:"key,omitempty"`
	Rune      string   `yaml:"rune,omitempty"`
}

type Modes map[string]Mode

type Mode struct {
	Rune  string `yaml:"rune,omitempty"`
	Binds Binds  `yaml:"binds,omitempty"`
}

func (m *Mode) GetBind(r rune) *Bind {
	for _, b := range m.Binds {
		if rune(b.Rune[0]) == r {
			return &b
		}
	}
	return nil
}

var (
	configDir  string
	configPath string
)

var defaultConfig = Config{
	Actions: Actions{
		Click: "edit",
		Enter: "edit",
		PostEdictEdicts: map[string]string{
			"create": "edit",
		},
	},
	Shortcuts: []Shortcut{
		{
			Edict:   "edit",
			Keyword: "e",
		},
		{
			Edict:   "open",
			Keyword: "o",
		},
	},
	Binds: []Bind{
		{
			Edict: "quit",
			Rune:  "q",
		},
		{
			Edict: "edit",
			Rune:  "e",
		},
		{
			Edict: "open",
			Rune:  "o",
		},
	},
	UseMouse: true,
	Systems: Systems{
		JavaScriptPlugins: true,
	},
}

func ensureConfig() (cfg Config) {
	cdir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configDir = filepath.Join(cdir, "treefiddy")
	configPath = filepath.Join(configDir, "config.yaml")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		panic(err)
	}
	if _, err := os.Stat(configPath); err != nil && !os.IsNotExist(err) {
		panic(err)
	} else if err != nil && os.IsNotExist(err) {
		bytes, err := yaml.Marshal(defaultConfig)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(configPath, bytes, 0o644); err != nil {
			panic(err)
		}
	}

	// Read it up.
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
