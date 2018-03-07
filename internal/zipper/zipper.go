package zipper

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/martskins/zipsrv/internal/types"
)

func ProcessFiles(p types.ZipRequest, filename string, workers int) error {
	output := "/tmp/" + filename + ".zip"
	newfile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	wCh := make(chan types.WriteTask)
	dCh := make(chan string, workers)
	var wg sync.WaitGroup
	wg.Add(len(p.Files))

	for w := 0; w < workers; w++ {
		go downloadWorker(&wg, dCh, wCh)
	}
	go scheduleDownloads(p, dCh)

	go func() {
		wg.Wait()
		close(wCh)
	}()

	writerWorker(zipWriter, wCh)
	return nil
}

func scheduleDownloads(p types.ZipRequest, dCh chan string) {
	for _, f := range p.Files {
		dCh <- f
	}
	close(dCh)
}

func writerWorker(zipWriter *zip.Writer, wCh chan types.WriteTask) {
	for t := range wCh {
		err := WriteFile(zipWriter, t)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func downloadWorker(wg *sync.WaitGroup, dCh chan string, wCh chan types.WriteTask) {
	for url := range dCh {
		resp, err := GetFile(url)
		if err != nil {
			log.Fatal(err)
		}

		wCh <- types.WriteTask{resp, url}
		(*wg).Done()
	}
}

func WriteFile(zipWriter *zip.Writer, t types.WriteTask) error {
	defer t.Resp.Body.Close()

	parts := strings.Split(t.URL, "/")
	fn := parts[len(parts)-1]

	w, err := zipWriter.Create(fn)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, t.Resp.Body)
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
