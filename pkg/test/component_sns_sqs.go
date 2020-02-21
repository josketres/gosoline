package test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/ory/dockertest/docker"
	"log"
	"regexp"
	"strings"
	"sync"
)

type snsSqsConfig struct {
	Host    string `mapstructure:"host"`
	SnsPort int    `mapstructure:"sns_port"`
	SqsPort int    `mapstructure:"sqs_port"`
}

var snsClients map[string]*sns.SNS
var snsSqsConfigs map[string]*snsSqsConfig
var snsLck sync.Mutex

func init() {
	snsSqsConfigs = map[string]*snsSqsConfig{}
	snsClients = map[string]*sns.SNS{}
}

func onDestroy() {
	snsSqsConfigs = map[string]*snsSqsConfig{}
	snsClients = map[string]*sns.SNS{}
	sqsClients = map[string]*sqs.SQS{}
}

func runSnsSqs(name string, config configInput) {
	wait.Add(1)
	go doRunSnsSqs(name, config)
}

func doRunSnsSqs(name string, configMap configInput) {
	defer wait.Done()
	defer log.Printf("%s component of type %s is ready", name, "sns_sqs")

	localConfig := &snsSqsConfig{}
	unmarshalConfig(configMap, localConfig)
	snsSqsConfigs[name] = localConfig

	services := "SERVICES=" + strings.Join([]string{
		"sns",
		"sqs",
	}, ",")

	runContainer("gosoline_test_sns_sqs", ContainerConfig{
		Repository: "localstack/localstack",
		Tag:        "0.10.7",
		Env:        []string{services, "DEBUG=1"},
		PortBindings: PortBinding{
			"4575/tcp": fmt.Sprint(localConfig.SnsPort),
			"4576/tcp": fmt.Sprint(localConfig.SqsPort),
		},
		HealthCheck: snsSqsHealthcheck,
		OnDestroy:   onDestroy,
	})

}

func snsSqsHealthcheck() error {
	logs := bytes.NewBufferString("")

	err := dockerPool.Client.Logs(docker.LogsOptions{
		Container:    "gosoline_test_sns_sqs",
		OutputStream: logs,
		Stdout:       true,
		Stderr:       true,
	})

	if err != nil {
		return err
	}

	ready, err := regexp.MatchString("Ready\\.", logs.String())

	if err != nil {
		return err
	}

	if !ready {
		return errors.New("localstack services not ready yet")
	}

	return nil
}
