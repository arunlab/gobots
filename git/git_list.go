package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Please set the GITHUB_TOKEN environment variable.")
		return
	}

	url := "https://api.github.com/user/repos"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		return
	}

	var repos []map[string]interface{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		fmt.Printf("Error decoding JSON response: %s\n", err)
		return
	}

	for _, repo := range repos {
		fmt.Println(repo["full_name"])
	}
}
