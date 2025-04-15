package cfg

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"os"
	"path/filepath"
	"strings"
)

// Validatable is config which could be validate
type Validatable interface {
	Validate() error
}

var vp *viper.Viper

// Init config file
func Init(name, file string, config Validatable) error {
	vp = viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_")))

	vp.SetConfigName(name) // name of config file (without extension)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "failed to get current dir")
	}

	vp.AddConfigPath(dir)
	vp.SetEnvPrefix("COMMERCE")
	vp.AutomaticEnv()

	if file != "" { // enable ability to specify config file via flag
		vp.SetConfigFile(file)
	}

	if err := vp.ReadInConfig(); err != nil {
		return errors.Wrap(err, "viper: failed to read config")
	}

	if err = vp.Unmarshal(config); err != nil {
		return errors.Wrap(err, "failed to unmarshal config to obj")
	}

	return config.Validate()
}

// ValidateConfigs validates configs
func ValidateConfigs(configs ...Validatable) error {
	var errs []string

	for _, c := range configs {
		if err := c.Validate(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}
