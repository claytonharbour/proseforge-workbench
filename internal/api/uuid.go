package api

import "github.com/google/uuid"

// uuidParse parses a UUID string. Returns uuid.Nil on invalid input.
func uuidParse(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
