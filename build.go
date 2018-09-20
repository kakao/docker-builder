package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	githubProtocol = "https"
)

var (
	dockerClient    *docker.Client
	pushRequestCh   = make(chan *pushImageRequest)
	pushImageReqMap = map[string]chan *pushImageRequest{}
)

type pushImageRequest struct {
	BuildSpec   *BuildSpec
	BuildResult *BuildResult
	OutBuffer   *bytes.Buffer
	ResultCh    chan error
}

func init() {
	if dockerCertPath != "" {
		dockerClient, _ = docker.NewTLSClient(dockerHost,
			fmt.Sprintf("%s/cert.pem", dockerCertPath),
			fmt.Sprintf("%s/key.pem", dockerCertPath),
			fmt.Sprintf("%s/ca.pem", dockerCertPath),
		)
	} else {
		dockerClient, _ = docker.NewClient(dockerHost)
	}

	go func() {
		for {
			pushReq := <-pushRequestCh

			imageName := pushReq.BuildSpec.dockerImageName()
			if _, ok := pushImageReqMap[imageName]; !ok {
				pushImageReqMap[imageName] = make(chan *pushImageRequest)
				go func() {
					for {
						select {
						case pushImgReq := <-pushImageReqMap[imageName]:
							err := pushDockerImage(pushImgReq.BuildSpec, os.Stdout)
							if err == nil {
								pushImgReq.BuildResult.sendSuccess(pushImgReq.OutBuffer.Bytes())
							} else {
								log.Printf("push docker image error: %s\n", err)
								pushImgReq.BuildResult.sendError(err, pushImgReq.OutBuffer.Bytes())
							}
						}
					}
				}()
			}
			pushImageReqMap[imageName] <- pushReq

			pushReq.ResultCh <- nil
		}
	}()
}

type multipleWriter struct {
	Writers []io.Writer
}

func (s *multipleWriter) Write(p []byte) (n int, err error) {
	for _, w := range s.Writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return
}

func build(buildSpec *BuildSpec) error {
	buildResult := &BuildResult{
		URL:             buildSpec.ResultCallbackURL,
		BuildID:         buildSpec.BuildID,
		DockerImageName: buildSpec.dockerImageNameWithTag(),
		BuildSpec:       buildSpec,
	}

	if buildSpec.WatchCenterID != "" {
		notifyMsg := fmt.Sprintf(`
[D2Hub] 이미지 빌드 시작!

<Docker>
이미지: %s

<GitHub>
URL: %s
Branch/Tag: %s

<빌드 결과>
URL: %s
	`,
			buildResult.DockerImageName,
			fmt.Sprintf("http://%s/%s/%s", buildSpec.GithubHost, buildSpec.GithubOrgName, buildSpec.GithubRepoName),
			buildSpec.GitBranchName,
			fmt.Sprintf("%s/%s", buildHistoryPageURL, buildResult.BuildID),
		)
		sendWatchCenterMessage(buildSpec.WatchCenterID, notifyMsg)
	}

	projectDir, err := ioutil.TempDir("", fmt.Sprintf("docker%s", buildSpec.BuildID))
	defer os.RemoveAll(projectDir)
	if err != nil {
		return err
	}

	outBuffer := &bytes.Buffer{}

	if err := cloneGit(buildSpec, projectDir, os.Stdout); err != nil {
		log.Printf("clone git error: %s\n", err)
		buildResult.sendError(err, outBuffer.Bytes())
		return err
	}

	if err := buildDockerImage(buildSpec, projectDir, &multipleWriter{
		Writers: []io.Writer{
			outBuffer,
			os.Stdout,
		},
	}); err != nil {
		log.Printf("build docker image error: %s\n", err)
		buildResult.sendError(err, outBuffer.Bytes())
		return err
	}

	pushReq := &pushImageRequest{
		BuildSpec:   buildSpec,
		BuildResult: buildResult,
		OutBuffer:   outBuffer,
		ResultCh:    make(chan error),
	}

	pushRequestCh <- pushReq
	return <-pushReq.ResultCh
}

func cloneGit(buildSpec *BuildSpec, projectDir string, outputStream io.Writer) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: buildSpec.GithubToken},
	)

	githubClient := github.NewClient(oauth2.NewClient(ctx, ts))
	githubClient.BaseURL, _ = url.Parse(fmt.Sprintf("%s://api.%s/", githubProtocol, buildSpec.GithubHost))

	repo, resp, err := githubClient.Repositories.Get(buildSpec.GithubOrgName, buildSpec.GithubRepoName)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad status code")
	}

	cloneURL := fmt.Sprintf("%s://%s@%s/%s.git", githubProtocol, buildSpec.GithubToken, buildSpec.GithubHost, *repo.FullName)
	log.Printf("git cloneURL: %s\n", cloneURL)
	if buildSpec.GitBranchType == "Branch" {
		cmd := exec.Command("git", "clone", "-b", buildSpec.GitBranchName, cloneURL, projectDir)
		cmd.Stdout = outputStream
		cmd.Stderr = outputStream
		if err = cmd.Run(); err != nil {
			return err
		}
	}
	if buildSpec.GitBranchType == "Tag" {
		cmd := exec.Command("git", "clone", cloneURL, projectDir)
		cmd.Stdout = outputStream
		cmd.Stderr = outputStream
		if err = cmd.Run(); err != nil {
			return err
		}
		cmd = exec.Command("git", "checkout", buildSpec.GitBranchName)
		cmd.Dir = projectDir
		cmd.Stdout = outputStream
		cmd.Stderr = outputStream
		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if existsGitmodulesFile(projectDir) {
		log.Println("exists .gitmodules file")
		if err = updateSubmodules(projectDir, buildSpec.GithubHost, buildSpec.GithubToken, outputStream); err != nil {
			return err
		}
	}

	return nil
}

func buildDockerImage(buildSpec *BuildSpec, projectDir string, outputStream io.Writer) error {
	dockerFileFullPath := filepath.Join(projectDir, buildSpec.dockerfilePath())
	if _, err := os.Stat(dockerFileFullPath); os.IsNotExist(err) {
		return err
	}

	cmd := exec.Command("tar", "cC", filepath.Dir(dockerFileFullPath), ".")
	tarRead, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	buildOpts := docker.BuildImageOptions{
		Name:         buildSpec.dockerImageNameWithTag(),
		BuildArgs:    buildSpec.BuildArg,
		NoCache:      true,
		Pull:         true,
		InputStream:  tarRead,
		OutputStream: outputStream,
	}

	if buildSpec.DockerfileName != "" {
		buildOpts.Dockerfile = buildSpec.DockerfileName
	}

	if err := dockerClient.BuildImage(buildOpts); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func pushDockerImage(buildSpec *BuildSpec, outputStream io.Writer) error {
	opts := docker.PushImageOptions{
		Name:         buildSpec.dockerImageName(),
		Tag:          buildSpec.DockerTagName,
		Registry:     dockerRegistryAddr,
		OutputStream: outputStream,
	}
	if err := dockerClient.PushImage(opts, buildSpec.d2hubAuth()); err != nil {
		return err
	}
	return nil
}

func existsGitmodulesFile(projectDir string) bool {
	filePath := filepath.Join(projectDir, ".gitmodules")
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func updateSubmodules(projectDir, githubHost, githubToken string, outputStream io.Writer) error {
	configCmd := exec.Command("bash", "-c", "git config --file=.gitmodules -l | grep .url")
	configCmd.Dir = projectDir
	submoduleBytes, err := configCmd.Output()
	if err != nil {
		log.Printf("searching submodule URL error: %s", err.Error())
	}
	if len(submoduleBytes) == 0 {
		return nil
	}

	submoduleStrs := strings.Split(fmt.Sprintf("%s", bytes.TrimSpace(submoduleBytes)), "\n")
	log.Printf("submoduleStrs: %s", submoduleStrs)

	for _, submoduleStr := range submoduleStrs {
		splitValues := strings.Split(submoduleStr, "=")
		if len(splitValues) != 2 {
			return fmt.Errorf("Invalid submodule spec: %s", submoduleStr)
		}
		subKey := splitValues[0]
		gitURL := splitValues[1]
		if strings.HasPrefix(gitURL, "git@") {
			gitURL = strings.Replace(gitURL, "git@", "https://", 1)
			gitURL = strings.Replace(gitURL, fmt.Sprintf("%s:", githubHost), fmt.Sprintf("%s/", githubHost), 1)
		}
		gitURL = strings.Replace(gitURL, fmt.Sprintf("https://%s", githubHost), fmt.Sprintf("https://%s@%s", githubToken, githubHost), 1)
		log.Printf("Submodule github URL: %s\n", gitURL)

		updateSubmoduleCmd := exec.Command("git", "config", "--file=.gitmodules", subKey, gitURL)
		updateSubmoduleCmd.Dir = projectDir
		updateSubmoduleCmd.Stdout = outputStream
		updateSubmoduleCmd.Stderr = outputStream
		if err := updateSubmoduleCmd.Run(); err != nil {
			return err
		}
	}

	subUpdateCmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	subUpdateCmd.Dir = projectDir
	subUpdateCmd.Stdout = outputStream
	subUpdateCmd.Stderr = outputStream
	return subUpdateCmd.Run()
}
