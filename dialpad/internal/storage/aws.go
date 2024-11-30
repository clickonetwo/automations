/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package storage

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func S3GetBlob(ctx context.Context, blobname string, blobfile *os.File) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(GetConfig().AwsRegion))
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	env := GetConfig()
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(env.AwsBucket),
		Key:    aws.String(env.AwsDialpadFolder + "/" + blobname),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(blobfile, resp.Body)
	return err
}

func S3PutBlob(ctx context.Context, blobname string, blobfile *os.File) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	env := GetConfig()
	stat, err := blobfile.Stat()
	if err != nil {
		return err
	}
	bloblen := stat.Size()
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(env.AwsBucket),
		Key:           aws.String(env.AwsDialpadFolder + "/" + blobname),
		ContentType:   aws.String("application/octet-stream"),
		ContentLength: aws.Int64(bloblen),
		Body:          blobfile,
	})
	return err
}
