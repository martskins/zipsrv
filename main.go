package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type writeTask struct {
	resp *http.Response
	url  string
}

type zipRequest struct {
	Files  []string
	Output string
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/zipfiles", handleZipCall).Methods("POST")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleZipCall(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var p zipRequest
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(err.Error()))
		return
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(err.Error()))
		return
	}

	go func() {
		err := downloadAndProcessFiles(p)
		if err != nil {
			log.Println(err)
		}
	}()

	res.WriteHeader(204)
	res.Write([]byte("Received"))
}

func downloadAndProcessFiles(p zipRequest) error {
	rand.Seed(time.Now().UnixNano())
	zipName := randString(24)
	output := p.Output + "/" + zipName + ".zip"

	newfile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	tasks := make(chan writeTask)
	var wg sync.WaitGroup

	wg.Add(len(p.Files))
	go queueTasks(p.Files, &wg, tasks)
	go func() {
		wg.Wait()
		close(tasks)
	}()

	for t := range tasks {
		err = writeFile(zipWriter, t)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func queueTasks(files []string, wg *sync.WaitGroup, ch chan writeTask) {
	for _, u := range files {
		go func(url string) {
			resp, err := getFile(url)
			if err != nil {
				log.Fatal(err)
			}

			ch <- writeTask{resp, url}
			(*wg).Done()
		}(u)
	}
}

func writeFile(zipWriter *zip.Writer, t writeTask) error {
	defer t.resp.Body.Close()

	parts := strings.Split(t.url, "/")
	fn := parts[len(parts)-1]

	w, err := zipWriter.Create(fn)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, t.resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getFile(url string) (*http.Response, error) {
	var err error
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func randString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
