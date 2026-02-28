package entity

// Scene represents the top-level Excalidraw JSON document.
type Scene struct {
	Type     string                            `json:"type"`
	Version  int                               `json:"version"`
	Source   string                            `json:"source"`
	Elements []map[string]interface{}          `json:"elements"`
	AppState map[string]interface{}            `json:"appState"`
	Files    map[string]map[string]interface{} `json:"files"`
}

// NewScene returns an empty Excalidraw scene with sane defaults.
func NewScene() *Scene {
	return &Scene{
		Type:     "excalidraw",
		Version:  2,
		Source:   "https://excalidraw.com",
		Elements: []map[string]interface{}{},
		AppState: map[string]interface{}{
			"gridSize":            nil,
			"viewBackgroundColor": "#ffffff",
		},
		Files: map[string]map[string]interface{}{},
	}
}
