package test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/ory/dockertest/docker"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

type snsSqsConfig struct {
	Host    string `mapstructure:"host"`
	SnsPort int    `mapstructure:"sns_port"`
	SqsPort int    `mapstructure:"sqs_port"`
}

var snsSqsConfigs map[string]*snsSqsConfig

var clients = simpleCache{}

func init() {
	snsSqsConfigs = map[string]*snsSqsConfig{}
	clients = simpleCache{}
}

// do we really need this ?
func onDestroy() {
	snsSqsConfigs = map[string]*snsSqsConfig{}
	clients = simpleCache{}
}

func ProvideSnsClient(name string) *sns.SNS {
	return clients.New("sns-"+name, func() interface{} {
		sess, err := getSession(snsSqsConfigs[name].Host, snsSqsConfigs[name].SnsPort)

		if err != nil {
			logErr(err, "could not create sns client: %s")
		}

		return sns.New(sess)
	}).(*sns.SNS)
}

func ProvideSqsClient(name string) *sqs.SQS {
	return clients.New("sqs-"+name, func() interface{} {
		sess, err := getSession(snsSqsConfigs[name].Host, snsSqsConfigs[name].SqsPort)

		if err != nil {
			logErr(err, "could not create sqs client: %s")
		}

		return sqs.New(sess)
	}).(*sqs.SQS)
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

	if isReachable(address + ":4575") {
		fmt.Println("overriding host", address)
		snsSqsConfigs[name].Host = address
	}
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

func isReachable(address string) bool {
	timeout := time.Duration(5) * time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("connection established between localhost and", address)
	fmt.Println("remote address", conn.RemoteAddr().String())
	fmt.Println("local address", conn.LocalAddr().String())

	return true
}
