package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AdRoll/goamz/s3"
	jwt "github.com/dgrijalva/jwt-go"
)

type StreamDownloadHandler struct {
	config    *Configuration
	awsBucket *s3.Bucket
}

func NewStreamDownloadHandler(config *Configuration, awsBucket *s3.Bucket) *StreamDownloadHandler {
	ret := &StreamDownloadHandler{}
	ret.config = config
	ret.awsBucket = awsBucket

	return ret
}

func (st *StreamDownloadHandler) handle(w http.ResponseWriter, r *http.Request) {
	authorizationToken := r.Header.Get("Authorization")

	if err := st.authenticate(authorizationToken); err != nil {
		http.Error(w, "Not authorized", 403)
		fmt.Println("Authentication error")
		return
	}

	start := time.Now()
	decoder := json.NewDecoder(r.Body)
	var payload RequestPayload
	err := decoder.Decode(&payload)
	if err != nil {
		http.Error(w, "Incorrect Payload", 500)
		return
	}
	var makeSafeFileName = regexp.MustCompile(`[#<>:"/\|?*\\]`)
	if len(payload.DownloadAs) > 0 {
		payload.DownloadAs = makeSafeFileName.ReplaceAllString(payload.DownloadAs, "")
		if payload.DownloadAs == "" {
			payload.DownloadAs = "download.zip"
		}
	} else {
		payload.DownloadAs = "download.zip"
	}

	files := payload.Files

	// Start processing the response
	w.Header().Add("Content-Disposition", "attachment; filename=\""+payload.DownloadAs+"\"")
	w.Header().Add("Content-Type", "application/zip")

	// Loop over files, add them to the
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
	for _, file := range files {

		if file.Path == "" {
			log.Printf("Missing path for file: %v", file)
			continue
		}

		// Build safe file file name
		safeFileName := makeSafeFileName.ReplaceAllString(file.FileName, "")
		if safeFileName == "" { // Unlikely but just in case
			safeFileName = "file"
		}

		// Read file from S3, log any errors
		rdr, err := st.awsBucket.GetReader(file.Path)
		if err != nil {
			switch t := err.(type) {
			case *s3.Error:
				if t.StatusCode == 404 {
					log.Printf("File not found. %s", file.Path)
				}
			default:
				log.Printf("Error downloading \"%s\" - %s", file.Path, err.Error())
			}
			continue
		}

		// Build a good path for the file within the zip
		zipPath := ""
		// Prefix project Id and name, if any (remove if you don't need)
		if file.ProjectId > 0 {
			zipPath += strconv.FormatInt(file.ProjectId, 10) + "."
			// Build Safe Project Name
			file.ProjectName = makeSafeFileName.ReplaceAllString(file.ProjectName, "")
			if file.ProjectName == "" { // Unlikely but just in case
				file.ProjectName = "Project"
			}
			zipPath += file.ProjectName + "/"
		}
		// Prefix folder name, if any
		if file.Folder != "" {
			zipPath += file.Folder
			if !strings.HasSuffix(zipPath, "/") {
				zipPath += "/"
			}
		}
		zipPath += safeFileName
		h := &zip.FileHeader{
			Name:   zipPath,
			Method: zip.Store,
			Flags:  0x800,
		}

		if file.Modified != "" {
			h.SetModTime(file.ModifiedTime)
		}

		f, _ := zipWriter.CreateHeader(h)

		io.Copy(f, rdr)
		rdr.Close()
	}

	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

func (st *StreamDownloadHandler) authenticate(bearerToken string) error {
	authorizationToken := strings.TrimPrefix(bearerToken, "Bearer ")
	token, err := jwt.Parse(authorizationToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Incorrect token")
		}
		return st.config.SignatureKey, nil
	})

	// In production, plase. Validate only that err != nil so that we make sure the signature is valid before we proceed.
	if (0 == 1) && err != nil {
		return fmt.Errorf("Error authenticating")
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	if time.Now().Before(time.Unix(int64(claims["exp"].(float64)), 0)) {
		return nil
	}

	return fmt.Errorf("Error authenticating")
}

func (st *StreamDownloadHandler) ServeAndHandle() {

	http.HandleFunc("/", st.handle)
	http.ListenAndServe(":"+strconv.Itoa(st.config.Port), nil)
}
