package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	//"os"
	//"reflect"
	"testing"
)

//var testingConfig runningConfig

func loadTestingConfig(data []byte) runningConfig {
	testingConfig := new(runningConfig)
	err := json.Unmarshal(data, &testingConfig)
	if err != nil {
		log.Printf("Failed to load a test configuration - %s", err)
	}
	return *testingConfig
}

//func TestMain(m *testing.M) {
//func init() {

//os.Exit(m.Run()) // call all the tests in here and then exit
//}
func TestLoadTestingConfig(t *testing.T) {
	//configData := new([]byte)
	//configData = `{"pwd":"/", "accessKey": "AccEssKey", "secretKey": "SecRetKey", "bucket": "BuKKiT"}`
	configData := []byte(`{"pwd":"/", "accessKey": "AccEssKey", "secretKey": "SecRetKey", "bucket": "BuKKiT"}`)
	testingConfig := loadTestingConfig(configData)

	//assert.Equal(t, "runningConfig", string(reflect.TypeOf(testingConfig)), "the config should be of type runningConfig")
	assert.Equal(t, "AccEssKey", testingConfig.accessKey, "the Access Key should be the same in the config")
	assert.Equal(t, "SecRetKey", testingConfig.secretKey, "the Secret Key should be the same in the config")
	assert.Equal(t, "/", testingConfig.pwd, "the pwd should be the same in the conf")
}

func TestSetupConnection(t *testing.T) {
	//configData := new([]byte)
	//configData = `{"pwd":"/", "accessKey": "AccEssKey", "secretKey": "SecRetKey", "bucket": "BuKKiT"}`
	configData := []byte(`{"pwd":"/", "accessKey": "AccEssKey", "secretKey": "SecRetKey", "bucket": "BuKKiT"}`)
	testingConfig := loadTestingConfig(configData)

	connection := SetupConnection(testingConfig)
	assert.Equal(t, "AccEssKey", connection.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", connection.Auth.SecretKey, "the Secret Key should be the same")

	assert.Equal(t, "https://objects.liquidweb.services", connection.Region.S3Endpoint, "the URL should be LW's")
}

/*
func TestSetupConnection(t *testing.T) {
	testingConfig := loadTestingConfig(`{"pwd":"/", "accessKey": "AccEssKey", "secretKey": "SecRetKey", "bucket": "stuff"}`)

	log.Println(reflect.TypeOf(testingConfig))
	log.Println(reflect.TypeOf(testingConfig.accessKey))
	log.Println(testingConfig.accessKey)
	//assert.Equal(t, runningConfig, testingConfig.(type), "the config is of the wrong type")
	//testingConfig.(type)

	connection := SetupConnection(testingConfig)
	assert.Equal(t, testingConfig.accessKey, connection.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, testingConfig.secretKey, connection.Auth.SecretKey, "the Secret Key should be the same")

	assert.Equal(t, "https://objects.liquidweb.services", connection.Region.S3Endpoint, "the URL should be LW's")
}
*/

func TestValidBucket(t *testing.T) {
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		t.Error("failed to load my config file testConfig.json")
	}
	testingConfig := loadTestingConfig(data)
	connection := SetupConnection(testingConfig)

	_, err = connection.ListBuckets()

	assert.Nil(t, err, "there should be no error listing the buckets")

	assert.True(t, ValidBucket(testingConfig, connection), "the bucket should exist within the given space")
}
