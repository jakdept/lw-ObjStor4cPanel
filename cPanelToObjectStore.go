package main

import(
"os"
"io/ioutil"
"bufio"
"log"

"github.com/mitchellh/goamz/aws"
"github.com/mitchellh/goamz/s3"
)

const contentType = "binary/octet-stream"

const pagesize = 10000

var (
	Bucket s3.Bucket
	command, pwd, cmdParams, host, accessKey, secretKey string
	)

func getArguments() s3.Bucket{
	command, pwd := os.Argv[1:3]
	cmdParams := os.Argv[3:length(os.Argv)-2]
	host, accessKey := os.Argv[length(os.Argv)-2:]
	secretKey := os.Getenv("PASSWORD")

	bucketAuth := new(aws.Auth)

	bucketAuth.

	bucketRegion := new(aws.Region)

	bucketRegion.S3BucketEndpoint = "https://objects.liquidweb.services"

	return new(s3.Bucket())
}

func get(remoteSource string, localDestionation string) {
	data := new(Buffer)
	data, _ = Bucket.Get(remoteSource)
	ioutil.WriteFile(localDestionation, data, 0644)
}

func put(localSource string, remoteDestination string) {
	data := new(Buffer)
	data, _ = ioutil.ReadFile(localSource)
	Bucket.Put(remoteDestination, data, contentType, "0644")
}

func chdir(path string) string {
	return path
}

func mkdir(path string) {
	return
}

func rmdir(path string) {
	itemsToNuke, _ := bucket.List(path, "", "", pagesize)
	for length(itemsToNuke.Contents) > 0 {
	for _, target := range itemsToNuke.Contents {
		Bucket.Del(target)
	}
	itemsToNuke, _ := bucket.List(path, "", "", pagesize)
}
}

func ls(path string){
	bucketContents, _ := Bucket.List(path, "", "", pagesize)
}