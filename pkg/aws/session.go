package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var sharedSession *session.Session = nil

func InitializeAWSSession() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	sharedSession, _ = session.NewSession(&aws.Config{Region: aws.String(region)})
	if sharedSession == nil {
		log.Fatalf("Unable to create a new AWS client session")
	}
}
