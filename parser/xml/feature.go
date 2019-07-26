package xml

import (
	"encoding/json"
	"fmt"

	"github.com/agile-work/srv-shared/constants"
	"github.com/beevik/etree"
)

func createFeature(x *xml, element *etree.Element, taskSequence int, path string) error {
	elmModuleCode := element.SelectAttrValue("moduleCode", "")
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")
	permissions := make(map[string]interface{})

	path = fmt.Sprintf("%s/createFeature[@moduleCode='%s'][@code='%s']", path, elmModuleCode, elmCode)

	x.addTranslation([]string{path, "name", elmName})
	x.addTranslation([]string{path, "description", elmDescription})

	if err := x.loadTranslation(path, "name", &elmName); err != nil {
		return err
	}

	if err := x.loadTranslation(path, "description", &elmDescription); err != nil {
		return err
	}

	for _, p := range element.SelectElements("permission") {
		code := p.SelectAttrValue("code", "")
		name := p.SelectAttrValue("name", "")

		pathPermission := fmt.Sprintf("%s/permission[@code='%s']", path, code)

		x.addTranslation([]string{pathPermission, "name", name})

		if err := x.loadTranslation(pathPermission, "name", &name); err != nil {
			return err
		}

		permissions[code] = (json.RawMessage)([]byte(name))
	}
	permissionsByte, err := json.MarshalIndent(permissions, "", "  ")
	if err != nil {
		return err
	}

	task := task{
		Sequence:    taskSequence,
		ExecAction:  constants.ExecuteAPIPost,
		ExecAddress: fmt.Sprintf("{system.api_host}/api/v1/core/admin/modules/%s/features", elmModuleCode),
		ExecPayload: (json.RawMessage)([]byte(fmt.Sprintf(`{
			"%s": {
				"name": %s,
				"description": %s,
				"permissions": %s
			}
		}`, elmCode, elmName, elmDescription, string(permissionsByte)))),
	}

	x.Tasks = append(x.Tasks, task)

	if err := x.addTask(element.ChildElements(), taskSequence, path); err != nil {
		return err
	}
	return nil
}
