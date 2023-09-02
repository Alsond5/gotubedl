package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SearchQuery struct {
	Url        string `json:"url"`
	Title      string `json:"title"`
	Thumbnail  string `json:"thumbnail"`
	LengthText string `json:"lengthText"`
	Author     string `json:"author"`
}

func Search(query string) ([]SearchQuery, error) {
	url := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", strings.ReplaceAll(query, " ", "+"))

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	ytInitialData := string(body)[strings.Index(string(body), "var ytInitialData = {"):]
	ytInitialData = ytInitialData[20 : strings.Index(ytInitialData, "</script>")-1]

	var jsonData map[string]interface{}

	ok := json.Unmarshal([]byte(ytInitialData), &jsonData)

	if ok != nil {
		fmt.Println("error", ok)
		return nil, ok
	}

	contents := jsonData["contents"].(map[string]interface{})["twoColumnSearchResultsRenderer"].(map[string]interface{})["primaryContents"].(map[string]interface{})["sectionListRenderer"].(map[string]interface{})["contents"].([]interface{})[0].(map[string]interface{})["itemSectionRenderer"].(map[string]interface{})["contents"].([]interface{})

	var videos []SearchQuery

	for i, v := range contents {
		if i == 0 {
			continue
		}

		videoRenderer, ok := v.(map[string]interface{})["videoRenderer"].(map[string]interface{})

		if !ok {
			continue
		}

		title, ok := videoRenderer["title"].(map[string]interface{})["runs"].([]interface{})

		if !ok || len(title) == 0 {
			continue
		}

		titleText, ok := title[0].(map[string]interface{})["text"].(string)

		if !ok {
			continue
		}

		video_url := videoRenderer["videoId"].(string)
		video_url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", video_url)

		thumbnail := videoRenderer["thumbnail"].(map[string]interface{})["thumbnails"].([]interface{})[0].(map[string]interface{})["url"].(string)

		lengthText := videoRenderer["lengthText"].(map[string]interface{})["simpleText"].(string)

		author := videoRenderer["longBylineText"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)

		videos = append(videos, SearchQuery{
			Url:        video_url,
			Title:      titleText,
			Thumbnail:  thumbnail,
			LengthText: lengthText,
			Author:     author,
		})
	}

	return videos, nil
}
