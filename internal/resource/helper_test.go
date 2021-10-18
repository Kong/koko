package resource

import "github.com/google/uuid"

func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
