package xml

import (
	"encoding/json"
	"fmt"

	"github.com/agile-work/srv-shared/constants"
	"github.com/beevik/etree"
)

func createSchema(x *xml, element *etree.Element, taskSequence int, path string) error {
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")

	path = fmt.Sprintf("%s/createSchema[@code='%s']", path, elmCode)

	x.addTranslation([]string{path, "name", elmName})
	x.addTranslation([]string{path, "description", elmDescription})

	elmName, err := x.loadTranslation(path, "name", elmName)
	if err != nil {
		return err
	}

	elmDescription, err = x.loadTranslation(path, "description", elmDescription)
	if err != nil {
		return err
	}

	task := task{
		Sequence:    taskSequence,
		ExecAction:  constants.ExecuteAPIPost,
		ExecAddress: "{system.api_host}/api/v1/core/admin/schemas",
		ExecPayload: (json.RawMessage)([]byte(fmt.Sprintf(`{
			"code": "%s",
			"content_code": "%s",
			"name": %s,
			"description": %s
		}`, elmCode, x.ContentCode, elmName, elmDescription))),
	}

	x.Tasks = append(x.Tasks, task)

	if err := x.addTask(element.ChildElements(), taskSequence, path); err != nil {
		return err
	}
	return nil
}
