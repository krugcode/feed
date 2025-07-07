package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
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
		ID        string `json:"id"`
		Title     string `json:"title"`
		Slug      string `json:"slug"`
		Permalink string `json:"permalink"`
		Created   string `json:"created"`
	} `json:"post"`
	Message string `json:"message"`
}

type UploadResponse struct {
	ID          string `json:"id"`
	File        string `json:"file"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Frontmatter struct for parsing YAML
type Frontmatter struct {
	Title              string   `yaml:"title"`
	Subtitle           string   `yaml:"subtitle"`
	Tags               []string `yaml:"tags"`
	Contexts           []string `yaml:"contexts"`
	Collections        []string `yaml:"collections"`
	IsVisible          bool     `yaml:"is_visible"`
	FeaturedImage      string   `yaml:"featured_image"`
	CrosspostInstagram bool     `yaml:"crosspost_instagram"`
	CrosspostThreads   bool     `yaml:"crosspost_threads"`
	Summary            string   `yaml:"summary"`
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
		fmt.Printf("%sPart of the PocketBase feed application%s\n", ColorBlue, ColorReset)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("%sError: Unknown command '%s'%s\n", ColorRed, os.Args[1], ColorReset)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("%sfeed%s - a markdown blogging companion\n", ColorPurple, ColorReset)
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

func parseFrontmatter(content string) (*Frontmatter, string, error) {
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return nil, content, nil
	}

	var frontmatterLines []string
	var contentStart int

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			contentStart = i + 1
			break
		}
		frontmatterLines = append(frontmatterLines, lines[i])
	}

	if contentStart == 0 {
		return nil, "", fmt.Errorf("frontmatter not properly closed")
	}

	frontmatterContent := strings.Join(frontmatterLines, "\n")
	markdownContent := strings.Join(lines[contentStart:], "\n")

	var frontmatter Frontmatter
	if err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter); err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	return &frontmatter, markdownContent, nil
}

func reconstructContent(frontmatter *Frontmatter, markdownContent string) (string, error) {
	frontmatterBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %v", err)
	}

	return fmt.Sprintf("---\n%s---\n%s", string(frontmatterBytes), markdownContent), nil
}

func processFeaturedImage(frontmatter *Frontmatter, appURL, token string) error {
	if frontmatter.FeaturedImage == "" {
		return nil
	}

	if strings.Contains(frontmatter.FeaturedImage, "/api/files/") || (len(frontmatter.FeaturedImage) == 15 && !strings.Contains(frontmatter.FeaturedImage, "/")) {
		fmt.Printf("%sSkipping featured image (already processed)%s\n", ColorYellow, ColorReset)
		return nil
	}

	fmt.Printf("%sProcessing featured image: %s%s\n", ColorBlue, frontmatter.FeaturedImage, ColorReset)

	// Upload the featured image
	uploadResp, err := uploadAsset(frontmatter.FeaturedImage, "Featured image", "", appURL, token)
	if err != nil {
		return fmt.Errorf("failed to upload featured image: %v", err)
	}

	// Replace with upload ID (you can also use the full PocketBase URL if preferred)
	frontmatter.FeaturedImage = uploadResp.ID
	fmt.Printf("%sFeatured image uploaded: %s%s\n", ColorGreen, uploadResp.ID, ColorReset)

	return nil
}

func processAssets(content string, appURL, token string) (string, error) {
	// regex to match markdown image syntax: ![alt](url "title")
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+?)(?:\s+"([^"]*)")?\)`)

	updatedContent := content
	matches := imageRegex.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		fmt.Printf("%sNo images found to process%s\n", ColorBlue, ColorReset)
		return content, nil
	}

	fmt.Printf("%sFound %d images to process...%s\n", ColorBlue, len(matches), ColorReset)

	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		altText := match[1]
		originalURL := strings.TrimSpace(match[2])
		originalURL = strings.Trim(originalURL, `"'`) // Remove quotes
		title := ""
		if len(match) > 3 && match[3] != "" {
			title = match[3]
		}

		// skip if already a pocketbase url
		if strings.Contains(originalURL, "/api/files/") {
			fmt.Printf("%s  [%d/%d] Skipping PocketBase URL%s\n", ColorYellow, i+1, len(matches), ColorReset)
			continue
		}

		// skip relative paths that aren't local files
		if strings.HasPrefix(originalURL, "/") && !strings.HasPrefix(originalURL, "/home/") && !strings.HasPrefix(originalURL, "/Users/") && !strings.HasPrefix(originalURL, "/tmp/") {
			fmt.Printf("%s  [%d/%d] Skipping relative path: %s%s\n", ColorYellow, i+1, len(matches), originalURL, ColorReset)
			continue
		}

		fmt.Printf("%s  [%d/%d] Processing: %s%s\n", ColorBlue, i+1, len(matches), originalURL, ColorReset)

		// upload the asset
		uploadResp, err := uploadAsset(originalURL, altText, title, appURL, token)
		if err != nil {
			fmt.Printf("%s    Failed: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

		// generate pocketbase url
		pbURL := fmt.Sprintf("/api/files/uploads/%s/%s", uploadResp.ID, uploadResp.File)

		// replace in content
		originalMarkdown := match[0]
		var newMarkdown string
		if title != "" {
			newMarkdown = fmt.Sprintf(`![%s](%s "%s")`, altText, pbURL, title)
		} else {
			newMarkdown = fmt.Sprintf(`![%s](%s)`, altText, pbURL)
		}

		updatedContent = strings.Replace(updatedContent, originalMarkdown, newMarkdown, 1)
		fmt.Printf("%s    Uploaded: %s%s\n", ColorGreen, pbURL, ColorReset)
	}

	return updatedContent, nil
}

func uploadAsset(assetURL, altText, title, appURL, token string) (*UploadResponse, error) {
	var fileData []byte
	var filename string
	var err error

	// determine if it's a local file or remote url
	if strings.HasPrefix(assetURL, "http://") || strings.HasPrefix(assetURL, "https://") {
		// remote file
		fileData, filename, err = downloadRemoteFile(assetURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download remote file: %v", err)
		}
	} else {
		// local file
		fileData, err = os.ReadFile(assetURL)
		if err != nil {
			return nil, fmt.Errorf("failed to read local file: %v", err)
		}
		filename = filepath.Base(assetURL)
	}

	// set description from alt text and title
	description := altText
	if title != "" && title != altText {
		if altText != "" {
			description = fmt.Sprintf("%s - %s", altText, title)
		} else {
			description = title
		}
	}
	if description == "" {
		description = "Featured image"
	}

	// determine file type based on extension
	fileType := "Image"
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		fileType = "Video"
	case ".md", ".txt":
		fileType = "Markdown"
	}

	// Create multipart form
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Add form fields
	writer.WriteField("description", description)
	writer.WriteField("type", fileType)

	// Add URL fields if it's a remote file
	if strings.HasPrefix(assetURL, "http") {
		writer.WriteField("url", assetURL)
		writer.WriteField("credit_source_url", assetURL)
	}

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := fileWriter.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %v", err)
	}

	writer.Close()

	// Create request
	url := appURL + "/api/collections/uploads/records"
	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(body))
	}

	var uploadResp UploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %v", err)
	}

	return &uploadResp, nil
}

func downloadRemoteFile(url string) ([]byte, string, error) {
	// Create request with proper headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	// Add common headers that many sites expect
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PocketBase-Feed/1.0)")
	req.Header.Set("Accept", "image/*,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Extract filename from URL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// If no extension, try to guess from content-type
	if filepath.Ext(filename) == "" || len(filename) < 3 {
		contentType := resp.Header.Get("Content-Type")
		baseFilename := "download"

		switch {
		case strings.Contains(contentType, "image/jpeg"):
			filename = baseFilename + ".jpg"
		case strings.Contains(contentType, "image/png"):
			filename = baseFilename + ".png"
		case strings.Contains(contentType, "image/gif"):
			filename = baseFilename + ".gif"
		case strings.Contains(contentType, "image/webp"):
			filename = baseFilename + ".webp"
		default:
			filename = baseFilename + ".jpg" // Default to jpg
		}
	}

	// Clean filename
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(filename, "")

	return data, filename, nil
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

	content := string(data)

	// Parse frontmatter
	frontmatter, markdownContent, err := parseFrontmatter(content)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	// Process featured image if present
	if frontmatter != nil {
		if err := processFeaturedImage(frontmatter, appURL, token); err != nil {
			return fmt.Errorf("failed to process featured image: %v", err)
		}
	}

	// Process assets in markdown content
	processedMarkdown, err := processAssets(markdownContent, appURL, token)
	if err != nil {
		return fmt.Errorf("failed to process assets: %v", err)
	}

	// Reconstruct content with updated frontmatter
	var processedContent string
	if frontmatter != nil {
		processedContent, err = reconstructContent(frontmatter, processedMarkdown)
		if err != nil {
			return fmt.Errorf("failed to reconstruct content: %v", err)
		}
	} else {
		processedContent = processedMarkdown
	}

	url := appURL + "/api/markdown/posts"
	req, err := http.NewRequest("POST", url, strings.NewReader(processedContent))
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

	// Show permalink if available
	if postResp.Post.Permalink != "" {
		fmt.Printf("%sURL: %s%s\n", ColorPurple, postResp.Post.Permalink, ColorReset)
	}

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

	content := string(data)

	// Parse frontmatter
	frontmatter, markdownContent, err := parseFrontmatter(content)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	// Process featured image if present
	if frontmatter != nil {
		if err := processFeaturedImage(frontmatter, appURL, token); err != nil {
			return fmt.Errorf("failed to process featured image: %v", err)
		}
	}

	// Process assets in markdown content
	processedMarkdown, err := processAssets(markdownContent, appURL, token)
	if err != nil {
		return fmt.Errorf("failed to process assets: %v", err)
	}

	// Reconstruct content with updated frontmatter
	var processedContent string
	if frontmatter != nil {
		processedContent, err = reconstructContent(frontmatter, processedMarkdown)
		if err != nil {
			return fmt.Errorf("failed to reconstruct content: %v", err)
		}
	} else {
		processedContent = processedMarkdown
	}

	url := appURL + "/api/markdown/posts/" + postID
	req, err := http.NewRequest("PUT", url, strings.NewReader(processedContent))
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

	// Show permalink if available
	if postResp.Post.Permalink != "" {
		fmt.Printf("%sURL: %s%s\n", ColorPurple, postResp.Post.Permalink, ColorReset)
	}

	fmt.Printf("%sAdmin: %s/_/#/collections/posts/records/%s%s\n",
		ColorBlue, appURL, postResp.Post.ID, ColorReset)

	return nil
}
