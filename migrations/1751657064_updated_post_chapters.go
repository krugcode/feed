package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1719698224")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("url4068916274")

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1719698224")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"exceptDomains": [],
			"hidden": false,
			"id": "url4068916274",
			"name": "permalink",
			"onlyDomains": [],
			"presentable": false,
			"required": true,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
