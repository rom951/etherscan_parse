package etherscan_parse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	//"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Answer struct {
	ErrorStr  string `json:"error"`
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
}

const etherscan_url = "https://etherscan.io/token/"

func init() {
	// Register an HTTP function with the Functions Framework
	//functions.HTTP("EtherscanParse", etherscanParse)
}

func etherscanParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := getInputToken(r)
	if err != nil {
		json.NewEncoder(w).Encode(Answer{err.Error(), "", ""})
		return
	}

	res, err := ethercsanGet(token)
	if err != nil {
		json.NewEncoder(w).Encode(Answer{err.Error(), "", ""})
		return
	}

	answer, err := parseBody(res)
	if err != nil {
		json.NewEncoder(w).Encode(Answer{err.Error(), "", ""})
		return
	}
	if answer.ErrorStr != "" {
		answer.ErrorStr = "(" + token + ") " + answer.ErrorStr
		json.NewEncoder(w).Encode(*answer)
		return
	} else if answer.Token != token {
		json.NewEncoder(w).Encode(Answer{
			fmt.Sprintf("Requested (%s) and received (%s) token do not match.", token, answer.Token),
			"", ""})
		return
	}

	json.NewEncoder(w).Encode(*answer)
}

func ethercsanGet(token string) ([]byte, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	etherscanURL, err := url.Parse(etherscan_url + token)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		//		Body:   ioutil.NopCloser(strings.NewReader(token)),
		URL: etherscanURL,
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func parseBody(body []byte) (*Answer, error) {
	answer := Answer{}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		answer.ErrorStr = err.Error()
		return &answer, nil
	}

	// Find the review items
	// //*/span[@class="text-secondary small"] [0-token, 1 - type]
	// span[class='text-secondary small'] [0-token, 1 - type]
	// h2[class='card-header-title'] span[class='text-secondary small'] - type
	// div[class='col-md-6'] a[class='text-truncate d-block mr-2'] - contract
	var res = make([]string, 0)
	doc.Find("div[class='col-md-6'] a[class='text-truncate d-block mr-2']").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Text()
		if title != "" {
			res = append(res, title)
		}
	})
	doc.Find("h2[class='card-header-title'] span[class='text-secondary small']").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Text()
		if title != "" {
			res = append(res, title)
		}
	})
	if len(res) != 2 {
		answer.ErrorStr = "The response does not contain information about the token or the response format has been changed or timeOut."
		return &answer, nil
	}

	answer.Token = res[0]
	answer.TokenType = res[1]
	return &answer, nil
}

type InputParams map[string]string

func getToInputValues(v url.Values) InputParams {
	res := make(InputParams)
	for key, val := range v {
		res[key] = val[0]
	}
	return res
}
func getInputToken(r *http.Request) (string, error) {
	var v InputParams
	var token string
	if r.Method == "GET" {
		token = r.URL.Query().Get("token")
		if token != "" {
			return token, nil
		} else {
			if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
				return "", err
			}

			if token == "" {
				return "", errors.New("Not input params")
			}
			return token, nil
		}
	} else if r.Method == "POST" {
		if r.Header.Get("Content-Type") != "" {
			if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
				body, _ := ioutil.ReadAll(r.Body)
				if err := json.Unmarshal(body, &v); err != nil {
					return "", err
				}
			} else {
				r.ParseMultipartForm(r.ContentLength + 1)
				v = getToInputValues(r.PostForm)
			}
		}
		if len(v) != 0 && v["token"] != "" {
			return v["token"], nil
		}
	}

	return "", errors.New("Not input params")
}
