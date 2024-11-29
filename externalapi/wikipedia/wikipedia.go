package wikipedia

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
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
const titleCategoriesEndpoint = "https://en.wikipedia.org/w/api.php"

// https://en.wikipedia.org/api/rest_v1/feed/onthisday/events/12/01
const restV1OnThisDayEndpoint = "https://en.wikipedia.org/api/rest_v1/feed/onthisday/%s/%s/%s"

var titleCategoriesBaseParameters = map[string]string{
	"format":  "json",
	"action":  "query",
	"prop":    "categories",
	"clprop":  "timestamp",
	"clshow":  "!hidden",
	"cllimit": "5",
}

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

func (c Client) GetSummary(ctx context.Context, title string) (RestV1SummaryResponse, error) {
	resp, err := c.httpClient.Get(ctx, c.summaryURL(title))
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

func (c Client) Categories(ctx context.Context, title string) ([]string, error) {
	resp, err := c.httpClient.Get(ctx, titleCategoriesEndpoint, httpclient.WithQueryParameters(map[string]string{
		"titles": title,
	}), httpclient.WithQueryParameters(titleCategoriesBaseParameters))
	if err != nil {
		return nil, err
	}
	var categoryResponse TitleCategoriesResponse
	if err := httpclient.DeserializeJSON(resp, &categoryResponse); err != nil {
		return nil, err
	}
	var categories []string
	for _, page := range categoryResponse.Query.Pages {
		for _, category := range page.Categories {
			categories = append(categories, strings.TrimPrefix(category.Title, "Category:"))
		}
	}
	slices.Sort(categories)
	return categories, nil
}

type TitleCategoriesResponse struct {
	Query struct {
		Pages map[string]struct {
			Pageid     int    `json:"pageid"`
			Ns         int    `json:"ns"`
			Title      string `json:"title"`
			Categories []struct {
				Ns        int       `json:"ns"`
				Title     string    `json:"title"`
				Timestamp time.Time `json:"timestamp"`
			} `json:"categories"`
		} `json:"pages"`
	} `json:"query"`
}

func (c Client) OnThisDay(ctx context.Context) (RestV1EventsOnThisDayResponse, error) {
	now := time.Now().UTC()
	resp, err := c.httpClient.Get(ctx, fmt.Sprintf(restV1OnThisDayEndpoint, "events", now.Format("01"), now.Format("02")))
	if err != nil {
		return RestV1EventsOnThisDayResponse{}, err
	}
	var v RestV1EventsOnThisDayResponse
	if err := httpclient.DeserializeJSON(resp, &v); err != nil {
		return RestV1EventsOnThisDayResponse{}, err
	}
	return v, nil
}

type RestV1EventsOnThisDayResponse struct {
	Events []struct {
		Text  string `json:"text"`
		Pages []struct {
			Title  string `json:"title"`
			Titles struct {
				Canonical  string `json:"canonical"`
				Normalized string `json:"normalized"`
				Display    string `json:"display"`
			} `json:"titles"`
			Pageid      int    `json:"pageid"`
			Extract     string `json:"extract"`
			ExtractHtml string `json:"extract_html"`
			Thumbnail   struct {
				Source string `json:"source"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"thumbnail"`
			Originalimage struct {
				Source string `json:"source"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"originalimage"`
			Lang        string    `json:"lang"`
			Dir         string    `json:"dir"`
			Timestamp   time.Time `json:"timestamp"`
			Description string    `json:"description"`
			ContentUrls struct {
				Desktop struct {
					Page string `json:"page"`
				}
			} `json:"content_urls"`
		} `json:"pages"`
	} `json:"events"`
}
