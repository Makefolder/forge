<div align="left">
  <img src="https://img.shields.io/badge/golang-242424?logo=go&style=for-the-badge&logoColor=00ADD8"/>
  <img src="https://img.shields.io/badge/github-242424?logo=github&style=for-the-badge&logoColor=ffffff"/>
  <img src="https://img.shields.io/badge/gitlab-242424?logo=gitlab&style=for-the-badge&logoColor=FC6D26"/>
</div>

# Forge 🔥

Forge is an automated Docker container deployment tool designed for VPS environments. It monitors Git repositories and automatically redeploys containers upon detecting new commits.

## Table of Contents 🗃️
- [Configuration File](#configuration-file-)
- [How to Run](#how-to-run-)
- [Logs](#logs-)
- [Developer Notes](#dev-)

## Configuration File ⚙️

The configuration file can be generated using the following command:

```sh
forge -g -d <directory>
```

The `directory` parameter can be either global or relative. Note that only `~/` is supported; `./` is not supported.

## How to Run 🐉

After generating and configuring the `.yaml` file, you can start Forge with the following command:

```sh
ACCESS_TOKEN="your-access-token" forge -d <directory>
```

The `access-token` must be from one of the supported platforms:
- GitHub
- GitLab

## Logs 🪵

Logs are stored in files within the specified directory. Valid types for logs are:
- `json`
- `text` (default)

You can specify the log type in the Forge arguments as follows:

```sh
forge -fmt <type> -d <directory>
```

## Dev 🧑🏻‍💻

This section contains notes for developers about the project structure.

### Patterns Used 🖇️

This project utilises the Observer pattern. It monitors the repository, and once changes are detected, the Observer notifies the Deployer module, which handles the rest of the process.
