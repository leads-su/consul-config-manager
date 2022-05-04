package tasks

// EventStructure describes event structure
type EventStructure struct {
	Source    int    `json:"source"`
	Type      int    `json:"type"`
	Line      string `json:"line"`
	Timestamp int64  `json:"timestamp"`
}
