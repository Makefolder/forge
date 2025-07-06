# Forge

Forge - Automated Docker container deployment tool for VPS environments.  
Monitors Git repositories and redeploys containers on new commits.

## Configuration file

The configuration file can be generated with the following command

```sh
forge -g -d <directory>
```

The `directory` can either be global or relative (only `~/`; `./` is unsupported).

## How to run

Once you have generated and configured the `.yaml` file, it's about the time
to get the thing up & running which can be done with this command:

```sh
ACCESS_TOKEN="your-access-token" forge -d <directory>
```

The `access-token` must be of one of the supported platforms which are:

- GitHub
- GitLab

## Logs

Logs are being stored in files in the specified directory.
There're two types of logs:

- TEXT
- JSON

The type can be specified in the forge arguments like this:

```sh
forge -fmt <type> -d <directory>
```

Valid types:

- `json`
- `text` - the default one

## Dev

This section contains developer notes about the project structure.

### Patterns used

This project uses "Observer pattern". It is monitoring repository. Once changes found,
the Observer notifies deployer structure which does everything else.

