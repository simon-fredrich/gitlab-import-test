package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	// "log"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Print("Hello from zerolog.")

	projectPathWithNamespace := flag.String("path", "gitlab-org/api/client-go", "Path with namespace to identify project.")
	projectName := flag.String("name", "client-go", "Name of the project to identify project.")
	flag.Parse()

	token := os.Getenv("GITLAB_API_KEY")
	url := os.Getenv("GITLAB_URL")

	project, err := getProject(token, url, *projectName, *projectPathWithNamespace)

	if err != nil {
		log.Error().Err(err).Msg("Error getting the project details: ")
	}

	// Print project details
	fmt.Printf("ID: %v, Name: %v, PathWithNamespace: %v", project.ID, project.Name, project.PathWithNamespace)
}

func getProject(token string, url string, projectName string, projectPathWithNamespace string) (Project, error) {
	if token == "" {
		log.Error().Msg("GITLAB_API_KEY is not set.")
	}

	if url == "" {
		log.Error().Msg("GITLAB_URL is not set.")
	}

	if projectName == "" {
		log.Print("projectName not provided.")
	}

	if projectPathWithNamespace == "" {
		log.Error().Msg("projectPathWithNamespace not provided.")
	}

	// Create new search utilizing the name of the project
	req, err := http.NewRequest("GET", url+"/api/v4/search?scope=projects&search="+projectPathWithNamespace, nil)
	if err != nil {
		log.Error().Msg("Error creating requests: ")
	}

	// Providing private token
	req.Header.Set("PRIVATE-TOKEN", token)

	// Creating a http client and performing request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error making request:")
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body:")
	}

	// Unmarshal response body
	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		log.Error().Err(err).Msg("Error unmarshaling body: ")
	}

	// Find project correspondint to `Name` and `PathWithNamespace`
	for _, project := range projects {
		if (project.Name == projectName) && (project.PathWithNamespace == projectPathWithNamespace) {
			return project, nil
		}
	}

	return Project{}, errors.New("Project not found")
}
