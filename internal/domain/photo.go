package domain

import "time"

// Photo represents a photo attached to a risk.
type Photo struct {
	RiskID      int64
	PhotoNumber int
	FilePath    string
	UploadedAt  time.Time
}
