package main

import (
	"github.com/frhwang/gopkg/env"
)

var (
	serverPort          = env.GetOrDefault("SERVER_PORT", "3000")
	dockerCertPath      = env.GetOrDefault("DOCKER_CERT_PATH", "")
	dockerHost          = env.GetOrDefault("DOCKER_HOST", "unix:///var/run/docker.sock")
	dockerRegistryAddr  = env.GetOrDefault("DOCKER_REGISTRY_ADDR", "")
	watchCenterURL      = env.GetOrDefault("WATCH_CENTER_URL", "")
	emailDomain         = env.GetOrDefault("EMAIL", "kakaocorp.com")
	buildHistoryPageURL = env.GetOrDefault("BUILD_HISTORY_URL", "")
)
