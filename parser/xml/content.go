package xml

import (
	"encoding/json"
	"fmt"

	"github.com/agile-work/srv-shared/constants"
	"github.com/beevik/etree"
)

func createContent(x *xml, element *etree.Element, taskSequence int, path string) error {
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")
	elmPrefix := element.SelectAttrValue("prefix", "")
	elmModule := element.SelectAttrValue("module", "false")
	elmSystem := element.SelectAttrValue("system", "false")

	path = fmt.Sprintf("%s/createContent[@code='%s']", path, elmCode)

	x.addTranslation([]string{path, "name", elmName})
	x.addTranslation([]string{path, "description", elmDescription})

	if err := x.loadTranslation(path, "name", &elmName); err != nil {
		return err
	}

	if err := x.loadTranslation(path, "description", &elmDescription); err != nil {
		return err
	}

	task := task{
		Sequence:    taskSequence,
		ExecAction:  constants.ExecuteAPIPost,
		ExecAddress: "{system.api_host}/api/v1/core/admin/contents",
		ExecPayload: (json.RawMessage)([]byte(fmt.Sprintf(`{
			"code": "%s",
			"name": %s,
			"description": %s,
			"prefix": "%s",
			"is_module": %s,
			"is_system": %s
		}`, elmCode, elmName, elmDescription, elmPrefix, elmModule, elmSystem))),
	}

	x.Tasks = append(x.Tasks, task)

	if err := x.addTask(element.ChildElements(), taskSequence, path); err != nil {
		return err
	}
	return nil
}
