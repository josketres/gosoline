package test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"log"
	"sync"
)

type kinesisConfig struct {
	Port int `mapstructure:"port"`
}

var kinesisClients map[string]*kinesis.Kinesis
var kinesisConfigs map[string]*kinesisConfig
var kinesisLck sync.Mutex

func init() {
	kinesisConfigs = map[string]*kinesisConfig{}
	kinesisClients = map[string]*kinesis.Kinesis{}
}

func ProvideKinesisClient(name string) *kinesis.Kinesis {
	kinesisLck.Lock()
	defer kinesisLck.Unlock()

	_, ok := kinesisClients[name]
	if ok {
		return kinesisClients[name]
	}

	sess, err := getSession(kinesisConfigs[name].Port)

	if err != nil {
		logErr(err, "could not create kinesis client: %s")
	}

	kinesisClients[name] = kinesis.New(sess)

	return kinesisClients[name]
}

func runKinesis(name string, config configInput) {
	wait.Add(1)
	go doRunKinesis(name, config)
}

func doRunKinesis(name string, configMap configInput) {
	defer wait.Done()
	defer log.Printf("%s component of type %s is ready", name, "kinesis")

	localConfig := &kinesisConfig{}
	unmarshalConfig(configMap, localConfig)
	kinesisConfigs[name] = localConfig

	runContainer("gosoline-test-kinesis", ContainerConfig{
		Repository: "localstack/localstack",
		Tag:        "0.10.3",
		Env: []string{
			"SERVICES=kinesis",
		},
		PortBindings: PortBinding{
			"4568/tcp": fmt.Sprint(localConfig.Port),
		},
		HealthCheck: func() error {
			kinesisClient := ProvideKinesisClient(name)
			_, err := kinesisClient.ListStreams(&kinesis.ListStreamsInput{})

			return err
		},
	})
}
