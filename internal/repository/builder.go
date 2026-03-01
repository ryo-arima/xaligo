package repository

import "math"

// base builds the common fields shared by all Excalidraw elements.
func base(id string, x, y, w, h float64, seed int) map[string]interface{} {
	return map[string]interface{}{
		"id":           id,
		"x":            math.Round(x),
		"y":            math.Round(y),
		"width":        math.Round(w),
		"height":       math.Round(h),
		"angle":        0,
		"opacity":      100,
		"groupIds":     []interface{}{},
		"frameId":      nil,
		"roundness":    nil,
		"seed":         seed,
		"version":      1,
		"versionNonce": seed,
		"isDeleted":    false,
		"boundElements": []interface{}{},
		"updated":      int64(1709000000000),
		"link":         nil,
		"locked":       false,
	}
}

// MakeText builds an Excalidraw text element.
// textAlign should be "left", "center", or "right".
func MakeText(id string, x, y, w, h float64, text string, fontSize int, color string, bold bool, textAlign string, seed int) map[string]interface{} {
	el := base(id, x, y, w, h, seed)
	el["type"] = "text"
	el["text"] = text
	el["rawText"] = text
	el["originalText"] = text
	el["fontSize"] = fontSize
	el["fontFamily"] = 4
	el["textAlign"] = textAlign
	el["verticalAlign"] = "top"
	el["strokeColor"] = color
	el["backgroundColor"] = "transparent"
	el["fillStyle"] = "solid"
	el["strokeWidth"] = 1
	el["roughness"] = 0
	el["containerId"] = nil
	el["lineHeight"] = 1.25
	if bold {
		el["fontStyle"] = "bold"
	}
	return el
}

// MakeImage builds an Excalidraw image element referencing a file by its ID.
func MakeImage(id string, x, y, w, h float64, fileID string, backgroundColor string, seed int) map[string]interface{} {
	el := base(id, x, y, w, h, seed)
	el["type"] = "image"
	el["status"] = "saved"
	el["fileId"] = fileID
	el["scale"] = []int{1, 1}
	el["roughness"] = 0
	el["strokeColor"] = "transparent"
	el["backgroundColor"] = backgroundColor
	el["fillStyle"] = "solid"
	el["strokeWidth"] = 1
	return el
}
