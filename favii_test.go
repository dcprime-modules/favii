package favii

import (
	"fmt"
	"log"
	"testing"
)

func TestFavii(t *testing.T) {
	f := New(true)
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "pkg.go.dev",
			url:  "https://pkg.go.dev/golang.org/x/net/html",
			want: "https://pkg.go.dev/favicon.ico",
		},
		{
			name: "duckduckgo.com",
			url:  "https://duckduckgo.com/?q=golang",
			want: "https://duckduckgo.com/favicon.ico",
		},
		{
			name: "git.dcpri.me",
			url:  "https://git.dcpri.me/modules/favii",
			want: "https://git.dcpri.me/img/favicon.svg",
		},
		{
			name: "git.dcpri.me-2",
			url:  "https://git.dcpri.me/modules",
			want: "https://git.dcpri.me/img/favicon.svg",
		},
		{
			name: "grafana.com",
			url:  "https://grafana.com/oss/grafana/",
			want: "https://grafana.com/static/assets/img/fav32.png",
		},
		{
			name: "firefox.com",
			url:  "https://firefox.com",
			want: "https://www.mozilla.org/media/img/favicons/firefox/browser/favicon.f093404c0135.ico",
		},
		{
			name:    "invalid-url",
			url:     "lol-lol.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotm, err := f.GetMetaInfo(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Favii.GetMetaInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := gotm.GetFaviconURL()
			if got != tt.want {
				t.Errorf("MetaInfo.GetFaviconURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ExampleFavii shows how to get the MetaInfo from an URL and also from that
// MetaInfo get favicon URL from that.
func ExampleFavii() {
	f := New(true)
	metainfo, err := f.GetMetaInfo("https://git.dcpri.me/modules/favii")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(metainfo.GetFaviconURL())
	//Output:
	// https://git.dcpri.me/img/favicon.svg
}
