package svc

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ZoranCalic/deps-list/internal/dto"
	"github.com/ZoranCalic/deps-list/internal/github"
	"github.com/minus5/svckit/log"
	"github.com/pkg/errors"
)

type DependancySvc struct {
	githubSvc           *github.Service
	ignoredRepositories []string
}

func NewDependancySvc(gs *github.Service, ignoredRepositories []string) *DependancySvc {
	return &DependancySvc{
		githubSvc:           gs,
		ignoredRepositories: ignoredRepositories,
	}
}

func (s *DependancySvc) ExtractDependancies() error {
	repos, err := s.githubSvc.GetOrgRepos()
	if err != nil {
		return errors.Wrap(err, "failed getting github org repos")
	}

	var dependancies []dto.Dependancy

	for _, r := range repos {
		log.Info("Processing repository " + r.Name)
		if r.Archived {
			// skip archived repositories
			continue
		}

		ignored := false
		for _, ignoredRepo := range s.ignoredRepositories {
			if r.Name == ignoredRepo {
				ignored = true
			}
		}
		if ignored {
			continue
		}

		langs, err := s.githubSvc.GetRepoLanguages(r.LanguagesURL)
		if err != nil {
			return errors.Wrap(err, "failed getting github repo languages")
		}

		// TODO - extract git clone into separate process - now if repo has both Go and Ruby code, git clone is execute twice
		if _, ok := langs["Go"]; ok {
			log.Info("Go repo found...")
			goDeps, err := s.extractGoDependancies(r)
			if err != nil {
				return errors.Wrap(err, "failed extracting go dependancies for repo "+r.Name)
			}
			dependancies = append(dependancies, goDeps...)
		}
		if _, ok := langs["Ruby"]; ok {
			log.Info("Ruby repo found...")
			rubyDeps, err := s.extractRubyDependancies(r)
			if err != nil {
				return errors.Wrap(err, "failed extracting ruby dependancies for repo "+r.Name)
			}
			dependancies = append(dependancies, rubyDeps...)
		}
	}

	err = s.writeDataToCSVFile(dependancies)
	if err != nil {
		return errors.Wrap(err, "failed writing data to csv file")
	}

	err = s.writeDataToSQLFile(dependancies)
	return errors.Wrap(err, "failed writing data to sql file")
}

func (s *DependancySvc) extractGoDependancies(r dto.GithubRepo) ([]dto.Dependancy, error) {
	cmd, err := exec.Command("/bin/sh", "../scripts/list_go_deps.sh", r.SshURL, r.Name).Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed executing go dependancies script")
	}
	output := string(cmd)
	log.Debug("Extracted dependancies: " + output)

	var dependancies []dto.Dependancy

	dependancyLines := strings.Split(output, "\n")
	for _, dependancyLine := range dependancyLines {
		if dependancyLine != "" {
			dep := strings.Split(dependancyLine, " ")
			if len(dep) == 2 {
				dependancies = append(dependancies, dto.Dependancy{
					ProgrammingLanguage: "Go",
					Repository:          r.HtmlURL,
					DependancyName:      dep[0],
					DependancyVersion:   dep[1],
				})
			} else {
				log.Info("Invalid dependancy: " + dependancyLine)
			}
		}
	}
	return dependancies, nil
}

func (s *DependancySvc) extractRubyDependancies(r dto.GithubRepo) ([]dto.Dependancy, error) {
	cmd, err := exec.Command("/bin/sh", "../scripts/list_ruby_deps.sh", r.SshURL, r.Name).Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed executing ruby dependancies script")
	}
	output := string(cmd)
	log.Debug("Extracted dependancies: " + output)

	var dependancies []dto.Dependancy

	dependancyLines := strings.Split(output, "\n")
	for _, dependancyLine := range dependancyLines {
		// first line of the sh script output is "Gems included by the bundle", it should be ignored
		if dependancyLine != "" && !strings.HasPrefix(dependancyLine, "Gems included by the bundle") {
			// dependancy line example: "  * actioncable (5.1.1)"
			line := strings.TrimLeft(strings.TrimSpace(dependancyLine), "* ")
			dep := strings.Split(line, " ")
			if len(dep) == 2 {
				dependancies = append(dependancies, dto.Dependancy{
					ProgrammingLanguage: "Ruby",
					Repository:          r.HtmlURL,
					DependancyName:      dep[0],
					DependancyVersion:   strings.TrimRight(strings.TrimLeft(dep[1], "("), ")"),
				})
			} else {
				log.Info("Invalid dependancy: " + dependancyLine)
			}
		}
	}
	return dependancies, nil
}

func (s *DependancySvc) writeDataToSQLFile(dependancies []dto.Dependancy) error {
	file, err := os.Create("../dependancies.sql")
	if err != nil {
		return errors.Wrap(err, "failed creating dependancies file")
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	sqlCommand := "INSERT INTO [schema_name].[table_name]\n(programming_language, repository, dependancy_name, dependancy_version)\nVALUES\n"
	if _, err = w.WriteString(sqlCommand + "\n"); err != nil {
		return errors.Wrap(err, "failed writing to dependancies file")
	}

	for i, dependancy := range dependancies {
		row := fmt.Sprintf("('%s', '%s', '%s', '%s')", dependancy.ProgrammingLanguage, dependancy.Repository, dependancy.DependancyName, dependancy.DependancyVersion)
		if i+1 < len(dependancies) {
			row += ","
		} else {
			row += ";"
		}

		if _, err = w.WriteString(row + "\n"); err != nil {
			return errors.Wrap(err, "failed writing to dependancies file")
		}
	}
	return nil
}

func (s *DependancySvc) writeDataToCSVFile(dependancies []dto.Dependancy) error {
	file, err := os.Create("../dependancies.csv")
	if err != nil {
		return errors.Wrap(err, "failed creating dependancies file")
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()
	for _, dependancy := range dependancies {
		row := []string{dependancy.ProgrammingLanguage, dependancy.Repository, dependancy.DependancyName, dependancy.DependancyVersion}
		if err := w.Write(row); err != nil {
			return errors.Wrap(err, "failed writing to dependancies file")
		}
	}
	return nil
}
