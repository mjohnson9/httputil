package httputil

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

type AcceptHeader []*AcceptItem

func (a AcceptHeader) Len() int {
	return len(a)
}

func (a AcceptHeader) Less(i, j int) bool {
	iElem, jElem := a[i], a[j]
	if iElem.Quality > jElem.Quality {
		return true
	}
	if iElem.Position < jElem.Position {
		return true
	}
	return false
}

func (a AcceptHeader) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AcceptHeader) FindBestType(knownTypes []string) string {
	if len(a) <= 0 {
		return knownTypes[0]
	}

	for _, wantedType := range a {
		for _, knownType := range knownTypes {
			if wantedType.MIME == knownType {
				return knownType
			} else if matched, err := path.Match(wantedType.MIME, knownType); matched && err == nil {
				return knownType
			}
		}
	}

	return ""
}

func (a AcceptHeader) String() string {
	buf := new(bytes.Buffer)

	first := true
	for _, acceptItem := range a {
		if !first {
			buf.WriteRune(',')
		} else {
			first = false
		}

		buf.WriteString(acceptItem.MIME)
		if acceptItem.Quality != 1 {
			fmt.Fprintf(buf, ";q=%.1f", acceptItem.Quality)
		}
	}

	return buf.String()
}

type AcceptItem struct {
	MIME     string
	Position int
	Quality  float32
}

func ParseAccept(acceptHeader string) AcceptHeader {
	headerSplit := strings.Split(strings.ToLower(strings.TrimSpace(acceptHeader)), ",")
	parsed := make(AcceptHeader, 0, len(headerSplit))
outerLoop:
	for i, acceptableUnsplit := range headerSplit {
		tags := strings.Split(acceptableUnsplit, ";")

		current := &AcceptItem{
			MIME:     tags[0],
			Position: i,
			Quality:  1,
		}
		parsed = append(parsed, current)

		tags = tags[1:]

		for _, tag := range tags {
			const (
				qualityPrefix    = "q="
				qualityPrefixLen = len(qualityPrefix)
			)

			if strings.HasPrefix(tag, qualityPrefix) {
				quality, err := strconv.ParseFloat(tag[qualityPrefixLen:], 32)
				if err != nil {
					current.Quality = 0
					continue outerLoop
				}

				current.Quality = float32(quality)

				continue outerLoop // We've found everything we can handle
			}
		}
	}

	sort.Sort(parsed)

	return parsed
}
