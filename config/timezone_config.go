package config

import (
	"time"
)

var NEW_YORK *time.Location

func init() {
	NEW_YORK, _ = time.LoadLocation("America/New_York")
}
