//+build integration

package test_test

import (
	pkgTest "github.com/applike/gosoline/pkg/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_only_container(t *testing.T) {
	err := os.Setenv("AWS_ACCESS_KEY_ID", "a")
	assert.NoError(t, err)

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", "b")
	assert.NoError(t, err)

	pkgTest.Boot("test_configs/config.sns_sqs.test.yml")
	defer pkgTest.Shutdown()
}
