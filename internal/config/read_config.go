package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ReadConfig[T any](path string, config *T) (err error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		return err
	}

	defer func() {
		if clerr := f.Close(); clerr != nil {
			err = clerr
		}
	}()

	return yaml.NewDecoder(f).Decode(config)
}
