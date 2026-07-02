package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultBindAddress = "0.0.0.0:9000"
	defaultPort        = 9000
	defaultTickRate    = 20
	defaultMaxPlayers  = 1000
	defaultMaxRooms    = 100
	defaultLogLevel    = "debug"
)

type Config struct {
	BindAddress string `yaml:"bind_address"`
	Port        int    `yaml:"port"`
	TickRate    int    `yaml:"tick_rate"`
	MaxPlayers  int    `yaml:"max_players"`
	MaxRooms    int    `yaml:"max_rooms"`
	LogLevel    string `yaml:"log_level"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		BindAddress: defaultBindAddress,
		Port:        defaultPort,
		TickRate:    defaultTickRate,
		MaxPlayers:  defaultMaxPlayers,
		MaxRooms:    defaultMaxRooms,
		LogLevel:    defaultLogLevel,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config %q: %w", path, err)
	}

	cfg.applyEnvironment()
	cfg.applyDefaults()

	return cfg, nil
}

func (c *Config) applyEnvironment() {
	c.BindAddress = parseEnvString("SERVER_BIND_ADDRESS", c.BindAddress)
	c.Port = parseEnvInt("SERVER_PORT", c.Port)
	c.TickRate = parseEnvInt("SERVER_TICK_RATE", c.TickRate)
	c.MaxPlayers = parseEnvInt("SERVER_MAX_PLAYERS", c.MaxPlayers)
	c.MaxRooms = parseEnvInt("SERVER_MAX_ROOMS", c.MaxRooms)
	c.LogLevel = parseEnvString("LOG_LEVEL", c.LogLevel)
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.BindAddress) == "" {
		c.BindAddress = fmt.Sprintf("0.0.0.0:%d", c.Port)
	}

	if c.Port == 0 {
		c.Port = defaultPort
	}

	if c.TickRate == 0 {
		c.TickRate = defaultTickRate
	}

	if c.MaxPlayers == 0 {
		c.MaxPlayers = defaultMaxPlayers
	}

	if c.MaxRooms == 0 {
		c.MaxRooms = defaultMaxRooms
	}

	if strings.TrimSpace(c.LogLevel) == "" {
		c.LogLevel = defaultLogLevel
	}
}

func parseEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}
	return fallback
}

func parseEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
