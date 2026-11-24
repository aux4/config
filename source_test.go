package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// aux4 interprets extra arguments after a leaf command as command inputs.
// Inspecting an invalid path must never run the partially matched command.
func TestInspectLocalVariables(t *testing.T) {
	if _, err := exec.LookPath("aux4"); err != nil {
		t.Skip("aux4 CLI unavailable")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(cwd) })
	manifest := `{"profiles":[{"name":"main","commands":[{"name":"spec-check","execute":["profile:spec-check"]}]},{"name":"spec-check","commands":[{"name":"task","execute":["touch executed"],"help":{"variables":[{"name":"tls.enabled","type":"boolean","default":"false"}]}}]}]}`
	if err = os.WriteFile(".aux4", []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	vars, err := commandVariables([]string{"spec-check", "task"})
	if err != nil || len(vars) != 1 || vars[0].Type != "boolean" || vars[0].Default == nil || *vars[0].Default != "false" {
		t.Fatalf("metadata: %#v, %v", vars, err)
	}
	if _, err = commandVariables([]string{"spec-check", "task", "extra"}); err == nil {
		t.Fatal("accepted extra action")
	}
	if _, err = os.Stat(filepath.Join(dir, "executed")); !os.IsNotExist(err) {
		t.Fatal("target command executed")
	}
}
