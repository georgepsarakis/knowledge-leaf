package main

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type WikiSummaryMetadata struct {
	Description string `json:"description"`
	URL         string `json:"url"`
	Image       Image  `json:"image"`
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

type OnThisDayEvent struct {
	Title       string                    `json:"title"`
	ShortTitle  string                    `json:"short_title"`
	Description string                    `json:"description"`
	Image       Image                     `json:"image"`
	Extract     string                    `json:"extract"`
	URL         string                    `json:"url"`
	References  []OnThisDayEventReference `json:"references"`
	Year        int                       `json:"year"`
}

type OnThisDayEventReference struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type EventsOnThisDayResponse struct {
	Titles []OnThisDayEvent `json:"titles"`
}
