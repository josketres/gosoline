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

var httpClient = &http.Client{
	Timeout: time.Minute,
}

func getSession(host string, port int) (*session.Session, error) {
	endpoint := fmt.Sprintf("http://%s:%d", host, port)

	s := sessions.New(endpoint, func() interface{} {
		return createNewSession(endpoint)
	})

	return s.(*session.Session), nil
}

func createNewSession(endpoint string) interface{} {
	log.Println("creating new aws session for endpoint : " + endpoint)

	config := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		MaxRetries:                    mdl.Int(30),
		Region:                        aws.String(endpoints.EuCentral1RegionID),
		Endpoint:                      aws.String(endpoint),
		HTTPClient:                    httpClient,
		LogLevel:                      aws.LogLevel(aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors),
	}

	newSession, err := session.NewSession(config)

	if err != nil {
		panic(err)
	}

	return newSession
}
