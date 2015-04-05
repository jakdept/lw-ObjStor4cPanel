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

func TestLoadTestingConfig(t *testing.T) {
	configData := []byte(`{"Pwd":"/", "AccessKey": "AccEssKey", "SecretKey": "SecRetKey", "Bucket": "BuKKiT"}`)
	testingConfig := loadTestingConfig(configData)

	assert.Equal(t, "AccEssKey", testingConfig.AccessKey, "the Access Key should be the same in the config")
	assert.Equal(t, "SecRetKey", testingConfig.SecretKey, "the Secret Key should be the same in the config")
	assert.Equal(t, "/", testingConfig.Pwd, "the pwd should be the same in the conf")
	assert.Equal(t, "BuKKiT", testingConfig.Bucket, "the pwd should be the same in the conf")
}

func TestSetupConnection(t *testing.T) {
	configData := []byte(`{"Pwd":"/", "AccessKey": "AccEssKey", "SecretKey": "SecRetKey", "Bucket": "BuKKiT"}`)
	testingConfig := loadTestingConfig(configData)
	connection := SetupConnection(testingConfig)

	assert.Equal(t, "AccEssKey", connection.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", connection.Auth.SecretKey, "the Secret Key should be the same")
	assert.Equal(t, "https://objects.liquidweb.services", connection.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", connection.Region.Name, "the URL should be LW's")
}

func TestSetupBucket(t *testing.T) {
	configData := []byte(`{"Pwd":"/", "AccessKey": "AccEssKey", "SecretKey": "SecRetKey", "Bucket": "BuKKiT"}`)
	testingConfig := loadTestingConfig(configData)
	//connection := SetupConnection(testingConfig)
	bucket := SetupBucket(testingConfig)

	assert.Equal(t, "AccEssKey", bucket.S3.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", bucket.S3.Auth.SecretKey, "the Secret Key should be the same")
	//assert.Equal(t, "https://BuKKiT.objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "https://objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", bucket.S3.Region.Name, "the URL should be LW's")
	assert.Equal(t, "bukkit", bucket.Name, "the name of the bucket is not being set correctly")
}

func TestHiddenConfig(t *testing.T) {
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		t.Error("failed to load my config file testConfig.json")
	}
	testingConfig := loadTestingConfig(data)
	//connection := SetupConnection(testingConfig)
	bucket := SetupBucket(testingConfig)

	assert.Equal(t, testingConfig.AccessKey, bucket.S3.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, testingConfig.SecretKey, bucket.S3.Auth.SecretKey, "the Secret Key should be the same")
	assert.Equal(t, testingConfig.Bucket, bucket.Name, "the name of the bucket is not being set correctly")
	assert.Equal(t, "https://objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", bucket.S3.Region.Name, "the URL should be LW's")
}

func TestValidBucket(t *testing.T) {
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		t.Error("failed to load my config file testConfig.json")
	}
	testingConfig := loadTestingConfig(data)
	connection := SetupConnection(testingConfig)

	_, err = connection.ListBuckets()
	assert.Nil(t, err, "there should be no error listing the buckets")

	bucketExists := ValidBucket(testingConfig, connection)
	assert.True(t, bucketExists, "the bucket should exist within the given space")

	testingConfig.Bucket = "BadBucket"
	bucketExists = ValidBucket(testingConfig, connection)
	assert.False(t, bucketExists, "the BadBucket should not exist within the given space")
}

func ExampleChdir() {
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		return
	}
	testingConfig := loadTestingConfig(data)
	bucket := SetupBucket(testingConfig)

	testingConfig.CmdParams = []string{"/"}
	Chdir(testingConfig, bucket)

	testingConfig.CmdParams = []string{"/folderthatdoesnotexist"}
	Chdir(testingConfig, bucket)

	testingConfig.CmdParams = []string{"/testing"}
	Chdir(testingConfig, bucket)
	// Output:
	// /
	// /folderthatdoesnotexist
	// /testing
}

func ExampleLsdir() {
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		return
	}
	testingConfig := loadTestingConfig(data)
	bucket := SetupBucket(testingConfig)

	//testingConfig.CmdParams = []string{"/"}
	//Lsdir(testingConfig, bucket)

	//testingConfig.CmdParams = []string{"/folderthatdoesnotexist"}
	//Lsdir(testingConfig, bucket)

	testingConfig.CmdParams = []string{"/stuff"}
	Lsdir(testingConfig, bucket)
	// Output
	// this is not the correct output
}
