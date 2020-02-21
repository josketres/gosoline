package test

import (
	"fmt"
	"github.com/applike/gosoline/pkg/mdl"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"net/http"
	"time"
)

var sessions = simpleCache{}

func getSession(host string, port int) (*session.Session, error) {
	endpoint := fmt.Sprintf("http://%s:%d", host, port)

	s := sessions.New(endpoint, func() interface{} {
		return createNewSession(endpoint)
	})

	return s.(*session.Session), nil
}

func createNewSession(endpoint string) interface{} {
	log.Println("creating new aws session for endpoint : " + endpoint)

	httpTimeout := time.Minute

	logLevel := aws.LogDebug

	config := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		MaxRetries:                    mdl.Int(30),
		Region:                        aws.String(endpoints.EuCentral1RegionID),
		Endpoint:                      aws.String(endpoint),
		HTTPClient: &http.Client{
			Timeout: httpTimeout,
		},
		LogLevel: &logLevel,
	}

	newSession, err := session.NewSession(config)

	if err != nil {
		panic(err)
	}

	return newSession
}
