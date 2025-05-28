package uuidhelpers

import (
	"github.com/google/uuid"
)

func IsValidUUID(uid string) bool {
	_, err := uuid.Parse(uid)
	return err == nil
}
