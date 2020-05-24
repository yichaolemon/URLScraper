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
)

// regex 
var r_css = regexp.MustCompile(`href="([^"]*\.css)"`)

func main() {
	fmt.Println("Pusheens!")
	url := "https://ocw.mit.edu/courses/find-by-topic/#cat=engineering&subcat=computerscience"
	fn := "output/ocw.html"
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

	// process the lines as they are pushed to the channel 
	go processLine(line_chan)
	
	writecloser := writeToFile (fn)
	defer writecloser.Close()

	bytes, err := io.Copy(writecloser, p_reader)

	if err != nil {
		panic(err)
	}
	fmt.Printf("%d bytes written to file %s\n", bytes, fn)
}

func processLine (lines chan string) {
	for line := range lines {
		lineProcessor(line)
	}
}

func lineProcessor (line string) {
	var strList []string 
	strList = r_css.FindStringSubmatch(line)
	if len(strList) == 2 {
		cssFilename := strList[1]
		fmt.Printf("used CSS file %s\n", cssFilename)
	}
}

