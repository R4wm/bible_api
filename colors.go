package kjv

import "math/rand"

// Colors soft colors for html background
var (
	Colors = []string{
		"AntiqueWhite ",
		"BlanchedAlmond",
		"BurlyWood",
		"Coral",
		"DarkKhaki",
		"DarkSeaGreen",
		"GoldenRod",
		"Silver",
		"LightSalmon",
		"DarkSalmon",
		"MediumSpringGreen",
		"Olive",
		"Orange",
		"SpringGreen",
		"YellowGreen",
		"LightPink",
		"PapayaWhip",
		"Moccasin",
		"Khaki",
		"DarkKhaki",
		"Lavender",
		"Plum",
		"GreenYellow",
		"PaleGreen",
		"MediumSeaGreen",
		"LightCyan",
		"AquaMarine",
		"CadetBlue",
		"PowderBlue",
		"Peru",
		"Chocolate",
		"HoneyDew",
		"SeaShell",
		"Ivory",
		"FloralWhite",
		"LavenderBlush",
		"MistyRose",
	}
)

// GetRandomColor returns a random named html color supported by all browsers
func GetRandomColor() string {
	return Colors[rand.Intn(len(Colors))]
}
