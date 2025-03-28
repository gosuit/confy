# Confy

Confy is Go library for reading configuration settings from environment variables and YAML or other files. It is based on <a href="https://github.com/ilyakaznacheev/cleanenv">CleanEnv</a>. It provides a simple way to manage application configurations while ensuring that the configurations are valid.

## Installation

```zsh
go get github.com/gosuit/confy
```

## Features
 
- **Files Support**: Load configuration settings from files. Types are supported: 
  - **YAML**
  - **JSON**
  - **TOML**
- **Environment Variables**: Override configuration settings with environment variables.
- **Env Names Expand**: Set the names of environment variables through files to get the values
- **Validation**: Validate configuration structures.

## Usage

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
