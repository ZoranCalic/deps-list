package dto

import "time"

type GithubRepo struct {
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	HtmlURL       string    `json:"html_url"`
	LanguagesURL  string    `json:"languages_url"`
	CreatedAt     time.Time `json:"created_at"`
	SshURL        string    `json:"ssh_url"`
	Language      string    `json:"language"`
	Visibility    string    `json:"visibility"`
	Archived      bool      `json:"archived"`
	DefaultBranch string    `json:"default_branch"`
}
