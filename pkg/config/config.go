package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const appConfigDirName = "pm"

// PlannerAnchors holds user-configurable HH:MM defaults for the four planner slots.
type PlannerAnchors struct {
	In1  string `json:"in1" yaml:"in1,omitempty"`
	Out1 string `json:"out1" yaml:"out1,omitempty"`
	In2  string `json:"in2" yaml:"in2,omitempty"`
	Out2 string `json:"out2" yaml:"out2,omitempty"`
}

// File mirrors ~/.config/pm/config.yaml keys used by pm-cli.
type File struct {
	Email         string          `json:"email" yaml:"email"`
	Password      string          `json:"password" yaml:"password"`
	CacheTTLHours int             `json:"cache_ttl_hours,omitempty" yaml:"cache_ttl_hours,omitempty"`
	Planner       *PlannerAnchors `json:"planner,omitempty" yaml:"planner,omitempty"`
}

// DefaultDir returns the platform-native config directory for PM.
// On Linux this remains $XDG_CONFIG_HOME/pm, or $HOME/.config/pm by default.
func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appConfigDirName), nil
}

// DefaultPath returns the platform-native config.yaml path.
func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// ResolvePath returns override if non-empty, otherwise DefaultPath().
func ResolvePath(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return filepath.Clean(override), nil
	}
	return DefaultPath()
}

func applyEnvAndViperBase() {
	viper.SetEnvPrefix("PM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

// Init configures viper the same way the CLI historically did (YAML + PM_ env vars).
func Init(cfgFileOverride string) error {
	viper.SetConfigType("yaml")
	if strings.TrimSpace(cfgFileOverride) != "" {
		viper.SetConfigFile(filepath.Clean(cfgFileOverride))
	} else {
		viper.SetConfigName("config")
		if dir, err := DefaultDir(); err == nil {
			viper.AddConfigPath(dir)
		}
	}

	applyEnvAndViperBase()
	_ = viper.ReadInConfig()
	return nil
}

// Read loads config from disk. If the file is missing, returns an empty File and nil error.
func Read(cfgFileOverride string) (*File, error) {
	path, err := ResolvePath(cfgFileOverride)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &File{}, nil
		}
		return nil, err
	}
	var f File
	if len(b) > 0 {
		if err := yaml.Unmarshal(b, &f); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}
	return &f, nil
}

// Save writes YAML atomically with tight permissions when credentials are stored.
func Save(cfgFileOverride string, f *File) error {
	if f != nil && f.Planner != nil && f.Planner.anySet() {
		anchors, err := mergePlannerAnchors(f, true)
		if err != nil {
			return err
		}
		if err := ValidatePlannerAnchors(anchors); err != nil {
			return fmt.Errorf("planner: %w", err)
		}
	}
	path, err := ResolvePath(cfgFileOverride)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	out, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".pm-config-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_, writeErr := tmp.Write(out)
	closeErr := tmp.Close()
	if writeErr != nil {
		_ = os.Remove(tmpPath)
		return writeErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	applyEnvAndViperBase()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)
	return viper.ReadInConfig()
}
