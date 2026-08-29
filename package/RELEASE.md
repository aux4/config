# Release Notes

## Features

- `--file` now reads the `AUX4_CONFIG_FILE` environment variable. Point it at a
  `config.yaml` once — for example an S3-synced config on a cloud VM — instead of
  passing `--file` on every call. Precedence stays argument > env > default.
