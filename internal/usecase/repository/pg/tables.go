package pg

import "fmt"

const (
	TableUsers  = "users"
	TablePhotos = "photos"
)

func usersField(field string) string {
	return fmt.Sprintf("%s.%s", TableUsers, field)
}

func photosField(field string) string {
	return fmt.Sprintf("%s.%s", TablePhotos, field)
}
