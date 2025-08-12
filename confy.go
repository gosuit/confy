package confy

func Read(to any, from string) error {
	fileData, fileTag, err := getFileData(from)
	if err != nil {
		return err
	}

	err = fillConfig(to, fileData, fileTag)
	if err != nil {
		return err
	}

	return nil
}

func ReadMany(to any, from ...string) error {
	fileData, fileTag, err := getMultipleFilesData(from)
	if err != nil {
		return err
	}

	err = fillConfig(to, fileData, fileTag)
	if err != nil {
		return err
	}

	return nil
}

func ReadEnv(to any) error {
	fileData := make(map[string]any)

	err := fillConfig(to, fileData, confyTag)
	if err != nil {
		return err
	}

	return nil
}
