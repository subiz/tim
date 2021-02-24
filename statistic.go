package tim

import (
	"strconv"
	"strings"
)

func formatLabel(label string, length int) string {
	if len(label) < length {
		return strings.Repeat(" ", length-len(label)) + label
	}

	return label[:length-3] + "..."
}

func topK(m map[string]int, k int) map[string]int {
	out := make(map[string]int)

	for i := 0; i < k; i++ {
		maxi := ""
		for word, freq := range m {
			if _, has := out[word]; has {
				continue
			}

			if maxi == "" {
				maxi = word
			}
			if m[maxi] < freq {
				maxi = word
			}
		}
		if maxi != "" {
			out[maxi] = m[maxi]
		}
	}
	return out
}

func drawGraph(full map[string]int) string {
	// convert data to sort
	labels := []string{}
	data := []int{}
	for k, v := range full {
		labels = append(labels, k)
		data = append(data, v)
	}

	// sort data
	for i := 0; i < len(data); i++ {
		for j := i + 1; j < len(data); j++ {
			if data[i] < data[j] { // swap
				data[i], data[j] = data[j], data[i]
				labels[i], labels[j] = labels[j], labels[i]
			}
		}
	}

	s := ""
	// draw graph
	for i, d := range data {
		l := formatLabel(labels[i], 50)
		numStroke := d * 120 / data[0] // max 60 strokes
		line := strings.Repeat("#", numStroke) + strings.Repeat(" ", 120-numStroke)
		s += "\n" + l + " " + line + "  " + strconv.Itoa(d)
	}
	return s
}
