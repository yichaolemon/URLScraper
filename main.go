package main

import (
	"fmt"
	"io"
	"net/http"
	// "io/ioutil"
	"os"
)

func main() {
	fmt.Println("Pusheens!")
	url := "http://leedanilek.com/"
	fn := "output/totosla.html"
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

	go func() {
		defer p_writer.Close() // so that the reader will get closed 

		readcloser := downloadURL (url)
		defer readcloser.Close()
		bytes, err := io.Copy(p_writer, readcloser)
		if err != nil {
			panic (err)
		}
		fmt.Printf("%d bytes written to pipe\n", bytes)
	}()
	
	writecloser := writeToFile (fn)
	defer writecloser.Close()

	bytes, err := io.Copy(writecloser, p_reader)

	if err != nil {
		panic(err)
	}
	fmt.Printf("%d bytes written to file %s\n", bytes, fn)
}

