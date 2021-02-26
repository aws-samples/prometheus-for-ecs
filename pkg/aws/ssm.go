package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func getSSMClient() *ssm.SSM {
	service := ssm.New(sharedSession)
	return service
}

//
// Retrive the value for a given SSM Parameter name
//
func GetParameter(parameterName string) *string {
	ssmService := ssm.New(sharedSession)
	getParameterOutput, err := ssmService.GetParameter(&ssm.GetParameterInput{Name: &parameterName})
	if err != nil {
		log.Println(err)
		return aws.String("")
	}
	return getParameterOutput.Parameter.Value
}
