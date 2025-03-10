# aux4 config

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
> aux4 config get --name dev
```
```json
{
  "host": "localhost",
  "port": 3000
}
```


```bash
> aux4 config get --name dev/host
```
```bash
localhost
```

### Set configuration

The command [aux4 config set](./commands/config/set) can be used to set the configuration.

```bash
> aux4 config set --name dev/host --value dev.aux4.io
```

```bash
> aux4 config get --name dev
```
```json
{
  "host": "dev.aux4.io",
  "port": 3000
}
```
