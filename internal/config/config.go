package config

import (
	"strings"

	"github.com/spf13/viper"
	pkgconfig "github.com/zqr233qr/story-trim/pkg/config"
)

type Config struct {
	FileStorage FileStorageConfig   `mapstructure:"file_storage"`
	LLM         LLM                 `mapstructure:"llm"`
	Database    DatabaseConfig      `mapstructure:"database"`
	Memory      MemoryConfig        `mapstructure:"memory"`
	Log         pkgconfig.LogConfig `mapstructure:"log"`
	Auth        AuthConfig          `mapstructure:"auth"`
	Parser      ParserConfig        `mapstructure:"parser"`
}

type ParserConfig struct {
	Version int          `mapstructure:"version" json:"version"`
	Rules   []ParserRule `mapstructure:"rules" json:"rules"`
}

type ParserRule struct {
	Name    string `mapstructure:"name" json:"name"`
	Pattern string `mapstructure:"pattern" json:"pattern"`
	Weight  int    `mapstructure:"weight" json:"weight"`
}

type FileStorageConfig struct {
	UploadDir string `mapstructure:"upload_dir"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

type MemoryConfig struct {
	SummaryLimit         int `mapstructure:"summary_limit"`
	EncyclopediaInterval int `mapstructure:"encyclopedia_interval"`
	MockStreamSpeed      int `mapstructure:"mock_stream_speed"`
}

type DatabaseConfig struct {
	Source string `mapstructure:"source"`
}

type LLM struct {
	Use       string               `mapstructure:"use"`
	LLMConfig map[string]LLMConfig `mapstructure:"llm_config"`
}

type LLMConfig struct {
	BaseURL     string  `mapstructure:"base_url"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	InputPrice  float64 `mapstructure:"input_price"`  // 输入价格(百万token)元
	OutputPrice float64 `mapstructure:"output_price"` // 输出价格(百万token)元
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
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
