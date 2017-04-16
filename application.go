package main

import "golang.org/x/net/html"
import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func findElement(n *html.Node, tag string, tagIdKey string, tagIdValue string, attrName string) (token string, err error) {
	if n.Type == html.ElementNode && n.Data == tag {
		for _, a := range n.Attr {
			if a.Key == tagIdKey && a.Val == tagIdValue {
				for _, a := range n.Attr {
					if a.Key == attrName {
						return a.Val, nil
					}
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		token, err := findElement(c, tag, tagIdKey, tagIdValue, attrName)
		if err == nil {
			return token, err
		}
	}
	return "", errors.New("couldnt find token")
}

func main() {
	content, err := getPageContent("https://magnetis.com.br/users/sign_in")
	//content, err := ioutil.ReadFile("login.html")
	if err != nil {
		log.Fatal(err)
	}
	pageHtml := string(content)
	doc, err := html.Parse(strings.NewReader(pageHtml))
	if err != nil {
		log.Fatal(err)
	}
	token, err := findElement(doc, "input", "name", "authenticity_token", "value")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Token: " + token)
}

func getPageContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
