package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/stretchr/testify/assert"
	//"os"
	//"reflect"
	"testing"
)

//var testingConfig runningConfig

func loadTestingConfig(t *testing.T) runningConfig {
	data, err := ioutil.ReadFile("testConfig.json")
	if err != nil {
		t.Skip("failed to load my config file testConfig.json")
	}
	testingConfig := new(runningConfig)
	err = json.Unmarshal(data, &testingConfig)
	if err != nil {
		t.Skip("Failed to parse test configuration -", err.Error())
	}
	return *testingConfig
}

func TestGetConfig(t *testing.T) {
	os.Args = []string{"junk", "hackers", "/pwd/", "command", "args", "are", "here", "hack", "the", "gibson", "bucket", "access"}
	err := os.Setenv("PASSWORD", "sekret")
	assert.NoError(t, err)

	config := getConfig()

	expectedConfig := runningConfig{
		Command:   "hackers",
		Pwd:       "",
		Bucket:    "bucket",
		AccessKey: "access",
		SecretKey: "sekret",
		CmdParams: []string{
			"command",
			"args",
			"are",
			"here",
			"hack",
			"the",
			"gibson",
		},
	}

	assert.Equal(t, expectedConfig, config)
}

func TestSetupConnection(t *testing.T) {
	testingConfig := runningConfig{
		Pwd:       "/",
		AccessKey: "AccEssKey",
		SecretKey: "SecRetKey",
		Bucket:    "BuKKiT",
	}
	connection, err := SetupConnection(testingConfig)
	assert.NoError(t, err)

	assert.Equal(t, "AccEssKey", connection.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", connection.Auth.SecretKey, "the Secret Key should be the same")
	assert.Equal(t, "https://objects.liquidweb.services", connection.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", connection.Region.Name, "the URL should be LW's")
}

func TestSetupBucket(t *testing.T) {
	testingConfig := runningConfig{
		Pwd:       "/",
		AccessKey: "AccEssKey",
		SecretKey: "SecRetKey",
		Bucket:    "BuKKiT",
	}
	bucket, err := SetupBucket(testingConfig)
	assert.NoError(t, err)

	assert.Equal(t, "AccEssKey", bucket.S3.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", bucket.S3.Auth.SecretKey, "the Secret Key should be the same")
	//assert.Equal(t, "https://BuKKiT.objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "https://objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", bucket.S3.Region.Name, "the URL should be LW's")
	assert.Equal(t, "bukkit", bucket.Name, "the name of the bucket is not being set correctly")
}

func TestHiddenConfig(t *testing.T) {
	testingConfig := loadTestingConfig(t)
	//connection := SetupConnection(testingConfig)
	bucket, err := SetupBucket(testingConfig)
	assert.NoError(t, err)

	assert.Equal(t, testingConfig.AccessKey, bucket.S3.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, testingConfig.SecretKey, bucket.S3.Auth.SecretKey, "the Secret Key should be the same")
	assert.Equal(t, testingConfig.Bucket, bucket.Name, "the name of the bucket is not being set correctly")
	assert.Equal(t, "https://objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", bucket.S3.Region.Name, "the URL should be LW's")
}

func TestValidBucket(t *testing.T) {
	testingConfig := loadTestingConfig(t)
	connection, err := SetupConnection(testingConfig)
	assert.NoError(t, err)

	_, err = connection.ListBuckets()
	assert.NoError(t, err)

	bucketExists, err := ValidBucket(testingConfig, connection)
	assert.True(t, bucketExists, "the bucket should exist within the given space")
	assert.NoError(t, err)

	testingConfig.Bucket = "BadBucket"
	bucketExists, err = ValidBucket(testingConfig, connection)
	assert.False(t, bucketExists, "the BadBucket should not exist within the given space")
	assert.NoError(t, err)
}

func ExampleChdir() {
	testingConfig := runningConfig{
		Pwd:       "/",
		AccessKey: "AccEssKey",
		SecretKey: "SecRetKey",
		Bucket:    "BuKKiT",
	}
	bucket, _ := SetupBucket(testingConfig)

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

// func TestLsdir(t *testing.T) {
// 	log.Println("in ExampleLsdir")
// 	data, err := ioutil.ReadFile("testConfig.json")
// 	if err != nil {
// 		return
// 	}
// 	testingConfig := loadTestingConfig(data)
// 	bucket := SetupBucket(testingConfig)

// 	testingConfig.CmdParams = []string{"/"}
// 	Lsdir(testingConfig, bucket)

// 	//testingConfig.CmdParams = []string{"/folderthatdoesnotexist"}
// 	//Lsdir(testingConfig, bucket)

// 	//testingConfig.CmdParams = []string{"/stuff"}
// 	//Lsdir(testingConfig, bucket)
// 	// Output
// 	// this is not the correct output
// }
