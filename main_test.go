package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	//"os"
	"testing"
)

var testingConfig runningConfig

func loadTestingConfig() *runningConfig {
	configContents, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		log.Printf("Failed to load my test configuration file at testConfig.json - %s", err)
	}
	testConfig := new(runningConfig)
	err = json.Unmarshal(configContents, &testingConfig)
	if err != nil {
		log.Printf("Failed to load my test configuration file at testConfig.json - %s", err)
	}
	return testConfig
}

//func TestMain(m *testing.M) {
//func init() {

//os.Exit(m.Run()) // call all the tests in here and then exit
//}

func TestSetupConnection(t *testing.T) {
	testingConfig := loadTestingConfig()

	connection := SetupConnection(testingConfig)
	assert.Equal(t, testingConfig.accessKey, connection.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, testingConfig.secretKey, connection.Auth.SecretKey, "the Secret Key should be the same")

	assert.Equal(t, "https://objects.liquidweb.services", connection.Region.S3Endpoint, "the URL should be LW's")
}

func TestValidBucket(t *testing.T) {
	testingConfig := loadTestingConfig()
	connection := SetupConnection(testingConfig)

	_, err := connection.ListBuckets()

	assert.Nil(t, err, "there should be no error listing the buckets")

	assert.True(t, ValidBucket(connection, testingConfig), "the bucket should exist within the given space")
}
