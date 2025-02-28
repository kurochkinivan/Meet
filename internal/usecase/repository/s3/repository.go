package s3

import "github.com/aws/aws-sdk-go-v2/service/s3"

type Repositories struct {
	*PhotoRepository
}

func NewRepositories(client *s3.Client, bucketName string) *Repositories {
	return &Repositories{
		PhotoRepository: NewPhotoRepository(client, bucketName),
	}
}
