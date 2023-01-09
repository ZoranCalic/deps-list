package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/ZoranCalic/deps-list/internal/dto"
	"github.com/minus5/svckit/log"
	"github.com/pkg/errors"
)

const (
	apiBaseURL = "https://api.github.com/"

	apiPathGetOrgRepos = "/orgs/{org}/repos"
)

type Service struct {
	client           *http.Client
	organisationName string
	authToken        string
}

func NewGitHubService(orgName, authToken string) *Service {
	service := &Service{
		organisationName: orgName,
		authToken:        authToken,
	}

	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(10000) * time.Millisecond,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(10000) * time.Millisecond,
	}

	service.client = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(20000) * time.Millisecond,
	}

	return service
}

func (s *Service) GetOrgRepos() ([]dto.GithubRepo, error) {
	apiPath := strings.ReplaceAll(apiPathGetOrgRepos, "{org}", s.organisationName)
	qParams := map[string]string{}
	qParams["per_page"] = "100" // TODO: If organisation has more than 100 repos, pagination should be used!

	var repos []dto.GithubRepo
	err := s.executeAPICall(apiPath, http.MethodGet, qParams, nil, &repos)
	return repos, errors.Wrap(err, "executing api call failed")
}

func (s *Service) GetRepoLanguages(url string) (map[string]int, error) {
	var langs map[string]int
	err := s.executeAPICall(strings.TrimPrefix(url, apiBaseURL), http.MethodGet, nil, nil, &langs)
	return langs, errors.Wrap(err, "executing api call failed")
}

func (s *Service) executeAPICall(apiPath, httpMethod string, queryParams map[string]string, reqBody, rspBody interface{}) error {
	apiURL, err := url.ParseRequestURI(apiBaseURL)
	if err != nil {
		return errors.Wrap(err, "failed parsing request uri")
	}
	apiURL.Path = path.Join(apiURL.Path, apiPath)
	urlStr := apiURL.String()

	var body io.Reader
	if reqBody != nil {
		bs, err := json.Marshal(reqBody)
		if err != nil {
			return errors.Wrap(err, "failed marshalling request body")
		}
		body = bytes.NewBuffer(bs)
	}

	req, err := http.NewRequest(httpMethod, urlStr, body)
	if err != nil {
		return errors.Wrap(err, "failed creating request")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.authToken)

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	log.Info(fmt.Sprintf("GitHub API [%s] Request to %s", req.Method, req.URL.String()))

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed executing request")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		log.Info("GitHub API call executed successfully [200 OK].")
		if rspBody != nil {
			if err := unmarshalBody(resp, &rspBody); err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed unmarshalling response body for HTTP status code %d", resp.StatusCode))
			}
			log.Info(fmt.Sprintf("GitHub API Response: %v", rspBody))
		}
	default:
		log.Info(fmt.Sprintf("GitHub API call unsuccessful [HTTP Status Code %d].", resp.StatusCode))
		return errors.Wrap(errors.New("github api error"), fmt.Sprintf("github failed to process the request with HTTP status code %d", resp.StatusCode))
	}

	return nil
}

func unmarshalBody(r *http.Response, rsp interface{}) error {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, "failed reading response body")
	}
	if err := json.Unmarshal(buf, rsp); err != nil {
		return errors.Wrap(err, "failed unmarshalling response body")
	}
	return nil
}
