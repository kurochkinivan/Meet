package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/sirupsen/logrus"
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

func (r *PhotoRepository) UploadPhoto(ctx context.Context, userID string, file io.Reader) (url, objectKey string, err error) {
	objectKey = fmt.Sprintf("users/%s/photos/%s.jpg", userID, uuid.New().String())

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
		ACL:    types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", "", apperr.WithHTTPStatus(fmt.Errorf("can't upload file with objectkey %s, err: %w", objectKey, err), http.StatusInternalServerError)
	}

	err = s3.NewObjectExistsWaiter(r.client).Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	}, time.Minute)
	if err != nil {
		return "", "", apperr.WithHTTPStatus(fmt.Errorf("failed attempt to wait for object %s to exist", objectKey), http.StatusInternalServerError)
	}

	url = fmt.Sprintf("https://storage.yandexcloud.net/%s/%s", r.bucketName, objectKey)
	return url, objectKey, nil
}

func (r *PhotoRepository) DeletePhoto(ctx context.Context, objectKey string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError

		if errors.As(err, &noKey) {
			logrus.WithField("objectKey", objectKey).Error(fmt.Sprintf("object does not exist in %s", r.bucketName))
			return apperr.WithHTTPStatus(noKey, http.StatusBadRequest)
		}

		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "AccessDenied" {
				logrus.WithField("objectKey", objectKey).Error(fmt.Sprintf("access denied: cannot delete object from %s", r.bucketName))
			}
			return apperr.WithHTTPStatus(apiErr, http.StatusInternalServerError)
		}

		return err
	}

	return nil
}
