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

var snsSqsConfigs map[string]*snsSqsConfig

var (
	snsClients map[string]*sns.SNS
	snsLck     sync.Mutex
)

var (
	sqsClients map[string]*sqs.SQS
	sqsLck     sync.Mutex
)

func init() {
	snsSqsConfigs = map[string]*snsSqsConfig{}
	snsClients = map[string]*sns.SNS{}
	sqsClients = map[string]*sqs.SQS{}
}

func ProvideSnsClient(name string) *sns.SNS {
	snsLck.Lock()
	defer snsLck.Unlock()

	_, ok := snsClients[name]
	if ok {
		return snsClients[name]
	}

	sess, err := getSession(snsSqsConfigs[name].Host, snsSqsConfigs[name].SnsPort)

	if err != nil {
		logErr(err, "could not create sns client: %s")
	}

	snsClients[name] = sns.New(sess)

	return snsClients[name]
}

func ProvideSqsClient(name string) *sqs.SQS {
	sqsLck.Lock()
	defer sqsLck.Unlock()

	_, ok := sqsClients[name]
	if ok {
		return sqsClients[name]
	}

	sess, err := getSession(snsSqsConfigs[name].Host, snsSqsConfigs[name].SqsPort)

	if err != nil {
		logErr(err, "could not create sqs client: %s")
	}

	sqsClients[name] = sqs.New(sess)

	return sqsClients[name]
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
		Env:        []string{services},
		PortBindings: PortBinding{
			"4575/tcp": fmt.Sprint(localConfig.SnsPort),
			"4576/tcp": fmt.Sprint(localConfig.SqsPort),
		},
		HealthCheck: snsSqsHealthcheck,
		OnDestroy:   onDestroy,
	})

	c, err := dockerPool.Client.InspectContainer("gosoline_test_sns_sqs")

	if err != nil {
		logErr(err, "could not inspect container")
	}

	address := c.NetworkSettings.Networks["bridge"].IPAddress

	fmt.Println("using container address", address)

	localConfig.Host = address
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
