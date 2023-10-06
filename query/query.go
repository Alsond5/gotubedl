package query

import (
	"errors"

	"github.com/Alsond5/gotubedl/stream"
)

type StreamQuery struct {
	Streams []stream.Stream
}

func CreateStreamQuery(streams []stream.Stream) *StreamQuery {
	newQuery := &StreamQuery{
		Streams: streams,
	}

	return newQuery
}

func (sq *StreamQuery) Filter(filterFunc func(stream.Stream) bool) *StreamQuery {
	var filteredStreams []stream.Stream

	for _, v := range sq.Streams {
		if filterFunc(v) {
			filteredStreams = append(filteredStreams, v)
		}
	}

	return CreateStreamQuery(filteredStreams)
}

func (sq *StreamQuery) First() *stream.Stream {
	return &sq.Streams[0]
}

func (sq *StreamQuery) GetByItag(itag float64) (stream.Stream, error) {
	for _, v := range sq.Streams {
		if v.Itag == itag {
			return v, nil
		}
	}

	return stream.Stream{}, errors.New("not found")
}

func (sq *StreamQuery) GetAudioOnly() *stream.Stream {
	for _, v := range sq.Streams {
		if v.Type == "audio" {
			return &v
		}
	}

	return &stream.Stream{}
}

func (sq *StreamQuery) GetHighestResolution() *stream.Stream {
	return sq.First()
}
