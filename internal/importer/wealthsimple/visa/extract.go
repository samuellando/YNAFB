package visa

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

// extract reads a PDF from data and returns its text as a structured document.
func extract(data []byte) (document, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read pdf: %w", err)
	}

	var doc document
	for i := 1; i <= reader.NumPage(); i++ {
		rows, err := reader.Page(i).GetTextByRow()
		if err != nil {
			return nil, fmt.Errorf("extract page %d: %w", i, err)
		}

		// Position increases top-to-bottom, so sort ascending for reading order.
		sort.SliceStable(rows, func(a, b int) bool {
			return rows[a].Position < rows[b].Position
		})

		for _, row := range rows {
			doc = append(doc, groupRow(row))
		}
	}

	return doc, nil
}

// groupRow groups a row's text fragments into columns by their X position and
// concatenates fragments within a column in reading order.
func groupRow(row *pdf.Row) textLine {
	type bucket struct {
		x   float64
		sb  strings.Builder
		has bool
	}

	buckets := make([]bucket, 0, 4)
	for _, t := range row.Content {
		key := math.Round(t.X)
		idx := -1
		for i := range buckets {
			if buckets[i].has && buckets[i].x == key {
				idx = i
				break
			}
		}
		if idx < 0 {
			buckets = append(buckets, bucket{x: key, has: true})
			idx = len(buckets) - 1
		}
		buckets[idx].sb.WriteString(t.S)
	}

	line := textLine{y: float64(row.Position)}
	for i := range buckets {
		line.columns = append(line.columns, strings.TrimSpace(buckets[i].sb.String()))
	}
	return line
}
