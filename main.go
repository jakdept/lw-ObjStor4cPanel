package main

// cPanel backup transport helper for Liquidweb Object Storage
// By Jack Hayhurst

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	//"log"
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

func SetupConnection(config runningConfig) (*s3.S3, error) {
	bucketRegion := aws.Region{
		Name:              "liquidweb",
		S3Endpoint:        "https://objects.liquidweb.services",
		S3LowercaseBucket: true,
	}

	bucketAuth, err := aws.GetAuth(config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("Problem creating Authentication %s - %v", config.AccessKey, err)
	}

	return s3.New(bucketAuth, bucketRegion), nil
}

func ValidBucket(config runningConfig, connection *s3.S3) (bool, error) {
	allBuckets, err := connection.ListBuckets()
	if err != nil {
		return false, fmt.Errorf("problem listing buckets - %v", err)
	}
	bucketExists := false

	for _, Bucket := range allBuckets.Buckets {
		if Bucket.Name == config.Bucket {
			bucketExists = true
		}
	}
	return bucketExists, nil
}

func SetupBucket(config runningConfig) (s3.Bucket, error) {
	connection, err := SetupConnection(config)
	if err != nil {
		return s3.Bucket{}, err
	}

	b := *connection.Bucket(config.Bucket)
	return b, nil
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

// does almost nothing - not required, but must return the path
// cli: `binary` `chdir` `Pwd` `path` `bucketName` `username`
func Chdir(config runningConfig, Bucket s3.Bucket) error {
	_, err := fmt.Println(config.CmdParams[0])
	if err != nil {
		return fmt.Errorf("failed to print the given path %s - %v", config.CmdParams[9], err)
	}
	return nil
}

// lists the content of a directory on the remote system
// cli: `binary` `ls` `Pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func Lsdir(config runningConfig, Bucket s3.Bucket) error {
	items, err := Bucket.List(config.CmdParams[0], "/", "", pagesize)
	if err != nil {
		return fmt.Errorf("failed to list the contents of path %s - %v", config.CmdParams[0], err)
	}
	for _, target := range items.Contents {
		// prints out in the format defined by:
		// "-rwxr-xr-1 root root 3171 Jan 18 12:23 temp.txt"
		_, err = fmt.Printf("-rwxr-xr-1 %s %s %d Jan 18 12:23 %s", target.Owner, target.Owner, target.Size, target.Key)
		if err != nil {
			panic(fmt.Sprintf("failed display the file %s - %s", target.Key, err.Error()))
		}
	}
	return nil
}

// Gets a file from the remote location and puts it on the local system
// cli: `binary` `get` `Pwd `Remote file` `local file` `bucketName` `username`
// passed to this is ["remote file", "local file"]
func get(config runningConfig, Bucket s3.Bucket) {
	//data := new(Buffer)
	data, err := Bucket.Get(config.CmdParams[0])
	if err != nil {
		panic(fmt.Sprintf("error loading remote file %s - %s", config.CmdParams[0], err.Error()))
	}
	err = ioutil.WriteFile(config.CmdParams[1], data, 0644)
	if err != nil {
		panic(fmt.Sprintf("error writing local file %s - %s", config.CmdParams[1], err.Error()))
	}
}

func magicGet(config runningConfig, Bucket s3.Bucket) {
	// open up the output file for writing
	outFile, err := os.Create(config.CmdParams[1])
	defer outFile.Close()
	if err != nil {
		panic(fmt.Sprintf("error writing to local file %s - %s", config.CmdParams[1], err.Error()))
	}

	// open up the remote file for reading
	dataResponse, err := Bucket.GetResponse(config.CmdParams[0])
	defer dataResponse.Body.Close()
	if err != nil {
		panic(fmt.Sprintf("error loading remote file %s - %s", config.CmdParams[0], err.Error()))
	}

	// copy all bytes, without loading stuff in memory, then defer close happen
	_, err = io.Copy(outFile, dataResponse.Body)
	if err != nil {
		panic(fmt.Sprintf("error writing to local file %s - %s", config.CmdParams[1], err.Error()))
	}
}

// puts a file from the local location to a remote location
// cli: `binary` `put` `Pwd `local file` `remote file` `bucketName` `username`
func put(config runningConfig, Bucket s3.Bucket) {
	//data := new(Buffer)
	data, err := ioutil.ReadFile(config.CmdParams[0])
	if err != nil {
		panic(fmt.Sprintf("error loading local file %s - %s", config.CmdParams[0], err.Error()))
	}
	err = Bucket.Put(config.CmdParams[1], data, contentType, "0644")
	if err != nil {
		panic(fmt.Sprintf("error writing remote file %s - %s", config.CmdParams[1], err.Error()))
	}
}

// puts a file from the local location to a remote location by pieces
// cli: `binary` `put` `Pwd `local file` `remote file` `bucketName` `username`
func magicPut(config runningConfig, Bucket s3.Bucket) {
	// open the file to be transferred
	file, err := os.Open(config.CmdParams[0])
	defer file.Close()
	if err != nil {
		panic(fmt.Sprintf("error loading local file %s - %s", config.CmdParams[0], err.Error()))
	}

	bytes := make([]byte, chunkSize)
	buffer := bufio.NewReader(file)
	// at most, buffer.Read can only read len(bytes) bytes
	_, err = buffer.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("error reading from local file %s - %s", config.CmdParams[0], err.Error()))
	}

	// determine the filetype based on the bytes you read
	filetype := http.DetectContentType(bytes)

	// set up for multipart upload
	multiUploader, err := Bucket.InitMulti(config.CmdParams[1], filetype, s3.ACL("private"))
	if err != nil {
		panic(fmt.Sprintf("error opening remote file %s - %s", config.CmdParams[1], err.Error()))
	}

	// upload all of the file in pieces
	parts, err := multiUploader.PutAll(file, chunkSize)
	if err != nil {
		panic(fmt.Sprintf("error writing to remote file %s - %s", config.CmdParams[1], err.Error()))
	}

	// complete the file
	err = multiUploader.Complete(parts)
	if err != nil {
		panic(fmt.Sprintf("error completing file %s - %s", config.CmdParams[1], err.Error()))
	}

	return
}

// removes everything under the given path on the remote Bucket
// cli: `binary` `rmdir` `Pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func rmdir(config runningConfig, Bucket s3.Bucket) {
	items, err := Bucket.List(config.CmdParams[0], "", "", pagesize)
	if err != nil {
		panic(fmt.Sprintf("error listing path %s - %s", config.CmdParams[0], err.Error()))
	}
	for len(items.Contents) > 0 {
		for _, target := range items.Contents {
			err = Bucket.Del(target.Key)
			if err != nil {
				panic(fmt.Sprintf("error removing remote %s - %s", target.Key, err.Error()))
			}
		}
		// check to make sure everything is gone
		items, err = Bucket.List(config.CmdParams[0], "", "", pagesize)
		if err != nil {
			panic(fmt.Sprintf("error listing path %s - %s", config.CmdParams[0], err.Error()))
		}
	}
}

// removes a file at a given location
// cli: `delete` `rmdir` `Pwd` `path` `bucketName` `username`
// passed to this is ["path"]
func delete(config runningConfig, Bucket s3.Bucket) {
	err := Bucket.Del(config.CmdParams[0])
	if err != nil {
		panic(fmt.Sprintf("failed to delete %s", config.CmdParams[0], err.Error()))
	}
}

func main() {

	config := getConfig()

	//connection := SetupConnection(config)
	bucket, err := SetupBucket(config)
	if err != nil {
		log.Fatal(err)
	}

	callFunc(config, bucket)
}
