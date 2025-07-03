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
					"cascadeDelete": false,
					"collectionId": "pbc_601157786",
					"hidden": false,
					"id": "relation4232930610",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "collection",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": false,
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
				},
				{
					"hidden": false,
					"id": "number4113142680",
					"max": null,
					"min": null,
					"name": "order",
					"onlyInt": false,
					"presentable": false,
					"required": true,
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
			"id": "pbc_3519724588",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_nm5DJNctzE` + "`" + ` ON ` + "`" + `collection_posts` + "`" + ` (` + "`" + `collection` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_O5OVI8VYLH` + "`" + ` ON ` + "`" + `collection_posts` + "`" + ` (` + "`" + `post` + "`" + `)"
			],
			"listRule": null,
			"name": "collection_posts",
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
		collection, err := app.FindCollectionByNameOrId("pbc_3519724588")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
