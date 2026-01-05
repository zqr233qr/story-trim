package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Memory   MemoryConfig   `mapstructure:"memory"`
	Protocol ProtocolConfig `mapstructure:"protocol"`
	Log      LogConfig      `mapstructure:"log"`
}

type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	TokenDuration int    `mapstructure:"token_duration"` // hours
}

type MemoryConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	SummaryLimit         int  `mapstructure:"summary_limit"`         // 每次携带的前章摘要数
	EncyclopediaInterval int  `mapstructure:"encyclopedia_interval"` // 百科更新频率
	ContextMode          int  `mapstructure:"context_mode"`          // 0:无, 1:摘要, 2:摘要+百科
	MockStreamSpeed      int  `mapstructure:"mock_stream_speed"`     // 拟真流延迟(ms)
}

type ProtocolConfig struct {
	BaseInstruction string `mapstructure:"base_instruction"` // 系统底层逻辑指令
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"` // sqlite or mysql
	Source string `mapstructure:"source"` // file path or dsn
}

type LLMConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	Timeout int    `mapstructure:"timeout"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json or console
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("llm.base_url", "https://api.deepseek.com")
	viper.SetDefault("llm.model", "deepseek-chat")
	viper.SetDefault("llm.timeout", 120)
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.source", "storytrim.db")
	viper.SetDefault("auth.jwt_secret", "storytrim-secret-key-default")
	viper.SetDefault("auth.token_duration", 24)
	
	// Memory 默认值
	viper.SetDefault("memory.enabled", true)
	viper.SetDefault("memory.summary_limit", 1)
	viper.SetDefault("memory.encyclopedia_interval", 50)
	viper.SetDefault("memory.context_mode", 2)
	viper.SetDefault("memory.mock_stream_speed", 35)

	// Protocol 默认值
	viper.SetDefault("protocol.base_instruction", "你是一个具备逻辑追踪能力的文学编辑。请参考提供的[全局背景]和[前情提要]进行精简。如果原文中出现了背景中提到的重要人物或设定，请务必保留相关描写。")

	viper.SetDefault("log.level", "info")
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