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
				"CREATE INDEX ` + "`" + `idx_FLcgWqytYx` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `tags` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_zZQC4x0xyG` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `context` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(10, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_3961493164",
			"hidden": false,
			"id": "relation3797779838",
			"maxSelect": 999,
			"minSelect": 0,
			"name": "context",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
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
				"CREATE INDEX ` + "`" + `idx_RQpao89B94` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `title` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_FLcgWqytYx` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `tags` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation3797779838")

		return app.Save(collection)
	})
}
