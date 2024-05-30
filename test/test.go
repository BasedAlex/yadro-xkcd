package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"slices"
	"sync"

	"github.com/basedalex/yadro-xkcd/pkg/config"
)

func main() {
	cmnd := exec.Command("../xkcd.exe", "")

	err := cmnd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service started")

	cfg, err := config.Load("./config.yaml")

	if err != nil {
		log.Fatal(err)
	}

	type creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	requestBody := creds{
		Login: "admin",
		Password: "admin",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/login", cfg.SrvPort), bytes.NewReader(jsonBody))
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	type HTTPResponse struct {
		Data map[string]string
	}

	var response HTTPResponse

	body, err := io.ReadAll(res.Body)
	json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal("test", err)
	}
	
	token := response.Data["token"]

	var wg sync.WaitGroup

	wg.Add(1)

	go func () {
		defer wg.Done()

		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/update", cfg.SrvPort), nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("token", token)
		log.Println("Request created", token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		log.Println("Response received")

		var updateResp any
		body, err := io.ReadAll(res.Body)
		json.Unmarshal(body, &updateResp)
		if err != nil {
			log.Fatal("test", err)
		}
		log.Println("Response unmarshalled:", updateResp)
	} ()

	log.Println("updating...")
	wg.Wait()
	log.Println("done updating!")


	getReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/pics?search=apple,doctor", cfg.SrvPort), nil)
	if err != nil {
		log.Fatal(err)
	}

	getReq.Header.Add("token", token)
	log.Println("Request created", token)

	res, err = http.DefaultClient.Do(getReq)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	log.Println("Response received")

	type HTTPResponseQuery struct {
		Data []string
	}

	var updateResp HTTPResponseQuery

	body, err = io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println("Response body:", string(body))
	json.Unmarshal(body, &updateResp)
	if err != nil {
		log.Fatal("test", err)
	}
	log.Println("Response unmarshalled:", updateResp.Data)
	if slices.Contains(updateResp.Data, "https://xkcd.com/2161") {
		log.Println("comic found!")
	} else {
		log.Println("comic not found...")
	}
}