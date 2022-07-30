package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func main() {
	sesnConfig := aws.Config{
		Region:   aws.String("s3.us-west-001"),
		Endpoint: aws.String("s3.us-west-001.backblazeb2.com"),
	}
	sesn, err := session.NewSession(&sesnConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	svc := s3.New(sesn)

	newbucket := s3.CreateBucketInput{
		Bucket: aws.String("go-aws-s3-course-for-me"), // this name needs to be globally unique, so I hope this will be unique.
	}
	input := &newbucket

	result, err := svc.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok { // this is a type assertion test.
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				log.Fatal("Error: Bucket already exists")
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				log.Fatal("Error: Bucket already owned by you, as if you already ran this code.")
			default:
				log.Fatal(aerr.Error())
			}
		} else {
			log.Fatal(err.Error())
		}
	}

	log.Print(result)
}

/*
  This code doesn't work on backblaze.  I'm getting an error
  invalid argument: location constraint specifies another region, Backblaze does not support cross-region requests.
  I wonder if the error message is not quite right.  I may have to create the bucket from the web interface, and d/l the keys, so I can then access the bucket.
*/
