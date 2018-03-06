package zipper

import (
	"archive/zip"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/martskins/zipsrv/internal/timer"
)

type writeTask struct {
	resp *http.Response
	url  string
}

type zipRequest struct {
	Files []string
}

func ProcessFiles(p zipRequest, filename string) error {
	stopTimer := timer.Timer("processFiles")
	defer stopTimer()
	output := "/tmp/" + filename + ".zip"

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

func QueueTasks(files []string, wg *sync.WaitGroup, ch chan writeTask) {
	stopTimer := timer.Timer("queueTasks")
	defer stopTimer()
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

func WriteFile(zipWriter *zip.Writer, t writeTask) error {
	stopTimer := timer.Timer("writeFile")
	defer stopTimer()
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

func GetFile(url string) (*http.Response, error) {
	var err error
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func RandString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
