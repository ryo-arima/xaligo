package controller_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/pkg/controller"
)

func TestInitVersionCmd_NotNil(t *testing.T) {
	cmd := controller.InitVersionCmd()
	if cmd == nil {
		t.Error("InitVersionCmd() returned nil")
	}
}

func TestInitVersionCmd_UseName(t *testing.T) {
	cmd := controller.InitVersionCmd()
	if cmd.Use != "version" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "version")
	}
}
