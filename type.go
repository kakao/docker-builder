package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	defaultDockerfileName = "Dockerfile"
)

// BuildSpec is docker build spec
type BuildSpec struct {
	Username             string            `json:"username"`
	BuildTagID           int               `json:"buildTagID"`
	GithubHost           string            `json:"githubHost"`
	GithubOrgName        string            `json:"githubOrgName"`
	GithubRepoName       string            `json:"githubRepoName"`
	GithubToken          string            `json:"githubToken"`
	GitBranchType        string            `json:"gitBranchType"`
	GitBranchName        string            `json:"gitBranchName"`
	DockerfileLocation   string            `json:"dockerfileLocation"`
	DockerfileName       string            `json:"dockerfileName"`
	DockerTagName        string            `json:"dockerTagName"`
	DockerOrgName        string            `json:"dockerOrgName"`
	DockerRepoName       string            `json:"dockerRepoName"`
	D2HUBPushKeyID       string            `json:"d2hubPushKeyID"`
	D2HUBPushKeyPassword string            `json:"d2hubPushKeyPassword"`
	WatchCenterID        string            `json:"watchCenterID"`
	BuildID              string            `json:"buildID"`
	ResultCallbackURL    string            `json:"resultCallbackURL"`
	BuildArgStr          string            `json:"dockerbuildArg"`
	BuildArg             []docker.BuildArg `json:"-"`
}

func (b *BuildSpec) dockerImageNameWithTag() string {
	return fmt.Sprintf("%s:%s", b.dockerImageName(), b.DockerTagName)
}

func (b *BuildSpec) dockerImageName() string {
	return fmt.Sprintf("%s/%s/%s", dockerRegistryAddr, b.DockerOrgName, b.DockerRepoName)
}

func (b *BuildSpec) dockerfilePath() string {
	dockerfileName := defaultDockerfileName
	if b.DockerfileName != "" {
		dockerfileName = b.DockerfileName
	}
	return filepath.Join(b.DockerfileLocation, dockerfileName)
}

func (b *BuildSpec) d2hubAuth() docker.AuthConfiguration {
	return docker.AuthConfiguration{
		Username: b.D2HUBPushKeyID,
		Password: b.D2HUBPushKeyPassword,
		Email:    fmt.Sprintf("%s@%s", b.Username, emailDomain),
	}
}

// BuildResult is result that build docker image
type BuildResult struct {
	URL             string     `json:"-"`
	BuildID         string     `json:"buildID"`
	IsSuccess       bool       `json:"isSuccess"`
	ErrorReason     string     `json:"errorReason"`
	Logs            string     `json:"logs"`
	DockerImageName string     `json:"dockerImageName"`
	BuildSpec       *BuildSpec `json:"buildSpec"`
}

func (b *BuildResult) sendError(err error, logs []byte) {
	b.IsSuccess = false
	b.ErrorReason = err.Error()
	b.Logs = bytes.NewBuffer(logs).String()
	jsonBytes, _ := json.Marshal(b)
	http.DefaultClient.Post(b.URL, "application/json", bytes.NewReader(jsonBytes))

	if b.BuildSpec.WatchCenterID != "" {
		notifyMsg := fmt.Sprintf(`
[D2Hub] 이미지 빌드 실패 ㅠㅠ

<Docker>
이미지: %s

<GitHub>
URL: %s
Branch/Tag: %s

<빌드 결과>
URL: %s
	`,
			b.DockerImageName,
			fmt.Sprintf("http://%s/%s/%s", b.BuildSpec.GithubHost, b.BuildSpec.GithubOrgName, b.BuildSpec.GithubRepoName),
			b.BuildSpec.GitBranchName,
			fmt.Sprintf("%s/%s", buildHistoryPageURL, b.BuildID),
		)
		sendWatchCenterMessage(b.BuildSpec.WatchCenterID, notifyMsg)
	}
}

func (b *BuildResult) sendSuccess(logs []byte) {
	b.IsSuccess = true
	b.Logs = bytes.NewBuffer(logs).String()
	jsonBytes, _ := json.Marshal(b)
	http.DefaultClient.Post(b.URL, "application/json", bytes.NewReader(jsonBytes))

	if b.BuildSpec.WatchCenterID != "" {
		notifyMsg := fmt.Sprintf(`
[D2Hub] 이미지 빌드 성공 :)

<Docker>
이미지: %s

<GitHub>
URL: %s
Branch/Tag: %s

<빌드 결과>
URL: %s
	`,
			b.DockerImageName,
			fmt.Sprintf("http://github.com/%s/%s", b.BuildSpec.GithubOrgName, b.BuildSpec.GithubRepoName),
			b.BuildSpec.GitBranchName,
			fmt.Sprintf("%s/%s", buildHistoryPageURL, b.BuildID),
		)
		sendWatchCenterMessage(b.BuildSpec.WatchCenterID, notifyMsg)
	}
}
