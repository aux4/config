package main

import (
	"os"
	"strings"
	"testing"
)

func TestMergeSpecCreatesNestedTypedSchema(t *testing.T) {
	file := t.TempDir() + "/config.spec.yaml"
	defaultPort := "5432"
	if err := mergeSpec(file, "database", []specVariable{
		{Name: "host", Text: "Database host", Type: "string", Default: stringPtr("localhost")},
		{Name: "port", Type: "integer", Default: &defaultPort},
		{Name: "password", Type: "string", Hide: true},
		{Name: "mode", Options: []string{"read", "write"}},
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	for _, expected := range []string{"type: object", "database:", "properties:", "host:", "description: Database host", "default: 5432", "type: integer", "hide: true", "options:", "- read"} {
		if !strings.Contains(output, expected) {
			t.Errorf("generated spec missing %q:\n%s", expected, output)
		}
	}
}

func TestMergeSpecAppendsPrefixAndPreservesMetadata(t *testing.T) {
	file := t.TempDir() + "/config.spec.yaml"
	initial := "# user schema\ntype: object\nproperties:\n  database:\n    type: object\n    properties:\n      host:\n        type: string\n        title: Hostname\n        # keep this comment\n"
	if err := os.WriteFile(file, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}
	if err := mergeSpec(file, "database", []specVariable{{Name: "host", Type: "string"}, {Name: "port", Type: "integer"}}); err != nil {
		t.Fatal(err)
	}
	if err := mergeSpec(file, "notifications.smtp", []specVariable{{Name: "host", Type: "string"}}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(file)
	output := string(data)
	for _, expected := range []string{"# user schema", "title: Hostname", "# keep this comment", "port:", "notifications:", "smtp:", "host:"} {
		if !strings.Contains(output, expected) {
			t.Errorf("merged spec missing %q:\n%s", expected, output)
		}
	}
}

func TestMergeSpecSkipsPrivateAndDetectsConflictAtomically(t *testing.T) {
	file := t.TempDir() + "/config.spec.yaml"
	if err := mergeSpec(file, "", []specVariable{{Name: "secret", Type: "string", Private: true}, {Name: "port", Type: "integer"}}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(file)
	if err := mergeSpec(file, "", []specVariable{{Name: "port", Type: "string"}, {Name: "new", Type: "string"}}); err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected port conflict, got %v", err)
	}
	after, _ := os.ReadFile(file)
	if string(before) != string(after) {
		t.Fatalf("conflict changed file:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if strings.Contains(string(after), "secret:") {
		t.Fatal("private variable was generated")
	}
}

func TestMergeSpecMultipleUsesArrayItems(t *testing.T) {
	file := t.TempDir() + "/config.spec.yaml"
	if err := mergeSpec(file, "", []specVariable{{Name: "hosts", Type: "string", Multiple: true, Default: stringPtr("a")}}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(file)
	output := string(data)
	if !strings.Contains(output, "type: array") || !strings.Contains(output, "items:") || !strings.Contains(output, "default: a") {
		t.Fatalf("unexpected array schema:\n%s", output)
	}
}

func stringPtr(value string) *string { return &value }
