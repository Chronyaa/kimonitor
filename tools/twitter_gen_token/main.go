package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: ./twitter_gen_token ${consumerKey} ${consumerSecret}")
		return
	}
	consumerKey := os.Args[1]
	consumerSecret := os.Args[2]
	client := &http.Client{}
	req, err := http.NewRequest("POST", `https://api.twitter.com/oauth2/token`, bytes.NewReader([]byte("grant_type=client_credentials")))
	if err != nil {
		log.Println("Create request failed.", err)
		return
	}
	encodedBearTokenCredentials := "Basic " + base64.StdEncoding.EncodeToString([]byte(consumerKey+":"+consumerSecret))
	req.Header.Add("Authorization", encodedBearTokenCredentials)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Get access token failed.", err)
		return
	}
	auth := &AuthorizationResponse{}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read all failed.", err)
		return
	}
	err = json.Unmarshal(buf, auth)
	if err != nil {
		log.Println("Parse json failed.", err)
		return
	}
	fmt.Println("Bearer " + auth.AccessToken)
}

// AuthorizationResponse autogen
type AuthorizationResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}
