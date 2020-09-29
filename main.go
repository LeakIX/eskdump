package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/schollz/progressbar"
	"gitlab.nobody.run/tbi/core"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient *http.Client

func main() {
	// Usage
	if len(os.Args) < 2 {
		log.Fatal("Usage: eskdump \"http(s)://<ip>:<port>\" \"<index-name/pattern>\" [size]")
	}
	// Parse args
	url := os.Args[1]
	index := os.Args[2]
	size := "100"
	if len(os.Args) > 3 {
		size = os.Args[3]
	}

	log.Println("ESKDump starting...")
	log.Println("Kibana endpoint : " + shellYellow(url))
	log.Println("Index : " + shellYellow(index))

	//Create proxied http client ( Allow ALL_PROXY, HTTP_PROXY environment support )
	proxiedPlugin := &core.ProxiedPlugin{}
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext:       proxiedPlugin.DialContext,
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// make sure URL trailing / is removed
	url = strings.TrimLeft(url, "/")
	// Create first scroll request, use /api/console/proxy endpoint to proxy request to ES
	kibanaScroll := fmt.Sprintf("/api/console/proxy?method=POST&path=/%s/_search?scroll=5m", index)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", url, kibanaScroll), bytes.NewBuffer([]byte("{\"size\":"+size+",\"sort\":\"_doc\"}")))
	if err != nil {
		log.Fatal("can't create request:", shellRed(err))
	}
	// https://www.elastic.co/guide/en/kibana/current/using-api.html
	req.Header["kbn-xsrf"] = []string{"true"}
	req.Header["User-Agent"] = []string{"ESKDump/0.0.1"}
	req.Header["Content-Type"] = []string{"application/json"}
	// Fetch and parse :
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("can't GET page:", shellRed(err))
	}
	scrollResponse := ScrollResponse{}
	jsonDecoder := json.NewDecoder(resp.Body)
	err = jsonDecoder.Decode(&scrollResponse)
	if err != nil {
		log.Fatal("Json decode error. Make sure you are targeting a Kibana >5 instance : ", shellRed(err))
	}
	log.Println("Got scrollId : " + shellYellow(scrollResponse.ScrollId))
	log.Println(fmt.Sprintf("Dumping %s documents to stdout :", shellYellow(scrollResponse.Hits.Total.Value)))
	// Progress bar, yeaaah
	bar := progressbar.Default(scrollResponse.Hits.Total.Value, "Docs")
	for _, hit := range scrollResponse.Hits.Hits {
		bar.Add(1)
		hit, _ := json.Marshal(hit)
		fmt.Println(string(hit))
	}

	// now we loop and do the same using our scrollID
	for {
		kibanaScroll := "/api/console/proxy?method=POST&path=/_search/scroll"
		req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", url, kibanaScroll), bytes.NewBuffer([]byte(
			fmt.Sprintf("{\"scroll\":\"5m\",\"scroll_id\":\"%s\"}", scrollResponse.ScrollId))))
		if err != nil {
			log.Fatal("can't create request:", err)
		}
		req.Header["kbn-xsrf"] = []string{"true"}
		req.Header["User-Agent"] = []string{"ESKDump/0.0.1"}
		req.Header["Content-Type"] = []string{"application/json"}
		// use the http client to fetch the page
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatal("can't GET page:", err)
		}
		if err != nil {
			log.Println("Timed out, continuing in 10s ...")
			time.Sleep(10 * time.Second)
			continue
		}
		scrollResponse := ScrollResponse{}
		jsonDecoder := json.NewDecoder(resp.Body)
		err = jsonDecoder.Decode(&scrollResponse)
		if err != nil {
			log.Println("Couldn't parse sleeping 10s ...")
			time.Sleep(10 * time.Second)
			log.Fatal(err)
		}
		if len(scrollResponse.Hits.Hits) < 1 {
			// Seems like there's no more results, exit
			log.Println("Dump completed")
			os.Exit(0)
		}
		for _, hit := range scrollResponse.Hits.Hits {
			bar.Add(1)
			hit, _ := json.Marshal(hit)
			fmt.Println(string(hit))
		}
	}

}

var shellYellow = color.New(color.FgYellow).SprintFunc()
var shellRed = color.New(color.FgRed).SprintFunc()

type ScrollResponse struct {
	Scroll   string `json:"scroll"`
	ScrollId string `json:"_scroll_id"`
	Hits     struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []interface{} `json:"hits"`
	} `json:"hits"`
}
