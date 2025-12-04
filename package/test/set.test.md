# aux4 config set

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

### set dev/host

#### should set dev host value

```execute
aux4 config set --name dev/host --value 127.0.0.1
cat config.json | jq .
```

```expect
{
  "config": {
    "dev": {
      "host": "127.0.0.1",
      "port": 3000
    },
    "prod": {
      "host": "aux4.io",
      "port": 80
    }
  }
}
```

### set to a nested property that does not exist

#### should create the nested property

```execute
aux4 config set --name dev/protocol/type --value http
cat config.json | jq .
```

```expect
{
  "config": {
    "dev": {
      "host": "localhost",
      "port": 3000,
      "protocol": {
        "type": "http"
      }
    },
    "prod": {
      "host": "aux4.io",
      "port": 80
    }
  }
}
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

### set dev/host

#### should set dev host value

```execute
aux4 config set --name dev/host --value 127.0.0.1
cat config.yaml
```

```expect
config:
  dev:
    host: 127.0.0.1
    port: 3000
  prod:
    host: aux4.io
    port: 80
```

### set to a nested property that does not exist

#### should create the nested property

```execute
aux4 config set --name dev/protocol/type --value http
cat config.yaml
```

```expect
config:
  dev:
    host: localhost
    port: 3000
    protocol:
      type: http
  prod:
    host: aux4.io
    port: 80
```

