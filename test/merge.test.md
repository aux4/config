# aux4 config merge

## when file is json

```file:file1.json
{
  "config": {
    "dev": {
      "host": "localhost",
      "port": 3000
    }
  }
}
```


```file:file2.json
{
  "config": {
    "prod": {
      "host": "aux4.io",
      "port": 80
    }
  }
}
```

### merge file2 into file1

#### should merge file2 into file1 without saving

```execute
aux4 config get --file file2.json | aux4 config merge --file file1.json
```

```expect
{
  "dev": {
    "host": "localhost",
    "port": 3000
  },
  "prod": {
    "host": "aux4.io",
    "port": 80
  }
}
```

##### file1 should not be changed

```execute
cat file1.json
```

```expect
{
  "config": {
    "dev": {
      "host": "localhost",
      "port": 3000
    }
  }
}
```

### merge file2 into file1 and save

#### should merge file2 into file1 and save

```execute
aux4 config get --file file2.json | aux4 config merge --file file1.json --save
cat file1.json
```

```expect
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
