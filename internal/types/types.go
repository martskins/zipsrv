package types

import "net/http"

type WriteTask struct {
	Resp *http.Response
	URL  string
}

type ZipRequest struct {
	Files []string
}
