<div align="left">
  <img src="https://img.shields.io/badge/golang-242424?logo=go&style=for-the-badge&logoColor=00ADD8"/>
  <img src="https://img.shields.io/badge/github-242424?logo=github&style=for-the-badge&logoColor=ffffff"/>
  <img src="https://img.shields.io/badge/gitlab-242424?logo=gitlab&style=for-the-badge&logoColor=FC6D26"/>
  <img src="https://img.shields.io/badge/docker-242424?logo=docker&style=for-the-badge&logoColor=2496ED"/>
  <img src="https://img.shields.io/badge/kubernetes-242424?logo=kubernetes&style=for-the-badge&logoColor=326CE5"/>
</div>

# Forge 🔥

Forge is an automated Docker container deployment tool designed for VPS environments.
It monitors Git repositories and automatically redeploys containers upon detecting new commits.

## Table of Contents 🗃️

- [Configuration File](#configuration-file-)
- [How to Run](#how-to-run-)
- [Logs](#logs-)
- [Developer Notes](#dev-)
- [Project Goals](#project-goals-)

## Prerequisites

Before running Forge, make sure you have Docker installed on your VPS and a Dockerfile in your repository.  
Your projects **must** contain `Dockerfile` in its root directory

## Configuration File 🔧

The configuration file can be generated using the following command:

```sh
forge -g -d <directory>
```

### Note on Directory Paths

- The directory parameter can be either global or relative.
- Both `~/` and `./` are supported in the input.
- In the configuration file, use only the `global path` or paths that start with `~/`.

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

- `text` (default)
- `json`

You can specify the log type in the Forge arguments as follows:

```sh
forge -fmt <type> -d <directory>
```

### Log Levels

Log level can be specified as ENV Variable (case insensitive). Here are the options:

- `INFO`
- `DEBUG`
- `WARN`
- `ERROR` - default option

## Dev 🧑🏻‍💻

This section contains notes for developers about the project structure.

### Patterns Used

This project utilises the Observer pattern. It monitors the repository, and once changes are detected, the Observer notifies the Deployer module, which handles the rest of the process.  

## Project Goals 🎯

The final architecture of the project is designed as follows:

```mermaid
---
config:
  theme: redux
  layout: elk
---
flowchart TD
 subgraph INotifier["INotifier"]
        S("Slack")
        T("Telegram")
        E("Email")
        N["Notify()"]
  end
 subgraph IDeployer["IDeployer"]
        K("Deploy Kubernetes")
        DC("Deploy Docker compose")
        DF("Deploy Dockerfile")
        DP["Deploy()"]
  end
    S --> N
    T --> N
    E --> N
    F[("Forge")] --> O(["Observer"])
    O --> S & T & E & D("Selector")
    D --> K & DC & DF
    K --> DP
    DC --> DP
    DF --> DP
```

Currently, the `Deploy()` functionality using Dockerfile has been implemented. Future plans include extending the functionality to incorporate additional features and deployment methods.
