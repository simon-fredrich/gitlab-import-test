package main

import (
	"errors"
	"flag"
	"fmt"

	// "log"

	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Print("Hello from zerolog.")

	// projectPathWithNamespace := flag.String("projectPathWithNamespace", "gitlab-org/api/client-go", "Path with namespace to identify project.")
	projectPath := flag.String("path", "", "Provide the path of desired project.")
	projectName := flag.String("name", "", "Name of the project to identify desired project.")
	parentId := flag.Int("parentId", 0, "Provide the parentId of desired project.")
	flag.Parse()

	// get environment variables and check if they are initilized
	token := os.Getenv("GITLAB_API_KEY")
	if token == "" {
		log.Error().Msg("GITLAB_API_KEY is not set.")
	}

	url := os.Getenv("GITLAB_URL")
	if url == "" {
		log.Error().Msg("GITLAB_URL is not set.")
	}

	// Create a new instance of the gitlab api "client-go"
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(url+"/api/v4"))
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong creating a new api client.")
	}

	projects, err := getProjectsByParentId(git, *parentId, *projectPath, *projectName)
	if err != nil {
		log.Error().Err(err)
	}

	for _, project := range projects {
		fmt.Println(project.Name)
	}
}

func getProjectsByParentId(client *gitlab.Client, parentId int, path string, name string) ([]*gitlab.Project, error) {
	// Get subgroups to later display their projects
	subgroups, _, err := client.Groups.ListSubGroups(parentId, &gitlab.ListSubGroupsOptions{})
	if err != nil {
		return nil, err
	}

	for _, subgroup := range subgroups {
		fmt.Println(subgroup.Name)

		// search projects of subgroup by path or name
		if path != "" {
			log.Print("Trying path to find project ...")
			projects, _, err := client.Groups.ListGroupProjects(subgroup.ID, &gitlab.ListGroupProjectsOptions{Search: gitlab.Ptr(path)})
			if err != nil {
				return nil, err
			}

			// utilize name parameter if path based approach did not work
			if len(projects) == 0 && name != "" {
				log.Print("No Project found using path. Trying name now...")
				projects, _, err := client.Groups.ListGroupProjects(subgroup.ID, &gitlab.ListGroupProjectsOptions{Search: gitlab.Ptr(name)})
				if err != nil {
					return nil, err
				}

				return projects, nil
			} else {
				return projects, nil
			}
		}
	}

	return nil, errors.New("No Projects have been found.")
}
