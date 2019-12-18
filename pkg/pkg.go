package pkg

import (
	"time"
)

// FileInfoDetails contains basic stat + permission details about a file
type FileInfoDetails struct {
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
	Perm    int       `json:"permission"`
	ModTime time.Time `json:"updated_at"`
}
