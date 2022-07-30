package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func main() {
	sesnConfig := aws.Config{
		Region:   aws.String("s3.us-west"),
		Endpoint: aws.String("s3.us-west-001.backblazeb2.com"),
	}

	sesn, err := session.NewSession(&sesnConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	svc := s3.New(sesn)

	result, er := svc.ListBuckets(nil)
	if er != nil {
		log.Fatalf(" Error listing buckets is: %v\n", er)
	}

	for _, bucket := range result.Buckets {
		log.Printf(" Bucket: %s\n", aws.StringValue(bucket.Name))
	}
}
