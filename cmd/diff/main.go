package main

import (
	"ImDevinC/plex-meta-manager-configs/internal/forgejo"
	"ImDevinC/plex-meta-manager-configs/internal/gh"
	"ImDevinC/plex-meta-manager-configs/internal/issueclient"
	"ImDevinC/plex-meta-manager-configs/internal/pmm"
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	var sourceFile string
	var dryrun bool
	flag.StringVar(&sourceFile, "source", "", "Source file")
	flag.BoolVar(&dryrun, "dryrun", false, "If enabled, will not create issue")
	flag.Parse()

	f, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Fatal(err)
	}

	var payload pmm.Config
	err = yaml.Unmarshal(f, &payload)
	if err != nil {
		log.Fatal(err)
	}

	client := buildClient()

	for k, v := range payload.Metadata {
		if v.PosterURL == "" {
			err := client.CheckForExistingMovieIssue(context.Background(), k)
			var existingErr issueclient.ErrAlreadyExists
			var ignoredErr issueclient.ErrIgnored
			if errors.As(err, &ignoredErr) {
				log.Printf("Issue for movie is marked ignored, skipping: %s\n", k)
				continue
			}
			if errors.As(err, &existingErr) {
				log.Printf("Issue already exists for missing movie: %s\n", k)
				continue
			}
			if err != nil {
				log.Printf("ERROR: %s\n", err)
				continue
			}
			if dryrun {
				log.Printf("Would create issue for movie: %s, but dryrun enabled.\n", k)
				continue
			}
			err = client.AddMissingMovie(context.Background(), k)
			if err != nil {
				log.Printf("ERROR: %s\n", err)
			}
			log.Printf("Created issue for movie: %s\n", k)
		}
	}
}

func buildClient() issueclient.IssueClient {
	serverType := strings.ToLower(os.Getenv("SERVER_TYPE"))
	if serverType == "" {
		serverType = "github"
	}

	switch serverType {
	case "forgejo":
		return buildForgejoClient()
	case "github":
		return buildGitHubClient()
	default:
		log.Fatalf("unsupported SERVER_TYPE: %s", serverType)
		return nil
	}
}

func buildGitHubClient() issueclient.IssueClient {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("missing required GITHUB_TOKEN environment variable")
	}

	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		log.Fatal("missing required GITHUB_REPO environment variable")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		log.Fatal("missing required GITHUB_OWNER environment variable")
	}

	return gh.NewGitHubClient(context.Background(), githubToken, owner, repo)
}

func buildForgejoClient() issueclient.IssueClient {
	url := os.Getenv("FORGEJO_URL")
	if url == "" {
		log.Fatal("missing required FORGEJO_URL environment variable")
	}

	token := os.Getenv("FORGEJO_TOKEN")
	if token == "" {
		log.Fatal("missing required FORGEJO_TOKEN environment variable")
	}

	repo := os.Getenv("FORGEJO_REPO")
	if repo == "" {
		log.Fatal("missing required FORGEJO_REPO environment variable")
	}

	owner := os.Getenv("FORGEJO_OWNER")
	if owner == "" {
		log.Fatal("missing required FORGEJO_OWNER environment variable")
	}

	assignee := os.Getenv("FOREGEJO_ASSIGNEE")
	if assignee == "" {
		log.Print("missing FOREGEJO_ASSIGNEE, using FORGEJO_OWNER")
		assignee = owner
	}

	client, err := forgejo.NewForgejoClient(context.Background(), url, token, owner, repo, assignee)
	if err != nil {
		log.Fatalf("failed to create forgejo client: %s", err)
	}
	return client
}
