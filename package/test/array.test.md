# aux4 config get - Array Parsing

## yaml

### when file contains array of strings

```file:config.yaml
config:
  environments:
    - development
    - staging
    - production
  tags:
    - frontend
    - backend
    - api
  features:
    - auth
    - dashboard
    - analytics
```

#### get environments array

##### should return array of environment strings in JSON format

```execute
aux4 config get environments | jq .
```

```expect
[
  "development",
  "staging",
  "production"
]
```

#### get tags array

##### should return array of tag strings in JSON format

```execute
aux4 config get tags | jq .
```

```expect
[
  "frontend",
  "backend",
  "api"
]
```

#### get features array

##### should return array of feature strings in JSON format

```execute
aux4 config get features | jq .
```

```expect
[
  "auth",
  "dashboard",
  "analytics"
]
```

### when file contains array of objects

```file:config.yaml
config:
  servers:
    - name: web-server
      host: localhost
      port: 8080
      ssl: false
    - name: api-server
      host: api.example.com
      port: 443
      ssl: true
    - name: db-server
      host: db.example.com
      port: 5432
      ssl: true
  users:
    - id: 1
      name: admin
      role: administrator
      permissions:
        - read
        - write
        - delete
    - id: 2
      name: user
      role: viewer
      permissions:
        - read
```

#### get servers array

##### should return array of server objects in JSON format

```execute
aux4 config get servers | jq .
```

```expect
[
  {
    "name": "web-server",
    "host": "localhost",
    "port": 8080,
    "ssl": false
  },
  {
    "name": "api-server",
    "host": "api.example.com",
    "port": 443,
    "ssl": true
  },
  {
    "name": "db-server",
    "host": "db.example.com",
    "port": 5432,
    "ssl": true
  }
]
```

#### get users array

##### should return array of user objects in JSON format

```execute
aux4 config get users | jq .
```

```expect
[
  {
    "id": 1,
    "name": "admin",
    "role": "administrator",
    "permissions": [
      "read",
      "write",
      "delete"
    ]
  },
  {
    "id": 2,
    "name": "user",
    "role": "viewer",
    "permissions": [
      "read"
    ]
  }
]
```

### when file contains nested arrays within objects

```file:config.yaml
config:
  deployment:
    stages:
      - name: test
        services:
          - web
          - api
        environments:
          - host: test.example.com
            replicas: 1
          - host: test2.example.com
            replicas: 1
      - name: production
        services:
          - web
          - api
          - worker
        environments:
          - host: prod1.example.com
            replicas: 3
          - host: prod2.example.com
            replicas: 3
          - host: prod3.example.com
            replicas: 2
```

#### get deployment stages array

##### should return array of stage objects with nested arrays

```execute
aux4 config get deployment/stages | jq .
```

```expect
[
  {
    "name": "test",
    "services": [
      "web",
      "api"
    ],
    "environments": [
      {
        "host": "test.example.com",
        "replicas": 1
      },
      {
        "host": "test2.example.com",
        "replicas": 1
      }
    ]
  },
  {
    "name": "production",
    "services": [
      "web",
      "api",
      "worker"
    ],
    "environments": [
      {
        "host": "prod1.example.com",
        "replicas": 3
      },
      {
        "host": "prod2.example.com",
        "replicas": 3
      },
      {
        "host": "prod3.example.com",
        "replicas": 2
      }
    ]
  }
]
```

### when file contains empty arrays

```file:config.yaml
config:
  empty_strings: []
  empty_objects: []
  mixed:
    - name: item1
      tags: []
    - name: item2
      tags:
        - tag1
        - tag2
```

#### get empty string array

##### should return empty array

```execute
aux4 config get empty_strings | jq .
```

```expect
[]
```

#### get empty object array

##### should return empty array

```execute
aux4 config get empty_objects | jq .
```

```expect
[]
```

#### get mixed array with empty nested arrays

##### should return array with objects containing empty and populated arrays

```execute
aux4 config get mixed | jq .
```

```expect
[
  {
    "name": "item1",
    "tags": []
  },
  {
    "name": "item2",
    "tags": [
      "tag1",
      "tag2"
    ]
  }
]
```

### when accessing non-existent array paths

```file:config.yaml
config:
  items:
    - first
    - second
```

#### get non-existent array

##### should return nothing

```execute
aux4 config get non_existent_array
```

```expect
```

## json

### when file contains array of strings

```file:config.json
{
  "config": {
    "environments": [
      "development",
      "staging",
      "production"
    ],
    "tags": [
      "frontend",
      "backend",
      "api"
    ],
    "features": [
      "auth",
      "dashboard",
      "analytics"
    ]
  }
}
```

#### get environments array

##### should return array of environment strings in JSON format

```execute
aux4 config get environments | jq .
```

```expect
[
  "development",
  "staging",
  "production"
]
```

#### get tags array

##### should return array of tag strings in JSON format

```execute
aux4 config get tags | jq .
```

```expect
[
  "frontend",
  "backend",
  "api"
]
```

#### get features array

##### should return array of feature strings in JSON format

```execute
aux4 config get features | jq .
```

```expect
[
  "auth",
  "dashboard",
  "analytics"
]
```

### when file contains array of objects

```file:config.json
{
  "config": {
    "servers": [
      {
        "name": "web-server",
        "host": "localhost",
        "port": 8080,
        "ssl": false
      },
      {
        "name": "api-server",
        "host": "api.example.com",
        "port": 443,
        "ssl": true
      },
      {
        "name": "db-server",
        "host": "db.example.com",
        "port": 5432,
        "ssl": true
      }
    ],
    "users": [
      {
        "id": 1,
        "name": "admin",
        "role": "administrator",
        "permissions": [
          "read",
          "write",
          "delete"
        ]
      },
      {
        "id": 2,
        "name": "user",
        "role": "viewer",
        "permissions": [
          "read"
        ]
      }
    ]
  }
}
```

#### get servers array

##### should return array of server objects in JSON format

```execute
aux4 config get servers | jq .
```

```expect
[
  {
    "host": "localhost",
    "name": "web-server",
    "port": 8080,
    "ssl": false
  },
  {
    "host": "api.example.com",
    "name": "api-server",
    "port": 443,
    "ssl": true
  },
  {
    "host": "db.example.com",
    "name": "db-server",
    "port": 5432,
    "ssl": true
  }
]
```

#### get users array

##### should return array of user objects in JSON format

```execute
aux4 config get users | jq .
```

```expect
[
  {
    "id": 1,
    "name": "admin",
    "permissions": [
      "read",
      "write",
      "delete"
    ],
    "role": "administrator"
  },
  {
    "id": 2,
    "name": "user",
    "permissions": [
      "read"
    ],
    "role": "viewer"
  }
]
```

### when file contains nested arrays within objects

```file:config.json
{
  "config": {
    "deployment": {
      "stages": [
        {
          "name": "test",
          "services": [
            "web",
            "api"
          ],
          "environments": [
            {
              "host": "test.example.com",
              "replicas": 1
            },
            {
              "host": "test2.example.com",
              "replicas": 1
            }
          ]
        },
        {
          "name": "production",
          "services": [
            "web",
            "api",
            "worker"
          ],
          "environments": [
            {
              "host": "prod1.example.com",
              "replicas": 3
            },
            {
              "host": "prod2.example.com",
              "replicas": 3
            },
            {
              "host": "prod3.example.com",
              "replicas": 2
            }
          ]
        }
      ]
    }
  }
}
```

#### get deployment stages array

##### should return array of stage objects with nested arrays

```execute
aux4 config get deployment/stages | jq .
```

```expect
[
  {
    "environments": [
      {
        "host": "test.example.com",
        "replicas": 1
      },
      {
        "host": "test2.example.com",
        "replicas": 1
      }
    ],
    "name": "test",
    "services": [
      "web",
      "api"
    ]
  },
  {
    "environments": [
      {
        "host": "prod1.example.com",
        "replicas": 3
      },
      {
        "host": "prod2.example.com",
        "replicas": 3
      },
      {
        "host": "prod3.example.com",
        "replicas": 2
      }
    ],
    "name": "production",
    "services": [
      "web",
      "api",
      "worker"
    ]
  }
]
```

### when file contains empty arrays

```file:config.json
{
  "config": {
    "empty_strings": [],
    "empty_objects": [],
    "mixed": [
      {
        "name": "item1",
        "tags": []
      },
      {
        "name": "item2",
        "tags": [
          "tag1",
          "tag2"
        ]
      }
    ]
  }
}
```

#### get empty string array

##### should return empty array

```execute
aux4 config get empty_strings | jq .
```

```expect
[]
```

#### get empty object array

##### should return empty array

```execute
aux4 config get empty_objects | jq .
```

```expect
[]
```

#### get mixed array with empty nested arrays

##### should return array with objects containing empty and populated arrays

```execute
aux4 config get mixed | jq .
```

```expect
[
  {
    "name": "item1",
    "tags": []
  },
  {
    "name": "item2",
    "tags": [
      "tag1",
      "tag2"
    ]
  }
]
```

### when accessing non-existent array paths

```file:config.json
{
  "config": {
    "items": [
      "first",
      "second"
    ]
  }
}
```

#### get non-existent array

##### should return nothing

```execute
aux4 config get non_existent_array
```

```expect
```