package conf

import (
	"log"
	"os"
	"sync"

	"github.com/arizon-dread/secret-syncer/internal/conf/models"
	"github.com/spf13/viper"
)

var (
	config *models.Config
	once   sync.Once
)

// GetConfig reads the config once from disk and env vars and then returns it.
// If already read, the read instance is returned immediately
func GetConfig() *models.Config {
	if config != nil {
		return config
	}
	once.Do(func() {
		configPath := os.Getenv("configPath")
		if configPath == "" {
			configPath = "./conf"
		}
		v := viper.New()

		v.SetConfigName("config")
		v.AddConfigPath(configPath)
		err := viper.ReadInConfig()
		if err != nil {
			log.Printf("unable to read configFile")
		}

		v.AutomaticEnv()

		config = &models.Config{}
		err = v.Unmarshal(config)
		if err != nil {
			panic("unable to unmarshal config into go struct, quitting")
		}
	})
	return config
}
