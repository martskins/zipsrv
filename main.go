package main

import (
	"archive/zip"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func main() {
	files := []string{"https://cloud.netlifyusercontent.com/assets/344dbf88-fdf9-42bb-adb4-46f01eedd629/68dd54ca-60cf-4ef7-898b-26d7cbe48ec7/10-dithering-opt.jpg"}
	zipName := randStringRunes(24)
	output := "/Users/martin/Desktop/" + zipName + ".zip"

	err := zipFiles(output, files)
	if err != nil {
		log.Fatal(err)
	}
}

func zipFiles(filename string, files []string) error {
	var err error
	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	for _, path := range files {
		resp, err := http.Get(path)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		parts := strings.Split(path, "/")
		fn := parts[len(parts)-1]
		w, err := zipWriter.Create(fn)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
