package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

type RequestPayload struct {
	Files      []FileInfo
	DownloadAs string
}

type Configuration struct {
	AccessKey          string
	SecretKey          string
	Bucket             string
	Region             string
	RedisServerAndPort string
	Port               int
	SignatureKey       string
}

type FileInfo struct {
	FileName     string
	Folder       string
	Path         string
	FileId       int64 `json:",string"`
	ProjectId    int64 `json:",string"`
	ProjectName  string
	Modified     string
	ModifiedTime time.Time
}

func readConfig() (*Configuration, error) {

	configFile, _ := os.Open("conf.json")
	decoder := json.NewDecoder(configFile)
	var config Configuration
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	awsBucket, err := initAwsBucket(config)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Running on port", config.Port)

	handlerManager := NewStreamDownloadHandler(config, awsBucket)
	handlerManager.ServeAndHandle()
}

func parseFileDates(files []*FileInfo) {
	layout := "2006-01-02T15:04:05Z"
	for _, file := range files {
		t, err := time.Parse(layout, file.Modified)
		if err != nil {
			fmt.Println(err)
			continue
		}
		file.ModifiedTime = t
	}
}

func initAwsBucket(config *Configuration) (*s3.Bucket, error) {
	expiration := time.Now().Add(time.Hour * 1)
	auth, err := aws.GetAuth(config.AccessKey, config.SecretKey, "", expiration) //"" = token which isn't needed
	if err != nil {
		return nil, err
	}

	awsBucket := s3.New(auth, aws.GetRegion(config.Region)).Bucket(config.Bucket)
	return awsBucket, nil
}
