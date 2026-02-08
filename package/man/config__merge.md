If *file* is not provided it will look for a file named `config.yaml`, `config.yml`, or `config.json` in the current directory.

In case you have multiple configuration files, you can merge them.

For example, if you have a `dev.yaml` file:

```yaml
config:
  dev:
    host: localhost
    port: 3000
```

And a `prod.yaml` file:

```yaml
config:
  prod:
    host: aux4.io
    port: 80
```

You can merge them with:

```bash
> aux4 config get --file dev.yaml | aux4 config merge --file prod.yaml | jq .
```

You can merge any JSON to your configuration file as well.

Instead of generating a new configuration file, you can also merge the configuration to an existing file
using the `--save` flag. It will keep the current file format (either JSON or YAML).

```bash
> aux4 config get --file dev.yaml | aux4 config merge --file prod.yaml --save
```

You can also merge into a specific path within the configuration.

For example, if you have a database configuration file:

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

And a main config file:

config.json:
```json
{
  "config": {
    "dev": {
      "host": "localhost",
      "port": 3000
    }
  }
}
```

You can merge the database config into the dev section:

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
