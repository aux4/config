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
> aux4 config get --file dev.yaml | aux4 config merge --file prod.yaml > config.json
```

You can merge any JSON to your configuration file as well.

Instead of generating a new configuration file, you can also merge the configuration to an existing file
using the `--save` flag. It will keep the current file format (either JSON or YAML).

```bash
> aux4 config get --file dev.yaml | aux4 config merge --file prod.yaml --save
```
