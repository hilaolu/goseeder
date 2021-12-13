package nexus

import (
	"fmt"
	"hash/fnv"
	"os/exec"
	"seeder/src/config"

	"github.com/mmcdole/gofeed"
)

func torrent2size(link string) string {
	cmd := fmt.Sprintf("curl %s | head -1 | grep -aoE '6:lengthi[0-9]+' | cut -di -f2", link)
	out, _ := exec.Command(cmd).Output()
	return string(out)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

type Client struct {
	baseURL string
	Rule    config.NodeRule
}

type Torrent struct {
	GUID  string
	Title string
	URL   string
	Size  string
}

func NewClient(source string, limit int, passkey string, Rule config.NodeRule) Client {
	var baseURL = "https://" + source
	return Client{
		baseURL: baseURL,
		Rule:    Rule,
	}
}

func (c *Client) Get() ([]Torrent, error) {
	var ts []Torrent
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(c.baseURL)
	if err == nil {
		for _, value := range feed.Items {
			ts = append(ts, Torrent{
				GUID:  value.GUID,
				Title: value.Title,
				URL:   value.Link,
				Size:  torrent2size(value.Link),
			})
		}
		return ts, nil
	}

	return nil, err
}
