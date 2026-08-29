package checking

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/ledongthuc/pdf"
)

// gapThreshold separates tokens within a line. Character advances within a
// word are ~0pt, while gaps between columns are >= ~3.5pt.
const gapThreshold = 1.0

// extract reads a PDF from data and returns its text as a structured document.
//
// Desjardins (AccesD) statements have a non-PDF prefix before the "%PDF"
// header and position text with cm/Tm matrices that GetTextByRow ignores, so
// this uses Content().Text which applies the full text matrix.
func extract(data []byte) (document, error) {
	if idx := bytes.Index(data, []byte("%PDF")); idx > 0 {
		data = data[idx:]
	}

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read pdf: %w", err)
	}

	var doc document
	for i := 1; i <= reader.NumPage(); i++ {
		texts := reader.Page(i).Content().Text
		sort.SliceStable(texts, func(a, b int) bool {
			if math.Abs(texts[a].Y-texts[b].Y) > 1 {
				return texts[a].Y > texts[b].Y
			}
			return texts[a].X < texts[b].X
		})

		var cur *textLine
		var rightEdge float64
		haveToken := false
		for _, t := range texts {
			if cur == nil || math.Abs(cur.y-t.Y) > 1 {
				doc = append(doc, textLine{y: t.Y})
				cur = &doc[len(doc)-1]
				haveToken = false
			}
			if !haveToken || t.X-rightEdge > gapThreshold {
				cur.tokens = append(cur.tokens, token{x: t.X, right: t.X + t.W, text: t.S})
				haveToken = true
			} else {
				last := &cur.tokens[len(cur.tokens)-1]
				last.text += t.S
				last.right = t.X + t.W
			}
			rightEdge = t.X + t.W
		}
	}

	return doc, nil
}
