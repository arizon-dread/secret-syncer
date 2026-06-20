package conf

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/arizon-dread/secret-syncer/internal/conf/models"
	"github.com/spf13/viper"
)

var (
	config  *models.Config
	once    sync.Once
	initErr error
)

// GetConfig reads the config once from disk, reads env vars and then returns the Config instance.
// If already read, the read instance is returned immediately
func GetConfig() (*models.Config, error) {
	once.Do(func() {
		configPath := os.Getenv("configPath")
		if configPath == "" {
			configPath = "./conf"
		}
		v := viper.New()

		v.SetConfigName("config")
		v.AddConfigPath(configPath)
		err := v.ReadInConfig()
		if err != nil {
			log.Printf("unable to read configFile")
		}

		v.AutomaticEnv()

		config = &models.Config{}
		err = v.Unmarshal(config)
		if err != nil {
			initErr = fmt.Errorf("unable to unmarshal config into go struct, quitting")
		}
		if config.SecretServer.BaseURL == "" {
			initErr = fmt.Errorf("secret server baseURL was not set, please set the config variables to run secret-syncer")
		}
	})
	return config, initErr
}
