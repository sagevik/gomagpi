/*
Simple go script for downloading MagPi magazine
*/
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const DOWNLOADPATH = "Documents/Magazines/MagPi/"

func isDownloaded(dir, pdf string) (bool, string) {
	var files []string
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	for _, f := range files {
		if path.Base(f) == pdf {
			return true, f
		}
	}
	return false, ""
}

func checkErr(e error) {
	if e != nil {
		fmt.Println("Error:", e.Error())
		os.Exit(1)
	}
}
func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return info.IsDir()
}

func createDirectoryIfNotExists(path string) {
	if !directoryExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("Could not create download directory")
			os.Exit(1)
		}
	}
}

func createDownloadLink(pre, s string) string {
	var pdfPart string
	linkParts := strings.Split(s, "href=\"")
	for _, pt := range linkParts {
		if strings.Contains(pt, ".pdf") {
			pdfPart = pt
		}
	}
	return pre + pdfPart
}

func checkArgs(s1, s2 string) (int, int) {
	i1, err := strconv.Atoi(s1)
	checkErr(err)
	i2, err := strconv.Atoi(s2)
	checkErr(err)
	if i1 > i2 {
		return i2, i1
	}
	return i1, i2
}

func downloadPath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Could not get user home dir")
		os.Exit(1)
	}
	return dirname
}

func parseArgs(a []string) []string {
	var issues []string
	if len(a) < 2 || len(a) > 3 {
		fmt.Println("Specify one MagPi issue or a range of MagPi issues to download")
		fmt.Println("for example: 'gomagpi 123' or 'gomagpi 123 132'")
		os.Exit(0)
	}
	if len(a) == 2 {
		issues = append(issues, a[1])
	}
	if len(a) == 3 {
		start, end := checkArgs(a[1], a[2])
		for i := start; i <= end; i++ {
			issues = append(issues, strconv.Itoa(i))
		}
	}

	return issues
}

func main() {
	issues := parseArgs(os.Args)

	dl_prefix := "https://magpi.raspberrypi.com"
	searchDl := "href=\"/downloads/"
	searchPdf := ".pdf"

	for _, nr := range issues {
		fmt.Println("Trying to download MagPi issue nr:", nr)
		dl_url := "https://magpi.raspberrypi.com/issues/" + nr + "/pdf/download"
		dl, err := http.Get(dl_url)
		checkErr(err)

		body, err := io.ReadAll(dl.Body)
		checkErr(err)

		bodyStr := string(body)
		bodyParts := strings.Split(bodyStr, "\">")
		var dlLink string
		for _, p := range bodyParts {
			if strings.Contains(p, searchDl) && strings.Contains(p, searchPdf) {
				dlLink = createDownloadLink(dl_prefix, p)
				break
			}
		}

		filenameParts := strings.Split(dlLink, "/")
		fileName := filenameParts[len(filenameParts)-1]
		if fileName == "" {
			fmt.Println("Could not find download link to MagPi issue nr:", nr)
			continue
		}

		dlPath := path.Join(downloadPath(), DOWNLOADPATH)

		createDirectoryIfNotExists(dlPath)

		check, localPath := isDownloaded(dlPath, fileName)
		if check {
			fmt.Printf("MagPi issue nr: %s is already downloaded -> %s\n", nr, localPath)
			continue
		}

		fmt.Println("Downloading MagPi issue nr:", nr)

		file, err := os.Create(path.Join(dlPath, fileName))
		checkErr(err)

		client := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}

		resp, err := client.Get(dlLink)
		checkErr(err)
		defer resp.Body.Close()

		_, err = io.Copy(file, resp.Body)
		checkErr(err)
		defer file.Close()
		fmt.Printf("Downloaded: %s\n", fileName)
	}
	fmt.Println("Finished.")
}
