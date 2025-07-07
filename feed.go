package main

import (
	"feed/views"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func (app *App) homePage(re *core.RequestEvent) error {
	posts, err := app.pb.FindRecordsByFilter(
		"posts",
		"is_visible = true",
		"-created",
		10,
		0,
	)
	if err != nil {
		app.pb.Logger().Error("Error fetching posts", "error", err)
		return re.InternalServerError("Failed to load posts", err)
	}

	// Expand direct relations (tags, featured_image)
	errs := app.pb.ExpandRecords(posts, []string{"tags", "featured_image"}, nil)
	if len(errs) > 0 {
		app.pb.Logger().Error("Failed to expand direct relations", "errors", errs)
	}

	// Load junction table relations for each post
	for _, post := range posts {
		if err := app.loadPostRelations(post); err != nil {
			app.pb.Logger().Error("Failed to load post relations", "post_id", post.Id, "error", err)
		}
	}

	app.pb.Logger().Info("Posts loaded", "count", len(posts))
	component := views.FeedPage(posts)
	return component.Render(re.Request.Context(), re.Response)
}

func (app *App) loadPostRelations(post *core.Record) error {
	// Load contexts through context_posts junction table
	if err := app.loadPostContexts(post); err != nil {
		return fmt.Errorf("failed to load contexts: %v", err)
	}

	// Load collections through collection_posts junction table
	if err := app.loadPostCollections(post); err != nil {
		return fmt.Errorf("failed to load collections: %v", err)
	}

	// Load chapters through post_chapters table
	if err := app.loadPostChapters(post); err != nil {
		return fmt.Errorf("failed to load chapters: %v", err)
	}

	return nil
}

func (app *App) loadPostContexts(post *core.Record) error {
	// Find context_posts records for this post
	contextPosts, err := app.pb.FindRecordsByFilter(
		"context_posts",
		"post = {:postId}",
		"-created",
		0,
		0,
		map[string]any{"postId": post.Id},
	)
	if err != nil {
		return err
	}

	// Expand both the context AND the context's logo
	errs := app.pb.ExpandRecords(contextPosts, []string{"context", "context.logo"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand contexts: %v", errs)
	}

	// Extract the expanded context records
	var contexts []*core.Record
	for _, cp := range contextPosts {
		if context := cp.ExpandedOne("context"); context != nil {
			contexts = append(contexts, context)
		}
	}

	post.Set("expanded_contexts", contexts)
	return nil
}

func (app *App) loadPostCollections(post *core.Record) error {
	// Find collection_posts records for this post, ordered by their order field
	collectionPosts, err := app.pb.FindRecordsByFilter(
		"collection_posts",
		"post = {:postId}",
		"order", // Order by the order field
		0,
		0,
		map[string]any{"postId": post.Id},
	)
	if err != nil {
		return err
	}

	// Expand the collection relation in collection_posts
	errs := app.pb.ExpandRecords(collectionPosts, []string{"collection"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand collections: %v", errs)
	}

	// Extract the expanded collection records (preserving order)
	var collections []*core.Record
	for _, cp := range collectionPosts {
		if collection := cp.ExpandedOne("collection"); collection != nil {
			collections = append(collections, collection)
		}
	}

	// Store collections in a custom field
	post.Set("expanded_collections", collections)
	return nil
}

func (app *App) loadPostChapters(post *core.Record) error {
	// Find post_chapters records for this post, ordered by their order field
	chapters, err := app.pb.FindRecordsByFilter(
		"post_chapters",
		"post = {:postId}",
		"order", // Order by the order field
		0,
		0,
		map[string]any{"postId": post.Id},
	)
	if err != nil {
		return err
	}

	// Expand parent_chapter relations if they exist
	errs := app.pb.ExpandRecords(chapters, []string{"parent_chapter"}, nil)
	if len(errs) > 0 {
		app.pb.Logger().Error("Failed to expand parent chapters", "errors", errs)
		// Don't return error as parent chapters are optional
	}

	// Store chapters in a custom field
	post.Set("expanded_chapters", chapters)
	return nil
}

// Helper functions to access the expanded data in your templates

// GetPostTags returns the expanded tags for a post
func GetPostTags(post *core.Record) []*core.Record {
	return post.ExpandedAll("tags")
}

// GetPostFeaturedImage returns the expanded featured image for a post
func GetPostFeaturedImage(post *core.Record) *core.Record {
	return post.ExpandedOne("featured_image")
}

// GetPostContexts returns the contexts for a post
func GetPostContexts(post *core.Record) []*core.Record {
	if contexts := post.Get("expanded_contexts"); contexts != nil {
		if ctxSlice, ok := contexts.([]*core.Record); ok {
			return ctxSlice
		}
	}
	return nil
}

// GetPostCollections returns the collections for a post (in order)
func GetPostCollections(post *core.Record) []*core.Record {
	if collections := post.Get("expanded_collections"); collections != nil {
		if collSlice, ok := collections.([]*core.Record); ok {
			return collSlice
		}
	}
	return nil
}

// GetPostChapters returns the chapters for a post (in order)
func GetPostChapters(post *core.Record) []*core.Record {
	if chapters := post.Get("expanded_chapters"); chapters != nil {
		if chapSlice, ok := chapters.([]*core.Record); ok {
			return chapSlice
		}
	}
	return nil
}
