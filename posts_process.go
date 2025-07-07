package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	FeaturedImage      string   `yaml:"featured_image"`
	CrosspostInstagram bool     `yaml:"crosspost_instagram"`
	CrosspostThreads   bool     `yaml:"crosspost_threads"`
	Summary            string   `yaml:"summary"`
}

type Chapter struct {
	Title    string
	Level    int
	Content  string
	Order    int
	ParentID string
}

func (app *App) processPost(re *core.RequestEvent, isUpdate bool, postID string) error {
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
	app.pb.Logger().Info(content)

	frontmatter, markdownContent, err := app.parseFrontmatter(content)
	if err != nil {
		return re.BadRequestError("Failed to parse frontmatter", err)
	}

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

	var slug string
	if !isUpdate || post.GetString("slug") == "" {
		slug = app.generateUniqueSlug(frontmatter.Title, "posts")
		post.Set("slug", slug)
	} else {
		slug = post.GetString("slug")
	}

	if appURL := os.Getenv("APP_URL"); appURL != "" {
		permalink := fmt.Sprintf("%s/%s", strings.TrimRight(appURL, "/"), slug)
		post.Set("permalink", permalink)
	}

	post.Set("title", frontmatter.Title)
	post.Set("subtitle", frontmatter.Subtitle)
	post.Set("content", markdownContent)
	post.Set("type", "Blog")
	post.Set("is_visible", frontmatter.IsVisible)
	post.Set("summary", frontmatter.Summary)

	if frontmatter.FeaturedImage != "" && len(frontmatter.FeaturedImage) == 15 && !strings.Contains(frontmatter.FeaturedImage, "/") {
		post.Set("featured_image", frontmatter.FeaturedImage)
	}

	if err := app.pb.Save(post); err != nil {
		return re.BadRequestError("Failed to save post", err)
	}

	// process tags
	if err := app.processTags(post, frontmatter.Tags); err != nil {
		log.Printf("Error processing tags: %v", err)
	}

	// process contexts
	if err := app.processContexts(post, frontmatter.Contexts); err != nil {
		log.Printf("Error processing contexts: %v", err)
	}

	// process collections
	if err := app.processCollections(post, frontmatter.Collections); err != nil {
		app.pb.Logger().Error("Error processing collections: %v", err)
	}

	// Process chapters
	if err := app.processChapters(post, markdownContent); err != nil {
		app.pb.Logger().Error("Error processing chapters", err)
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

func (app *App) parseFrontmatter(content string) (*PostFrontmatter, string, error) {
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
	app.pb.Logger().Info(frontmatterContent)
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

	// delete existing connections
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

		// find or create the collection by title
		collection, err := app.pb.FindFirstRecordByFilter("collections", "title = {:title}", map[string]any{"title": collectionName})
		if err != nil {
			// collection doesn't exist, create it
			log.Printf("Collection '%s' not found, creating it", collectionName)

			collectionsCollection, err := app.pb.FindCollectionByNameOrId("collections")
			if err != nil {
				log.Printf("Collections collection not found: %v", err)
				continue
			}

			collection = core.NewRecord(collectionsCollection)
			collection.Set("title", collectionName)
			collection.Set("slug", app.generateUniqueSlug(collectionName, "collections"))
			collection.Set("description", "")

			if err := app.pb.Save(collection); err != nil {
				log.Printf("Failed to create collection %s: %v", collectionName, err)
				continue
			}
			log.Printf("Successfully created collection: %s", collectionName)
		}

		// create collection_posts junction record
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
		} else {
			log.Printf("Successfully linked post to collection: %s", collectionName)
		}
	}

	return nil
}

func (app *App) processChapters(post *core.Record, markdownContent string) error {
	//  delete existing chapters for this post
	if post.Id != "" {
		existingChapters, _ := app.pb.FindRecordsByFilter("post_chapters", "post = {:postId}", "-created", 0, 0, map[string]any{"postId": post.Id})
		for _, chapter := range existingChapters {
			app.pb.Delete(chapter)
		}
	}

	chapters := app.parseChapters(markdownContent)
	if len(chapters) == 0 {
		log.Printf("No chapters found in post content")
		return nil
	}

	chapterRecords := make(map[string]*core.Record)

	collection, err := app.pb.FindCollectionByNameOrId("post_chapters")
	if err != nil {
		return fmt.Errorf("post_chapters collection not found: %v", err)
	}

	for i, chapter := range chapters {
		app.pb.Logger().Info(chapter.Title)
		chapterRecord := core.NewRecord(collection)
		chapterSlug := app.generateUniqueSlug(chapter.Title, "post_chapters")

		chapterRecord.Set("post", post.Id)
		chapterRecord.Set("title", chapter.Title)
		chapterRecord.Set("slug", chapterSlug)

		// set chapter permalink - just the fragment since it's within a post
		chapterRecord.Set("permalink", fmt.Sprintf("#%s", chapterSlug))
		chapterRecord.Set("order", i)

		// handle parent chapter relationships
		if chapter.Level > 1 && chapter.ParentID != "" {
			if parentRecord, exists := chapterRecords[chapter.ParentID]; exists {
				chapterRecord.Set("parent_chapter", parentRecord.Id)
				log.Printf("Setting parent for chapter '%s' to '%s'", chapter.Title, parentRecord.GetString("title"))
			}
		}

		if err := app.pb.Save(chapterRecord); err != nil {
			app.pb.Logger().Error("Failed to create chapter", chapter.Title, err)
			continue
		}

		chapterRecords[chapterSlug] = chapterRecord
		log.Printf("Successfully created chapter: %s (level %d)", chapter.Title, chapter.Level)
	}

	log.Printf("Successfully processed %d chapters", len(chapterRecords))
	return nil
}

func (app *App) parseChapters(content string) []Chapter {
	var chapters []Chapter
	lines := strings.Split(content, "\n")

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	var parentStack []string

	for _, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)

		if len(matches) == 3 {
			level := len(matches[1])
			title := strings.TrimSpace(matches[2])

			if title == "" {
				continue
			}

			if level-1 > len(parentStack) {
				for len(parentStack) < level-1 {
					parentStack = append(parentStack, "")
				}
			} else if level-1 < len(parentStack) {
				parentStack = parentStack[:level-1]
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

			currentSlug := app.generateSlugBase(title)
			if level-1 == len(parentStack) {
				parentStack = append(parentStack, currentSlug)
			} else {
				parentStack[level-1] = currentSlug
			}
		}
	}

	return chapters
}

func (app *App) processCrosspostQueue(post *core.Record, frontmatter *PostFrontmatter, queueType string) error {
	if !frontmatter.CrosspostInstagram && !frontmatter.CrosspostThreads {
		return nil
	}

	if frontmatter.CrosspostInstagram {
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

	return nil
}

func (app *App) generateUniqueSlug(title, collectionName string) string {
	baseSlug := app.generateSlugBase(title)

	if !app.slugExists(baseSlug, collectionName) {
		return baseSlug
	}

	// if it exists, try with a counter
	counter := 1
	for {
		candidateSlug := fmt.Sprintf("%s-%d", baseSlug, counter)
		if !app.slugExists(candidateSlug, collectionName) {
			return candidateSlug
		}
		counter++

		if counter > 5 {
			// fall back to timestamp
			timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
			return fmt.Sprintf("%s-%s", baseSlug, timestamp)
		}
	}
}

func (app *App) generateSlugBase(title string) string {

	slug := strings.ToLower(title)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	return slug
}

func (app *App) slugExists(slug, collectionName string) bool {
	var filter string
	var params map[string]any

	if collectionName == "posts" {
		// for posts, also check permalink conflicts if app_url is set
		if appURL := os.Getenv("APP_URL"); appURL != "" {
			expectedPermalink := fmt.Sprintf("%s/%s", strings.TrimRight(appURL, "/"), slug)
			filter = "slug = {:slug} || permalink = {:permalink}"
			params = map[string]any{
				"slug":      slug,
				"permalink": expectedPermalink,
			}
		} else {
			filter = "slug = {:slug}"
			params = map[string]any{"slug": slug}
		}
	} else {
		filter = "slug = {:slug}"
		params = map[string]any{"slug": slug}
	}

	record, err := app.pb.FindFirstRecordByFilter(collectionName, filter, params)
	return err == nil && record != nil
}
