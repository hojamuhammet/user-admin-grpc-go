package utils

import "database/sql"

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