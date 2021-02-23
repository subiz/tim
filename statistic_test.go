package tim

import (
	"fmt"
	"testing"
)

func TestGraph(t *testing.T) {
	D := map[string]int{
		"dog":                    123123,
		"cat":                    498,
		"func":                   4903,
		"string":                 403,
		"which":                  1939,
		"contains":               193,
		"a":                      1200,
		"repeated":               1203,
		"string contains":        12934,
		"and":                    129,
		"it":                     3903,
		"is":                     39023,
		"defined":                3023,
		"under":                  323,
		"the":                    323,
		"strings":                329,
		"package":                323,
		"So":                     133,
		"you":                    494,
		"have":                   18439,
		"to":                     30949,
		"import":                 8492,
		"this is a long strings": 15804,
		"package is also long":   49034,
		"in":                     23,
		"your":                   1343,
		"program":                230940,
		"for":                    3904,
		"accessing":              2390,
		"Repeat":                 4344,
		"function":               3434,
	}
	fmt.Println(drawGraph("T", D))
}
