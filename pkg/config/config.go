package config

import (
	"embed"
	"strings"

	"github.com/spf13/viper"
)

//go:embed *.tmpl
var Templates embed.FS

type Config struct {
	FileStorage FileStorageConfig `mapstructure:"file_storage"`
	LLM         LLMConfig         `mapstructure:"llm"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Memory      MemoryConfig      `mapstructure:"memory"`
	Log         LogConfig         `mapstructure:"log"`
	Auth        AuthConfig        `mapstructure:"auth"`
}

type FileStorageConfig struct {
	UploadDir string `mapstructure:"upload_dir"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // console, json
}

type MemoryConfig struct {
	SummaryLimit         int `mapstructure:"summary_limit"`         // 每次携带的前章摘要数
	EncyclopediaInterval int `mapstructure:"encyclopedia_interval"` // 百科更新频率
	MockStreamSpeed      int `mapstructure:"mock_stream_speed"`     // 拟真流延迟(ms)
}

type DatabaseConfig struct {
	Source string `mapstructure:"source"` // file path or dsn
}

type LLMConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
