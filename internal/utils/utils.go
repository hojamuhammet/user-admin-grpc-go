package utils

import (
	"database/sql"
	"time"
)

func CreateNullString(value string) sql.NullString {
	if value != "" {
		return sql.NullString{String: value, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func NullableStringToString(valid bool, value string) string {
    if valid {
        return value
    }
    return ""
}


func ToDate(year, month, day int32) time.Time {
	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
}