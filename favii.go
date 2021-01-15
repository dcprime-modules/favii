package favii

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// Favii with client and cache (in future)
type Favii struct {
	client *http.Client
	cache  map[string]*MetaInfo
}

// MetaInfo with metadata details
type MetaInfo struct {
	Metas []Meta
	Links []Link
}

// Meta is a simple struct to keep name and content attributes of an HTML Page
type Meta struct {
	Name    string
	Content string
}

// Link is a simple struct to keep rel and href attributes of an HTML Page
type Link struct {
	Rel  string
	Href string
}

// New for new Favii struct with new client
func New() *Favii {
	return &Favii{
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		cache: map[string]*MetaInfo{},
	}
}

// NewWithClient for new Favii struct with existing client
func NewWithClient(client *http.Client) *Favii {
	return &Favii{
		client: client,
	}
}

// GetMetaInfo for getting meta information
func (f *Favii) GetMetaInfo(url string) (*MetaInfo, error) {
	m, err := f.getMetaInfo(url)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetFaviconURL for getting favicon URL from the MetaInfo
func (m *MetaInfo) GetFaviconURL() string {
	faviconURLs := [2]string{"", ""}
	for _, l := range m.Links {
		// strict check on icon
		if l.Rel == "icon" || l.Rel == "shortcut icon" {
			faviconURLs[0] = l.Href
		}
		// loose check if anything with icon available
		if strings.Contains(l.Rel, "icon") {
			faviconURLs[1] = l.Href
		}
	}
	if faviconURLs[0] != "" {
		return faviconURLs[0]
	}
	return faviconURLs[1]
}

func (f *Favii) getMetaInfo(url string) (*MetaInfo, error) {
	if m, ok := f.cache[url]; ok {
		return m, nil
	}
	m := &MetaInfo{
		Metas: []Meta{},
		Links: []Link{},
	}
	defer func(m *MetaInfo) {
		f.cache[url] = m
	}(m)
	response, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	t := html.NewTokenizer(response.Body)

	for {
		tt := t.Next()
		if tt == html.ErrorToken {
			if t.Err() == io.EOF {
				break
			}
			fmt.Printf("Error: %v", t.Err())
			break
		}

		if tt != html.SelfClosingTagToken && tt != html.TextToken {
			continue
		}

		tagname, _ := t.TagName()

		if string(tagname[:]) == "meta" {
			mt := Meta{}
			for {
				tagattrkey, tagattrval, hasMore := t.TagAttr()
				if string(tagattrkey) == "name" {
					mt.Name = string(tagattrval)
				}
				if string(tagattrkey) == "content" {
					mt.Content = string(tagattrval)
				}
				if !hasMore {
					break
				}
			}
			m.Metas = append(m.Metas, mt)
		}
		if string(tagname[:]) == "link" {
			lk := Link{}
			for {
				tagattrkey, tagattrval, hasMore := t.TagAttr()
				if string(tagattrkey) == "rel" {
					lk.Rel = string(tagattrval)
				}
				if string(tagattrkey) == "href" {
					lk.Href = string(tagattrval)
				}
				if !hasMore {
					break
				}
			}
			m.Links = append(m.Links, lk)
		}
	}
	return m, nil
}
