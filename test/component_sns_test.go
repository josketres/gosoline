//+build integration

package test_test

import (
	"fmt"
	pkgTest "github.com/applike/gosoline/pkg/test"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func Test_sns_sqs(t *testing.T) {
	err := os.Setenv("AWS_ACCESS_KEY_ID", "a")
	assert.NoError(t, err)

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", "b")
	assert.NoError(t, err)

	pkgTest.Boot("test_configs/config.sns_sqs.test.yml")
	defer pkgTest.Shutdown()

	queueName := "my-queue"
	topicName := "my-topic"

	snsClient := pkgTest.ProvideSnsClient("sns_sqs")
	sqsClient := pkgTest.ProvideSqsClient("sns_sqs")

	// create a topic
	topic, err := snsClient.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String(topicName),
	})

	assert.NoError(t, err)
	assert.NotNil(t, topic)
	if assert.NotNil(t, topic.TopicArn) {
		assert.Equal(t, *topic.TopicArn, fmt.Sprintf("arn:aws:sns:us-east-1:000000000000:%s", topicName))
	}

	// create a queue
	queue, err := sqsClient.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})

	assert.NoError(t, err)
	assert.NotNil(t, queue)
	if assert.NotNil(t, queue.QueueUrl) {
		assert.Equal(t, *queue.QueueUrl, fmt.Sprintf("http://localhost:4576/queue/%s", queueName))
	}

	// wait for queue to be created
	time.Sleep(1 * time.Second)

	// create a topic subscription
	subscriptionOutput, err := snsClient.Subscribe(&sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		Endpoint: queue.QueueUrl,
		TopicArn: topic.TopicArn,
	})

	assert.NoError(t, err)
	assert.NotNil(t, subscriptionOutput)
	if assert.NotNil(t, subscriptionOutput.SubscriptionArn) {
		assert.Contains(t, *subscriptionOutput.SubscriptionArn, fmt.Sprintf("arn:aws:sns:us-east-1:000000000000:%s:", topicName))
	}

	// send a message to a topic
	publishOutput, err := snsClient.Publish(&sns.PublishInput{
		Message:  aws.String("Hello there."),
		TopicArn: topic.TopicArn,
	})

	assert.NoError(t, err)
	if assert.NotNil(t, publishOutput) {
		assert.NotNil(t, publishOutput.MessageId)
	}

	// wait for localstack to forward the message to sqs
	time.Sleep(1 * time.Second)

	// receive the message from sqs
	receiveOutput, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: queue.QueueUrl,
	})

	assert.NoError(t, err)
	if !assert.NotNil(t, receiveOutput) {
		return
	}
	if assert.Len(t, receiveOutput.Messages, 1) {
		assert.Contains(t, *receiveOutput.Messages[0].Body, "Hello there.")
	}
}
