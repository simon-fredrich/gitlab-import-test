package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	projectPathWithNamespace := flag.String("path", "gitlab-org/api/client-go", "Path with namespace to identify project.")
	projectName := flag.String("name", "client-go", "Name of the project to identify project.")
	flag.Parse()

	token := getEnvVar("GITLAB_API_KEY")
	url := getEnvVar("GITLAB_URL")

	project := getProject(token, url, *projectName, *projectPathWithNamespace)

	// Print project details
	fmt.Printf("ID: %v, Name: %v, PathWithNamespace: %v", project.ID, project.Name, project.PathWithNamespace)

}

func getProject(token string, url string, projectName string, projectPathWithNamespace string) Project {
	if token == "" {
		log.Fatal("GITLAB_API_KEY is not set.")
	}

	if url == "" {
		log.Fatal("GITLAB_URL is not set.")
	}

	// Create new search utilizing the name of the project
	req, err := http.NewRequest("GET", url+"/api/v4/search?scope=projects&search="+projectPathWithNamespace, nil)
	if err != nil {
		log.Fatal("Error creating requests: ", err)
	}

	// Providing private token
	req.Header.Set("PRIVATE-TOKEN", token)

	// Creating a http client and performing request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	// Unmarshal response body
	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		log.Fatal("Error unmarshaling body: ", err)
	}

	// Find project correspondint to `Name` and `PathWithNamespace`
	for _, project := range projects {
		if (project.Name == projectName) && (project.PathWithNamespace == projectPathWithNamespace) {
			return project
		}
	}

	return Project{}
}

func getEnvVar(key string) string {
	// Find .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	return os.Getenv(key)
}
