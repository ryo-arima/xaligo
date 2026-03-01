package entity_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/internal/entity"
)

func TestNewScene_Type(t *testing.T) {
	s := entity.NewScene()
	if s.Type != "excalidraw" {
		t.Errorf("Type = %q, want %q", s.Type, "excalidraw")
	}
}

func TestNewScene_Version(t *testing.T) {
	s := entity.NewScene()
	if s.Version != 2 {
		t.Errorf("Version = %d, want 2", s.Version)
	}
}

func TestNewScene_ElementsNotNil(t *testing.T) {
	s := entity.NewScene()
	if s.Elements == nil {
		t.Error("Elements should not be nil")
	}
	if len(s.Elements) != 0 {
		t.Errorf("Elements should be empty, got %d", len(s.Elements))
	}
}

func TestNewScene_FilesNotNil(t *testing.T) {
	s := entity.NewScene()
	if s.Files == nil {
		t.Error("Files should not be nil")
	}
}

func TestNewScene_AppStateNotNil(t *testing.T) {
	s := entity.NewScene()
	if s.AppState == nil {
		t.Error("AppState should not be nil")
	}
}

func TestNewScene_ViewBackgroundColor(t *testing.T) {
	s := entity.NewScene()
	if s.AppState["viewBackgroundColor"] != "#ffffff" {
		t.Errorf("viewBackgroundColor = %v, want %q", s.AppState["viewBackgroundColor"], "#ffffff")
	}
}
