package confy

// Get reads config from file and override values with environment variables.
func Get(path string, cfg any) error {
	err := parseFile(path, cfg)
	if err != nil {
		return err
	}

	return nil
}

// GetEnv reads environment variables into the structure.
func GetEnv(cfg any) error {
	return readEnvVars(cfg, false)
}
