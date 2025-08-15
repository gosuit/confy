# Confy

Confy is a Go library designed to make working with the configuration of your application as flexible as possible, but at the same time as simple as possible.

## Installation

```zsh
go get github.com/gosuit/confy
```

## Features
 
- **Files Support**: Load configuration settings from files. Types are supported: 
  - **YAML**
  - **JSON**
  - **TOML**
  - **DOTENV**
- **Environment Variables**: Override configuration settings with environment variables.
- **Env Names Expand**: Set the names of environment variables through files to get the values
- **Multiple files**: Load configuration settings from multiple files.
- **Reader**: High-level interface for flexible management of reading sources

## Documentation

- [Simple example](docs/simple)
- [Environment override](docs/env-override)
- [Env names expand](docs/env-names-expand)
- [Environment only](docs/env-only)
- [Multiple files read](docs/multiple-files)
- [Directory read](docs/directory)
- [Reader](docs/reader)

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## Thanks

We would like to thank [Ilya Kaznacheev](https://github.com/ilyakaznacheev) for his [Clean Env](https://github.com/ilyakaznacheev/cleanenv) library. Confy did not use the CleanEnv source code, but the idea of Confy was born on the basis of the CleanEnv project and has a similar API. In fact, we wanted to create our own CleanEnv with useful additional functionality.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
