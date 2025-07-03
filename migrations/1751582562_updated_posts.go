package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1125843985")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_lBuOjNGbrm` + "`" + ` ON ` + "`" + `posts` + "`" + ` (\n  ` + "`" + `slug` + "`" + `,\n  ` + "`" + `permalink` + "`" + `\n)",
				"CREATE INDEX ` + "`" + `idx_RQpao89B94` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `title` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_FLcgWqytYx` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `tags` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1125843985")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_lBuOjNGbrm` + "`" + ` ON ` + "`" + `posts` + "`" + ` (\n  ` + "`" + `slug` + "`" + `,\n  ` + "`" + `permalink` + "`" + `\n)",
				"CREATE INDEX ` + "`" + `idx_RQpao89B94` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `title` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
