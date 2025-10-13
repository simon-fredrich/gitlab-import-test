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

	// iterate over projects which satisfy conditions
	for _, project := range projects {
		// print project details
		fmt.Printf("ID: %v, Name: %v, Path: %v\n", project.ID, project.Name, project.Path)
	}
}

func getProjectsByParentId(client *gitlab.Client, parentId int, path string, name string) ([]*gitlab.Project, error) {
	// Get subgroups to later display their projects
	subgroups, _, err := client.Groups.ListSubGroups(parentId, &gitlab.ListSubGroupsOptions{})
	if err != nil {
		return nil, err
	}

	for _, subgroup := range subgroups {
		// search projects of subgroup by path
		if path != "" {
			log.Print("Trying path to find project ...")
			projects, _, err := client.Groups.ListGroupProjects(subgroup.ID, &gitlab.ListGroupProjectsOptions{Search: gitlab.Ptr(path)})
			if err != nil {
				return nil, err
			}

			// return if at least one project was found
			if len(projects) > 0 {
				log.Print("path was successful")
				return projects, nil
			}
		}

		// search projects of subgroup by name
		if name != "" {
			log.Print("path was unsufficient or empty, trying name...")
			projects, _, err := client.Groups.ListGroupProjects(subgroup.ID, &gitlab.ListGroupProjectsOptions{Search: gitlab.Ptr(name)})
			if err != nil {
				return nil, err
			}

			// return if at least one project was found
			if len(projects) > 0 {
				log.Print("name was successful")
				return projects, nil
			}
		}

		// both path and name where not provided
		log.Error().Msg("path and name where both unsufficient or empty.")
	}

	return nil, errors.New("no subgroups or projects have been found")
}
