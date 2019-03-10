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
	"MediumSlateBlue",
	"Olive",
	"Orange",
	"SpringGreen",
	"YellowGreen",
}

var BookChapterLimit = map[string]int{
	"GENESIS":         50,
	"EXODUS":          40,
	"LEVITICUS":       27,
	"NUMBERS":         36,
	"DEUTERONOMY":     34,
	"JOSHUA":          24,
	"JUDGES":          21,
	"RUTH":            4,
	"1SAMUEL":         31,
	"2SAMUEL":         24,
	"1KINGS":          22,
	"2KINGS":          25,
	"1CHRONICLES":     29,
	"2CHRONICLES":     36,
	"EZRA":            10,
	"NEHEMIAH":        13,
	"ESTHER":          10,
	"JOB":             42,
	"PSALMS":          150,
	"PROVERBS":        31,
	"ECCLESIASTES":    12,
	"SONG OF SOLOMON": 8,
	"ISAIAH":          66,
	"JEREMIAH":        52,
	"LAMENTATIONS":    5,
	"EZEKIEL":         48,
	"DANIEL":          12,
	"HOSEA":           14,
	"JOEL":            3,
	"AMOS":            9,
	"OBADIAH":         1,
	"JONAH":           4,
	"MICAH":           7,
	"NAHUM":           3,
	"HABAKKUK":        3,
	"ZEPHANIAH":       3,
	"HAGGAI":          2,
	"ZECHARIAH":       14,
	"MALACHI":         4,
	"MATTHEW":         28,
	"MARK":            16,
	"LUKE":            24,
	"JOHN":            21,
	"ACTS":            28,
	"ROMANS":          16,
	"1CORINTHIANS":    16,
	"2CORINTHIANS":    13,
	"GALATIANS":       6,
	"EPHESIANS":       6,
	"PHILIPPIANS":     4,
	"COLOSSIANS":      4,
	"1THESSALONIANS":  5,
	"2THESSALONIANS":  3,
	"1TIMOTHY":        6,
	"2TIMOTHY":        4,
	"TITUS":           3,
	"PHILEMON":        1,
	"HEBREWS":         13,
	"JAMES":           5,
	"1PETER":          5,
	"2PETER":          3,
	"1JOHN":           5,
	"2JOHN":           1,
	"3JOHN":           1,
	"JUDE":            1,
	"REVELATION":      22,
}

var seed = rand.NewSource(time.Now().UnixNano())
var random = rand.New(seed)

// GetRandomColor returns a random named html color supported by all browsers
func GetRandomColor() string {
	return Colors[random.Intn(len(Colors))]
}

func GetBookMaxChapter(book string) int {
	return BookChapterLimit[book]
}
