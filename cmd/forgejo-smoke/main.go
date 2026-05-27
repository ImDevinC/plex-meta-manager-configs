package main

import (
	"ImDevinC/plex-meta-manager-configs/internal/forgejo"
	"ImDevinC/plex-meta-manager-configs/internal/issueclient"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	var movie string
	flag.StringVar(&movie, "movie", "", "Movie title to use when creating the issue")
	flag.Parse()

	if movie == "" {
		log.Fatal("missing required -movie argument")
	}

	url := mustEnv("FORGEJO_URL")
	owner := mustEnv("FORGEJO_OWNER")
	repo := mustEnv("FORGEJO_REPO")
	token := mustEnv("FORGEJO_TOKEN")

	client, err := forgejo.NewForgejoClient(context.Background(), url, token, owner, repo)
	if err != nil {
		log.Fatalf("failed to create forgejo client: %v", err)
	}

	ctx := context.Background()
	if err := client.CheckForExistingMovieIssue(ctx, movie); err != nil {
		switch err.(type) {
		case issueclient.ErrIgnored:
			log.Fatalf("existing issue for %q is marked ignored", movie)
		case issueclient.ErrAlreadyExists:
			log.Fatalf("open issue for %q already exists", movie)
		default:
			log.Fatalf("failed to check for existing issue: %v", err)
		}
	}

	if err := client.AddMissingMovie(ctx, movie); err != nil {
		log.Fatalf("failed to create issue for %q: %v", movie, err)
	}

	fmt.Printf("created issue for %q\n", movie)
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required %s environment variable", key)
	}
	return value
}
