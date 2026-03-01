package command_test

import (
	"testing"

	command "github.com/ryo-arima/xaligo/pkg"
)

func TestNewRootCmd_NotNil(t *testing.T) {
	cmd := command.NewRootCmd()
	if cmd == nil {
		t.Error("NewRootCmd() returned nil")
	}
}

func TestNewRootCmd_HasSubcommands(t *testing.T) {
	cmd := command.NewRootCmd()
	if len(cmd.Commands()) == 0 {
		t.Error("root command should have subcommands")
	}
}

func TestNewRootCmd_SubcommandNames(t *testing.T) {
	cmd := command.NewRootCmd()
	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Use] = true
	}
	expected := []string{"render <input.xal>", "init", "version", "add", "generate"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}
