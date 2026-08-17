package application

import "github.com/google/uuid"

func newID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
