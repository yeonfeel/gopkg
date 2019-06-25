package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"

	"github.com/yeonfeel/gopkg/er"
	"github.com/yeonfeel/gopkg/logger"
)

var log = logger.Get("config")

// Config is a configuration of cli
type Config struct {
	Value          interface{}
	Filename       string
	SetFromEnvvars func(value interface{})
}

// New returns a new Config
func New(filename string, value interface{}, setFromEnvvars func(interface{})) *Config {
	conf := &Config{
		Value:          value,
		Filename:       filename,
		SetFromEnvvars: setFromEnvvars,
	}

	if e := conf.Read(); e != nil {
		log.Error(er.Error(e, "Failed to read a config file"))
		return nil
	}

	return conf
}

func (c *Config) createIfNotExist() (err error) {
	// create file if not exist
	configFile := filepath.Join(configDir(), c.Filename)
	_, err = os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			var file *os.File
			file, err = os.Create(configFile)
			if err != nil {
				return err
			}
			file.Close()
		} else {
			return err
		}
	}
	return nil
}

// Load set from environment values
func (c *Config) Load() {
	c.SetFromEnvvars(c.Value)
}

// Read configuration from config file
func (c *Config) Read() (err error) {
	err = c.createIfNotExist()
	if err != nil {
		return err
	}

	file, err := os.Open(c.getConfigFile())
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return yaml.Unmarshal(b, c.Value)
	}

	return nil
}

// Write a configuration file
func (c *Config) Write() error {
	b, err := yaml.Marshal(c.Value)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.getConfigFile(), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) getConfigFile() string {
	configFile := filepath.Join(configDir(), c.Filename)
	if _, err := os.Stat(configFile); err == nil {
		return configFile
	}

	return ""
}

func configDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		return home
	}
	return os.Getenv("HOME")
	//path := "/etc/kubewatch"a
	//if _, err := os.Stat(path); os.IsNotExist(err) {
	//	os.Mkdir(path, 755)
	//}
	//return path
}
