package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Event represents a single event from the GitHub API.
// We only define the fields we need to parse.
type Event struct {
	Type    string  `json:"type"`
	Repo    Repo    `json:"repo"`
	Payload Payload `json:"payload"`
}

// Repo contains information about the repository.
type Repo struct {
	Name string `json:"name"`
}

// Issue contains details about an issue or pull request.
type Issue struct {
	Title string `json:"title"`
}

// Forkee contains information about the forked repository.
type Forkee struct {
	FullName string `json:"full_name"`
}

// Payload contains event-specific details.
type Payload struct {
	Action      string `json:"action"`
	RefType     string `json:"ref_type"`
	Commits     []any  `json:"commits"` // We only need the count, so the type doesn't matter.
	Issue       Issue  `json:"issue"`
	Forkee      Forkee `json:"forkee"`
	PullRequest Issue  `json:"pull_request"`
}

func main() {
	// Check if a username was provided as a command-line argument
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run github_activity.go <username>")
		os.Exit(1)
	}

	githubUsername := os.Args[1]
	getGithubActivity(githubUsername)
}
func getGithubActivity(username string) {
	// Construct the API URL
	apiURL := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	// Make the HTTP GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("Error: Could not reach GitHub API. Reason: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Handle non-200 status codes
	if resp.StatusCode == 404 {
		fmt.Printf("Error: Could not find GitHub user '%s'.\n", username)
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Error: Received status code %d from GitHub API.\n", resp.StatusCode)
		return
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: Failed to read response body. Reason: %v\n", err)
		return
	}

	// Unmarshal the JSON data into a slice of Event structs
	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Printf("Error: Failed to parse the response from the GitHub API. Reason: %v\n", err)
		return
	}

	fmt.Printf("Recent Activity for %s:\n\n", username)

	if len(events) == 0 {
		fmt.Println("No recent public activity found.")
		return
	}

	// Process and display each event
	for _, event := range events {
		switch event.Type {
		case "PushEvent":
			fmt.Printf("- Pushed %d commit(s) to %s\n", len(event.Payload.Commits), event.Repo.Name)
		case "CreateEvent":
			fmt.Printf("- Created a new %s in %s\n", event.Payload.RefType, event.Repo.Name)
		case "IssuesEvent":
			fmt.Printf("- %s an issue in %s: \"%s\"\n", strings.Title(event.Payload.Action), event.Repo.Name, event.Payload.Issue.Title)
		case "IssueCommentEvent":
			fmt.Printf("- Commented on an issue in %s: \"%s\"\n", event.Repo.Name, event.Payload.Issue.Title)
		case "WatchEvent":
			fmt.Printf("- %s watching %s\n", strings.Title(event.Payload.Action), event.Repo.Name)
		case "ForkEvent":
			fmt.Printf("- Forked %s to %s\n", event.Repo.Name, event.Payload.Forkee.FullName)
		case "PullRequestEvent":
			fmt.Printf("- %s a pull request in %s: \"%s\"\n", strings.Title(event.Payload.Action), event.Repo.Name, event.Payload.PullRequest.Title)
		case "PublicEvent":
			fmt.Printf("- Made %s public\n", event.Repo.Name)
		default:
			fmt.Printf("- Performed a %s on %s\n", event.Type, event.Repo.Name)
		}
	}
}
