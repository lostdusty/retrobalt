package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

type MediaInfo struct {
	Url  string
	Size uint //Media size in bytes
	Name string
	Type string
}

func parseMedia(url string) (*MediaInfo, error) {
	log.Println("[info] (module: downloader) Getting media info before downloading...")
	req, err := genericRequest(url, "HEAD", nil)
	if err != nil {
		return nil, err
	}

	_, parsefilename, err := mime.ParseMediaType(req.Header.Get("Content-Disposition"))
	filename := parsefilename["filename"]
	if err != nil {
		filename = path.Base(req.Request.URL.Path)
	}
	size := req.Header.Get("Content-Length")
	if size == "" {
		size = "0"
	}
	parseSize, err := strconv.Atoi(size)
	if err != nil {
		return nil, err
	}

	return &MediaInfo{
		Url:  url,
		Size: uint(parseSize),
		Name: filename,
		Type: req.Header.Get("Content-Type"),
	}, nil
}

func downloadFile(media MediaInfo) (string, error) {
	normalizeFileName := regexp.MustCompile(`[^[:word:][:punct:]\s]`)
	newName := normalizeFileName.ReplaceAllString(media.Name, "")

	start := time.Now()

	resp, err := genericRequest(media.Url, "GET", nil)
	if err != nil {
		return "", err
	}

	dest, err := os.OpenFile(newName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	if media.Size > 2 {
		go func() {
			for {
				select {
				default:
					fi, err := dest.Stat()
					if err != nil {
						return
					}
					size := fi.Size()

					if size == 0 {
						size = 1
					}
					var percent float64 = float64(size) / float64(media.Size) * 100

					fmt.Printf("\r[info] (module: download) Downloading (%.0f%%)", percent)
				}
				time.Sleep(time.Second)
			}
		}()
	} else {
		log.Println("[info] (module: download) Starting to download... (cobalt didn't send the file size)")
	}

	_, err = io.Copy(dest, resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Printf("\r[info] (module: download) Finished downloading! (%.0f%%)\n", 100.0)

	pwd, _ := os.Getwd()

	return fmt.Sprintf("Downloaded %s\\%s in %v", pwd, dest.Name(), time.Since(start)), nil
}
