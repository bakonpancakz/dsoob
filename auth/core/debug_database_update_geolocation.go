package core

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"dsoob/backend/tools"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type IPV4Entry struct {
	RangeStart, CountryIndex, RegionIndex, CityIndex uint32
	TimezoneOffset                                   int32
}
type IPV6Entry struct {
	RangeStart                           [16]byte
	CountryIndex, RegionIndex, CityIndex uint32
	TimezoneOffset                       int32
}

var (
	ENTRIES_IPV4 = []IPV4Entry{}
	ENTRIES_IPV6 = []IPV6Entry{}
	ENTRIES_TEXT = []string{}
	INDEX_TEXT   = map[string]uint32{}
	INDEX_RENAME = map[string]string{
		"United States of America": "United States",
		"-":                        "Unknown",
	}
)

func trimString(givenString string) string {
	return strings.Trim(givenString, "\"")
}

func findString(givenString string) uint32 {
	if replace, ok := INDEX_RENAME[givenString]; ok {
		givenString = replace
	}
	if index, ok := INDEX_TEXT[givenString]; ok {
		// Duplicate String, return index
		return index
	} else {
		// Unique String Given, return new index
		ENTRIES_TEXT = append(ENTRIES_TEXT, givenString)
		index := uint32(len(ENTRIES_TEXT) - 1)
		INDEX_TEXT[givenString] = index
		return index
	}
}

func parseOffset(offset string) int32 {
	if offset == "-" {
		return 0
	}
	sign := 1
	if offset[0] == '-' {
		sign = -1
	}
	parts := strings.Split(offset[1:], ":")
	hours, _ := strconv.Atoi(parts[0])
	mins, _ := strconv.Atoi(parts[1])
	return int32(sign * ((hours * 3600) + (mins * 60)))
}

func download(code, filename, token string) (fs.File, error) {

	// Download Archive from API
	url := fmt.Sprintf("https://www.ip2location.com/download/?token=%s&file=%s", token, code)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request error: %s", err)
	}
	defer res.Body.Close()

	// Download Server is flawed and returns a 200 OK even if an error was thrown
	// Instead we can check to see if Received a Reasonable Amount of Bytes
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("response error: %s", err)
	}
	if len(body) < 1024 {
		return nil, fmt.Errorf("response invalid")
	}

	// Unzip Archive and Read CSV
	unzip, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("archive error: %s", err)
	}
	csv, err := unzip.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("inflate error: %s", err)
	}

	return csv, nil
}

func DebugDatabaseUpdateGeolocation() {
	t := time.Now()
	OUTPUT_FILE := "./include/DatabaseGeolocate.kani.gz"
	REQUEST_KEY := os.Getenv("API_KEY_IP2LOCATION")
	if REQUEST_KEY == "" {
		tools.LoggerGeolocation.Log(tools.FATAL,
			"Environment Variable API_KEY_IP2LOCATION was not set")
	}

	// Process IPV4 CSV
	{
		tools.LoggerGeolocation.Log(tools.INFO, "Downloading IPV4 Archive")
		f, err := download("DB11LITECSV", "IP2LOCATION-LITE-DB11.CSV", REQUEST_KEY)
		if err != nil {
			tools.LoggerGeolocation.Log(tools.FATAL, "%s", err)
			return
		}
		defer f.Close()

		tools.LoggerGeolocation.Log(tools.INFO, "Parsing IPV4 Archive")
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			splits := strings.SplitN(scanner.Text(), ",", 10)
			if len(splits) != 10 {
				tools.LoggerGeolocation.Log(tools.FATAL,
					"Error Decoding Line %d: Bad Split", len(ENTRIES_IPV4)+1)
			}

			rangeStart, err := strconv.Atoi(trimString(splits[0]))
			if err != nil {
				tools.LoggerGeolocation.Log(tools.FATAL,
					"Error Decoding Range on Line %d: %s", len(ENTRIES_IPV4)+1, err)
			}

			ENTRIES_IPV4 = append(ENTRIES_IPV4, IPV4Entry{
				RangeStart:     uint32(rangeStart),
				CountryIndex:   findString(trimString(splits[3])),
				RegionIndex:    findString(trimString(splits[4])),
				CityIndex:      findString(trimString(splits[5])),
				TimezoneOffset: parseOffset(trimString(splits[9])),
			})
		}
	}

	// Process IPV6 CSV
	// {
	// 	tools.LoggerGeolocation.Log(tools.INFO, "Downloading IPV6 Archive")
	// 	f, err := download("DB11LITECSVIPV6", "IP2LOCATION-LITE-DB11.IPV6.CSV", REQUEST_KEY)
	// 	if err != nil {
	// 		tools.LoggerGeolocation.Log(tools.FATAL, "%s", err)
	// 		return
	// 	}
	// 	defer f.Close()

	// 	tools.LoggerGeolocation.Log(tools.INFO, "Parsing IPV6 Archive")
	// 	scanner := bufio.NewScanner(f)
	// 	for scanner.Scan() {
	// 		splits := strings.SplitN(scanner.Text(), ",", 10)
	// 		if len(splits) != 10 {
	// 			tools.LoggerGeolocation.Log(tools.FATAL,
	// 				"Error Decoding Line %d: Bad Split", len(ENTRIES_IPV6)+1)
	// 		}
	// 		rangeStr := trimString(splits[0])
	// 		bigInt := new(big.Int)
	// 		if _, ok := bigInt.SetString(rangeStr, 10); !ok {
	// 			tools.LoggerGeolocation.Log(tools.FATAL,
	// 				"Error Decoding Range on Line %d", len(ENTRIES_IPV6)+1)
	// 		}
	// 		rangeBytes := bigInt.FillBytes(make([]byte, 16)) // big-endian by default
	// 		rangeStart := [16]byte{}
	// 		copy(rangeStart[:], rangeBytes)

	// 		ENTRIES_IPV6 = append(ENTRIES_IPV6, IPV6Entry{
	// 			RangeStart:     rangeStart,
	// 			CountryIndex:   findString(trimString(splits[3])),
	// 			RegionIndex:    findString(trimString(splits[4])),
	// 			CityIndex:      findString(trimString(splits[5])),
	// 			TimezoneOffset: parseOffset(trimString(splits[9])),
	// 		})
	// 	}
	// }

	// Begin Compression
	tools.LoggerGeolocation.Log(tools.INFO,
		"Compressing: %d IPV4 Ranges, %d IPV6 Ranges, %d Strings",
		len(ENTRIES_IPV4), len(ENTRIES_IPV6), len(ENTRIES_TEXT),
	)

	output, err := os.Create(OUTPUT_FILE)
	if err != nil {
		tools.LoggerGeolocation.Log(tools.FATAL, "Error Creating Archive File: %s", err)
	}
	defer output.Close()

	writer := gzip.NewWriter(output)
	defer writer.Close()

	// Repackage Contents
	binary.Write(writer, binary.LittleEndian, uint32(len(ENTRIES_IPV4)))
	binary.Write(writer, binary.LittleEndian, uint32(len(ENTRIES_IPV6)))
	binary.Write(writer, binary.LittleEndian, uint32(len(ENTRIES_TEXT)))
	for _, e := range ENTRIES_IPV4 {
		binary.Write(writer, binary.LittleEndian, e.RangeStart)
		binary.Write(writer, binary.LittleEndian, e.CountryIndex)
		binary.Write(writer, binary.LittleEndian, e.RegionIndex)
		binary.Write(writer, binary.LittleEndian, e.CityIndex)
		binary.Write(writer, binary.LittleEndian, e.TimezoneOffset)
	}
	for _, e := range ENTRIES_IPV6 {
		binary.Write(writer, binary.LittleEndian, e.RangeStart)
		binary.Write(writer, binary.LittleEndian, e.CountryIndex)
		binary.Write(writer, binary.LittleEndian, e.RegionIndex)
		binary.Write(writer, binary.LittleEndian, e.CityIndex)
		binary.Write(writer, binary.LittleEndian, e.TimezoneOffset)
	}
	for _, s := range ENTRIES_TEXT {
		binary.Write(writer, binary.LittleEndian, uint8(len(s)))
		io.WriteString(writer, s)
	}

	tools.LoggerGeolocation.Log(tools.INFO, "Completed in %s", time.Since(t))
}
