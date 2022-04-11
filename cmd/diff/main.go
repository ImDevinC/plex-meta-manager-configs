package main

import (
	"ImDevinC/plex-meta-manager-configs/internal/gh"
	"ImDevinC/plex-meta-manager-configs/internal/pmm"
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"

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

	client := gh.NewClient(context.Background(), githubToken, owner, repo)
	for k, v := range payload.Metadata {
		if v.PosterURL == "" {
			err := client.CheckForExistingMovieIssue(context.Background(), k)
			var existingErr gh.ErrAlreadyExists
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
