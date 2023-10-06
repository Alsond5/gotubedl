package gotubedl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Alsond5/gotubedl/extract"
	"github.com/Alsond5/gotubedl/query"
	"github.com/Alsond5/gotubedl/stream"
)

type YouTube struct {
	Url              string
	js               map[string]interface{}
	html             string
	streams          query.StreamQuery
	basejs_url       string
	basejs_content   string
	basejs_functions []string
	VideoId          string
	Title            string
	Length           int
	Author           string
	ViewCount        int
}

func Init(url string) (*YouTube, error) {
	re := regexp.MustCompile(`^(?:https?:\/\/)?(?:www\.)?(?:youtube\.com\/watch\?v=|youtu.be\/)([a-zA-Z0-9_-]{11})`)
	matches := re.FindSubmatch([]byte(url))

	if len(matches) == 0 {
		return nil, errors.New("invalid url")
	}

	youtube := &YouTube{
		Url: url,
	}

	return youtube, nil
}

func (y *YouTube) checkAvailability() string {
	status := y.Js()["playabilityStatus"].(map[string]interface{})

	if status["status"].(string) == "ERROR" {
		return "ERROR"
	}

	if _, ok := status["liveStreamability"]; ok {
		return "LIVE STREAM"
	}

	return ""
}

func (y *YouTube) Html() string {
	if y.html != "" {
		return y.html
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", y.Url, nil)

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

	y.html = string(body)

	return y.html
}

func (y *YouTube) Js() map[string]interface{} {
	if y.js != nil {
		return y.js
	}

	re := regexp.MustCompile(`var ytInitialPlayerResponse = ([\s\S]*?)<\/script>`)
	matches := re.FindSubmatch([]byte(y.Html()))

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

	y.js = jsonObj

	if y.checkAvailability() != "" {
		panic("Error: invalid video_id")
	}

	videoDetails := jsonObj["videoDetails"].(map[string]interface{})

	length, _ := strconv.Atoi(videoDetails["lengthSeconds"].(string))
	view_count, _ := strconv.Atoi(videoDetails["viewCount"].(string))

	y.VideoId = videoDetails["videoId"].(string)
	y.Title = videoDetails["title"].(string)
	y.Length = length
	y.Author = videoDetails["author"].(string)
	y.ViewCount = view_count

	return y.js
}

func (y *YouTube) BasejsUrl() string {
	if y.basejs_url != "" {
		return y.basejs_url
	}

	basejsPattern := regexp.MustCompile(`src="([^"]+base.js)"`)
	result := basejsPattern.FindStringSubmatch(y.Html())

	if len(result) > 1 {
		y.basejs_url = "https://youtube.com" + result[1]
	}

	return y.basejs_url
}

func (y *YouTube) BasejsContent() (string, error) {
	if y.basejs_url != "" {
		return y.basejs_content, nil
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", y.BasejsUrl(), nil)

	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	y.basejs_content = string(body)

	return y.basejs_content, nil
}

func (y *YouTube) Streams() *query.StreamQuery {
	if len(y.streams.Streams) > 0 {
		return &y.streams
	}

	if y.checkAvailability() != "" {
		panic("Error: invalid video_id")
	}

	var _streams []stream.Stream

	for _, s := range extract.ExtendStream(y.Js()["streamingData"].(map[string]interface{})) {
		video := stream.CreateStream(s, y.Title)

		if video.Sig != "" {
			sig := extract.DecodeSignature(y.BasejsFunctions(), video.Sig)

			i := strings.Index(video.Url, "&lsparams")

			video.Url = video.Url[:i] + "&sig=" + sig + video.Url[i:]
		}

		_streams = append(_streams, *video)
	}

	return query.CreateStreamQuery(_streams)
}

func (y *YouTube) BasejsFunctions() []string {
	if y.basejs_functions != nil {
		return y.basejs_functions
	}

	basejs, err := y.BasejsContent()

	if err != nil {
		return nil
	}

	basejsPattern := regexp.MustCompile(`[A-Za-z0-9]+=function\(a\)\{a=a\.split\(""[^"]*""\)\};`)
	result1 := basejsPattern.FindStringSubmatch(basejs)[0]

	keyPattern := regexp.MustCompile(`;(.*?)\.`)
	key := keyPattern.FindStringSubmatch(result1)[1]

	cleanedBasejs := strings.ReplaceAll(basejs, "\n", " ")

	slash := ""

	if key[0] == '#' || key[0] == '$' {
		slash = "\\"
	}

	pattern1 := fmt.Sprintf(`var %s%s=\{[A-Za-z0-9]+:function\([^)]*\)\{[^}]*\}, [A-Za-z0-9]+:function\([^)]*\)\{[^}]*\}, [A-Za-z0-9]+:function\([^)]*\)\{[^}]*\}\};`, slash, key)

	basejsPattern = regexp.MustCompile(pattern1)
	result2 := basejsPattern.FindStringSubmatch(cleanedBasejs)[0]

	y.basejs_functions = append(y.basejs_functions, result1, result2, key, slash)

	return y.basejs_functions
}
