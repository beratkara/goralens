package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestConfigStubIsValidGo(t *testing.T) {
	src := configStub("config", "example.com/app/facades", "facades")
	if _, err := parser.ParseFile(token.NewFileSet(), "goralens.go", src, parser.AllErrors); err != nil {
		t.Fatalf("stub does not parse: %v\n%s", err, src)
	}
}
