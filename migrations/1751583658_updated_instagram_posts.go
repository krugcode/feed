package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3545500662")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"exceptDomains": [],
			"hidden": false,
			"id": "url260122942",
			"name": "instagram_url",
			"onlyDomains": [],
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"exceptDomains": null,
			"hidden": false,
			"id": "url966623967",
			"name": "threads_url",
			"onlyDomains": null,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_1204830414",
			"hidden": false,
			"id": "relation4029274538",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "instagram_account",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1125843985",
			"hidden": false,
			"id": "relation1519021197",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "post",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3545500662")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("url260122942")

		// remove field
		collection.Fields.RemoveById("url966623967")

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_1204830414",
			"hidden": false,
			"id": "relation4029274538",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "instagram_account",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_1125843985",
			"hidden": false,
			"id": "relation1519021197",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "post",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
