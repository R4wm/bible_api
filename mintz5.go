package mintz5

import (
	"math/rand"
	"time"
)

// Colors soft colors for html background
var Colors = []string{
	"AntiqueWhite ",
	"BlanchedAlmond",
	"BurlyWood",
	"Coral",
	"DarkKhaki",
	"DarkSeaGreen",
	"GoldenRod",
	"IndianRed",
	"LightSalmon",
	"MediumSpringGreen",
	"Olive",
	"Orange",
	"SpringGreen",
	"YellowGreen",
}

var seed = rand.NewSource(time.Now().UnixNano())
var random = rand.New(seed)

// GetRandomColor returns a random named html color supported by all browsers
func GetRandomColor() string {
	return Colors[random.Intn(len(Colors))]
}
