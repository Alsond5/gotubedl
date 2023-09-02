package stream

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"gotubedl/extract"
)

type Stream struct {
	Itag       float64
	Url        string
	Mime_type  string
	Codecs     string
	Bitrate    float64
	Quality    string
	Type       string
	Extension  string
	Fps        float64
	Resolution string
	Sig        string
	file_size  int
	Title      string
}

func CreateStream(stream map[string]interface{}, title string) *Stream {
	mime_type, codecs := extract.MimeTypeCodec(stream["mimeType"].(string))
	parts := strings.Split(mime_type, "/")

	if _, ok := stream["fps"].(float64); !ok {
		stream["fps"] = 0.0
	}

	if _, ok := stream["s"].(string); !ok {
		stream["s"] = ""
	}

	quality := ""

	if parts[0] == "video" {
		quality = stream["qualityLabel"].(string)
	} else {
		quality = stream["audioQuality"].(string)
	}

	newStream := &Stream{
		Itag:       stream["itag"].(float64),
		Url:        stream["url"].(string),
		Mime_type:  mime_type,
		Codecs:     codecs,
		Bitrate:    stream["bitrate"].(float64),
		Quality:    stream["quality"].(string),
		Type:       parts[0],
		Extension:  parts[1],
		Sig:        stream["s"].(string),
		Fps:        stream["fps"].(float64),
		Resolution: quality,
		file_size:  0,
		Title:      title,
	}

	return newStream
}

func (s *Stream) FileSize() (int, error) {
	if s.file_size != 0 {
		return s.file_size, nil
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", s.Url, nil)

	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	defer res.Body.Close()

	s.file_size = int(res.ContentLength)

	return s.file_size, nil
}

func (s *Stream) downloadChunk(client *http.Client, from, to int64, wg *sync.WaitGroup, queue chan<- struct {
	Index int64
	Chunk []byte
}, index int64) {
	defer wg.Done()

	req, err := http.NewRequest("GET", s.Url, nil)

	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, to))

	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer res.Body.Close()

	chunk, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	queue <- struct {
		Index int64
		Chunk []byte
	}{Index: index, Chunk: chunk}
}

func (s *Stream) ReturnChunk(client *http.Client, from, to int64) []byte {
	req, err := http.NewRequest("GET", s.Url, nil)

	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, to))

	res, err := client.Do(req)

	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	defer res.Body.Close()

	chunk, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	return chunk
}

func (s *Stream) Download(filename string, filePath string, numChunks int) (string, error) {
	client := &http.Client{}

	file_size, _ := s.FileSize()

	chunkSize := int64(file_size) / int64(numChunks)

	queue := make(chan struct {
		Index int64
		Chunk []byte
	}, numChunks)
	var wg sync.WaitGroup

	for i := int64(0); i < int64(numChunks); i++ {
		from := i * chunkSize
		to := from + chunkSize - 1

		if i == int64(numChunks)-1 {
			to = int64(file_size) - 1
		}

		wg.Add(1)
		go s.downloadChunk(client, from, to, &wg, queue, i)
	}

	go func() {
		wg.Wait()
		close(queue)
	}()

	chunks := make([][]byte, numChunks)

	for q := range queue {
		chunks[q.Index] = q.Chunk
	}

	if filename == "" {
		if s.Type == "audio" && s.Extension == "mp4" {
			s.Extension = "m4a"
		}

		filename = filePath + s.Title + "." + s.Extension
	}

	file, err := os.Create(filename)

	if err != nil {
		return "", err
	}

	defer file.Close()

	for _, chunk := range chunks {
		_, err := file.Write(chunk)
		if err != nil {
			return "", err
		}
	}

	return filename, nil
}
