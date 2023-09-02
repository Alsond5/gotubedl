## gotubedl

Gotubedl is a Golang package used to download videos from YouTube.

### Features

* Support for downloading videos from YouTube.
* Download videos in various formats (MP4, MKV, MOV, MP3, M4A, AAC, FLAC, Opus, etc.).
* Choose video quality.
* List and preview videos.
* Adjust download speed.
* Split videos into parts during download.
* Save various metadata during download.

### Installation

To install Gotubedl, use the following command:

```
go get github.com/Alsond5/gotubedl
```

### Usage

To use Gotubedl, use the following code:

```go
yt, err := gotubedl.Init("https://www.youtube.com/watch?v=1uPr9a-Dnt0")

if err != nil {
    fmt.Println(err)
    return
}

yt.Streams().GetHighestResolution().Download("", "", 15)
```

### Help

To see Gotubedl's help menu, use the following command:

```
gotubedl -h
```

### Development

Gotubedl is developed as an open-source project on GitHub. To contribute, follow these steps:

1. Fork it.
2. Make changes.
3. Run tests.
4. Send a pull request.

### License

Gotubedl is licensed under The Unlicense.
