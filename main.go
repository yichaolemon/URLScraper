package main

import (
	"fmt"
	"io"
	"net/http"
	// "io/ioutil"
	"os"
	"regexp"
	"bufio"
	"strings"
	"net/url"
	"sync"
	"path/filepath"
	"flag"
	"runtime/debug"
)

// regex 
var r_links = regexp.MustCompile(`\bhref="([^"]*)"`)
var r_js = regexp.MustCompile(`\bsrc="([^"]*)"`)
var r_mailto = regexp.MustCompile(`\bhref="(mailto:[^"]*)"`)

var outputFilePaths = map[string]struct{}{}
var outputFilePathsMux sync.Mutex

var output_dir = "out"

func main() {
	url_name := flag.String("url", "https://yichaolemon.github.io/", "url to start the scraping")
	filename := flag.String("filename", "self.html", "filename for the intitial page download")
	flag.Parse()
	fn := filepath.Join(output_dir, *filename)
	downURLtoFile(*url_name, fn)
}

// download a single URL 
func downloadURL (url string) io.ReadCloser {
	resp, err := http.Get(url)
	if err != nil {
		logErr(err)
	}
	return resp.Body
}

func writeToFile (fn string) io.WriteCloser {
	file, err := os.OpenFile(fn, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0660)
	if err != nil {
		logErr(err)
	}
	return file
}

func downURLtoFile (url string, fn string) {
	// if it is already being downloaded
	outputFilePathsMux.Lock()
	if _, exists := outputFilePaths[fn]; exists {
		outputFilePathsMux.Unlock()
		return
	}
	outputFilePaths[fn] = struct{}{}
	outputFilePathsMux.Unlock()
	fmt.Printf("downloading file\t\t %s\n", fn)


	p_reader, p_writer := io.Pipe()
	defer p_reader.Close()

	// channel of lines 
	line_chan := make(chan string, 1)

	go func() {
		defer p_writer.Close() // so that the reader will get closed 
		defer close(line_chan) // so that the thread that processes the lines can finish 
		readcloser := downloadURL (url)
		if readcloser == nil {
			return
		}
		defer readcloser.Close()
		// line reader 
		linereader := bufio.NewReader(readcloser)

		done := false
		for !done {
			line, err := linereader.ReadString('\n')
			if err == io.EOF {
				done = true
			} else if err != nil {
				logErr(err)
				return
			}

			line_chan <- line

			_, err = io.Copy(p_writer, strings.NewReader(line))
			if err != nil {
				logErr (err)
				return
			}
			// fmt.Printf("%d bytes processed and written into pipe\n", bytes)
		}
	}()

	done := make(chan struct{})
	// process the lines as they are pushed to the channel 
	go func() {
		defer close(done)
		processLine(url, line_chan)
	}()
	defer func() {
		<-done // return only after either channel is closed or something is put into the channel
	}()
	
	writecloser := writeToFile (fn)
	if writecloser == nil {
		return
	}
	defer writecloser.Close()

	_, err := io.Copy(writecloser, p_reader)

	if err != nil {
		logErr(err)
		return
	}
	fmt.Printf("downloaded file\t\t %s\n", fn)
}

func processLine (urlName string, lines chan string) {
	parsed_url, err := url.Parse(urlName)
	if err != nil {
		logErr(err)
	}
	var wg sync.WaitGroup
	for line := range lines {
		lineProcessor(&wg, parsed_url, line)
	}
	wg.Wait()
}

func downloadLink(wg *sync.WaitGroup, parsed_url *url.URL, link string) {
	wg.Add(1)
	go func (){
		defer wg.Done()
		//fmt.Printf("linked to file %s\n", link)
		css_url, err := parsed_url.Parse(link)
		if err != nil {
			logErr (err)
			return
		}
		// want to know whether link is absolute 
		parsed_css_url, err := url.Parse(link)
		if len(parsed_css_url.Host) == 0 {
			// relative path
			css_path := filepath.Join(output_dir, css_url.Path)
			err := os.MkdirAll(filepath.Dir(css_path), 0777)
			if err != nil {
				logErr(err)
				return
			}
			downURLtoFile(css_url.String(), css_path)
		}
	}()
}

func lineProcessor (wg *sync.WaitGroup, parsed_url *url.URL, line string) {
	var strList []string 
	strList = r_links.FindStringSubmatch(line)
	jsList := r_js.FindStringSubmatch(line)
	mailList := r_mailto.FindStringSubmatch(line)
	if len(strList) == 2 && len(mailList) != 2 {
		downloadLink(wg, parsed_url, strList[1])
	}
	if len(jsList) == 2 {
		downloadLink(wg, parsed_url, jsList[1])
	}
}

func logErr(err error) {
	debug.PrintStack()
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
}



