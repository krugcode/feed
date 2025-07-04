package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorBlue   = "\033[34m"
	ColorYellow = "\033[33m"
	ColorPurple = "\033[35m"
)

type PostResponse struct {
	Post struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Created string `json:"created"`
	} `json:"post"`
	Message string `json:"message"`
}

var (
	appURL = "http://localhost:8090" // default dev fb
	token  = ""
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if token == "" {
		fmt.Printf("%sError: SUPERUSER_TOKEN not set at build time%s\n", ColorRed, ColorReset)
		fmt.Printf("%sUse: make build-cli and ensure .env file exists%s\n", ColorYellow, ColorReset)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "post":
		if len(os.Args) < 3 {
			fmt.Printf("%sError: Please specify a markdown file%s\n", ColorRed, ColorReset)
			printUsage()
			os.Exit(1)
		}
		err := postMarkdown(appURL, token, os.Args[2])
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}

	case "update":
		if len(os.Args) < 4 {
			fmt.Printf("%sError: Please specify post ID and markdown file%s\n", ColorRed, ColorReset)
			printUsage()
			os.Exit(1)
		}
		err := updateMarkdown(appURL, token, os.Args[2], os.Args[3])
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}

	case "version", "--version", "-v":
		fmt.Printf("%sfeed CLI v1.0.0%s\n", ColorPurple, ColorReset)
		fmt.Printf("%sPart of your PocketBase feed application%s\n", ColorBlue, ColorReset)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("%sError: Unknown command '%s'%s\n", ColorRed, os.Args[1], ColorReset)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("%sfeed%s - Your markdown blogging companion\n", ColorPurple, ColorReset)
	fmt.Println("")
	fmt.Println("Usage: feed <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Printf("  %spost%s <file.md>           Create a new post from markdown file\n", ColorBlue, ColorReset)
	fmt.Printf("  %supdate%s <post_id> <file>  Update existing post\n", ColorBlue, ColorReset)
	fmt.Printf("  %sversion%s                 Show version information\n", ColorBlue, ColorReset)
	fmt.Printf("  %shelp%s                     Show this help\n", ColorBlue, ColorReset)
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Printf("  %sfeed post quick-blog.md%s\n", ColorGreen, ColorReset)
	fmt.Printf("  %sfeed update abc123 updated-blog.md%s\n", ColorGreen, ColorReset)
	fmt.Println("")
	fmt.Printf("%sConfig is embedded at build time from .env file%s\n", ColorYellow, ColorReset)
}

func postMarkdown(appURL, token, filename string) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' not found", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	fmt.Printf("%sPosting %s to your feed...%s\n", ColorBlue, filename, ColorReset)

	url := appURL + "/markdown/posts"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	var postResp PostResponse
	err = json.Unmarshal(body, &postResp)
	if err != nil {
		// if we can't parse the response, just show success
		fmt.Printf("%sPost created successfully!%s\n", ColorGreen, ColorReset)
		fmt.Printf("%sServer response: %s%s\n", ColorBlue, string(body), ColorReset)
		return nil
	}

	fmt.Printf("%sPost created successfully!%s\n", ColorGreen, ColorReset)
	fmt.Printf("%sTitle: %s%s\n", ColorGreen, postResp.Post.Title, ColorReset)
	fmt.Printf("%sID: %s%s\n", ColorGreen, postResp.Post.ID, ColorReset)
	fmt.Printf("%sAdmin: %s/_/#/collections/posts/records/%s%s\n",
		ColorBlue, appURL, postResp.Post.ID, ColorReset)

	return nil
}

func updateMarkdown(appURL, token, postID, filename string) error {
	// check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' not found", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	fmt.Printf("%sUpdating post %s with %s...%s\n", ColorBlue, postID, filename, ColorReset)

	url := appURL + "/markdown/posts/" + postID
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	var postResp PostResponse
	err = json.Unmarshal(body, &postResp)
	if err != nil {
		fmt.Printf("%sPost updated successfully!%s\n", ColorGreen, ColorReset)
		fmt.Printf("%sServer response: %s%s\n", ColorBlue, string(body), ColorReset)
		return nil
	}

	fmt.Printf("%sPost updated successfully!%s\n", ColorGreen, ColorReset)
	fmt.Printf("%sTitle: %s%s\n", ColorGreen, postResp.Post.Title, ColorReset)
	fmt.Printf("%sID: %s%s\n", ColorGreen, postResp.Post.ID, ColorReset)
	fmt.Printf("%sAdmin: %s/_/#/collections/posts/records/%s%s\n",
		ColorBlue, appURL, postResp.Post.ID, ColorReset)

	return nil
}
