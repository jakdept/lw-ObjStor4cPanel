package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	//"os"
	//"reflect"
	"testing"

	"github.com/sebdah/goldie"
)

func init() {
	goldie.FixtureDir = "testdata/fixtures"
}

func loadTestingConfig(t *testing.T) (*runningConfig, bytes.Buffer) {
	testingConfig := new(runningConfig)
	testingConfig.Pwd = os.Getenv("PWD")
	testingConfig.AccessKey = os.Getenv("ACCESSKEY")
	testingConfig.SecretKey = os.Getenv("SECRETKEY")
	testingConfig.bucketName = os.Getenv("BUCKET")
	if testingConfig.Pwd == "" ||
		testingConfig.AccessKey == "" ||
		testingConfig.SecretKey == "" ||
		testingConfig.bucketName == "" {
		t.Skip("missing test configuration")
	}

	// set up something to capture output
	outputBuf := bytes.Buffer{}
	testingConfig.output = &outputBuf

	return testingConfig, outputBuf
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
		output:     os.Stdout,
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
	testingConfig, _ := loadTestingConfig(t)
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
	testingConfig, _ := loadTestingConfig(t)
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

func TestChdir(t *testing.T) {
	outputBuf := bytes.Buffer{}
	testingConfig := runningConfig{
		Pwd:        "/",
		AccessKey:  "AccEssKey",
		SecretKey:  "SecRetKey",
		bucketName: "BuKKiT",
		output:     &outputBuf,
	}

	testingConfig.SetupBucket()

	testingConfig.CmdParams = []string{"/"}
	testingConfig.Chdir(testingConfig.CmdParams[0])

	testingConfig.CmdParams = []string{"/folderthatdoesnotexist"}
	testingConfig.Chdir(testingConfig.CmdParams[0])

	testingConfig.CmdParams = []string{"/testing"}
	testingConfig.Chdir(testingConfig.CmdParams[0])

	goldie.Assert(t, t.Name(), outputBuf.Bytes())
}

func TestLsdir(t *testing.T) {
	testingConfig, outputBuf := loadTestingConfig(t)
	err := testingConfig.SetupBucket()
	assert.NoError(t, err)

	err = testingConfig.Lsdir("/")
	assert.NoError(t, err)

	goldie.Assert(t, t.Name(), outputBuf.Bytes())
}

func TestRemoteFolder(t *testing.T) {
	files := []string{
		"thank_you_for_not_loitering.jpg",
		"database.tar.gz",
	}

	for i := 0; i < len(files); i++ {
		files[i] = path.Clean(files[i])
	}

	testingConfig, outputBuf := loadTestingConfig(t)
	err := testingConfig.SetupBucket()
	assert.NoError(t, err)

	testingConfig.output = os.Stderr

	err = testingConfig.Rmdir("/testdata")
	assert.NoError(t, err)

	err = testingConfig.Mkdir("/testdata")
	assert.NoError(t, err)

	for _, file := range files {
		local := filepath.Join("testdata", file)
		remote := filepath.Join("testdata", file)
		fmt.Fprintf(testingConfig.output, "\nputting local [%s] into remote [%s]\n", local, remote)

		err = testingConfig.magicPut(remote, local)
		assert.NoError(t, err)
		err = testingConfig.Lsdir("testdata")
		assert.NoError(t, err)
	}
	// #TODO# remove me
	// ralph(testingConfig, "", "", "")

	tmpdir, err := ioutil.TempDir("", "cPanel_backup_transporter")
	assert.NoError(t, err)

	// #TODO# remove me
	// fmt.Fprintf(testingConfig.output, "working with temp directory [%s]\n", tmpdir)

	for _, file := range files {
		local := filepath.Join(tmpdir, file)
		remote := filepath.Join("testdata", file)
		// #TODO# remove me
		fmt.Fprintf(testingConfig.output, "pulling remote [%s] into local [%s]\n", remote, local)

		err = testingConfig.magicGet(local, remote)
		assert.NoError(t, err)
	}

	for _, file := range files {
		fi, err := os.Stat(filepath.Join(tmpdir, file))
		assert.NoError(t, err)

		fmt.Fprintf(testingConfig.output, "[%d] %s\n", fi.Size(), fi.Name())
	}

	fmt.Fprintln(testingConfig.output, "contents before removal")

	err = testingConfig.Lsdir("testdata")
	assert.NoError(t, err)

	err = testingConfig.Rmdir("testdata")
	assert.NoError(t, err)

	fmt.Fprintln(testingConfig.output, "contents after removal")

	err = testingConfig.Lsdir("testdata")
	assert.NoError(t, err)

	t.Skip()

	goldie.Assert(t, t.Name(), outputBuf.Bytes())
}

func ralph(c *runningConfig, prefix, delim, marker string) {
	stuff, err := c.bucket.List(prefix, delim, marker, 1000)
	if err == nil {
		spew.Dump(stuff)
	}
}
