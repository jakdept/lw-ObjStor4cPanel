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

const pagesize = 10000

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

	switch command {
	case "get":
		get(cmdParams[0], cmdParams[1])
	case "put":
		put(cmdParams[0], cmdParams[1])
	case "ls":
		ls(cmdParams[0])
	case "mkdir":
		mkdir(cmdParams[0])
	case "chdir":
		chdir(cmdParams[0])
	case "rmdir":
		rmdir(cmdParams[0])
	case "delete":
		delete(cmdParams[0])
	}
}

func get(remoteSource string, localDestionation string) {
	//data := new(Buffer)
	data, _ := Bucket.Get(remoteSource)
	ioutil.WriteFile(localDestionation, data, 0644)
}

func put(localSource string, remoteDestination string) {
	//data := new(Buffer)
	data, _ := ioutil.ReadFile(localSource)
	Bucket.Put(remoteDestination, data, contentType, "0644")
}

func ls(path string) {
	// ##todo## need to rework this yet
	items, _ := Bucket.List(path, "", "", pagesize)
	for _, target := range items.Contents {
		// prints out in the format defined by:
		// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
		fmt.Printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
	}
}
func mkdir(path string) {
	return
}

func chdir(path string) {
	fmt.Println(path)
}

func rmdir(path string) {
	items, _ := bucket.List(path, "", "", pagesize)
	for len(items.Contents) > 0 {
		for _, target := range items.Contents {
			Bucket.Del(target.Key)
		}
		items, _ = bucket.List(path, "", "", pagesize)
	}
}

func delete(path string) {
	bucket.Del(path)
}
