package main

import (
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func init() {
	publishProviders["s3"] = publishProviderS3{}
}

type publishProviderS3 struct{}

func (p publishProviderS3) Write(targetURL *url.URL, buf io.Reader) error {
	s3Svc := s3.New(session.New())

	_, err := s3Svc.PutObject(&s3.PutObjectInput{
		Body:        aws.ReadSeekCloser(buf),
		Bucket:      aws.String(targetURL.Host),
		ContentType: aws.String("application/json"),
		Key:         aws.String(targetURL.Path),
	})

	return err
}
