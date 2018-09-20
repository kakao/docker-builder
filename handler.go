package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write(bytes.NewBufferString("docker builder server").Bytes())
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("start build")
	defer r.Body.Close()

	buildSpec := BuildSpec{}
	if err := json.Unmarshal(body, &buildSpec); err != nil {
		log.Printf("json unmarshal has failed: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if buildSpec.BuildArgStr != "" {
		if err := json.Unmarshal([]byte(buildSpec.BuildArgStr), &buildSpec.BuildArg); err != nil {
			log.Printf("docker argument unmarshal has failed : %s\n", err)
			log.Println(buildSpec.BuildArgStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	log.Printf("buildSpec: %#+v\n", buildSpec)
	go build(&buildSpec)
	w.WriteHeader(http.StatusCreated)
}

func handleBuildResult(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	buildResult := &BuildResult{}
	if err := json.Unmarshal(body, buildResult); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !buildResult.IsSuccess {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
