package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ryo-arima/xaligo/internal/entity"
)

// ReadScene reads a .excalidraw JSON file and returns the parsed Scene.
func ReadScene(path string) (*entity.Scene, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scene %s: %w", path, err)
	}
	var scene entity.Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("parse scene %s: %w", path, err)
	}
	if scene.Files == nil {
		scene.Files = map[string]map[string]interface{}{}
	}
	if scene.Elements == nil {
		scene.Elements = []map[string]interface{}{}
	}
	return &scene, nil
}

// WriteScene serialises the Scene and writes it back to path.
func WriteScene(scene *entity.Scene, path string) error {
	data, err := json.MarshalIndent(scene, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal scene: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write scene %s: %w", path, err)
	}
	return nil
}
