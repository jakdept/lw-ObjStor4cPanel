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

func ls(path string){
	// ##todo## need to rework this yet
	items, _ := Bucket.List(path, "", "", pagesize)
	for _, target := range items.Contents {
	// prints out in the format defined by:
	// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
	fmt.printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
}
}
func mkdir(path string) {
	return
}

func chdir(path string) string {
	fmt.println(path)
}

func rmdir(path string) {
	items, _ := bucket.List(path, "", "", pagesize)
	for length(items.Contents) > 0 {
	for _, target := range items.Contents {
		Bucket.Del(target.Key)
	}
	items, _ := bucket.List(path, "", "", pagesize)
}
}

func delete(path string)