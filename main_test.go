package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var testingConfig runningConfig

func TestMain(m *testing.M) {

	configContents, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		log.Printf("Failed to load my test configuration file at testConfig.json - %s", err)
	}
	err = json.Unmarshal(configContents, &testingConfig)
	if err != nil {
		log.Printf("Failed to load my test configuration file at testConfig.json - %s", err)
	}
	os.Exit(m.Run()) // call all the tests in here and then exit
}

func TestSetupConnection(t *testing.T) {

	connection := SetupConnection(testingConfig)
	assert.Equal(t, connection.Auth.AccessKey, testingConfig.accessKey, "the Access Key should be the same")
	assert.Equal(t, connection.Auth.SecretKey, testingConfig.secretKey, "the Secret Key should be the same")

	assert.Equal(t, connection.Region.S3Endpoint, "https://objects.liquidweb.services", "the URL should be LW's")
}

func TestListBuckets(t *testing.T) {
	connection := SetupConnection(testingConfig)

	assert.True(t, ValidBucket(connection, testingConfig), "the bucket should exist within the given space")
}
