package config

import (
	"io/ioutil"

	"github.com/tsaikd/KDGoLib/errutil"
	yaml "gopkg.in/yaml.v2"
)

// errors
var (
	ErrUnmarshalYAMLConfig = errutil.NewFactory("Failed unmarshalling config in YAML format")
)

// Config contains all config
type Config struct {
	Interval uint64
	BigQuery struct {
		Mock      bool
		KeyFile   string
		ProjectID string
		DatasetID string
	}
	Memory struct {
		Enable bool
	}
	FileSystem struct {
		Paths []string
	}
	Docker struct {
		EndPoint string
		Stats    struct {
			Enable bool
		}
	}
}

// LoadFromFile load config from file path
func LoadFromFile(name string) (config Config, err error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return
	}
	return LoadFromYAML(data)
}

// LoadFromYAML load config from []byte in YAML format
func LoadFromYAML(data []byte) (config Config, err error) {
	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, ErrUnmarshalYAMLConfig.New(err)
	}
	return
}
