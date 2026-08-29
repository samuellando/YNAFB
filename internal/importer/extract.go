package importer

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

// TextLine is a single visual line from a PDF, with its text grouped into
// left-to-right columns by X position.
type TextLine struct {
	Y       float64
	Columns []string
}

// Document is an ordered (top-to-bottom) list of text lines.
type Document []TextLine

// Text renders the document back into a single whitespace-separated string.
// It is intended for parser detection (Match) rather than field extraction.
func (d Document) Text() string {
	var sb strings.Builder
	for _, line := range d {
		sb.WriteString(strings.Join(line.Columns, " "))
		sb.WriteString("\n")
	}
	return sb.String()
}

// ExtractFile reads a PDF from the given path and returns its text as a
// structured Document.
func ExtractFile(path string) (Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}

	return Extract(f, info.Size())
}

// Extract reads a PDF from r (of the given size) and returns its text as a
// structured Document.
func Extract(r io.ReaderAt, size int64) (Document, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("read pdf: %w", err)
	}

	var doc Document
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
func groupRow(row *pdf.Row) TextLine {
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

	line := TextLine{Y: float64(row.Position)}
	for i := range buckets {
		line.Columns = append(line.Columns, strings.TrimSpace(buckets[i].sb.String()))
	}
	return line
}
