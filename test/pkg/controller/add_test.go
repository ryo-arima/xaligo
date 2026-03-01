package controller_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/pkg/controller"
)

func TestInitAddCmd_NotNil(t *testing.T) {
	cmd := controller.InitAddCmd()
	if cmd == nil {
		t.Error("InitAddCmd() returned nil")
	}
}

func TestInitAddCmd_HasServiceSubcommand(t *testing.T) {
	cmd := controller.InitAddCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "service" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'service' subcommand under 'add'")
	}
}
