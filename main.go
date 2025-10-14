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
		log.Error().Err(err).Msg("did not get projects by parentId")
	}

	// iterate over projects which satisfy conditions
	for _, project := range projects {
		// print project details
		fmt.Printf("ID: %v, Name: %v, Path: %v\n", project.ID, project.Name, project.Path)
	}
}

func getSubGroups(client *gitlab.Client, groupId int) ([]*gitlab.Group, error) {
	subgroups, r, err := client.Groups.ListSubGroups(groupId, &gitlab.ListSubGroupsOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("gitlab resp: %+v", r)
		return nil, err
	}
	return subgroups, nil
}

func getProjects(client *gitlab.Client, groupId int, searchTerm string) ([]*gitlab.Project, error) {
	projects, r, err := client.Groups.ListGroupProjects(groupId, &gitlab.ListGroupProjectsOptions{Search: gitlab.Ptr(searchTerm)})
	if err != nil {
		log.Error().Err(err).Msgf("gitlab resp: %+v", r)
		return nil, err
	}
	return projects, nil
}

func getProjectsByParentId(client *gitlab.Client, parentId int, path string, name string) ([]*gitlab.Project, error) {
	projectsTotal := []*gitlab.Project{}
	var searchTerm string

	if path != "" {
		searchTerm = path
	} else if name != "" {
		searchTerm = name
	} else {
		return nil, errors.New("no path or name provided")
	}

	// search the parent group for projects and append to total if found
	projectsFromParent, err := getProjects(client, parentId, searchTerm)
	if err != nil {
		log.Error().Err(err).Msg("did not find projects inside parent")
	}
	if len(projectsFromParent) > 0 {
		projectsTotal = append(projectsTotal, projectsFromParent...)
	}

	subgroups, err := getSubGroups(client, parentId)
	if err != nil {
		log.Error().Err(err).Msg("did not find subgroups inside parent")
	}
	// return projects if no subgroup is found
	if len(subgroups) == 0 {
		return projectsTotal, nil
	}

	// search inside subgroups
	for _, subgroup := range subgroups {
		projects, err := getProjects(client, subgroup.ID, searchTerm)
		if err != nil {
			log.Error().Err(err).Msg("did not find projects inside subgroup")
		}
		if len(projects) > 0 {
			projectsTotal = append(projectsTotal, projects...)
		}
	}

	return projectsTotal, nil
}
