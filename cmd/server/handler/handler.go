package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/martskins/zipsrv/internal/randstring"
	"github.com/martskins/zipsrv/internal/timer"
	"github.com/martskins/zipsrv/internal/types"
	"github.com/martskins/zipsrv/internal/zipper"
)

func HandleGetZip(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tkn := vars["tkn"]

	res.Header().Set("Content-Type", "application/zip")
	res.Header().Set("Content-Disposition", "attachment; filename=result.zip")

	http.ServeFile(res, req, "/tmp/"+tkn+".zip")
}

func HandleZipRequest(res http.ResponseWriter, req *http.Request) {
	stopTimer := timer.Timer("handleZipRequest")
	defer stopTimer()
	defer req.Body.Close()

	var p types.ZipRequest
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

	reqID := randstring.New(24)
	go func() {
		err := zipper.ProcessFiles(p, reqID, 100)
		if err != nil {
			log.Println(err)
		}

		log.Println("Notify when this is done")
	}()

	res.WriteHeader(200)
	res.Write([]byte(reqID))
}
