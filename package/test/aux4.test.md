# using config on aux4 CLI

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

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "connect",
          "execute": [
            "log:connecting to $host:$port"
          ],
          "help": {
            "text": "Connect to a server",
            "variables": [
              {
                "name": "host",
                "text": "The server host"
              },
              {
                "name": "port",
                "text": "The server port"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### connecting to dev server

```execute
aux4 connect --config dev
```

```expect
connecting to localhost:3000
```

### connecting to prod server

```execute
aux4 connect --config prod
```

```expect
connecting to aux4.io:80
```
