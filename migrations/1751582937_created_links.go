package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text724990059",
					"max": 0,
					"min": 0,
					"name": "title",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": true,
					"system": false,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text888727361",
					"max": 0,
					"min": 0,
					"name": "href",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": true,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "bool3341446040",
					"name": "is_local",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_3446931122",
					"hidden": false,
					"id": "relation3309110367",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "image",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "number4032570574",
					"max": null,
					"min": null,
					"name": "click_count",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "bool1821142673",
					"name": "is_visible",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
				},
				{
					"hidden": false,
					"id": "number4113142680",
					"max": null,
					"min": null,
					"name": "order",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": false,
					"type": "autodate"
				},
				{
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": false,
					"type": "autodate"
				}
			],
			"id": "pbc_449060851",
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_rmUxsnboKo` + "`" + ` ON ` + "`" + `links` + "`" + ` (` + "`" + `title` + "`" + `)",
				"CREATE UNIQUE INDEX ` + "`" + `idx_yCPSDi3ERO` + "`" + ` ON ` + "`" + `links` + "`" + ` (` + "`" + `href` + "`" + `)"
			],
			"listRule": null,
			"name": "links",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_449060851")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
