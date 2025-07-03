package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_2079557661")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_iUxkYuv3ar` + "`" + ` ON ` + "`" + `crosspost_queue` + "`" + ` (` + "`" + `post` + "`" + `)"
			],
			"name": "crosspost_queue"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"hidden": false,
			"id": "select2063623452",
			"maxSelect": 1,
			"name": "status",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"Queued",
				"Success",
				"Failure"
			]
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text3194832890",
			"max": 0,
			"min": 0,
			"name": "status_message",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"hidden": false,
			"id": "date989355118",
			"max": "",
			"min": "",
			"name": "completed",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "date"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_2079557661")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [],
			"name": "crosspost_jobs"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("select2063623452")

		// remove field
		collection.Fields.RemoveById("text3194832890")

		// remove field
		collection.Fields.RemoveById("date989355118")

		return app.Save(collection)
	})
}
