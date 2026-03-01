package repository_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ryo-arima/xaligo/internal/entity"
	"github.com/ryo-arima/xaligo/internal/repository"
)

func TestReadScene_ValidFile(t *testing.T) {
	scene := entity.NewScene()
	scene.Elements = append(scene.Elements, map[string]interface{}{"id": "el1", "type": "rectangle"})

	data, err := json.MarshalIndent(scene, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "test.excalidraw")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	got, err := repository.ReadScene(tmpFile)
	if err != nil {
		t.Fatalf("ReadScene: %v", err)
	}
	if got.Type != "excalidraw" {
		t.Errorf("Type = %q, want %q", got.Type, "excalidraw")
	}
	if len(got.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(got.Elements))
	}
}

func TestReadScene_NonExistentFile_Error(t *testing.T) {
	_, err := repository.ReadScene("/nonexistent/path/file.excalidraw")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReadScene_InvalidJSON_Error(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.excalidraw")
	if err := os.WriteFile(tmpFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	_, err := repository.ReadScene(tmpFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestWriteScene_RoundTrip(t *testing.T) {
	scene := entity.NewScene()
	scene.Elements = append(scene.Elements, map[string]interface{}{"id": "el1", "type": "text"})

	tmpFile := filepath.Join(t.TempDir(), "out.excalidraw")
	if err := repository.WriteScene(scene, tmpFile); err != nil {
		t.Fatalf("WriteScene: %v", err)
	}

	got, err := repository.ReadScene(tmpFile)
	if err != nil {
		t.Fatalf("ReadScene after WriteScene: %v", err)
	}
	if got.Type != scene.Type {
		t.Errorf("Type = %q, want %q", got.Type, scene.Type)
	}
	if len(got.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(got.Elements))
	}
}

func TestWriteScene_InvalidDir_Error(t *testing.T) {
	scene := entity.NewScene()
	err := repository.WriteScene(scene, "/nonexistent/dir/out.excalidraw")
	if err == nil {
		t.Error("expected error writing to non-existent directory")
	}
}
