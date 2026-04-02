package model

import "time"

type RecordType string

const (
    RecordEntry RecordType = "entry"
    RecordExit  RecordType = "exit"
)

type TimeRecord struct {
    ID        int64      `json:"id"         db:"id"`
    UserID    int64      `json:"user_id"    db:"user_id"`
    Type      RecordType `json:"type"       db:"type"`
    Timestamp time.Time  `json:"timestamp"  db:"timestamp"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// DailyReport agrupa os registros de um dia e o total de horas
type DailyReport struct {
    Date         string        `json:"date"`
    Records      []TimeRecord  `json:"records"`
    HoursWorked  float64       `json:"hours_worked"`
}