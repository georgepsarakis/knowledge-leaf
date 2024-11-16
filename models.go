package main

type WikiSummaryImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type WikiSummaryMetadata struct {
	Description string           `json:"description"`
	URL         string           `json:"url"`
	Image       WikiSummaryImage `json:"image"`
}

type WikiSummary struct {
	Title      string              `json:"title"`
	Summary    string              `json:"summary"`
	Type       string              `json:"type"`
	Categories []string            `json:"categories"`
	Metadata   WikiSummaryMetadata `json:"metadata"`
}

type RandomTriviaResponse struct {
	Results []WikiSummary `json:"results"`
}
