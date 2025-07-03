package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1204830414")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_3446931122",
			"hidden": false,
			"id": "relation3311767829",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "profile_picture",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1784151356",
			"max": 0,
			"min": 0,
			"name": "at",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1204830414")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation3311767829")

		// remove field
		collection.Fields.RemoveById("text1784151356")

		return app.Save(collection)
	})
}
