package etherscan_parse

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type TokenStruct struct {
	token     string
	tokenType string
}

func TestGetInputParams(t *testing.T) {
	tokens := getTokensMap()
	for _, test := range tokens {
		// GET url/TOKEN
		token := test.token
		token_ := "\"" + token + "\""
		req := httptest.NewRequest("GET", "/", strings.NewReader(token_))
		req.Header.Add("Content-Type", "application/json")

		res, err := getInputToken(req)
		if err != nil {
			t.Errorf(err.Error())
		}
		if res != token {
			t.Errorf("getInputParams (%q), want %q", token, res)
		}

		// GET url?token=TOKEN

		//params := url.Values{}
		//params.Add("token", token)

		req = httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Content-Type", "application/json")

		q := req.URL.Query()
		q.Add("token", token)
		req.URL.RawQuery = q.Encode()

		res, err = getInputToken(req)
		if err != nil {
			log.Fatal(err)
		}
		if res != token {
			t.Errorf("getInputParams (%q), want %q", token, res)
		}

		// POST application/json
		values := map[string]string{"token": token}
		json_data, _ := json.Marshal(values)
		req = httptest.NewRequest("POST", "/", bytes.NewBuffer(json_data))
		req.Header.Add("Content-Type", "application/json")
		res, err = getInputToken(req)
		if err != nil {
			log.Fatal(err)
		}
		if res != token {
			t.Errorf("getInputParams (%q), want %q", token, res)
		}

		// POST multipart/form-data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		fw, _ := writer.CreateFormField("token")
		_, _ = io.Copy(fw, strings.NewReader(token))
		writer.Close()
		req = httptest.NewRequest("POST", "/", bytes.NewReader(body.Bytes()))
		req.Header.Add("Content-Type", writer.FormDataContentType())
		res, err = getInputToken(req)
		if err != nil {
			log.Fatal(err)
		}
		if res != token {
			t.Errorf("getInputParams (%q), want %q", token, res)
		}

		// POST application/x-www-form-urlencoded
		data := url.Values{}
		data.Set("token", token)
		req = httptest.NewRequest("POST", "/", strings.NewReader(data.Encode())) // URL-encoded payload
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res, err = getInputToken(req)
		if err != nil {
			log.Fatal(err)
		}
		if res != token {
			t.Errorf("getInputParams (%q), want %q", token, res)
		}
	}
}

func _TestParceBody(t *testing.T) {

	token := "0xcf39b7793512f03f2893c16459fd72e65d2ed00c"
	tokenType := "[ERC-20]"
	fileName := "./ERC-20.html"
	testParceBody(t, fileName, token, tokenType)

	token = "0x2af75676692817d85121353f0d6e8e9ae6ad5576"
	tokenType = "[ERC-1155]"
	fileName = "./ERC-1155.html"
	testParceBody(t, fileName, token, tokenType)

	token = "0x57f1887a8bf19b14fc0df6fd9b2acc9af147ea85"
	tokenType = "[ERC-721]"
	fileName = "./ERC-721.html"
	testParceBody(t, fileName, token, tokenType)
}

func TestEtherscanParce(t *testing.T) {
	tokens := getTokensMap()
	for _, test := range tokens {

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Content-Type", "application/json")

		q := req.URL.Query()
		q.Add("token", test.token)
		req.URL.RawQuery = q.Encode()

		rr := httptest.NewRecorder()
		etherscanParse(rr, req)

		got := rr.Body.String()
		var answ Answer
		err := json.Unmarshal([]byte(got), &answ)
		if err != nil {
			log.Fatal(err)
		}
		if answ.ErrorStr != "" {
			t.Error(answ.ErrorStr)
		} else if answ.Token != test.token || answ.TokenType != test.tokenType {
			t.Errorf("request(%q:%q), parse (%q:%q)", test.token, test.tokenType, answ.Token, answ.TokenType)
		}
		//time.Sleep(3 * time.Second)
	}
}

func testParceBody(t *testing.T, fileName, token, tokenType string) {
	body, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	answ, err := parseBody(body)
	if err != nil {
		log.Fatal(err)
	}
	if answ.ErrorStr != "" {
		log.Fatal(errors.New(answ.ErrorStr))
	}
	if answ.Token != token || answ.TokenType != tokenType {
		t.Errorf("in file(%q:%q), parse (%q:%q)", token, tokenType, answ.Token, answ.TokenType)
	}

}
func getTokensMap() []TokenStruct {

	tokens := make([]TokenStruct, 0)

	tokenFile, err := ioutil.ReadFile("./tokens.csv")

	if err != nil {
		log.Fatal(err)
	}

	tokenLines := strings.Split(string(tokenFile), "\n")

	for i := 0; i < len(tokenLines); i++ {

		if tokenLines[i] != "" {

			configLine := strings.Split(string(tokenLines[i]), ",")

			newConfig := TokenStruct{token: configLine[0], tokenType: configLine[1]}
			tokens = append(tokens, newConfig)
		}
	}
	return tokens
}
