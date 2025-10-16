package main

import (
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

	path := flag.String("path", "", "Provide the path of desired project.")
	namespaceId := flag.Int("namespaceId", 0, "namespaceId of desired project.")
	parentId := flag.Int("parentId", 0, "Provide the parentId of desired project.")
	flag.Parse()

	// check flag values
	if *path == "" {
		log.Fatal().Msg("please specify a path to compare against")
	}
	if *namespaceId == 0 && *parentId == 0 {
		log.Fatal().Msg("neither namespaceId nor parentId")
	} else if *namespaceId != 0 && *parentId != 0 {
		log.Fatal().Msg("specify namespaceId OR parentId")
	}

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

	// try to get project with namespaceId and path
	if *namespaceId != 0 {
		projectId, err := getProject(git, *namespaceId, *path)
		if err != nil || projectId == 0 {
			log.Fatal().Err(err).Msg("can't get projectId")
		}
		project, resp, err := git.Projects.GetProject(projectId, &gitlab.GetProjectOptions{})
		if err != nil {
			log.Error().Err(err).Msgf("can't get project, gitlab response: %+v", resp)
		}

		fmt.Println("Found a project:")
		fmt.Println("--------------")
		fmt.Printf("ID: %v,\nGroupID: %v,\nName: %v, Path: %v,\nPathWithNamespace: %v\n", project.ID, project.Namespace.ID, project.Name, project.Path, project.PathWithNamespace)
	} else {
		fmt.Println("namespaceId not specified, trying to find group now...")
	}

	// try to get group with parentId and path
	if *parentId != 0 {
		groupId, err := getGroup(git, *parentId, *path)
		if err != nil || groupId == 0 {
			log.Fatal().Err(err).Msg("can't get groupId")
		}
		group, resp, err := git.Groups.GetGroup(groupId, &gitlab.GetGroupOptions{})
		if err != nil {
			log.Error().Err(err).Msgf("can't get group, gitlab response: %+v", resp)
		}

		fmt.Println("Found a group:")
		fmt.Println("--------------")
		fmt.Printf("ID: %v,\nParentGroupID: %v,\nName: %v,\nPath: %v,\nNamespace: %v\n", group.ID, group.ParentID, group.Name, group.Path, group.FullPath)
	} else {
		fmt.Println("parentId not specified, please specify namespaceId/parentId to find project/group")
	}
}

// getProject returns the `projectId` for a given `namespaceId` and `path`
func getProject(client *gitlab.Client, namespaceId int, path string) (int, error) {
	// namespaceId is the ID of the group containing the desired project
	parentId := namespaceId

	// find project based on path
	projects, err := getProjects(client, parentId, "")
	if err != nil {
		log.Error().Err(err).Msgf("can't get projects")
		return 0, err
	}
	for _, project := range projects {
		if project.Path == path {
			return project.ID, nil
		}
	}
	return 0, fmt.Errorf("there is no project with matching path in namespace with ID %+v", namespaceId)
}

// getGroup returns the `groupId` for a given `parentId` and `path`
func getGroup(client *gitlab.Client, parentId int, path string) (int, error) {
	// find group based on path
	groups, err := getSubGroups(client, parentId)
	if err != nil {
		log.Error().Err(err).Msgf("can't get subgroups")
		return 0, err
	}
	for _, group := range groups {
		if group.Path == path {
			return group.ID, nil
		}
	}
	return 0, fmt.Errorf("there is no project with matching path in parent group with id: %+v", parentId)
}

// getSubGroups returns all groups of a given parent group
func getSubGroups(client *gitlab.Client, groupId int) ([]*gitlab.Group, error) {
	subgroupsTotal := []*gitlab.Group{}
	page := 1

	// iterate over all pages to retrieve all possible subgroups
	for {
		opt := &gitlab.ListSubGroupsOptions{
			AllAvailable: gitlab.Ptr(true),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    page,
			},
		}

		subgroups, resp, err := client.Groups.ListSubGroups(groupId, opt)
		if err != nil {
			log.Error().Err(err).Msgf("gitlab resp: %+v", resp)
			return nil, err
		}
		subgroupsTotal = append(subgroupsTotal, subgroups...)

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		page++
	}

	return subgroupsTotal, nil
}

// getProjects returns all projects of a given parent group and has
func getProjects(client *gitlab.Client, groupId int, searchTerm string) ([]*gitlab.Project, error) {
	projectsTotal := []*gitlab.Project{}
	page := 1

	// iterate over all pages to retrieve all possible projects in group with the given groupId
	for {
		opt := &gitlab.ListGroupProjectsOptions{
			Search: gitlab.Ptr(searchTerm),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    page,
			},
		}

		projects, resp, err := client.Groups.ListGroupProjects(groupId, opt)
		if err != nil {
			log.Error().Err(err).Msgf("gitlab resp: %+v", resp)
			return nil, err
		}

		projectsTotal = append(projectsTotal, projects...)

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		page++
	}

	return projectsTotal, nil
}

// function might not be needed anymore
func getProjectsByParentId(client *gitlab.Client, parentId int, path string, name string) ([]*gitlab.Project, error) {
	projectsTotal := []*gitlab.Project{}
	var searchTerm string

	if path != "" {
		searchTerm = path
	} else if name != "" {
		searchTerm = name
	} else {
		log.Print("no path or name provided")
		searchTerm = ""
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
		log.Print("no subgroups found")
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
