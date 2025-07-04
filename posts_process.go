package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"gopkg.in/yaml.v2"
)

type PostFrontmatter struct {
	Title              string   `yaml:"title"`
	Subtitle           string   `yaml:"subtitle"`
	Tags               []string `yaml:"tags"`
	Contexts           []string `yaml:"contexts"`
	Collections        []string `yaml:"collections"`
	IsVisible          bool     `yaml:"is_visible"`
	CrosspostInstagram bool     `yaml:"crosspost_instagram"`
	CrosspostThreads   bool     `yaml:"crosspost_threads"`
}

type Chapter struct {
	Title    string
	Level    int
	Content  string
	Order    int
	ParentID string
}

func (app *App) processPost(re *core.RequestEvent, isUpdate bool, postID string) error {
	// Read the markdown content from request body
	body := re.Request.Body
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return re.BadRequestError("Failed to read request body", err)
	}

	content := strings.Join(lines, "\n")

	// Parse frontmatter and content
	frontmatter, markdownContent, err := parseFrontmatter(content)
	if err != nil {
		return re.BadRequestError("Failed to parse frontmatter", err)
	}

	// Create or update the post
	var post *core.Record
	if isUpdate {
		post, err = app.pb.FindRecordById("posts", postID)
		if err != nil {
			return re.NotFoundError("Post not found", err)
		}
	} else {
		collection, err := app.pb.FindCollectionByNameOrId("posts")
		if err != nil {
			return re.BadRequestError("Posts collection not found", err)
		}
		post = core.NewRecord(collection)
	}

	// Generate slug from title (only for new posts or if slug is empty)
	if !isUpdate || post.GetString("slug") == "" {
		slug := generateSlug(frontmatter.Title)
		post.Set("slug", slug)
	}

	// Process assets (images/files) in markdown content and get updated content
	updatedMarkdownContent, uploadedAssets, err := app.processMarkdownAssets(markdownContent)
	if err != nil {
		log.Printf("Error processing markdown assets: %v", err)
		updatedMarkdownContent = markdownContent // Use original content if processing fails
	}

	// Set post fields
	post.Set("title", frontmatter.Title)
	post.Set("subtitle", frontmatter.Subtitle)
	post.Set("content", updatedMarkdownContent)
	post.Set("type", "Blog")
	post.Set("is_visible", frontmatter.IsVisible)

	// Save the post first to get an ID
	if err := app.pb.Save(post); err != nil {
		return re.BadRequestError("Failed to save post", err)
	}

	// Process tags
	if err := app.processTags(post, frontmatter.Tags); err != nil {
		log.Printf("Error processing tags: %v", err)
	}

	// Process contexts
	if err := app.processContexts(post, frontmatter.Contexts); err != nil {
		log.Printf("Error processing contexts: %v", err)
	}

	// Process collections
	if err := app.processCollections(post, frontmatter.Collections); err != nil {
		log.Printf("Error processing collections: %v", err)
	}

	// Process chapters (use updated content)
	if err := app.processChapters(post, updatedMarkdownContent); err != nil {
		log.Printf("Error processing chapters: %v", err)
	}

	// Link uploaded assets to the post
	if len(uploadedAssets) > 0 {
		var uploadIDs []string
		for _, upload := range uploadedAssets {
			uploadIDs = append(uploadIDs, upload.Id)
		}
		post.Set("uploads", uploadIDs)
		// Save the post again to update the uploads
		if err := app.pb.Save(post); err != nil {
			log.Printf("Failed to update post uploads: %v", err)
		}
	}

	// Process crosspost queue
	queueType := "Create"
	if isUpdate {
		queueType = "Update"
	}
	if err := app.processCrosspostQueue(post, frontmatter, queueType); err != nil {
		log.Printf("Error processing crosspost queue: %v", err)
	}

	return re.JSON(200, map[string]any{
		"post":    post,
		"message": "Post processed successfully",
	})
}

func parseFrontmatter(content string) (*PostFrontmatter, string, error) {
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return nil, "", fmt.Errorf("invalid frontmatter format")
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

	var frontmatter PostFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterContent), &frontmatter); err != nil {
		return nil, "", err
	}

	return &frontmatter, markdownContent, nil
}

func (app *App) processTags(post *core.Record, tagNames []string) error {
	if len(tagNames) == 0 {
		return nil
	}

	var tagIDs []string

	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		// Try to find existing tag
		tag, err := app.pb.FindFirstRecordByFilter("tags", "title = {:title}", map[string]any{"title": tagName})

		if err != nil {
			// Tag doesn't exist, create it
			collection, err := app.pb.FindCollectionByNameOrId("tags")
			if err != nil {
				log.Printf("Tags collection not found: %v", err)
				continue
			}

			tag = core.NewRecord(collection)
			tag.Set("title", tagName)
			tag.Set("search_count", 0)

			if err := app.pb.Save(tag); err != nil {
				log.Printf("Failed to create tag %s: %v", tagName, err)
				continue
			}
		}

		tagIDs = append(tagIDs, tag.Id)
	}

	if len(tagIDs) > 0 {
		post.Set("tags", tagIDs)
		// Save the post again to update the tags
		if err := app.pb.Save(post); err != nil {
			log.Printf("Failed to update post tags: %v", err)
		}
	}

	return nil
}

func (app *App) processMarkdownAssets(content string) (string, []*core.Record, error) {
	// Regex to match markdown image syntax: ![alt](url "title")
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)(?:\s+"([^"]*)")?\)`)

	var uploadedAssets []*core.Record
	updatedContent := content

	// Find all image matches
	matches := imageRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		altText := match[1]
		originalURL := strings.TrimSpace(match[2])
		title := ""
		if len(match) > 3 {
			title = match[3]
		}

		// Skip if URL is already a PocketBase URL or relative path
		if strings.Contains(originalURL, "/api/files/") || strings.HasPrefix(originalURL, "/") && !strings.HasPrefix(originalURL, "/home/") {
			continue
		}

		// Process the asset
		uploadRecord, err := app.processAsset(originalURL, altText, title)
		if err != nil {
			log.Printf("Failed to process asset %s: %v", originalURL, err)
			continue
		}

		if uploadRecord != nil {
			uploadedAssets = append(uploadedAssets, uploadRecord)

			// Generate PocketBase public URL for the upload
			pbURL := fmt.Sprintf("/api/files/%s/%s/%s",
				uploadRecord.Collection().Name,
				uploadRecord.Id,
				uploadRecord.GetString("file"))

			// Replace the original URL with the PocketBase URL in the content
			originalMarkdown := match[0]
			var newMarkdown string
			if title != "" {
				newMarkdown = fmt.Sprintf(`![%s](%s "%s")`, altText, pbURL, title)
			} else {
				newMarkdown = fmt.Sprintf(`![%s](%s)`, altText, pbURL)
			}

			updatedContent = strings.Replace(updatedContent, originalMarkdown, newMarkdown, 1)
		}
	}

	return updatedContent, uploadedAssets, nil
}

func (app *App) processAsset(assetURL, altText, title string) (*core.Record, error) {
	collection, err := app.pb.FindCollectionByNameOrId("uploads")
	if err != nil {
		return nil, fmt.Errorf("uploads collection not found: %v", err)
	}

	uploadRecord := core.NewRecord(collection)

	var fileData []byte
	var filename string

	// Determine if it's a local file or remote URL
	if strings.HasPrefix(assetURL, "http://") || strings.HasPrefix(assetURL, "https://") {
		// Remote file
		fileData, filename, err = app.downloadRemoteFile(assetURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download remote file: %v", err)
		}
		uploadRecord.Set("url", assetURL)
		uploadRecord.Set("credit_source_url", assetURL)
	} else {
		// Local file
		fileData, err = os.ReadFile(assetURL)
		if err != nil {
			return nil, fmt.Errorf("failed to read local file: %v", err)
		}
		filename = filepath.Base(assetURL)
	}

	// Set description from alt text and title
	description := altText
	if title != "" && title != altText {
		if altText != "" {
			description = fmt.Sprintf("%s - %s", altText, title)
		} else {
			description = title
		}
	}
	uploadRecord.Set("description", description)

	// Determine file type based on extension
	fileType := "Image" // Default
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		fileType = "Video"
	case ".md", ".txt":
		fileType = "Markdown"
	}
	uploadRecord.Set("type", fileType)

	// Create a temporary file that will persist during the upload
	tempFile, err := os.CreateTemp("", "upload_*_"+filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up after we're done

	// Write data to temp file
	if _, err := tempFile.Write(fileData); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Use filesystem approach to set the file
	form := map[string]any{
		"description": description,
		"type":        fileType,
	}

	if uploadRecord.GetString("url") != "" {
		form["url"] = uploadRecord.GetString("url")
		form["credit_source_url"] = uploadRecord.GetString("credit_source_url")
	}

	// Load form data
	uploadRecord.Load(form)

	// Open the temp file for reading
	file, err := os.Open(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %v", err)
	}
	defer file.Close()

	// Set file field using the file reader
	uploadRecord.Set("file", file)

	// Save the upload record
	if err := app.pb.Save(uploadRecord); err != nil {
		return nil, fmt.Errorf("failed to save upload record: %v", err)
	}

	log.Printf("Successfully uploaded asset: %s -> %s", assetURL, filename)
	return uploadRecord, nil
}

func (app *App) downloadRemoteFile(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
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

	// Extract filename from URL or content-disposition header
	filename := filepath.Base(url)
	if contentDisp := resp.Header.Get("Content-Disposition"); contentDisp != "" {
		// Try to extract filename from Content-Disposition header
		if matches := regexp.MustCompile(`filename="([^"]+)"`).FindStringSubmatch(contentDisp); len(matches) > 1 {
			filename = matches[1]
		}
	}

	// If no extension, try to guess from content-type
	if filepath.Ext(filename) == "" {
		contentType := resp.Header.Get("Content-Type")
		switch {
		case strings.Contains(contentType, "image/jpeg"):
			filename += ".jpg"
		case strings.Contains(contentType, "image/png"):
			filename += ".png"
		case strings.Contains(contentType, "image/gif"):
			filename += ".gif"
		case strings.Contains(contentType, "image/webp"):
			filename += ".webp"
		case strings.Contains(contentType, "video/mp4"):
			filename += ".mp4"
		default:
			filename += ".bin"
		}
	}

	return data, filename, nil
}

func (app *App) processContexts(post *core.Record, contextNames []string) error {
	if len(contextNames) == 0 {
		return nil
	}

	// First, delete existing context_posts for this post
	if post.Id != "" {
		existingContextPosts, _ := app.pb.FindRecordsByFilter("context_posts", "post = {:postId}", "-created", 0, 0, map[string]any{"postId": post.Id})
		for _, cp := range existingContextPosts {
			app.pb.Delete(cp)
		}
	}

	for _, contextName := range contextNames {
		contextName = strings.TrimSpace(contextName)
		if contextName == "" {
			continue
		}

		// Find the context by title
		context, err := app.pb.FindFirstRecordByFilter("contexts", "title = {:title}", map[string]any{"title": contextName})
		if err != nil {
			log.Printf("Context not found: %s", contextName)
			continue
		}

		// Create context_posts junction record
		collection, err := app.pb.FindCollectionByNameOrId("context_posts")
		if err != nil {
			log.Printf("Context_posts collection not found: %v", err)
			continue
		}

		contextPost := core.NewRecord(collection)
		contextPost.Set("context", context.Id)
		contextPost.Set("post", post.Id)

		if err := app.pb.Save(contextPost); err != nil {
			log.Printf("Failed to create context_post for %s: %v", contextName, err)
		}
	}

	return nil
}

func (app *App) processCollections(post *core.Record, collectionNames []string) error {
	if len(collectionNames) == 0 {
		return nil
	}

	// First, delete existing collection_posts for this post
	if post.Id != "" {
		existingCollectionPosts, _ := app.pb.FindRecordsByFilter("collection_posts", "post = {:postId}", "-created", 0, 0, map[string]any{"postId": post.Id})
		for _, cp := range existingCollectionPosts {
			app.pb.Delete(cp)
		}
	}

	for order, collectionName := range collectionNames {
		collectionName = strings.TrimSpace(collectionName)
		if collectionName == "" {
			continue
		}

		// Find the collection by title
		collection, err := app.pb.FindFirstRecordByFilter("collections", "title = {:title}", map[string]any{"title": collectionName})
		if err != nil {
			log.Printf("Collection not found: %s", collectionName)
			continue
		}

		// Create collection_posts junction record
		collectionPostsCollection, err := app.pb.FindCollectionByNameOrId("collection_posts")
		if err != nil {
			log.Printf("Collection_posts collection not found: %v", err)
			continue
		}

		collectionPost := core.NewRecord(collectionPostsCollection)
		collectionPost.Set("collection", collection.Id)
		collectionPost.Set("post", post.Id)
		collectionPost.Set("order", order)

		if err := app.pb.Save(collectionPost); err != nil {
			log.Printf("Failed to create collection_post for %s: %v", collectionName, err)
		}
	}

	return nil
}

func (app *App) processChapters(post *core.Record, markdownContent string) error {
	// First, delete existing chapters for this post
	if post.Id != "" {
		existingChapters, _ := app.pb.FindRecordsByFilter("post_chapters", "post = {:postId}", "-created", 0, 0, map[string]any{"postId": post.Id})
		for _, chapter := range existingChapters {
			app.pb.Delete(chapter)
		}
	}

	chapters := parseChapters(markdownContent)
	chapterRecords := make(map[string]*core.Record) // slug -> record for parent lookups

	collection, err := app.pb.FindCollectionByNameOrId("post_chapters")
	if err != nil {
		return fmt.Errorf("post_chapters collection not found: %v", err)
	}

	for i, chapter := range chapters {
		chapterRecord := core.NewRecord(collection)
		chapterSlug := generateSlug(chapter.Title)

		chapterRecord.Set("post", post.Id)
		chapterRecord.Set("title", chapter.Title)
		chapterRecord.Set("slug", chapterSlug)
		chapterRecord.Set("permalink", fmt.Sprintf("#%s", chapterSlug))
		chapterRecord.Set("order", i)

		// Handle parent chapter relationships
		if chapter.Level > 1 && chapter.ParentID != "" {
			if parentRecord, exists := chapterRecords[chapter.ParentID]; exists {
				chapterRecord.Set("parent_chapter", parentRecord.Id)
			}
		}

		if err := app.pb.Save(chapterRecord); err != nil {
			log.Printf("Failed to create chapter %s: %v", chapter.Title, err)
			continue
		}

		chapterRecords[chapterSlug] = chapterRecord
	}

	return nil
}

func parseChapters(content string) []Chapter {
	var chapters []Chapter
	lines := strings.Split(content, "\n")

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	var parentStack []string // Keep track of parent slugs by level

	for _, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			level := len(matches[1])
			title := strings.TrimSpace(matches[2])

			if title == "" {
				continue
			}

			// Adjust parent stack for current level
			if level <= len(parentStack) {
				parentStack = parentStack[:level-1]
			}

			// Pad parent stack if needed
			for len(parentStack) < level-1 {
				parentStack = append(parentStack, "")
			}

			var parentID string
			if level > 1 && len(parentStack) > 0 {
				parentID = parentStack[len(parentStack)-1]
			}

			chapter := Chapter{
				Title:    title,
				Level:    level,
				Order:    len(chapters),
				ParentID: parentID,
			}

			chapters = append(chapters, chapter)

			// Add current chapter slug to parent stack
			currentSlug := generateSlug(title)
			if level <= len(parentStack) {
				parentStack = parentStack[:level-1]
			}
			parentStack = append(parentStack, currentSlug)
		}
	}

	return chapters
}

func (app *App) processCrosspostQueue(post *core.Record, frontmatter *PostFrontmatter, queueType string) error {
	if !frontmatter.CrosspostInstagram && !frontmatter.CrosspostThreads {
		return nil
	}

	if frontmatter.CrosspostInstagram {
		// Find an Instagram account (you might want to make this configurable)
		instagramAccounts, err := app.pb.FindRecordsByFilter("instagram_accounts", "", "-created", 1, 0)
		if err == nil && len(instagramAccounts) > 0 {
			collection, err := app.pb.FindCollectionByNameOrId("crosspost_queue")
			if err != nil {
				log.Printf("Crosspost_queue collection not found: %v", err)
				return nil
			}

			queueRecord := core.NewRecord(collection)
			queueRecord.Set("platform", "Instagram")
			queueRecord.Set("type", queueType)
			queueRecord.Set("post", post.Id)
			queueRecord.Set("status", "Queued")
			queueRecord.Set("instagram_account", instagramAccounts[0].Id)

			if err := app.pb.Save(queueRecord); err != nil {
				log.Printf("Failed to create Instagram crosspost queue: %v", err)
			}
		}
	}

	// Note: Threads posting would be similar, but I don't see a threads-specific table
	// You might want to add crosspost_threads logic here if you have that setup

	return nil
}

func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Add timestamp to ensure uniqueness
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	slug = fmt.Sprintf("%s-%s", slug, timestamp)

	return slug
}
