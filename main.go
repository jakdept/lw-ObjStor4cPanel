package main

// cPanel backup transport helper for Liquidweb Object Storage
// By Jack Hayhurst

// ## TODO buffering needs to be added for file uploads and downloads
// https://www.socketloop.com/tutorials/golang-upload-big-file-larger-than-100mb-to-aws-s3-with-multipart-upload

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"bufio"
	"http"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const contentType = "binary/octet-stream"

const pagesize = 1000

const chunkSize = 33554432 // 32M in bytes

type runningConfig struct {
	command   string
	pwd       string
	host      string
	cmdParams []string
	accessKey string
	secretKey string
}

func getConfig() runningConfig {
	config := new(runningConfig)
	// parameters are passed as:
	// binary command pwd [cmdParams ...] host accessKey
	config.command = os.Args[1]
	//pwd := os.Args[2]
	config.host = os.Args[len(os.Args)-2]
	config.accessKey = os.Args[len(os.Args)-1]

	config.cmdParams = os.Args[3 : len(os.Args)-2]

	// secretKey is passed via enviroment variable
	config.secretKey = os.Getenv("PASSWORD")

	return *config
}

// ## TODO remove this stuff once you're sure you do not need it
/*
var (
	Bucket  s3.Bucket
	command string
	//pwd                  string
	cmdParams, host      string
	accessKey, secretKey string
)
*/

func createBucket(config runningConfig) s3.Bucket {

	// set up the required bucket
	bucketRegion := aws.Region{
		"LiquidWeb", //Name
		"",          //EC2Endpoint
		"",          //S3Endpoint
		"https://objects.liquidweb.services", //S3BucketEndpoint
		true, //S3LocationConstraint
		true, //S3LowercaseBucket
		"https://objects.liquidweb.services", //SDBEndpoint
		"", //SNSEndpoint
		"", //SQSEndpoint
		"", //IAMEndpoint
		"", //ELBEndpoint
		"", //AutoScalingEndpoint
		"", //RdsEndpoint
		"", //RouteS3Endpoint
	}

	bucketAuth := new(aws.Auth)
	bucketAuth.AccessKey = config.accessKey
	bucketAuth.SecretKey = config.secretKey
	//	bucketAuth := aws.Auth{accessKey, secretKey}

	connection := s3.New(*bucketAuth, bucketRegion)
	bucket := *connection.Bucket(config.host)

	return bucket
}

func main() {

	config := getConfig()

	bucket := createBucket(config)

	callFunc(config, bucket)
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
		get(config, bucket)
	case "put":
		put(config, bucket)
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
func magicPut(config runningConfig, bucket S3Bucket) {
	// open the file to be transferred
	file, err := os.Open(config.cmdParams[0])
	reportError("Caught an error opening the local file %s", config.cmdParams[0], err)

	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	bytes := make([]byte, fileSize)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)

	// determine the filetype
	http.DetectContentType(bytes)

	// set up for multipart upload
	multi, err := bucket.InitMulti(config.cmdParams[1], filetype, s3.ACL("private"))

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	totalPartsNum := uint64(math.Ceil(float64(filesize)/float64(chunkSize)))

	parts := []s3.Part{}

	for i :- uint64(1); i< totalPartsNum; i++ {
		partSize := int(mat.Min(chunkSize, float64(filesize-int64(i*chunkSize))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)
		part, err :=multi.putPart(int(i), file)
		fmt.Printf("Processing %d part of %d and uploaded %d bytes.", int(i), int(totalPartsNumb), int(part.Size))
		parts = append(parts,part)
		if err != nil {
			log.Printf("Uploading parts of file error :i %s \n ", err)
			os.Exit(1)
		}
	}
	err = multi.Complete(parts)

	if err != nil{
		log.Printf("Error completing parts %s", err)
		os.Exit(1)
	}
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
