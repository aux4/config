If *file* is not provided it will look for a file named `config.yaml`, `config.yml`, or `config.json` in the current directory.

The *name* of the configuration can be a nested property. For example, if the configuration file is:

```yaml
config:
  dev:
    host: localhost
    port: 3000
  prod:
    host: aux4.io
    port: 80
```

You can get the __dev__ configuration with:

```bash
> aux4 config get dev/host
```
```bash
localhost
```

You can also specify a different config file with the `--file` flag:

```bash
> aux4 config get --file second.yaml dev/host
```
```bash
127.0.0.1
```

