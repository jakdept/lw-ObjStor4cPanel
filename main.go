package main

// cPanel backup transport helper for Liquidweb Object Storage
// By Jack Hayhurst

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	// "gopkg.in/amz.v1/aws"
	// "gopkg.in/amz.v1/s3"
)

const contentType = "application/octet-stream"

const pagesize = 1000

const chunkSize = 33554432 // 32M in bytes

type runningConfig struct {
	Command   string
	Pwd       string
	Bucket    string
	CmdParams []string
	AccessKey string
	SecretKey string
}

func getConfig() runningConfig {
	config := new(runningConfig)
	// parameters are passed as:
	// binary Command Pwd [CmdParams ...] Bucket AccessKey
	config.Command = os.Args[1]
	//Pwd := os.Args[2]
	config.Bucket = os.Args[len(os.Args)-2]
	config.AccessKey = os.Args[len(os.Args)-1]

	config.CmdParams = os.Args[3 : len(os.Args)-2]

	// SecretKey is passed via enviroment variable
	config.SecretKey = os.Getenv("PASSWORD")

	return *config
}

func SetupConnection(config runningConfig) *s3.S3 {
	bucketRegion := aws.Region{
		Name:       "liquidweb",
		S3Endpoint: "https://objects.liquidweb.services",
		//S3Endpoint: config.url,
	}

	bucketAuth, err := aws.GetAuth(config.AccessKey, config.SecretKey)

	reportError("Ran into a problem creating the authentication with AccessKey %s", config.AccessKey, err)

	return s3.New(bucketAuth, bucketRegion)
}

func ValidBucket(config runningConfig, connection *s3.S3) bool {
	allBuckets, err := connection.ListBuckets()
	reportError("Could not retrieve buckets from %s", "objstor", err)

	bucketExists := false

	for _, Bucket := range allBuckets.Buckets {
		if Bucket.Name == config.Bucket {
			bucketExists = true
		}
	}
	return bucketExists
}

func SetupBucket(config runningConfig, connection s3.S3) s3.Bucket {

	Bucket := *connection.Bucket(config.Bucket)

	return Bucket
}

func callFunc(config runningConfig, Bucket s3.Bucket) {
	// call the function with the name of the Command that you got

	switch config.Command {
	case "get":
		magicGet(config, Bucket)
	case "put":
		magicPut(config, Bucket)
	case "ls":
		Lsdir(config, Bucket)
	case "mkdir":
	case "chdir":
		Chdir(config, Bucket)
	case "rmdir":
		rmdir(config, Bucket)
	case "delete":
		delete(config, Bucket)
	}
}

func reportError(message string, messageSub string, err error) {
	if err != nil {
		log.Printf(message, messageSub)
		log.Println(err.Error())
		os.Exit(1)
	}
	return
}

// does almost nothing - not required, but must return the path
// cli: `binary` `chdir` `Pwd` `path` `bucketName` `username`
func Chdir(config runningConfig, Bucket s3.Bucket) {
	_, err := fmt.Println(config.CmdParams[0])
	reportError("failed to print the given path %s", config.CmdParams[0], err)
}

// lists the content of a directory on the remote system
// cli: `binary` `ls` `Pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func Lsdir(config runningConfig, Bucket s3.Bucket) {
	items, err := Bucket.List(config.CmdParams[0], "", "", pagesize)
	reportError("Failed listing contents of the Bucket behind the path %s", config.CmdParams[0], err)
	for _, target := range items.Contents {
		// prints out in the format defined by:
		// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
		_, err = fmt.Printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
		reportError("Failed displaying the file %s", target.Key, err)
	}
}

// Gets a file from the remote location and puts it on the local system
// cli: `binary` `get` `Pwd `Remote file` `local file` `bucketName` `username`
// passed to this is ["remote file", "local file"]
func get(config runningConfig, Bucket s3.Bucket) {
	//data := new(Buffer)
	data, err := Bucket.Get(config.CmdParams[0])
	reportError("Caught an error loading the remote file %s", config.CmdParams[0], err)
	err = ioutil.WriteFile(config.CmdParams[1], data, 0644)
	reportError("Caught an writing the local file %s", config.CmdParams[1], err)
}

func magicGet(config runningConfig, Bucket s3.Bucket) {
	// open up the output file for writing
	outFile, err := os.Create(config.CmdParams[1])
	defer outFile.Close()
	reportError("Caught an writing the local file %s", config.CmdParams[1], err)

	// open up the remote file for reading
	dataResponse, err := Bucket.GetResponse(config.CmdParams[0])
	defer dataResponse.Body.Close()
	reportError("Caught an error loading the remote file %s", config.CmdParams[0], err)

	// copy all bytes, without loading stuff in memory, then defer close happen
	_, err = io.Copy(outFile, dataResponse.Body)
	reportError("Caught an writing the local file %s", config.CmdParams[1], err)
}

// puts a file from the local location to a remote location
// cli: `binary` `put` `Pwd `local file` `remote file` `bucketName` `username`
func put(config runningConfig, Bucket s3.Bucket) {
	//data := new(Buffer)
	data, err := ioutil.ReadFile(config.CmdParams[0])
	reportError("Caught an error loading the local file %s", config.CmdParams[0], err)
	err = Bucket.Put(config.CmdParams[1], data, contentType, "0644")
	reportError("Caught an error saving the remote file %s", config.CmdParams[1], err)
}

// puts a file from the local location to a remote location by pieces
// cli: `binary` `put` `Pwd `local file` `remote file` `bucketName` `username`
func magicPut(config runningConfig, Bucket s3.Bucket) {
	// open the file to be transferred
	file, err := os.Open(config.CmdParams[0])
	defer file.Close()
	reportError("Caught an error opening the local file %s", config.CmdParams[0], err)

	bytes := make([]byte, chunkSize)
	buffer := bufio.NewReader(file)
	// at most, buffer.Read can only read len(bytes) bytes
	_, err = buffer.Read(bytes)
	reportError("Had an issue reading bytes from the local file %s", config.CmdParams[0], err)

	// determine the filetype based on the bytes you read
	filetype := http.DetectContentType(bytes)

	// set up for multipart upload
	multiUploader, err := Bucket.InitMulti(config.CmdParams[1], filetype, s3.ACL("private"))
	reportError("Had an issue opening the remote file %s for writing", config.CmdParams[1], err)

	// upload all of the file in pieces
	parts, err := multiUploader.PutAll(file, chunkSize)
	reportError("Had an issue putting chunks in the remote file %s", config.CmdParams[1], err)

	// complete the file
	err = multiUploader.Complete(parts)
	reportError("Issue completing the file %s in the Bucket", config.CmdParams[1], err)

	return
}

// removes everything under the given path on the remote Bucket
// cli: `binary` `rmdir` `Pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func rmdir(config runningConfig, Bucket s3.Bucket) {
	items, err := Bucket.List(config.CmdParams[0], "", "", pagesize)
	reportError("Failed listing contents of the Bucket behind the path %s", config.CmdParams[0], err)
	for len(items.Contents) > 0 {
		for _, target := range items.Contents {
			err = Bucket.Del(target.Key)
			reportError("Failed removing the target %s from the Bucket", target.Key, err)
		}
		items, _ = Bucket.List(config.CmdParams[0], "", "", pagesize)
	}
}

// removes a file at a given location
// cli: `delete` `rmdir` `Pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func delete(config runningConfig, Bucket s3.Bucket) {
	err := Bucket.Del(config.CmdParams[0])
	reportError("failed to delete file %s", config.CmdParams[0], err)
}

func main() {

	config := getConfig()

	connection := SetupConnection(config)
	Bucket := SetupBucket(config, *connection)

	callFunc(config, Bucket)
}
