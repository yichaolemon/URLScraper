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
)

// regex 
var r_css = regexp.MustCompile(`href="([^"]*\.css[^"]*)"`)

func main() {
	fmt.Println("Pusheens!")
	url := "http://www.leedanilek.com"
	fn := "output/totola.html"
	downURLtoFile(url, fn)
}

// download a single URL 
func downloadURL (url string) io.ReadCloser {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	return resp.Body
}

func writeToFile (fn string) io.WriteCloser {
	file, err := os.OpenFile(fn, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0660)
	if err != nil {
		panic(err)
	}
	return file
}

func downURLtoFile (url string, fn string) {
	p_reader, p_writer := io.Pipe()
	defer p_reader.Close()

	// channel of lines 
	line_chan := make(chan string, 20)

	go func() {
		defer p_writer.Close() // so that the reader will get closed 
		defer close(line_chan) // so that the thread that processes the lines can finish 
		readcloser := downloadURL (url)
		defer readcloser.Close()
		// line reader 
		linereader := bufio.NewReader(readcloser)

		for {
			line, err := linereader.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}

			line_chan <- line

			_, err = io.Copy(p_writer, strings.NewReader(line))
			if err != nil {
				panic (err)
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
	defer writecloser.Close()

	bytes, err := io.Copy(writecloser, p_reader)

	if err != nil {
		panic(err)
	}
	fmt.Printf("%d bytes written to file %s\n", bytes, fn)
}

func processLine (urlName string, lines chan string) {
	parsed_url, err := url.Parse(urlName)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for line := range lines {
		lineProcessor(&wg, parsed_url, line)
	}
	wg.Wait()
}

func lineProcessor (wg *sync.WaitGroup, parsed_url *url.URL, line string) {
	var strList []string 
	strList = r_css.FindStringSubmatch(line)

	if len(strList) == 2 {
		cssFilename := strList[1]
		fmt.Printf("used CSS file %s\n", cssFilename)
		cssPath, err := parsed_url.Parse(cssFilename)
		if err != nil {
			panic (err)
		}
		// download css_path
		wg.Add(1)
		go func() {
			defer wg.Done()
			downURLtoFile(cssPath.String(), "output/style.css")
		}()
		fmt.Printf("needs to download CSS file %s\n", cssPath)
	}
}





