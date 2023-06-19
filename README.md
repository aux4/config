# config

## Install

```bash
npme install --global @aux4/config
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