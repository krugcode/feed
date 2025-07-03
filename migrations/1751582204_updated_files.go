package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3446931122")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"name": "uploads"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"hidden": false,
			"id": "select2363381545",
			"maxSelect": 1,
			"name": "type",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"Markdown",
				"Image",
				"Video"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3446931122")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"name": "files"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("select2363381545")

		return app.Save(collection)
	})
}
