package confy

func Read(to any, from string) error {
	fileData, fileTag, err := getFileData(from)
	if err != nil {
		return err
	}

	out, err := prepareStruct(to)
	if err != nil {
		return err
	}

	err = processStruct(out, fileData, fileTag)
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

	out, err := prepareStruct(to)
	if err != nil {
		return err
	}

	err = processStruct(out, fileData, fileTag)
	if err != nil {
		return err
	}

	return nil
}

func ReadEnv(to any) error {
	fileData := make(map[string]any)

	out, err := prepareStruct(to)
	if err != nil {
		return err
	}

	err = processStruct(out, fileData, confyTag)
	if err != nil {
		return err
	}

	return nil
}
