package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/kurochkinivan/Meet/internal/apperr"
)

type PhotoRepository struct {
	client     *s3.Client
	bucketName string
}

func NewPhotoRepository(client *s3.Client, bucketName string) *PhotoRepository {
	return &PhotoRepository{
		client:     client,
		bucketName: bucketName,
	}
}

func (r *PhotoRepository) UploadPhoto(ctx context.Context, userID string, file io.Reader) (string, error) {
	objectKey := fmt.Sprintf("users/%s/photos/%s.jpg", userID, uuid.New().String())

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return "", apperr.WithHTTPStatus(fmt.Errorf("can't upload file with objectkey %s, err: %w", objectKey, err), http.StatusInternalServerError)
	}

	err = s3.NewObjectExistsWaiter(r.client).Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	}, time.Minute)
	if err != nil {
		return "", apperr.WithHTTPStatus(fmt.Errorf("failed attempt to wait for object %s to exist", objectKey), http.StatusInternalServerError)
	}

	return objectKey, nil
}
