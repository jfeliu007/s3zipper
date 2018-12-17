# s3zipper
Microservice that serves a streaming zip file of files securely downloaded from S3

You can hit this service with a payload describing alist of S3 files and it will stream a zipfile back to you as fast as possible containing the files, regardless of the size of the files in question.

# How to integrate with share
## Obtaining the files

```
mkdir PATH_TO_ROOT_OF_APPS
git clone http://github.com/creativedrive/stream-download
cd stream-download
cp sample_conf.json conf.json
vim conf.json
// Update values for AssessKey, SecretKey, and Bucket with the correct values
docker-compose build
```

## How to run
cd PATH_TO_ROOT_OF_APPS/stream-download
docker-compose up

## How to use

This version will be listening on the localhost:8000

You can POST to localhost:8000 using the Authorization header with a valid Bearer token from Auth
The following is an example of a payload

```
{
    "Files": [
        {
            "Folder": "test1",
            "Path": "/account_folders/account_1/assets/201812/971daec51b6e5c0b10b8c178a7.43088177.jpg",
            "Filename": "test.jpg"
        },
        {
            "Folder": "",
            "Path": "/account_folders/account_1/assets/201812/p1cuf9rpk41bso1mu7rjr1fae1r1t3.jpeg",
            "Filename": "26579-40997 Test_09.jpeg"
        },
        {
            "Folder": "",
            "Path": "/account_folders/account_1/assets/201812/p1cufa1amj3v1s1jbpfhq2bvc3.jpeg",
            "Filename": "26579-40997 Test_01.jpeg"
        }
    ],
    "DownloadAs": "download.zip"
}
```
### Explanation of the payload
```
{
    "Files": [
        {
            "Folder": "FOLDER_NAME", // Folder of where this file must be inside the zipfile
            "Path": "FULL_FILEPATH", // Full path to the S3 file
            "Filename": "FILENAME" // Filename in the zipfile. 
        }, 
        .
        .
        .
    ],
    "DownloadAs": "filename.zip" // Name of the returned zipfile
}
```



