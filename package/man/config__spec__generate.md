Generate a configuration spec from the visible help variables of an aux4 command.
Pass the command path as positional arguments. The default output is
`config.spec.yaml` in the current directory.

```bash
aux4 config spec generate myapp database connect --prefix database
aux4 config spec generate myapp mail connect --prefix notifications.smtp
aux4 config spec generate myapp serve
```

A field named `host` from the mail command becomes
`notifications.smtp.host`. With an empty prefix fields start at the root.
Use `--file vm.spec.yaml` to select another spec.

Each invocation adds missing fields and preserves existing sections.
If definitions conflict, the command reports the field and leaves the file intact.
