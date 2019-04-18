package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Record struct {
	Data    string
	Host    string
	Ip      string
	Path    string
	Port    int32
	subject string
	Vhost   string
}

// decodedBase64 takes a string encoded as base64 and returns the string
// decoded as a UTF-8 string
// params:  base64 encoded string
// returns: decoded string
func decodeBase64(encoded string) string {
	decodedString, _ := base64.StdEncoding.DecodeString(encoded)
	return string(decodedString)
}

// Getbody contains a the data received from an HTTP GET response as
// a string, and treats it as if it came from an actual http response
// params:  data from HTTP GET request response
// returns: body of response
func getBody(resp string) string {
	strReader := strings.NewReader(resp)
	strBufioReader := bufio.NewReader(strReader)
	response, err := http.ReadResponse(strBufioReader, nil)
	// Skip response if error
	if err != nil {
		return ""
	}
	// Figure out if response is gzipped and uncompress if needed
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		// Skip the response if error
		if err != nil {
			return ""
		}
		defer reader.Close()
	default:
		reader = response.Body
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	body := buf.String()
	return body
}

// nextLine gets the next line a file
// params:  the open bufio.Reader for the file
// returns: the next line in the file
func nextLine(reader *bufio.Reader) (string, error) {
	var lineBuf bytes.Buffer
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil && err == io.EOF {
			return string(line), err
		}
		lineBuf.Write(line)
		if !isPrefix {
			return lineBuf.String(), err
		}
	}
}

// searchFile gets lines from file specified by fname, processes
// the json on each line, and returns a list of ip addresses whose
// GET response contains any of the strings in searchTerms
// params:
//   fname:       the name of the rapid7 dataset file
//   maxReaders:  the maximum number of goroutines that can be reading lines
//                at one time
//   searchTerms: a list of the terms you want to search for in the file
// returns: list of IP addresses whose HTTP GET response body contains one
//          of the specified search terms
func searchFile(fname string, maxReaders int, searchTerms []string) []string {
	// Open records for reading. Each line contains json with content
	// and some information about the GET request
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println("Failed to open file.")
	}
	reader := bufio.NewReader(file)

	var wg sync.WaitGroup
	readerSync := make(chan bool, maxReaders)
	for {
		line, err := nextLine(reader)
		// Will result in exiting for EOF
		if err != nil {
			break
		}
		// Add a bool to the channel that goroutines will receive at the end of
		// their execution
		readerSync <- true
		wg.Add(1)
		go func(line string) {
			defer func() {
				<-readerSync
				wg.Done()
			}()
			var record Record
			// Unpack json object and get the body of the HTTP Response
			json.Unmarshal([]byte(line), &record)
			decodedData := decodeBase64(record.Data)
			body := getBody(decodedData)
			if body == "" {
				return
			}
			// Search the body for the strings in searchTerms
			for _, term := range searchTerms {
				if strings.Contains(body, term) {
					fmt.Printf("IP %s's GET response contains %s\n", record.Ip, term)
				}
			}
		}(line)
	}
	wg.Wait()
	var junk []string
	return junk
}

func main() {
	filenamePtr := flag.String("i", "", "input file")
	maxReadersPtr := flag.Int("m", 1, "maximum number of readers at once")
	searchTermPtr := flag.String("s", "", "term(s) to search for")
	flag.Parse()
	searchTerms := append(flag.Args(), *searchTermPtr)

	results := searchFile(*filenamePtr, *maxReadersPtr, searchTerms)
	for ip := range results {
		fmt.Printf("%s\n", ip)
	}
}
