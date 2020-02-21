package test

import (
	"fmt"
	"github.com/applike/gosoline/pkg/mdl"
	"github.com/aws/aws-sdk-go/service/sns"
	"time"
)

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

func snsHealthcheck(name string) func() error {
	return func() error {
		snsClient := ProvideSnsClient(name)
		topicName := "healthcheck"

		topic, err := snsClient.CreateTopic(&sns.CreateTopicInput{
			Name: mdl.String(topicName),
		})

		if err != nil {
			return err
		}

		listTopics, err := snsClient.ListTopics(&sns.ListTopicsInput{})

		if err != nil {
			return err
		}

		if len(listTopics.Topics) != 1 {
			return fmt.Errorf("topic list should contain exactly 1 entry, but contained %d", len(listTopics.Topics))
		}

		_, err = snsClient.DeleteTopic(&sns.DeleteTopicInput{TopicArn: topic.TopicArn})

		if err != nil {
			return err
		}

		// wait for topic to be really deleted (race condition)
		for {
			listTopics, err := snsClient.ListTopics(&sns.ListTopicsInput{})

			if err != nil {
				return err
			}

			if len(listTopics.Topics) == 0 {
				return nil
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}
