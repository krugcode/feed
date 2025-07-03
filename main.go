package main

import (
	"feed/views"
	"log"
	"os"

	_ "feed/migrations"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

type App struct {
	pb *pocketbase.PocketBase
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
		godotenv.Load(".env")
	}
	pb := pocketbase.New()

	app := &App{pb: pb}

	autoMigrate := os.Getenv("AUTO_MIGRATE") != "false"
	migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{
		Automigrate: autoMigrate,
	})

	pb.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// custom routes (for views)
		app.setupRoutes(se)
		return se.Next()
	})
	app.setupHooks()
	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}
}

func (app *App) setupRoutes(se *core.ServeEvent) {
	se.Router.GET("/", app.homePage)
	se.Router.GET("/links", app.linksPage)
	se.Router.GET("/collections", app.collectionsPage)
	se.Router.GET("/about", app.aboutPage)

	// example api usage for later
	// se.Router.GET("/api/posts", app.apiGetPosts)
	// se.Router.GET("/api/posts/{slug}", app.apiGetPostBySlug)
}

func (app *App) setupHooks() {
	// example: validate posts before creation
	// app.pb.OnRecordCreateRequest("posts").BindFunc(func(re *core.RecordRequestEvent) error {
	// 	// Custom validation logic here
	// 	log.Printf("Creating new post: %s", re.Record.GetString("title"))
	// 	return re.Next()
	// })
	//
	// // example: update click count for collections
	// app.pb.OnRecordViewRequest("collections").BindFunc(func(re *core.RecordViewRequestEvent) error {
	// 	// Increment clicked_count when collection is viewed
	// 	currentCount := re.Record.GetInt("clicked_count")
	// 	re.Record.Set("clicked_count", currentCount+1)
	// 	app.pb.SaveRecord(re.Record)
	// 	return re.Next()
	// })
}

func (app *App) homePage(re *core.RequestEvent) error {
	posts, err := app.pb.FindRecordsByFilter(
		"posts",
		"visible = true",
		"-created",
		10,
		0,
	)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
	}

	component := views.FeedPage(posts)
	return component.Render(re.Request.Context(), re.Response)
}

func (app *App) linksPage(re *core.RequestEvent) error {
	links, err := app.pb.FindRecordsByFilter(
		"links",
		"is_visible = true",
		"order",
		50,
		0)
	if err != nil {
		log.Printf("Error fetching links: %v", err)
	}

	component := views.LinksPage(links)
	return component.Render(re.Request.Context(), re.Response)
}

func (app *App) collectionsPage(re *core.RequestEvent) error {
	collections, err := app.pb.FindRecordsByFilter(
		"collections",
		"",         // no filter
		"-created", // sort by created desc
		20,         // limit
		0,          // offset

	)
	if err != nil {
		log.Printf("Error fetching collections: %v", err)
	}

	component := views.CollectionsPage(collections)
	return component.Render(re.Request.Context(), re.Response)
}

func (app *App) aboutPage(re *core.RequestEvent) error {
	component := views.AboutPage()
	return component.Render(re.Request.Context(), re.Response)
}

// example of some api routes
// func (app *App) apiGetPosts(re *core.RequestEvent) error {
// 	posts, err := app.pb.FindRecordsByFilter(
// 		"posts",
// 		"visible = true",
// 		"-created",
// 		20,
// 		0,
// 		"tags,contexts", // Expand relations
// 	)
// 	if err != nil {
// 		return re.BadRequestError("Failed to fetch posts", err)
// 	}
//
// 	return re.JSON(200, map[string]any{
// 		"posts": posts,
// 	})
// }
//
// func (app *App) apiGetPostBySlug(re *core.RequestEvent) error {
// 	slug := re.Request().PathValue("slug")
//
// 	post, err := app.pb.FindFirstRecordByFilter(
// 		"posts",
// 		"slug = {:slug} && visible = true",
// 		map[string]any{"slug": slug},
// 		"tags,contexts", // Expand relations
// 	)
// 	if err != nil {
// 		return re.NotFoundError("Post not found", err)
// 	}
//
// 	return re.JSON(200, map[string]any{
// 		"post": post,
// 	})
// }
