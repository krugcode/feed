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
				"CREATE INDEX ` + "`" + `idx_RQpao89B94` + "`" + ` ON ` + "`" + `posts` + "`" + ` (` + "`" + `title` + "`" + `)"
			]
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(8, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_3446931122",
			"hidden": false,
			"id": "relation4036791755",
			"maxSelect": 999,
			"minSelect": 0,
			"name": "featured_images",
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
			"indexes": []
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation4036791755")

		return app.Save(collection)
	})
}
