package pmm

type Config struct {
	Metadata map[string]Metadata `yaml:"metadata"`
}

type Metadata struct {
	PosterURL string `yaml:"url_poster"`
	SortTitle string `yaml:"sort_title"`
}
