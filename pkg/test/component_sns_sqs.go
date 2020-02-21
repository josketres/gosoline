package test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/ory/dockertest/docker"
	"github.com/thoas/go-funk"
	"log"
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

	services := []string{
		"sns",
		"sqs",
	}

	envVariables := "SERVICES=" + strings.Join(services, ",")

	runContainer("gosoline_test_sns_sqs", ContainerConfig{
		Repository: "localstack/localstack",
		Tag:        "0.10.7",
		Env:        []string{envVariables},
		PortBindings: PortBinding{
			"4575/tcp": fmt.Sprint(localConfig.SnsPort),
			"4576/tcp": fmt.Sprint(localConfig.SqsPort),
		},
		HealthCheck: snsSqsHealthcheck(name),
		OnDestroy:   onDestroy,
	})

	c, _ := dockerPool.Client.InspectContainer("gosoline_test_sns_sqs")

	funk.ForEach(c.NetworkSettings.Networks, func(_ string, elem docker.ContainerNetwork) {
		localConfig.Host = elem.IPAddress
		log.Println(fmt.Sprintf("set Host to %s", localConfig.Host))
	})
}

func snsSqsHealthcheck(name string) func() error {
	return func() error {

		c, _ := dockerPool.Client.InspectContainer("gosoline_test_sns_sqs")

		log.Println("networks: ", funk.Keys(c.NetworkSettings.Networks))

		err := snsHealthcheck(name)()

		if err != nil {
			return err
		}

		return sqsHealthcheck(name)()
	}
}
