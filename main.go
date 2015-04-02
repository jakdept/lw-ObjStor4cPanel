package main

// cPanel backup transport helper for Liquidweb Object Storage
// By Jack Hayhurst

// ## TODO buffering needs to be added for file uploads and downloads
// https://www.socketloop.com/tutorials/golang-upload-big-file-larger-than-100mb-to-aws-s3-with-multipart-upload

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
	command   string
	pwd       string
	bucket    string
	cmdParams []string
	accessKey string
	secretKey string
}

func getConfig() *runningConfig {
	config := new(runningConfig)
	// parameters are passed as:
	// binary command pwd [cmdParams ...] bucket accessKey
	config.command = os.Args[1]
	//pwd := os.Args[2]
	config.bucket = os.Args[len(os.Args)-2]
	config.accessKey = os.Args[len(os.Args)-1]

	config.cmdParams = os.Args[3 : len(os.Args)-2]

	// secretKey is passed via enviroment variable
	config.secretKey = os.Getenv("PASSWORD")

	return config
}

func SetupConnection(config *runningConfig) *s3.S3 {

	/*
		// set up the required bucket
		bucketRegion := aws.Region{
			"LiquidWeb", //Name
			"",          //EC2Endpoint
			"https://objects.liquidweb.services", //S3Endpoint
			"https://objects.liquidweb.services", //S3BucketEndpoint
			false, //S3LocationConstraint
			false, //S3LowercaseBucket
			"https://objects.liquidweb.services", //SDBEndpoint
			"", //SNSEndpoint
			"", //SQSEndpoint
			"", //IAMEndpoint
			"", //ELBEndpoint
			"", //AutoScalingEndpoint
			"", //RdsEndpoint
			"", //RouteS3Endpoint
		}
	*/

	bucketRegion := aws.Region{
		Name:       "liquidweb",
		S3Endpoint: "https://objects.liquidweb.services",
	}

	//bucketAuth := new(aws.Auth)
	//bucketAuth.AccessKey = config.accessKey
	//bucketAuth.SecretKey = config.secretKey

	bucketAuth, err := aws.GetAuth(config.accessKey, config.secretKey)

	//fmt.Printf("The script has the AccessKey %s", bucketAuth.AccessKey)

	reportError("Ran into a problem creating the authentication with accessKey %s", config.accessKey, err)

	return s3.New(bucketAuth, bucketRegion)
}

func ValidBucket(connection *s3.S3, config *runningConfig) bool {
	allBuckets, err := connection.ListBuckets()
	reportError("Could not retrieve buckets from %s", "objstor", err)

	bucketExists := false

	for _, bucket := range allBuckets.Buckets {
		if bucket.Name == config.bucket {
			bucketExists = true
		}
	}
	return bucketExists
}

func SetupBucket(config runningConfig, connection s3.S3) s3.Bucket {

	bucket := *connection.Bucket(config.bucket)

	return bucket
}

/*
// in case I want to try to go back to setting up a function map with reflection
type map[string]commandFunc interface {
	run(args []string)
}
*/
func callFunc(config runningConfig, bucket s3.Bucket) {
	// call the function with the name of the command that you got

	switch config.command {
	case "get":
		magicGet(config, bucket)
	case "put":
		magicPut(config, bucket)
	case "ls":
		ls(config, bucket)
	case "mkdir":
		mkdir(config, bucket)
	case "chdir":
		chdir(config, bucket)
	case "rmdir":
		rmdir(config, bucket)
	case "delete":
		delete(config, bucket)
	}

	// ## TODO - setting up a function map for all of the functions to call
	// looks like this won't work without:
	// https://bitbucket.org/mikespook/golib/src/27c65cdf8a77/funcmap/
	// maybe read more here:
	// http://blog.golang.org/laws-of-reflection
	/*
		cmdFuncs := map[string]commandFunc{}{
			"get":    get,
			"put":    put,
			"ls":     ls,
			"mkdir":  mkdir,
			"chdir":  chdir,
			"rmdir":  rmdir,
			"delete": delete,
		}

		cmdFuncs[command].run(cmdParams)
	*/
}

func reportError(message string, messageSub string, err error) {
	if err != nil {
		log.Printf(message, messageSub)
		log.Println(err.Error())
		os.Exit(1)
	}
	return
}

// Gets a file from the remote location and puts it on the local system
// cli: `binary` `get` `pwd `Remote file` `local file` `bucketName` `username`
// passed to this is ["remote file", "local file"]
func get(config runningConfig, bucket s3.Bucket) {
	//data := new(Buffer)
	data, err := bucket.Get(config.cmdParams[0])
	reportError("Caught an error loading the remote file %s", config.cmdParams[0], err)
	err = ioutil.WriteFile(config.cmdParams[1], data, 0644)
	reportError("Caught an writing the local file %s", config.cmdParams[1], err)
}

func magicGet(config runningConfig, bucket s3.Bucket) {
	// open up the output file for writing
	outFile, err := os.Create(config.cmdParams[1])
	defer outFile.Close()
	reportError("Caught an writing the local file %s", config.cmdParams[1], err)

	// open up the remote file for reading
	dataResponse, err := bucket.GetResponse(config.cmdParams[0])
	defer dataResponse.Body.Close()
	reportError("Caught an error loading the remote file %s", config.cmdParams[0], err)

	// copy all bytes, without loading stuff in memory, then defer close happen
	_, err = io.Copy(outFile, dataResponse.Body)
	reportError("Caught an writing the local file %s", config.cmdParams[1], err)
}

// puts a file from the local location to a remote location
// cli: `binary` `put` `pwd `local file` `remote file` `bucketName` `username`
func put(config runningConfig, bucket s3.Bucket) {
	//data := new(Buffer)
	data, err := ioutil.ReadFile(config.cmdParams[0])
	reportError("Caught an error loading the local file %s", config.cmdParams[0], err)
	err = bucket.Put(config.cmdParams[1], data, contentType, "0644")
	reportError("Caught an error saving the remote file %s", config.cmdParams[1], err)
}

// puts a file from the local location to a remote location by pieces
// cli: `binary` `put` `pwd `local file` `remote file` `bucketName` `username`
func magicPut(config runningConfig, bucket s3.Bucket) {
	// open the file to be transferred
	file, err := os.Open(config.cmdParams[0])
	defer file.Close()
	reportError("Caught an error opening the local file %s", config.cmdParams[0], err)

	bytes := make([]byte, chunkSize)
	buffer := bufio.NewReader(file)
	// at most, buffer.Read can only read len(bytes) bytes
	_, err = buffer.Read(bytes)
	reportError("Had an issue reading bytes from the local file %s", config.cmdParams[0], err)

	// determine the filetype based on the bytes you read
	filetype := http.DetectContentType(bytes)

	// set up for multipart upload
	multiUploader, err := bucket.InitMulti(config.cmdParams[1], filetype, s3.ACL("private"))
	reportError("Had an issue opening the remote file %s for writing", config.cmdParams[1], err)

	// upload all of the file in pieces
	parts, err := multiUploader.PutAll(file, chunkSize)
	reportError("Had an issue putting chunks in the remote file %s", config.cmdParams[1], err)

	// complete the file
	err = multiUploader.Complete(parts)
	reportError("Issue completing the file %s in the bucket", config.cmdParams[1], err)

	return
}

// lists the content of a directory on the remote system
// cli: `binary` `ls` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func ls(config runningConfig, bucket s3.Bucket) {
	items, err := bucket.List(config.cmdParams[0], "", "", pagesize)
	reportError("Failed listing contents of the Bucket behind the path %s", config.cmdParams[0], err)
	for _, target := range items.Contents {
		// prints out in the format defined by:
		// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
		_, err = fmt.Printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
		reportError("Failed displaying the file %s", target.Key, err)
	}
}

// does nothing - making of directories is not required, but is required for cPanel transport
// cli: `binary` `mkdir` `pwd` `path` `bucketName` `username`
func mkdir(config runningConfig, bucket s3.Bucket) {
	return
}

// does almost nothing - not required, but must return the path
// cli: `binary` `chdir` `pwd` `path` `bucketName` `username`
func chdir(config runningConfig, bucket s3.Bucket) {
	_, err := fmt.Println(config.cmdParams[0])
	reportError("failed to print the given path %s", config.cmdParams[0], err)
}

// removes everything under the given path on the remote bucket
// cli: `binary` `rmdir` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func rmdir(config runningConfig, bucket s3.Bucket) {
	items, err := bucket.List(config.cmdParams[0], "", "", pagesize)
	reportError("Failed listing contents of the Bucket behind the path %s", config.cmdParams[0], err)
	for len(items.Contents) > 0 {
		for _, target := range items.Contents {
			err = bucket.Del(target.Key)
			reportError("Failed removing the target %s from the Bucket", target.Key, err)
		}
		items, _ = bucket.List(config.cmdParams[0], "", "", pagesize)
	}
}

// removes a file at a given location
// cli: `delete` `rmdir` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func delete(config runningConfig, bucket s3.Bucket) {
	err := bucket.Del(config.cmdParams[0])
	reportError("failed to delete file %s", config.cmdParams[0], err)
}

func main() {

	config := getConfig()

	connection := SetupConnection(config)
	bucket := SetupBucket(*config, *connection)

	callFunc(*config, bucket)
}
