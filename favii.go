package favii

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Favii is basically a client to use the struct. It has an http.Client for
// doing all HTTP Requests for fetching the HTML Pages
type Favii struct {
	client   *http.Client
	cache    map[string]*MetaInfo
	useCache bool
}

// MetaInfo with metadata details, this includes the URL used for requesting
// or calling GetMetaInfo() as well.
type MetaInfo struct {
	u     *url.URL
	Metas []Meta
	Links []Link
}

// Meta is a simple struct to keep name and content attributes of an HTML Page
// mostly contains the details about meta tag
type Meta struct {
	Name    string
	Content string
}

// Link is a simple struct to keep rel and href attributes of an HTML Page
// mostly contains the details about link tag
type Link struct {
	Rel  string
	Href string
}

// New creates a new Favii struct with http.DefaultClient and empty map, also
// an optional cache map
func New(useCache bool) *Favii {
	return &Favii{
		client:   http.DefaultClient,
		cache:    map[string]*MetaInfo{},
		useCache: useCache,
	}
}

// NewWithClient creates a new Favii struct with provided http.Client and all
// other things similar to New()
func NewWithClient(client *http.Client, useCache bool) *Favii {
	return &Favii{
		client:   client,
		cache:    map[string]*MetaInfo{},
		useCache: useCache,
	}
}

// GetMetaInfo for getting meta information, it is mainly a wrapper around
// unexported method getMetaInfo().
func (f *Favii) GetMetaInfo(url string) (*MetaInfo, error) {
	m, err := f.getMetaInfo(url)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetFaviconURL for getting favicon URL from the MetaInfo, using link tags,
// or use default /favicon.ico.
func (m *MetaInfo) GetFaviconURL() string {
	faviconURLs := [2]string{"", ""}
	if m == nil || m.Links == nil || m.Metas == nil {
		return ""
	}
	for _, l := range m.Links {
		// strict check on icon
		if l.Rel == "icon" || l.Rel == "shortcut icon" {
			if strings.HasPrefix(l.Href, "http") {
				faviconURLs[0] = l.Href
			} else if strings.HasPrefix(l.Href, "/") {
				faviconURLs[0] = m.u.Scheme + "://" + m.u.Hostname() + l.Href
			} else {
				faviconURLs[0] = m.u.Scheme + "://" + m.u.Hostname() + "/" + l.Href
			}
		}
		// loose check if anything with icon available
		if strings.Contains(l.Rel, "icon") {
			if strings.HasPrefix(l.Href, "http") {
				faviconURLs[1] = l.Href
			} else if strings.HasPrefix(l.Href, "/") {
				faviconURLs[1] = m.u.Scheme + "://" + m.u.Hostname() + l.Href
			} else {
				faviconURLs[1] = m.u.Scheme + "://" + m.u.Hostname() + "/" + l.Href
			}
		}
	}
	if faviconURLs[0] != "" {
		return faviconURLs[0]
	}
	if faviconURLs[1] != "" {
		return faviconURLs[1]
	}
	// in case if nothing is available go for the default one.
	return m.u.Scheme + "://" + m.u.Hostname() + "/favicon.ico"
}

func (f *Favii) getMetaInfo(u string) (*MetaInfo, error) {
	up, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	if m, ok := f.cache[up.Hostname()]; ok && f.useCache { // skip this if useCache is false
		return m, nil
	}
	m := &MetaInfo{
		Metas: []Meta{},
		Links: []Link{},
		u:     up,
	}
	defer func(m *MetaInfo) {
		f.cache[m.u.Hostname()] = m
	}(m)
	response, err := f.client.Get(u)
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

		if tt != html.SelfClosingTagToken && tt != html.TextToken && tt != html.StartTagToken {
			// fmt.Println("Skipping:", string(tagname))
			continue
		}

		tagname, hasAttr := t.TagName()
		if !hasAttr {
			continue
		}

		// fmt.Println("Processing:", string(tagname))
		switch string(tagname[:]) {
		case "link":
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
		case "meta":
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
	}
	return m, nil
}
