package mintz5

import (
	"fmt"
	"testing"
)

func TestGetRandomColor(t *testing.T) {
	color := GetRandomColor()
	colorType := fmt.Sprintf("%T", color)

	if colorType != "string" {
		t.Errorf("Expected string got %v\n", colorType)
	}
}
