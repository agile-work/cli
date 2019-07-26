package xml

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agile-work/srv-shared/constants"
	"github.com/beevik/etree"
)

func createDataset(x *xml, element *etree.Element, taskSequence int, path string) error {
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmType := element.SelectAttrValue("type", "")
	elmDescription := element.SelectAttrValue("desc", "")

	path = fmt.Sprintf("%s/createDataset[@code='%s']", path, elmCode)

	x.addTranslation([]string{path, "name", elmName})
	x.addTranslation([]string{path, "description", elmDescription})

	if err := x.loadTranslation(path, "name", &elmName); err != nil {
		return err
	}

	if err := x.loadTranslation(path, "description", &elmDescription); err != nil {
		return err
	}

	payload := fmt.Sprintf(`
		"code": "%s",
		"name": %s,
		"type": "%s",
		"description": %s
	`, elmCode, elmName, elmType, elmDescription)

	if elmType == constants.DatasetStatic {
		elmOptions := element.SelectElement("options").SelectElements("option")
		orders := []string{}
		options := make(map[string]map[string]interface{})

		for _, elmOption := range elmOptions {
			option := make(map[string]interface{})
			code := elmOption.SelectAttrValue("code", "")
			name := elmOption.SelectAttrValue("name", "")
			orders = append(orders, code)

			pathOption := fmt.Sprintf("%s/options/option[@code='%s']", path, code)

			if err := x.loadTranslation(pathOption, "name", &name); err != nil {
				return err
			}

			option["code"] = code
			option["name"] = (json.RawMessage)([]byte(name))
			option["active"] = "true"
			options[code] = option

			x.addTranslation([]string{pathOption, "name", name})
		}
		ordersByte, err := json.MarshalIndent(orders, "", "  ")
		if err != nil {
			return err
		}
		optionsByte, err := json.MarshalIndent(options, "", "  ")
		if err != nil {
			return err
		}
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"order": %s,
				"options": %s
			}
		}`, payload, string(ordersByte), string(optionsByte))
	} else {
		elmQuery := element.SelectElement("query")
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
        "query": "%s"
    	}
		}`, payload, strings.Trim(elmQuery.Text(), " \n\r"))
	}

	task := task{
		Sequence:    taskSequence,
		ExecAction:  constants.ExecuteAPIPost,
		ExecAddress: "{system.api_host}/api/v1/core/admin/datasets",
		ExecPayload: (json.RawMessage)([]byte(payload)),
	}

	x.Tasks = append(x.Tasks, task)

	if err := x.addTask(element.ChildElements(), taskSequence, path); err != nil {
		return err
	}
	return nil
}
