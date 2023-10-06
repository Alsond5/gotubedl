package gotubedl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Playlist struct {
	PlaylistUrl  string
	playlistData map[string]interface{}
	html         string
	youtube      []*YouTube
	searchQuery  []SearchQuery
}

func InitPlaylist(url string) *Playlist {
	re := regexp.MustCompile(`^https:\/\/www\.youtube\.com\/playlist\?list=[a-zA-Z0-9_-]+$`)
	matches := re.FindSubmatch([]byte(url))

	if len(matches) == 0 {
		return nil
	}

	return &Playlist{
		PlaylistUrl: url,
	}
}

func (p *Playlist) Html() string {
	if p.html != "" {
		return p.html
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", p.PlaylistUrl, nil)

	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}

	p.html = string(body)

	return p.html
}

func (p *Playlist) getPlaylistData() map[string]interface{} {
	if p.playlistData != nil {
		return p.playlistData
	}

	re := regexp.MustCompile(`var ytInitialData = ([\s\S]*?)<\/script>`)
	matches := re.FindSubmatch([]byte(p.Html()))

	if len(matches) == 0 {
		panic("Error: Js not found")
	}

	jsonMatch := matches[1]
	jsonMatch = jsonMatch[:len(jsonMatch)-1]

	var jsonObj map[string]interface{}

	err := json.Unmarshal([]byte(jsonMatch), &jsonObj)

	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	p.playlistData = jsonObj

	return p.playlistData
}

func (p *Playlist) GetVideos() []*YouTube {
	if p.youtube != nil {
		return p.youtube
	}

	contents := p.getPlaylistData()["contents"].(map[string]interface{})["twoColumnBrowseResultsRenderer"].(map[string]interface{})["tabs"].([]interface{})[0].(map[string]interface{})["tabRenderer"].(map[string]interface{})["content"].(map[string]interface{})["sectionListRenderer"].(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["itemSectionRenderer"].(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["playlistVideoListRenderer"].(map[string]interface{})["contents"].([]interface{})

	for _, v := range contents {
		video_url := "https://www.youtube.com/watch?v=" + v.(map[string]interface{})["playlistVideoRenderer"].(map[string]interface{})["videoId"].(string)

		video, err := Init(video_url)

		if err != nil {
			return nil
		}

		p.youtube = append(p.youtube, video)
	}

	return p.youtube
}

func (p *Playlist) Results() []SearchQuery {
	contents := p.getPlaylistData()["contents"].(map[string]interface{})["twoColumnBrowseResultsRenderer"].(map[string]interface{})["tabs"].([]interface{})[0].(map[string]interface{})["tabRenderer"].(map[string]interface{})["content"].(map[string]interface{})["sectionListRenderer"].(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["itemSectionRenderer"].(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["playlistVideoListRenderer"].(map[string]interface{})["contents"].([]interface{})

	for _, v := range contents {
		videoRenderer, ok := v.(map[string]interface{})["playlistVideoRenderer"].(map[string]interface{})

		if !ok {
			continue
		}

		video_url := "https://www.youtube.com/watch?v=" + v.(map[string]interface{})["playlistVideoRenderer"].(map[string]interface{})["videoId"].(string)
		title := videoRenderer["title"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)
		thumbnail := videoRenderer["thumbnail"].(map[string]interface{})["thumbnails"].([]interface{})[0].(map[string]interface{})["url"].(string)
		length := videoRenderer["lengthSeconds"].(string)
		author := videoRenderer["shortBylineText"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)

		p.searchQuery = append(p.searchQuery, SearchQuery{
			Url:        video_url,
			Title:      title,
			Thumbnail:  thumbnail,
			LengthText: length,
			Author:     author,
		})
	}

	return p.searchQuery
}
