# aux4 config get

## when file is json

```file:config.json
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

### get dev

#### should return dev value

```execute
aux4 config get dev
```

```expect
{
  "host": "localhost",
  "port": 3000
}
```

### get dev/host

#### should return dev host value

```execute
aux4 config get dev/host
```

```expect
localhost
```

### get undefined value

#### should return nothing

```execute
aux4 config get dev/undefined
```

```expect
```

## when file is yaml

```file:config.yaml
config:
  dev:
    host: localhost
    port: 3000
  prod:
    host: aux4.io
    port: 80
```

### get dev

#### should return dev value

```execute
aux4 config get dev
```

```expect
{
  "host": "localhost",
  "port": 3000
}
```

## when file is wrong json

```file:config.json
{
  wrong: json
}
```

```execute
aux4 config get dev
```

```error
Error loading config: invalid character 'w' looking for beginning of object key string
```

## when file is wrong yaml

```file:config.yaml
config: wrong
  other: wrong
```

```execute
aux4 config get dev
```

```error
Error loading config: yaml: line 2: mapping values are not allowed in this context
```

## when there are multiple config files

```file:config.yaml
config:
  dev:
    host: localhost
    port: 3000
  prod:
    host: aux4.io
    port: 80
```

```file:second.yaml
config:
  dev:
    host: 127.0.0.1
    port: 3000
```

### when file is not specified

#### get dev host

##### should return localhost

```execute
aux4 config get dev/host
```

```expect
localhost
```

### when file is specified

#### get dev host

##### should return 127.0.0.1

```execute
aux4 config get --file second.yaml dev/host 
```

```expect
127.0.0.1
```
