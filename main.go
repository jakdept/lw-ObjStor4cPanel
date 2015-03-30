package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	//"bufio"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const contentType = "binary/octet-stream"

const pagesize = 1000

var (
	Bucket  s3.Bucket
	command string
	//pwd                  string
	cmdParams, host      string
	accessKey, secretKey string
)

/*
type map[string]commandFunc interface {
	run(args []string)
}
*/

var bucket s3.Bucket

var bucketRegion = aws.Region{
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

func main() {

	// parameters are passed as:
	// binary command pwd [cmdParams ...] host accessKey
	command := os.Args[1]
	//pwd := os.Args[2]
	host := os.Args[len(os.Args)-2]
	accessKey := os.Args[len(os.Args)-1]

	cmdParams := os.Args[3 : len(os.Args)-2]

	// secretKey is passed via enviroment variable
	secretKey := os.Getenv("PASSWORD")

	bucketAuth := new(aws.Auth)
	bucketAuth.AccessKey = accessKey
	bucketAuth.SecretKey = secretKey
	//	bucketAuth := aws.Auth{accessKey, secretKey}

	connection := s3.New(*bucketAuth, bucketRegion)
	bucket = *connection.Bucket(host)

	// ##TODO## buffering needs to be added for file uploads and downloads

	// ##TODO##
	// setting up a function map for all of the functions to call
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

	// call the function with the name of the command that you got
	switch command {
	case "get":
		get(cmdParams)
	case "put":
		put(cmdParams)
	case "ls":
		ls(cmdParams)
	case "mkdir":
		mkdir(cmdParams)
	case "chdir":
		chdir(cmdParams)
	case "rmdir":
		rmdir(cmdParams)
	case "delete":
		delete(cmdParams)
	}
	//command(cmdParams)
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
func get(args []string) {
	//data := new(Buffer)
	data, err := Bucket.Get(args[0])
	reportError("Caught an error loading the remote file %s", args[0], err)
	err = ioutil.WriteFile(args[1], data, 0644)
	reportError("Caught an writing the local file %s", args[1], err)
}

// puts a file from the local location to a remote location
// cli: `binary` `put` `pwd `local file` `remote file` `bucketName` `username`
// passed to this is ["local file", "remote file"]
func put(args []string) {
	//data := new(Buffer)
	data, err := ioutil.ReadFile(args[0])
	reportError("Caught an error loading the local file %s", args[0], err)
	err = Bucket.Put(args[1], data, contentType, "0644")
	reportError("Caught an error loading the local file %s", args[0], err)
}

// lists the content of a directory on the remote system
// cli: `binary` `ls` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func ls(args []string) {
	// ##todo## need to rework this yet
	items, err := Bucket.List(args[0], "", "", pagesize)
	reportError("Failed listing contents of the Bucket behind the path %s", args[0], err)
	for _, target := range items.Contents {
		// prints out in the format defined by:
		// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
		err = fmt.Printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
		reportError("Failed displaying the file %s", target.Key, err)
	}
}

// does nothing - making of directories is not required, but is required for cPanel transport
// cli: `binary` `mkdir` `pwd` `path` `bucketName` `username`
func mkdir(args []string) {
	return
}

// does almost nothing - not required, but must return the path
// cli: `binary` `chdir` `pwd` `path` `bucketName` `username`
func chdir(args []string) {
	_, err := fmt.Println(args[0])
	reportError("failed to print the given path %s", args[0], err)
}

// removes everything under the given path on the remote bucket
// cli: `binary` `rmdir` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func rmdir(args []string) {
	items, err := bucket.List(args[0], "", "", pagesize)
	reportError("Failed listing contents of the Bucket behind the path %s", args[0], err)
	for len(items.Contents) > 0 {
		for _, target := range items.Contents {
			err = Bucket.Del(target.Key)
			reportError("Failed removing the target %s from the Bucket", target.Key, err)
		}
		items, _ = bucket.List(args[0], "", "", pagesize)
	}
}

// removes a file at a given location
// cli: `delete` `rmdir` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func delete(args []string) {
	err := bucket.Del(args[0])
	reportError("failed to delete file %s", args[0], err)
}
