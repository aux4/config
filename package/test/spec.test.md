# Generate configuration specs

## Generate at two independent locations

```file:.aux4
{
  "profiles": [{
    "name": "main",
    "commands": [{
      "name": "spec-example",
      "execute": ["log:unexpected execution"],
      "help": {
        "text": "Example connection",
        "variables": [{"name":"tls.enabled","text":"Enable TLS","type":"boolean","default":"false"}]
      }
    }]
  }]
}
```

### Preserve both sections and support repeated generation

```execute
set -e
work=$(mktemp -d ./config-spec-test.XXXXXX)
trap 'rm -rf "$work"' EXIT
(
cd "$work"
aux4 config spec generate spec-example
aux4 config spec generate spec-example --prefix notifications.smtp
aux4 config spec generate spec-example --prefix notifications.smtp
grep -c 'type: boolean' config.spec.yaml
)
```

```expect
2
```
