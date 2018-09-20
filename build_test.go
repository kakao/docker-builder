package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"github.com/fsouza/go-dockerclient"
)

var (
	buildSpec *BuildSpec
)

func init() {
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

	buildSpec.BuildArg = []docker.BuildArg{
		docker.BuildArg{
			Name:  "secret",
			Value: "xxx",
		},
	}
}

func TestBuild1(t *testing.T) {
	projectDir, err := ioutil.TempDir("", fmt.Sprintf("docker%s", buildSpec.BuildID))
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	if err := cloneGit(buildSpec, projectDir, os.Stdout); err != nil {
		t.Error(err)
		return
	}
	if err := buildDockerImage(buildSpec, projectDir, os.Stdout); err != nil {
		t.Fatal(err)
	}
	if err := pushDockerImage(buildSpec, os.Stdout); err != nil {
		t.Fatal(err)
	}
}

func TestBuild2(t *testing.T) {
	if err := build(buildSpec); err != nil {
		t.Fatal(err)
	}
}
