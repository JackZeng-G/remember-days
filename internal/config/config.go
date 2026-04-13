package config

import (
    "fmt"

    "github.com/caarlos0/env/v10"
)

// Config 应用配置
type Config struct {
    Port          int    `env:"PORT" envDefault:"8080"`
    DataDir       string `env:"DATA_DIR" envDefault:"data"`
    LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
    CSRFKey       string `env:"CSRF_KEY" envDefault:"change-me-in-production-csrf"`
    SessionSecret string `env:"SESSION_SECRET" envDefault:"change-me-in-production-session"`
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }
    return cfg, nil
}
