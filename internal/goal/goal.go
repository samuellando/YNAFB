package goal

import (
	"time"
)

type Goal interface {
	ForMonth(time.Time) int
}
