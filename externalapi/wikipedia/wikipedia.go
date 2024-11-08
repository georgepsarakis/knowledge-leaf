package wikipedia

import (
	"context"
	"net/url"
	"time"

	"github.com/georgepsarakis/go-httpclient"
)

type RestV1SummaryResponse struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	Displaytitle string `json:"displaytitle"`
	Namespace    struct {
		ID   int    `json:"id"`
		Text string `json:"text"`
	} `json:"namespace"`
	WikibaseItem string `json:"wikibase_item"`
	Titles       struct {
		Canonical  string `json:"canonical"`
		Normalized string `json:"normalized"`
		Display    string `json:"display"`
	} `json:"titles"`
	Pageid    int `json:"pageid"`
	Thumbnail struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail"`
	Originalimage struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"originalimage"`
	Lang              string    `json:"lang"`
	Dir               string    `json:"dir"`
	Revision          string    `json:"revision"`
	Tid               string    `json:"tid"`
	Timestamp         time.Time `json:"timestamp"`
	Description       string    `json:"description"`
	DescriptionSource string    `json:"description_source"`
	ContentUrls       struct {
		Desktop struct {
			Page      string `json:"page"`
			Revisions string `json:"revisions"`
			Edit      string `json:"edit"`
			Talk      string `json:"talk"`
		} `json:"desktop"`
		Mobile struct {
			Page      string `json:"page"`
			Revisions string `json:"revisions"`
			Edit      string `json:"edit"`
			Talk      string `json:"talk"`
		} `json:"mobile"`
	} `json:"content_urls"`
	Extract     string `json:"extract"`
	ExtractHTML string `json:"extract_html"`
}

const userAgent = "knowledge-leaf-client/1.0"

// URL endpoints
const restV1SummaryEndpoint = "https://en.wikipedia.org/api/rest_v1/page/summary"

type Client struct {
	httpClient *httpclient.Client
}

func NewClient() *Client {
	c := httpclient.New()
	c = c.WithDefaultHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   userAgent,
	})
	return &Client{
		httpClient: httpclient.New(),
	}
}

func (w Client) GetSummary(ctx context.Context, title string) (RestV1SummaryResponse, error) {
	resp, err := w.httpClient.Get(ctx, w.summaryURL(title))
	if err != nil {
		return RestV1SummaryResponse{}, err
	}
	summaryResponse := RestV1SummaryResponse{}
	if err := httpclient.DeserializeJSON(resp, &summaryResponse); err != nil {
		return RestV1SummaryResponse{}, err
	}
	return summaryResponse, nil
}

func (c Client) summaryURL(title string) string {
	p, _ := url.JoinPath(restV1SummaryEndpoint, title)
	return p
}
