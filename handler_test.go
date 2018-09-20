package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"log"
)

func TestHandleDockerBuild(t *testing.T) {
	buildSpec = &BuildSpec{
		Username:             "heimer.j",
		GithubOrgName:        "heimer-j",
		GithubRepoName:       "tmp-build-arg-test",
		GithubToken:          "c4d1e3034655ca7ddf590cea63694fb413dee82e",
		GitBranchType:        "Branch",
		GitBranchName:        "master",
		DockerfileLocation:   "/",
		DockerfileName:       "Dockerfile",
		DockerTagName:        "latest",
		DockerOrgName:        "heimer_j",
		DockerRepoName:       "tmp-build-arg-test",
		D2HUBPushKeyID:       "heimer.j",
		D2HUBPushKeyPassword: "",
		WatchCenterID:        "4958",
		BuildID:              "999999",
		ResultCallbackURL:    "http://localhost:3000/result",
		GithubHost: "github.com",
	}

	buildSpec.BuildArgStr = "[{\"name\":\"secret\", \"value\":\"434\"}]"

	jsonBytes, err := json.Marshal(buildSpec)
	if err != nil {
		t.Error(err)
	}
	resp, err := http.DefaultClient.Post("http://localhost:3000/build", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Error(resp.StatusCode)
	}
}
