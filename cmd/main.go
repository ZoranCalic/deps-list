package main

import (
	"flag"
	"strings"

	"github.com/ZoranCalic/deps-list/internal/github"
	"github.com/ZoranCalic/deps-list/internal/svc"
	"github.com/minus5/svckit/log"
)

var (
	githubOrgName      = "minus5"
	githubAuthToken    = ""
	ignoreRepositories = ""
)

func init() {
	flag.StringVar(&githubOrgName, "o", githubOrgName, "GitHub organisation name")
	flag.StringVar(&githubAuthToken, "a", githubAuthToken, "GitHub auth token")
	flag.StringVar(&ignoreRepositories, "ir", ignoreRepositories, "comma separated list of repository names to ignore")
	flag.Parse()
}

func main() {
	ignoredRepositories := strings.Split(ignoreRepositories, ",")
	for i := range ignoredRepositories {
		ignoredRepositories[i] = strings.TrimSpace(ignoredRepositories[i])
	}

	githubSvc := github.NewGitHubService(githubOrgName, githubAuthToken)
	depSvc := svc.NewDependancySvc(githubSvc, ignoredRepositories)
	err := depSvc.ExtractDependancies()
	if err != nil {
		log.Fatal(err)
	}
}
