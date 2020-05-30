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
var r_links = regexp.MustCompile(`\b(?i)href="([^"]*)"`)
var r_js = regexp.MustCompile(`\b(?i)src="([^"]*)"`)
var r_mailto = regexp.MustCompile(`\b(?i)href="(mailto:[^"]*)"`)

var outputFilePaths = map[string]struct{}{}
var outputFilePathsMux sync.Mutex

var output_dir = "SeniorProjects"

var dir_lock sync.Mutex
var https_pool_size = 10 // DON'T INCREASE THIS TOO MUCH IT CAN CRASH YOUR COMPUTER  

func main() {
	url_name := flag.String("url", "https://yichaolemon.github.io/", "url to start the scraping")
	filename := flag.String("filename", "self.html", "filename for the intitial page download")
	flag.Parse()
	fn := filepath.Join(output_dir, *filename)
	retryDownURLtoFile(*url_name, fn)
}

// set of http clients so that don't create too many
func initHttpPool () chan *http.Client {
  var http_thread_pool = make (chan *http.Client, https_pool_size)
  for i:=0; i<https_pool_size; i++ {
    http_thread_pool <- &http.Client{}
  }
  return http_thread_pool
}

var http_thread_pool = initHttpPool()

// download a single URL 
func downloadURL (url string) (io.ReadCloser, string) {
  httpClient := <- http_thread_pool
  defer func(){ http_thread_pool <- httpClient }()
  finalURL := url
	
  httpClient.CheckRedirect = 
		func(req *http.Request, via []*http.Request) error {
			finalURL = req.URL.String()
			return nil
		}

	resp, err := httpClient.Get(url)
	if err != nil {
		// logErr(err)
		return nil, ""
	}
  // if CheckRedirect is ever called after httpClient.Get returns,
  // then we have an issue because we don't get the correct finalURL.
  httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
    panic("AAAAAHHHHH")
  }
	return resp.Body, finalURL
}

func createFileWriter (fn string) (io.WriteCloser, error) {
	dir_lock.Lock()
	defer dir_lock.Unlock()

	fileinfo, err := os.Stat(fn)
	if err == nil && fileinfo.IsDir() {
		fn = filepath.Join(fn, "index.html")
	}
	file, err := os.OpenFile(fn, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0660)
	if err != nil {
    return nil, err
	}
	return file, nil
}

func retryDownURLtoFile (url string, fn string) {
	// if it is already being downloaded, skip
	outputFilePathsMux.Lock()
	if _, exists := outputFilePaths[fn]; exists {
		outputFilePathsMux.Unlock()
		return
	}
	outputFilePaths[fn] = struct{}{}
	outputFilePathsMux.Unlock()

  var err error 
  for i := 0; i < 10; i++ {
    err = downURLtoFile(url, fn)
    if err == nil { return }
    fmt.Printf("RETRY downloading file [%s]\n", fn)
  }
  fmt.Printf("GAVE UP downloading file [%s]\n", fn)
  logErr(err)
}

func downURLtoFile (url string, fn string) error {
	fmt.Printf("downloading file\t\t %s\n", fn)

	p_reader, p_writer := io.Pipe()
	defer p_reader.Close()

	// channel of lines 
	line_chan := make(chan string, 100)
	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	// pointer used across goroutines, does not need to be
	// protected because as one channel writes it before starting
	// output to line_chan, and the other only reads it after reading from line_chan.
	finalURL := &url

	go func() {
		defer wg.Done() // so that this go routine exits 
		defer p_writer.Close() // so that the reader will get closed 
		defer close(line_chan) // so that the thread that processes the lines can finish 
		readcloser, newURL := downloadURL(url)
		if readcloser == nil {
			return
		}
		defer readcloser.Close()
		*finalURL = newURL
		// line reader 
		linereader := bufio.NewReader(readcloser)

		done := false
		for !done {
			line, err := linereader.ReadString('\n')
			if err == io.EOF {
				done = true
			} else if err != nil {
				p_writer.CloseWithError(err)
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

	// process the lines as they are pushed to the channel 
	go func() {
		defer wg.Done() // so that this go routine exits 
		processLine(finalURL, line_chan)
	}()
	
	writecloser, err := createFileWriter (fn)
	if err != nil {
		return err
	}
	defer writecloser.Close()

  // errors from the p_writer.CloseWithError pop out here
	_, err = io.Copy(writecloser, p_reader)

	if err != nil {
		return err
	}
	fmt.Printf("downloaded file\t\t %s\n", fn)
  return nil
}

func processLine (finalURL *string, lines chan string) {
	var wg sync.WaitGroup
	for line := range lines {
		parsed_url, err := url.Parse(*finalURL)
		if err != nil {
			logErr(err)
			return
		}
		lineProcessor(&wg, parsed_url, line)
	}
	wg.Wait()
}

func convertFileToDirectory(dir_path string) error {
  fmt.Printf("switch files %s\n", dir_path)
  err := os.Rename(dir_path, dir_path+"(conflict)")
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir_path, 0777)
  if err != nil {
    return err 
  }
  err = os.Rename(dir_path+"(conflict)", filepath.Join(dir_path, "index.html"))
  if err != nil {
    return err
  }
  return nil
}

// creates a directory, does a swap if already exists as a file 
func createDir(dir_path string) error {
	dir_lock.Lock()
	defer dir_lock.Unlock()
  return createDirRecursive(dir_path)
}

func createDirRecursive(dir_path string) error {
	fileinfo, err := os.Stat(dir_path)

	// ignore errors about the path not existing, including errors about parent components
	// of the path not existing.
  // There are two types of "errors" which we can handle, i.e. they are not really errors.
  // 1. nothing exists at that path. (that's fine because we're trying to create something there)
  // 2. a parent path does not exist or is a file.
	if err != nil {
    pathIsAvailable := os.IsNotExist(err)
    parentPathIsInvalid := strings.Contains(err.Error(), "not a directory")
    if pathIsAvailable {
      // okay, continue
    } else if parentPathIsInvalid {
      err = createDirRecursive(filepath.Dir(dir_path))
      if err != nil {
        return err
      }
    } else {
      return err
    }
	}

	if fileinfo != nil && fileinfo.Mode().IsRegular() {
    convertFileToDirectory(dir_path)
	} else {
    // the path is available.
    // there's either nothing there or there's already a directory there.
		err = os.MkdirAll(dir_path, 0777)
		if err != nil {
			return err 
		}
	}
	return nil 
}

func downloadLink(wg *sync.WaitGroup, parsed_url *url.URL, link string) {
	wg.Add(1)
	go func (){
		defer wg.Done()
		//fmt.Printf("linked to file %s\n", link)
		css_url, err := parsed_url.Parse(link)
		if err != nil {
			//logErr (err)
			return
		}

		// only download relative paths (i.e., files within the same domain)
		parsed_css_url, err := url.Parse(link)
		if len(parsed_css_url.Host) == 0 {
			// relative path
			css_path := filepath.Join(output_dir, css_url.Path)
			// need to know where there is already a file there
			dir_path := filepath.Dir(css_path)

			err := createDir(dir_path)
			if err != nil {
				logErr(err)
				return
			}
			retryDownURLtoFile(css_url.String(), css_path)
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



