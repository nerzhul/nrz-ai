package config

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds all configuration options
type Config struct {
	// Audio & Speech
	WhisperModel string `mapstructure:"whisper_model" yaml:"whisper_model"`
	Language     string `mapstructure:"language" yaml:"language"`
	AudioSource  string `mapstructure:"audio_source" yaml:"audio_source"`

	// Wake Word
	WakeWordEnabled bool   `mapstructure:"wake_word_enabled" yaml:"wake_word_enabled"`
	WakeWord        string `mapstructure:"wake_word" yaml:"wake_word"`
	WakeWordSound   string `mapstructure:"wake_word_sound" yaml:"wake_word_sound"`

	// AI Configuration
	AIEnabled    bool   `mapstructure:"ai_enabled" yaml:"ai_enabled"`
	OllamaURL    string `mapstructure:"ollama_url" yaml:"ollama_url"`
	OllamaModel  string `mapstructure:"ollama_model" yaml:"ollama_model"`
	SystemPrompt string `mapstructure:"system_prompt" yaml:"system_prompt"`

	// Advanced
	LogLevel   string `mapstructure:"log_level" yaml:"log_level"`
	MaxHistory int    `mapstructure:"max_history" yaml:"max_history"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		// Audio & Speech defaults
		WhisperModel: "./models/ggml-large-v3.bin",
		Language:     "fr",
		AudioSource:  "default",

		// Wake Word defaults
		WakeWordEnabled: false,
		WakeWord:        "Jack",
		WakeWordSound:   "./sounds/pop-cartoon-328167.mp3",

		// AI defaults
		AIEnabled:    false,
		OllamaURL:    "http://localhost:11434",
		OllamaModel:  "llama3.2:3b",
		SystemPrompt: "Tu es un assistant vocal français intelligent et concis. Réponds brièvement et naturellement.",

		// Advanced defaults
		LogLevel:   "info",
		MaxHistory: 10,
	}
}

// LoadConfig loads configuration from YAML file following XDG Base Directory Specification
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// XDG Base Directory Specification
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logrus.WithError(err).Warn("Failed to get user home directory, using current directory")
		} else {
			configHome = filepath.Join(homeDir, ".config")
		}
	}

	configDir := filepath.Join(configHome, "nrz-ai")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Environment variable support
	viper.SetEnvPrefix("NRZ_AI")
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Info("No config file found, creating default configuration")
			if err := createDefaultConfigFile(configDir); err != nil {
				logrus.WithError(err).Warn("Failed to create default config file")
			}
		} else {
			logrus.WithError(err).Warn("Error reading config file")
		}
	} else {
		logrus.WithField("file", viper.ConfigFileUsed()).Info("Using config file")
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SaveConfig saves the current configuration to the XDG config directory
func (c *Config) SaveConfig() error {
	// Get XDG config directory
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	configDir := filepath.Join(configHome, "nrz-ai")
	configFile := filepath.Join(configDir, "config.yaml")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Set configuration in viper
	viper.Set("whisper_model", c.WhisperModel)
	viper.Set("language", c.Language)
	viper.Set("audio_source", c.AudioSource)
	viper.Set("wake_word_enabled", c.WakeWordEnabled)
	viper.Set("wake_word", c.WakeWord)
	viper.Set("wake_word_sound", c.WakeWordSound)
	viper.Set("ai_enabled", c.AIEnabled)
	viper.Set("ollama_url", c.OllamaURL)
	viper.Set("ollama_model", c.OllamaModel)
	viper.Set("system_prompt", c.SystemPrompt)
	viper.Set("log_level", c.LogLevel)
	viper.Set("max_history", c.MaxHistory)

	// Write configuration file
	return viper.WriteConfigAs(configFile)
}

// createDefaultConfigFile creates a default configuration file
func createDefaultConfigFile(configDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")
	
	// Don't overwrite existing file
	if _, err := os.Stat(configFile); err == nil {
		return nil
	}

	defaultConfig := DefaultConfig()
	
	// Set default values in viper
	viper.Set("whisper_model", defaultConfig.WhisperModel)
	viper.Set("language", defaultConfig.Language)
	viper.Set("audio_source", defaultConfig.AudioSource)
	viper.Set("wake_word_enabled", defaultConfig.WakeWordEnabled)
	viper.Set("wake_word", defaultConfig.WakeWord)
	viper.Set("wake_word_sound", defaultConfig.WakeWordSound)
	viper.Set("ai_enabled", defaultConfig.AIEnabled)
	viper.Set("ollama_url", defaultConfig.OllamaURL)
	viper.Set("ollama_model", defaultConfig.OllamaModel)
	viper.Set("system_prompt", defaultConfig.SystemPrompt)
	viper.Set("log_level", defaultConfig.LogLevel)
	viper.Set("max_history", defaultConfig.MaxHistory)

	return viper.WriteConfigAs(configFile)
}