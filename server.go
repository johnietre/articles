package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type Article struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type Articles struct {
	Articles []Article `json:"articles"`
	s        *http.Server
	mtx      sync.Mutex
}

var expr = regexp.MustCompile(``)

func (a *Articles) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	strings.Split(r.URL.Path, "/")
}

var (
	scriptExpr = regexp.MustCompile(`<script.*?>.*?</script>`)
	titleExpr  = regexp.MustCompile(`<title.*?>(.*)</title>`)
	bodyExpr   = regexp.MustCompile(`<body.*?>`)
)

func getArticle(link string) (string, error) {
	resp, err := http.Get("https://blog.devgenius.io/why-are-developers-slow-5038e1755012")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Remove all the script tags
	data = scriptExpr.ReplaceAll(data, []byte{})
	// Get the indexes of the opening and closing body tags
	first, last := bodyExpr.FindIndex(data), bytes.LastIndex(data, []byte("</body>"))
	fmt.Println(data[first[1]:last])
	// Get the title
	matches := titleExpr.FindAll(data, 1)
	if len(matches) == 0 {
		return "", nil
	}

	return string(matches[0]), nil
}
