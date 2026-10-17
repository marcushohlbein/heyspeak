package main

import "testing"

func TestCLIModelRejectsMissingAndUnknown(t *testing.T) {
	if err := cliModel(nil); err == nil {
		t.Fatal("missing subcommand")
	}
	if err := cliModel([]string{"nope"}); err == nil {
		t.Fatal("unknown subcommand")
	}
}
