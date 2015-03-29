package main

import (
	"fmt"
	"io/ioutil"
	"os"
	//"bufio"
	//"log"

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

var bucket s3.Bucket

/*
region format:
type Region struct {
	Name                 string // the canonical name of this region.
	EC2Endpoint          string
	S3Endpoint           string
	S3BucketEndpoint     string // Not needed by AWS S3. Use ${bucket} for bucket name.
	S3LocationConstraint bool   // true if this region requires a LocationConstraint declaration.
	S3LowercaseBucket    bool   // true if the region requires bucket names to be lower case.
	SDBEndpoint          string
	SNSEndpoint          string
	SQSEndpoint          string
	IAMEndpoint          string
	ELBEndpoint          string
	AutoScalingEndpoint  string
	RdsEndpoint          string
	Route53Endpoint      string
}
*/

var bucketRegion = aws.Region{
	"LiquidWeb",
	"",
	"",
	"https://objects.liquidweb.services",
	true,
	true,
	"https://objects.liquidweb.services",
	"",
	"",
	"",
	"",
	"",
	"",
	"",
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

	// call the function with the name of the command that you got
	command(cmdParams)
}

// Gets a file from the remote location and puts it on the local system
// cli: `binary` `get` `pwd `Remote file` `local file` `bucketName` `username`
// passed to this is ["remote file", "local file"]
func get(args []string) {
	//data := new(Buffer)
	data, _ := Bucket.Get(args[0])
	ioutil.WriteFile(args[1], data, 0644)
}

// puts a file from the local location to a remote location
// cli: `binary` `put` `pwd `local file` `remote file` `bucketName` `username`
// passed to this is ["local file", "remote file"]
func put(args []string) {
	//data := new(Buffer)
	data, _ := ioutil.ReadFile(args[0])
	Bucket.Put(args[1], data, contentType, "0644")
}

// lists the content of a directory on the remote system
// cli: `binary` `ls` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func ls(args []string) {
	// ##todo## need to rework this yet
	items, _ := Bucket.List(args[0], "", "", pagesize)
	for _, target := range items.Contents {
		// prints out in the format defined by:
		// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
		fmt.Printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
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
	fmt.Println(args[0])
}

// removes everything under the given path on the remote bucket
// cli: `binary` `rmdir` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func rmdir(args []string) {
	items, _ := bucket.List(args[0], "", "", pagesize)
	for len(items.Contents) > 0 {
		for _, target := range items.Contents {
			Bucket.Del(target.Key)
		}
		items, _ = bucket.List(args[0], "", "", pagesize)
	}
}

// removes a file at a given location
// cli: `delete` `rmdir` `pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func delete(args []string) {
	bucket.Del(args[0])
}
