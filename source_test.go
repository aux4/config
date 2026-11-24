package main

import (
	"reflect"
	"testing"
)

func TestParseCommandVariablesParams(t *testing.T) {
	data := []byte(`{"name":"connect","params":[{"name":"database.host","text":"Database host","default":"localhost","type":"string"},{"name":"port","text":"Port","default":5432,"type":"integer","options":["5432","6543"]},{"name":"password","text":"Password","hide":true},{"name":"internal","private":true}]}`)
	variables, err := parseCommandVariables(data)
	if err != nil {
		t.Fatal(err)
	}
	expectedDefault := "localhost"
	expectedNumber := "5432"
	expected := []specVariable{
		{Name: "database.host", Text: "Database host", Type: "string", Default: &expectedDefault},
		{Name: "port", Text: "Port", Type: "integer", Default: &expectedNumber, Options: []string{"5432", "6543"}},
		{Name: "password", Text: "Password", Hide: true},
	}
	if !reflect.DeepEqual(variables, expected) {
		t.Fatalf("variables = %#v, want %#v", variables, expected)
	}
}

func TestParseCommandVariablesSupportsVariablesKey(t *testing.T) {
	variables, err := parseCommandVariables([]byte(`{"variables":[{"name":"enabled","type":"boolean","default":true,"multiple":true}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(variables) != 1 || variables[0].Name != "enabled" || variables[0].Type != "boolean" || !variables[0].Multiple {
		t.Fatalf("unexpected variables: %#v", variables)
	}
	if variables[0].Default == nil || *variables[0].Default != "true" {
		t.Fatalf("default = %v, want true", variables[0].Default)
	}
}

func TestParseCommandVariablesRejectsInvalidDefault(t *testing.T) {
	_, err := parseCommandVariables([]byte(`{"params":[{"name":"value","default":}]}`))
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
