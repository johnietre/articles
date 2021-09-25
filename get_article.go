package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func createArticleFile() {
	resp, err := http.Get("https://blog.devgenius.io/why-are-developers-slow-5038e1755012")
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	data = scriptExpr.ReplaceAll(data, []byte{})
	fmt.Println(titleExpr.FindStringSubmatch(string(data))[1])
	if err := ioutil.WriteFile("test-article.html", data, 0755); err != nil {
		panic(err)
	}
}
