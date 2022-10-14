package util

import "github.com/spf13/viper"

type Config struct {
	NewsAPIKey    string `mapstructure:"NEWS_API_KEY"`
	RedisUrl      string `mapstructure:"REDIS_URL"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	HeadlinesUrl  string `mapstructure:"HEADLINES_URL"`
	QueryUrl      string `mapstructure:"QUERY_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
