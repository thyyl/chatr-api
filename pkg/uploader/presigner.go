package uploader

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Presigner struct {
	presignClient  *s3.PresignClient
	lifetimeSecond int64
}

func (presigner *Presigner) GetObject(context context.Context, bucketName string, objectKey string) (*v4.PresignedHTTPRequest, error) {
	request, err := presigner.presignClient.PresignGetObject(context, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Duration(presigner.lifetimeSecond * int64(time.Second))
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't get a presigned request to get %v:%v, reason: %v", bucketName, objectKey, err)
	}

	return request, nil
}

func (presigner *Presigner) PutObject(context context.Context, bucketName string, objectKey string) (*v4.PresignedHTTPRequest, error) {
	request, err := presigner.presignClient.PresignPutObject(context, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Duration(presigner.lifetimeSecond * int64(time.Second))
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't get a presigned request to put %v:%v, reason: %v", bucketName, objectKey, err)
	}

	return request, nil
}
