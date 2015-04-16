package maker

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type PageLoader struct {
	Referer    string
	Cookie     string
	NumTries   int
	TimeoutSec int
}

type HttpService interface {
	executeRequest(url string) []byte
}

func (pageLoader PageLoader) executeRequest(url string) ([]byte, error) {
	var result []byte

	var timeout time.Duration
	if pageLoader.TimeoutSec > 0 {
		timeout = time.Duration(time.Duration(pageLoader.TimeoutSec) * time.Second)
	} else {
		timeout = time.Duration(10 * time.Second)
	}
	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.2.2) Gecko/20100316 Firefox/3.6.2")
	if pageLoader.Referer != "" {
		req.Header.Add("Referer", pageLoader.Referer)
	}
	if pageLoader.Cookie != "" {
		req.Header.Add("Cookie", pageLoader.Cookie)
	}

	var n int
	if pageLoader.NumTries > 0 {
		n = pageLoader.NumTries
	} else {
		n = 1
	}
	for i := 0; i < n; i++ {
		log.Printf("Attempt#%d to load from %s\n", (i + 1), url)
		resp, err := client.Do(req)
		defer resp.Body.Close()
		if err == nil {
			result, err = ioutil.ReadAll(resp.Body)
		}
		resp.Body.Close()
		if result != nil {
			break
		}
	}
	if result == nil || len(result) == 0 {
		return result, errors.New("No html data loaded")
	}
	return result, err
}

func GetUrl(target string, request string) string {
	url := strings.TrimRight(target, "/") + "/" + strings.TrimLeft(request, "/")
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	return url
}
