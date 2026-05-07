# aux4/config

Configuration utility for aux4. Manage YAML and JSON configuration files with get, set, and merge operations.

## Installation

```bash
aux4 aux4 pkger install aux4/config
```

## Usage

### Create Configuration File
It can be a YAML or JSON file.

config.yaml
```yaml
config:
  dev:
    host: localhost
    port: 3000
  prod:
    host: aux4.io
    port: 80
```

config.json
```json
{
  "config": {
    "dev": {
      "host": "localhost",
      "port": 3000
    },
    "prod": {
      "host": "aux4.io",
      "port": 80
    }
  }
}
```

### Get configuration

The command [aux4 config get](./commands/config/get) can be used to get the configuration.

```bash
> aux4 config get dev
```
```json
{
  "host": "localhost",
  "port": 3000
}
```

Or specify a config file:

```bash
> aux4 config get --file config.yaml dev
```
```json
{
  "host": "localhost",
  "port": 3000
}
```


```bash
> aux4 config get dev/host
```
```bash
localhost
```

Or with a specific config file:

```bash
> aux4 config get --file config.yaml dev/host
```
```bash
localhost
```

### Set configuration

The command [aux4 config set](./commands/config/set) can be used to set the configuration.

```bash
> aux4 config set --name dev/host --value dev.aux4.io
```

Or with a specific config file:

```bash
> aux4 config set --file config.yaml --name dev/host --value dev.aux4.io
```

```bash
> aux4 config get dev
```
```json
{
  "host": "dev.aux4.io",
  "port": 3000
}
```

### Merge configuration

The command [aux4 config merge](./commands/config/merge) can be used to merge configurations.

```bash
> aux4 config get --file dev.yaml | aux4 config merge --file prod.yaml | jq .
```
```json
{
  "dev": {
    "host": "localhost",
    "port": 3000
  },
  "prod": {
    "host": "aux4.io",
    "port": 80
  }
}
```

You can also save the merged result directly to a file:

```bash
> aux4 config get --file dev.yaml | aux4 config merge --file prod.yaml --save
```

Or merge into a specific path within the configuration:

db.json:
```json
{
  "host": "localhost",
  "port": 5432,
  "user": "postgres",
  "password": "postgres",
  "database": "postgres"
}
```

```bash
> cat db.json | aux4 config merge --file config.json --name dev/db | jq .
```
```json
{
  "dev": {
    "db": {
      "database": "postgres",
      "host": "localhost",
      "password": "postgres",
      "port": 5432,
      "user": "postgres"
    },
    "host": "localhost",
    "port": 3000
  }
}
```
