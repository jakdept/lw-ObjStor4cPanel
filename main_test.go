package main

import (
	"os"

	"github.com/stretchr/testify/assert"
	//"os"
	//"reflect"
	"testing"
)

//var testingConfig runningConfig

func loadTestingConfig(t *testing.T) *runningConfig {
	testingConfig := new(runningConfig)
	testingConfig.Pwd = os.Getenv("PWD")
	testingConfig.AccessKey = os.Getenv("ACCESSKEY")
	testingConfig.SecretKey = os.Getenv("SECRETKEY")
	testingConfig.bucketName = os.Getenv("BUCKET")
	return testingConfig
}

func TestGetConfig(t *testing.T) {
	os.Args = []string{"junk", "hackers", "/pwd/", "command", "args", "are", "here", "hack", "the", "gibson", "bucket", "access"}
	err := os.Setenv("PASSWORD", "sekret")
	assert.NoError(t, err)

	config := getConfig()

	expectedConfig := runningConfig{
		Command:    "hackers",
		Pwd:        "",
		bucketName: "bucket",
		AccessKey:  "access",
		SecretKey:  "sekret",
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
	testingConfig := &runningConfig{
		Pwd:        "/",
		AccessKey:  "AccEssKey",
		SecretKey:  "SecRetKey",
		bucketName: "BuKKiT",
	}
	connection, err := testingConfig.SetupConnection()
	assert.NoError(t, err)

	assert.Equal(t, "AccEssKey", connection.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", connection.Auth.SecretKey, "the Secret Key should be the same")
	assert.Equal(t, "https://objects.liquidweb.services", connection.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", connection.Region.Name, "the URL should be LW's")
}

func TestSetupBucket(t *testing.T) {
	testingConfig := &runningConfig{
		Pwd:        "/",
		AccessKey:  "AccEssKey",
		SecretKey:  "SecRetKey",
		bucketName: "BuKKiT",
	}
	err := testingConfig.SetupBucket()
	assert.NoError(t, err)

	assert.Equal(t, "AccEssKey", testingConfig.bucket.S3.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, "SecRetKey", testingConfig.bucket.S3.Auth.SecretKey, "the Secret Key should be the same")
	//assert.Equal(t, "https://BuKKiT.objects.liquidweb.services", bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "https://objects.liquidweb.services", testingConfig.bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", testingConfig.bucket.S3.Region.Name, "the URL should be LW's")
	assert.Equal(t, "bukkit", testingConfig.bucket.Name, "the name of the bucket is not being set correctly")
}

func TestHiddenConfig(t *testing.T) {
	testingConfig := loadTestingConfig(t)
	//connection := SetupConnection(testingConfig)
	err := testingConfig.SetupBucket()
	assert.NoError(t, err)

	assert.Equal(t, testingConfig.AccessKey, testingConfig.bucket.S3.Auth.AccessKey, "the Access Key should be the same")
	assert.Equal(t, testingConfig.SecretKey, testingConfig.bucket.S3.Auth.SecretKey, "the Secret Key should be the same")
	assert.Equal(t, testingConfig.bucketName, testingConfig.bucket.Name, "the name of the bucket is not being set correctly")
	assert.Equal(t, "https://objects.liquidweb.services", testingConfig.bucket.S3.Region.S3Endpoint, "the URL should be LW's")
	assert.Equal(t, "liquidweb", testingConfig.bucket.S3.Region.Name, "the URL should be LW's")
}

func TestValidBucket(t *testing.T) {
	testingConfig := loadTestingConfig(t)
	connection, err := testingConfig.SetupConnection()
	assert.NoError(t, err)

	_, err = connection.ListBuckets()
	assert.NoError(t, err)

	bucketExists, err := ValidBucket(testingConfig.bucketName, connection)
	assert.True(t, bucketExists, "the bucket should exist within the given space")
	assert.NoError(t, err)

	bucketExists, err = ValidBucket("BadBucket", connection)
	assert.False(t, bucketExists, "the BadBucket should not exist within the given space")
	assert.NoError(t, err)
}

func ExampleChdir() {
	testingConfig := runningConfig{
		Pwd:        "/",
		AccessKey:  "AccEssKey",
		SecretKey:  "SecRetKey",
		bucketName: "BuKKiT",
	}

	testingConfig.SetupBucket()

	testingConfig.CmdParams = []string{"/"}
	testingConfig.Chdir(testingConfig.CmdParams[0])

	testingConfig.CmdParams = []string{"/folderthatdoesnotexist"}
	testingConfig.Chdir(testingConfig.CmdParams[0])

	testingConfig.CmdParams = []string{"/testing"}
	testingConfig.Chdir(testingConfig.CmdParams[0])
	// Output:
	// /
	// /folderthatdoesnotexist
	// /testing
}

func TestLsdir(t *testing.T) {
	testingConfig := loadTestingConfig(t)
	err := testingConfig.SetupBucket()
	assert.NoError(t, err)

	testingConfig.CmdParams = []string{"/"}
	err = Lsdir(*testingConfig, testingConfig.bucket)
	assert.NoError(t, err)

	//testingConfig.CmdParams = []string{"/folderthatdoesnotexist"}
	//Lsdir(testingConfig, bucket)

	//testingConfig.CmdParams = []string{"/stuff"}
	//Lsdir(testingConfig, bucket)
	// Output
	// this is not the correct output
}
