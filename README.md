# Plex Meta Manager Configs

This repository contains configuration files and automation tools for managing Plex media metadata using [Kometa](https://kometa.wiki/) (formerly Plex Meta Manager).

## Overview

This repository automates the management of Plex media library metadata including:
- Custom movie posters from ThePosterDB
- TV show and anime metadata
- Media collections and playlists
- Automated detection of missing posters

## Features

- **Automated Metadata Management**: Uses Kometa to apply custom posters and metadata to Plex libraries
- **Change Detection**: Go-based tool that compares configuration changes and identifies missing posters
- **GitHub Integration**: Automatically creates GitHub issues for movies missing poster URLs
- **Docker Support**: Containerized workflow for easy deployment and consistency
- **CI/CD Pipeline**: GitHub Actions workflow for building and deploying the Docker image

## Repository Structure

```
.
├── config/              # Kometa configuration files
│   ├── movies.yml       # Movie metadata and poster URLs
│   ├── tv.yml           # TV show configurations
│   ├── anime.yml        # Anime configurations
│   ├── movie-collections.yml  # Movie collection definitions
│   └── playlist.yaml    # Playlist configurations
├── assets/              # Custom asset files
├── cmd/
│   └── diff/            # Go tool for detecting config changes
├── internal/            # Internal Go packages
├── add_movie.sh         # Script to add movie poster URLs
├── entrypoint.sh        # Docker container entrypoint script
├── Dockerfile           # Multi-stage Docker build
└── docker-compose.yaml  # Docker Compose configuration
```

## Configuration Files

### Movies Configuration
The `config/movies.yml` file contains metadata for movies including custom poster URLs from ThePosterDB. Each entry specifies:
- `url_poster`: Custom poster URL
- `sort_title`: (optional) Custom sort title for proper ordering

Example:
```yaml
metadata:
  "The Matrix":
    url_poster: https://theposterdb.com/api/assets/12345
  "The Matrix Reloaded":
    url_poster: https://theposterdb.com/api/assets/12346
    sort_title: Matrix 2
```

## Tools

### Config Diff Tool
The Go-based diff tool (`cmd/diff/main.go`) compares configuration changes and automatically creates GitHub issues for movies that are missing poster URLs. This runs as part of the Docker workflow to ensure all media has proper metadata.

### Add Movie Script
The `add_movie.sh` script validates and adds poster URLs to the movies configuration:
```bash
./add_movie.sh "Movie Name" "https://theposterdb.com/api/assets/12345"
```

## Usage

### Using Docker Compose
```bash
docker-compose up
```

This will:
1. Build the Docker image with Kometa and the diff tool
2. Run Kometa to update Plex metadata
3. Compare configuration changes
4. Create GitHub issues for missing posters

### Manual Kometa Run
```bash
docker build -t pmm .
docker run -it --rm \
  --env-file .env \
  -v $(pwd)/config:/config \
  -v $(pwd)/assets:/assets \
  pmm
```

## Environment Variables

Required environment variables (configured in `.env`):
- `GITHUB_TOKEN`: GitHub personal access token for API access
- `GITHUB_OWNER`: GitHub repository owner
- `GITHUB_REPO`: GitHub repository name
- `KOMETA_PLEX_SECRET_TOKEN`: Plex authentication token
- `KOMETA_RADARR_TOKEN`: (optional) Radarr API token
- `KOMETA_TMDB_TOKEN`: The Movie Database API token

## CI/CD

The repository includes GitHub Actions workflows:
- **docker.yaml**: Builds and pushes Docker images to GitHub Container Registry
- **test.yaml**: Runs tests and validation
- **comment.yaml**: Handles GitHub issue comments

## Development

### Building the Diff Tool
```bash
go build -o config-diff ./cmd/diff/main.go
```

### Running the Diff Tool
```bash
./config-diff -source /path/to/movies-backup.yml
```

## Contributing

When adding new movies or updating metadata:
1. Use the `add_movie.sh` script to ensure proper formatting
2. Ensure poster URLs are valid (JPEG or PNG content type)
3. Use ThePosterDB URLs when possible for consistency
4. Submit changes via pull request

## License

This is a personal configuration repository. Feel free to use it as a reference for your own Plex Meta Manager setup.
