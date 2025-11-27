package mangadex

// API response types for MangaDex v5 API

type searchResponse struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []mangaData `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

type mangaResponse struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     mangaData `json:"data"`
}

type mangaData struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    mangaAttributes `json:"attributes"`
	Relationships []relationship  `json:"relationships"`
}

type mangaAttributes struct {
	Title                  map[string]string   `json:"title"`
	AltTitles              []map[string]string `json:"altTitles"`
	Description            map[string]string   `json:"description"`
	Status                 string              `json:"status"`
	Year                   *int                `json:"year"`
	ContentRating          string              `json:"contentRating"`
	Tags                   []tagData           `json:"tags"`
	OriginalLanguage       string              `json:"originalLanguage"`
	LastVolume             string              `json:"lastVolume"`
	LastChapter            string              `json:"lastChapter"`
	PublicationDemographic string              `json:"publicationDemographic"`
}

type relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type tagData struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	Attributes tagAttributes `json:"attributes"`
}

type tagAttributes struct {
	Name  map[string]string `json:"name"`
	Group string            `json:"group"`
}

type chapterFeedResponse struct {
	Result   string        `json:"result"`
	Response string        `json:"response"`
	Data     []chapterData `json:"data"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
	Total    int           `json:"total"`
}

type chapterData struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Attributes chapterAttributes `json:"attributes"`
}

type chapterAttributes struct {
	Volume             string `json:"volume"`
	Chapter            string `json:"chapter"`
	Title              string `json:"title"`
	TranslatedLanguage string `json:"translatedLanguage"`
	PublishAt          string `json:"publishAt"`
	Pages              int    `json:"pages"`
}
