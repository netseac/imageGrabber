package main

import (
	"bufio"
	"fmt"
	"image"
	"io/ioutil"
	"log"

	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"

	"image/jpeg"

	"github.com/google/uuid"
)

var (
	fileName, imgUrl string
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter URL")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Replace(text, "\r", "", -1)
		fullUrlFile, err := url.Parse(text)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := httpClient().Get(fullUrlFile.String())

		checkError(err)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		doc, err := html.Parse(strings.NewReader(string(body)))
		if err != nil {
			log.Fatal(err)
		}
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "meta" {
				for _, a := range n.Attr {
					if a.Key == "property" && a.Val == "twitter:image:src" {
						for _, t := range n.Attr {
							if t.Key == "content" {
								imgUrl = t.Val
							}
						}
						break
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)

		// create uploads dir if not exits
		if _, err := os.Stat("uploads"); os.IsNotExist(err) {
			// uploads does not exist
			pathErr := os.Mkdir("uploads", 0777)

			//check for error
			checkError(pathErr)
		}

		// Generate fileName
		fileName = uuid.New().String()

		// Create blank file
		file := createFile()

		// Put content on file
		putFile(file, httpClient(), imgUrl)
	}
}

func putFile(file *os.File, client *http.Client, fullUrlFile string) {
	resp, err := client.Get(fullUrlFile)

	checkError(err)

	defer resp.Body.Close()

	// Decode image to JPEG
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 100}); err != nil {
		log.Fatal(err)
	}

	fi, err := file.Stat()
	if err != nil {
		// Could not obtain stat, handle error
		log.Fatal(err)
	}

	fmt.Printf("\nJust Downloaded a file %s with size %d bytes\n", fileName, fi.Size())
}

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func createFile() *os.File {
	file, err := os.Create("uploads/" + fileName + ".jpg")

	checkError(err)
	return file
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
