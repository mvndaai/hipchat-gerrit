package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var gerritURL string
var gerritUsername string
var gerritPassword string
var hipchatWebhookURL string
var useHipchat = true

var env = map[string]string{}

func init() {
	getGlobals()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovering from panic: %s", err)
		}
	}()
	if len(os.Args) != 3 {
		log.Println("Please provide a gerrit change ID and a user to add to the review")
		return
	}
	gerritChangeID := os.Args[1]
	gerritUsernameToAdd := os.Args[2]
	fmt.Printf("Attempting to add '%s' as a reviewer on gerrit change '%s'\n", gerritUsernameToAdd, gerritChangeID)
	if !postToGerrit(gerritChangeID, gerritUsernameToAdd) {
		return
	}
	msg := fmt.Sprintf("%s requested to review %s/#/c/%s", gerritUsernameToAdd, gerritURL, gerritChangeID)
	if useHipchat {
		postToHipchat("purple", "@here "+msg)
	}
	fmt.Println(msg)
}

func postToGerrit(changeID string, user string) bool {
	uri := "/a/changes/" + changeID + "/reviewers"
	url := gerritURL + uri
	method := "POST"
	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		log.Printf("Recieved status code '%v' auth skipped", resp.StatusCode)
		return true
	}
	digestParts := digestParts(resp)
	digestParts["uri"] = uri
	digestParts["method"] = method
	digestParts["username"] = gerritUsername
	digestParts["password"] = gerritPassword
	var reqBody = []byte(`{"reviewer": "` + user + `" }`)
	req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", getDigestAuthrization(digestParts))
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Recieved status code '%v' not '%v' from Gerrit\n", resp.StatusCode, http.StatusOK)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Println("Gerrit response body: ", string(body))
		return false
	}
	return true
}

func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	return result
}

func postToHipchat(color string, message string) {
	var jsonStr = []byte(`{"color":"` + color + `","message":"` + message + `","notify":true,"message_format":"text"}`)
	req, err := http.NewRequest("POST", hipchatWebhookURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Hipchat status code '%v' was not '%v'", resp.StatusCode, http.StatusNoContent)
	}
}

func getMD5(text string) string {
	//https://gist.github.com/sergiotapia/8263278
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthrization(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := 00000001
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response)
	return authorization
}

func getGlobals() {
	e := "gerritURL"
	gerritURL = os.Getenv(e)
	if len(gerritURL) == 0 {
		log.Fatalf("Environment variable named '%s' required", e)
	}

	e = "gerritUsername"
	gerritUsername = os.Getenv(e)
	if len(gerritUsername) == 0 {
		log.Fatalf("Environment variable named '%s' required", e)
	}

	e = "gerritPassword"
	gerritPassword = os.Getenv(e)
	if len(gerritPassword) == 0 {
		log.Fatalf("Environment variable named '%s' required", e)
	}

	e = "hipchatWebhookURL"
	hipchatWebhookURL = os.Getenv(e)
	if len(hipchatWebhookURL) == 0 {
		useHipchat = false
		log.Printf("You need an environment variable named '%s' for full features", e)
	}
}
