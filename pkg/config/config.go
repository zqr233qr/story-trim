package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LLM      LLMConfig      `mapstructure:"llm"`
	Database DatabaseConfig `mapstructure:"database"`
	Memory   MemoryConfig   `mapstructure:"memory"`
	Protocol ProtocolConfig `mapstructure:"protocol"`
	Log      LogConfig      `mapstructure:"log"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // console, json
}

type MemoryConfig struct {
	SummaryLimit         int  `mapstructure:"summary_limit"`         // 每次携带的前章摘要数
	EncyclopediaInterval int  `mapstructure:"encyclopedia_interval"` // 百科更新频率
	MockStreamSpeed      int  `mapstructure:"mock_stream_speed"`     // 拟真流延迟(ms)
}

type ProtocolConfig struct {
	BaseInstruction string `mapstructure:"base_instruction"` // 系统底层逻辑指令
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
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	viper.SetDefault("llm.base_url", "https://api.deepseek.com")
	viper.SetDefault("llm.model", "deepseek-chat")
	viper.SetDefault("database.source", "storytrim.db")
	
	// Memory 默认值
	viper.SetDefault("memory.summary_limit", 1)
	viper.SetDefault("memory.encyclopedia_interval", 50)
	viper.SetDefault("memory.mock_stream_speed", 35)

	// Protocol 默认值
	viper.SetDefault("protocol.base_instruction", "你是一个文学处理助手。")

	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.format", "console")

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
