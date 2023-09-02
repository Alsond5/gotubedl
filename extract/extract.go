package extract

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func MimeTypeCodec(mime_type string) (string, string) {
	pattern := regexp.MustCompile(`(.+?);.*codecs=(".*?")`)
	result := pattern.FindStringSubmatch(mime_type)

	return result[1], result[2]
}

func ExtendStream(streams map[string]interface{}) []map[string]interface{} {
	var formats []map[string]interface{}

	if formatsList, ok := streams["formats"].([]interface{}); ok {
		for i := len(formatsList) - 1; i >= 0; i-- {
			format := formatsList[i]

			if formatMap, ok := format.(map[string]interface{}); ok {
				formats = append(formats, formatMap)
			}
		}
	}
	if adaptiveFormatsList, ok := streams["adaptiveFormats"].([]interface{}); ok {
		for _, format := range adaptiveFormatsList {
			if formatMap, ok := format.(map[string]interface{}); ok {
				formats = append(formats, formatMap)
			}
		}
	}

	for _, stream := range formats {
		if _, urlExists := stream["url"]; !urlExists {
			if signatureCipher, signatureExists := stream["signatureCipher"].(string); signatureExists {
				cipher, _ := url.ParseQuery(signatureCipher)

				stream["url"] = cipher.Get("url")
				stream["s"] = cipher.Get("s")
			}
		}
	}

	return formats
}

func DecodeSignature(functions []string, sig string) string {
	s := []byte(sig)
	str := ""

	function1 := functions[0][strings.Index(functions[0], "{"):]

	functions2 := strings.Split(functions[1][strings.Index(functions[1], "{")+1:len(functions[1])-2], "},")

	for _, v := range strings.Split(function1, ";") {
		if strings.Contains(v, "split") {
			continue
		}

		index := getIndex(v, functions2, functions[2], functions[3])

		if index == 2 {
			for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
				s[i], s[j] = s[j], s[i]
			}
		} else if index == 1 {
			_op := v[strings.Index(v, "("):]

			pattern := regexp.MustCompile(`\d+`)
			b := pattern.FindStringSubmatch(_op)[0]

			num, err := strconv.Atoi(b)

			if err != nil {
				return ""
			}

			s = s[num:]
		} else if index == 0 {
			_op := v[strings.Index(v, "("):]

			pattern := regexp.MustCompile(`\d+`)
			b := pattern.FindStringSubmatch(_op)[0]

			num, err := strconv.Atoi(b)

			if err != nil {
				return ""
			}

			c := s[0]
			s[0] = s[num%len(s)]
			s[num%len(s)] = c
		} else {
			str = string(s)
		}
	}

	return str
}

func getIndex(v string, function []string, key string, slash string) int {
	_key := fmt.Sprintf(`%s%s\.(\w+)\(`, key, slash)

	opPattern := regexp.MustCompile(_key)
	_op := opPattern.FindStringSubmatch(v)

	op := ""

	if len(_op) > 1 {
		op = _op[1]
	} else {
		return 3
	}

	for _, v := range function {
		v = strings.TrimSpace(v)

		if v[:strings.Index(v, ":")] == op {
			if strings.Contains(v, "reverse") {
				return 2
			} else if strings.Contains(v, "splice") {
				return 1
			} else {
				return 0
			}
		}
	}

	return 3
}
