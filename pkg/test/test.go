package test

import (
	"fmt"
	"github.com/ory/dockertest"
	"log"
	"sync"
	"time"
)

var err error
var wait sync.WaitGroup
var dockerPool *dockertest.Pool
var dockerResources []*dockertest.Resource
var configs []*ContainerConfig
var cfgFilename = "config.test.yml"

func init() {
	dockerPool, err = dockertest.NewPool("")
	dockerPool.MaxWait = 2 * time.Minute
	dockerResources = make([]*dockertest.Resource, 0)

	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
}

func Pool() *dockertest.Pool {
	return dockerPool
}

func logErr(err error, msg string) {
	Shutdown()
	log.Println(msg)
	log.Fatal(err)
}

func Boot(configFilenames ...string) {
	if len(configFilenames) == 0 {
		configFilenames = append(configFilenames, cfgFilename)
	}

	for _, filename := range configFilenames {
		log.Println(fmt.Sprintf("booting configuration %s", filename))
		bootFromFile(filename)
	}

	log.Println("test environment up and running")
	fmt.Println()
}

func bootFromFile(filename string) {
	config := readConfig(filename)

	for _, mockConfig := range config.Mocks {
		bootComponent(mockConfig)

		wait.Wait()
	}
}

func bootComponent(mockConfig configInput) {
	component := mockConfig["component"]
	name := mockConfig["name"].(string)

	switch component {
	case "cloudwatch":
		runCloudwatch(name, mockConfig)
	case "dynamodb":
		runDynamoDb(name, mockConfig)
	case "elasticsearch":
		runElasticsearch(name, mockConfig)
	case "kinesis":
		runKinesis(name, mockConfig)
	case "mysql":
		runMysql(name, mockConfig)
	case "redis":
		runRedis(name, mockConfig)
	case "sns_sqs":
		runSnsSqs(name, mockConfig)
	case "wiremock":
		runWiremock(name, mockConfig)
	default:
		err := fmt.Errorf("unknown component '%s'", component)
		logErr(err, err.Error())
	}
}

func Shutdown() {
	for _, res := range dockerResources {
		if err := dockerPool.Purge(res); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
	for _, cfg := range configs {
		if cfg.OnDestroy == nil {
			continue
		}
		cfg.OnDestroy()
	}

	dockerResources = make([]*dockertest.Resource, 0)
}
