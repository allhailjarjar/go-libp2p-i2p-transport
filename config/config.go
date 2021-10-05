package config

type Config struct {
	//i2p config stuff...
}

type Configurator func(*Config) error

// ConfMerge Merges different configs, starting at the first ending at the last.
func Merge(cs ...Configurator) Configurator {
	return func(c *Config) error {
		for _, v := range cs {
			if err := v(c); err != nil {
				return err
			}
		}
		return nil
	}
}
