package confy

// ReadConfig reads configuration file and parses it depending on tags in structure provided.
// Then it reads and parses
//
// Example:
//
//	type ConfigDatabase struct {
//		Port     string `yaml:"port" env:"PORT" env-default:"5432"`
//		Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
//		Name     string `yaml:"name" env:"NAME" env-default:"postgres"`
//		User     string `yaml:"user" env:"USER" env-default:"user"`
//		Password string `yaml:"password" env:"PASSWORD"`
//	}
//
//	var cfg ConfigDatabase
//
//	err := cleanenv.ReadConfig("config.yml", &cfg)
//	if err != nil {
//	    ...
//	}
func Get(path string, cfg interface{}) error {
	err := parseFile(path, cfg)
	if err != nil {
		return err
	}

	return nil
}

// ReadEnv reads environment variables into the structure.
func GetEnv(cfg interface{}) error {
	return readEnvVars(cfg, false)
}

// UpdateEnv rereads (updates) environment variables in the structure.
func UpdateEnv(cfg interface{}) error {
	return readEnvVars(cfg, true)
}
