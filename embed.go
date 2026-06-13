package dhook

import "time"

type Embed struct {
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	URL         string     `json:"url,omitempty"`
	Color       int        `json:"color,omitempty"`
	Timestamp   string     `json:"timestamp,omitempty"`
	Footer      *Footer    `json:"footer,omitempty"`
	Image       *Image     `json:"image,omitempty"`
	Thumbnail   *Thumbnail `json:"thumbnail,omitempty"`
	Author      *Author    `json:"author,omitempty"`
	Fields      []*Field   `json:"fields,omitempty"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

type Image struct {
	URL string `json:"url"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

func NewEmbed() *Embed {
	return &Embed{}
}

func (e *Embed) SetTitle(title string) *Embed {
	e.Title = title
	return e
}

func (e *Embed) SetDescription(desc string) *Embed {
	e.Description = desc
	return e
}

func (e *Embed) SetURL(url string) *Embed {
	e.URL = url
	return e
}

func (e *Embed) SetColor(color int) *Embed {
	e.Color = color
	return e
}

func (e *Embed) SetTimestamp(t time.Time) *Embed {
	e.Timestamp = t.UTC().Format(time.RFC3339)
	return e
}

func (e *Embed) SetFooter(text, iconURL string) *Embed {
	e.Footer = &Footer{
		Text:    text,
		IconURL: iconURL,
	}
	return e
}

func (e *Embed) SetImage(url string) *Embed {
	e.Image = &Image{URL: url}
	return e
}

func (e *Embed) SetThumbnail(url string) *Embed {
	e.Thumbnail = &Thumbnail{URL: url}
	return e
}

func (e *Embed) SetAuthor(name, url, iconURL string) *Embed {
	e.Author = &Author{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return e
}

func (e *Embed) AddField(name, value string, inline bool) *Embed {
	e.Fields = append(e.Fields, &Field{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}
