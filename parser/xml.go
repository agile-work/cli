package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/beevik/etree"
)

const (
	executeQuery     string = "exec_query"
	executeAPIGet    string = "api_get"
	executeAPIPost   string = "api_post"
	executeAPIDelete string = "api_delete"
	executeAPIUpdate string = "api_patch"

	fieldText       string = "text"
	fieldNumber     string = "number"
	fieldDate       string = "date"
	fieldLookup     string = "lookup"
	fieldAttachment string = "attachment"
)

type Job struct {
	LanguageCode string                 `json:"language_code"`
	ContentCode  string                 `json:"content_code"`
	Params       map[string]interface{} `json:"params"`
	Tasks        []Task                 `json:"tasks"`
}

type Task struct {
	Sequence    int         `json:"sequence"`
	ExecAction  string      `json:"exec_action"`
	ExecAddress string      `json:"exec_address"`
	ExecPayload interface{} `json:"exec_payload"`
}

// GenerateJobTasks receives a xml file to transform in a module json job tasks
func GenerateJobTasks(xmlPath string) error {
	if xmlPath == "" {
		return errors.New("The xml file path is required")
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(xmlPath); err != nil {
		return err
	}
	job := &Job{}

	root := doc.Root()
	definition := root.SelectElement("definition")
	job.LanguageCode = definition.SelectAttrValue("languageCode", "en")
	job.ContentCode = definition.SelectAttrValue("contentPackage", "")

	if err := job.addTask(root.SelectElement("tasks").ChildElements(), -1); err != nil {
		return err
	}

	jobByte, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile("test.json", jobByte, 0644); err != nil {
		return err
	}
	return nil
}

func (j *Job) addTask(childElements []*etree.Element, taskSequence int) error {
	taskSequence++
	for _, element := range childElements {
		switch element.Tag {
		case "createSchema":
			if err := createSchema(j, element, taskSequence); err != nil {
				return err
			}
			break
		case "createField":
			if err := createField(j, element, taskSequence); err != nil {
				return err
			}
			break
		case "createColumn":
			if err := createColumn(j, element, taskSequence); err != nil {
				return err
			}
			break
		case "createFeature":
			if err := createFeature(j, element, taskSequence); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func createSchema(job *Job, element *etree.Element, taskSequence int) error {
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")

	task := Task{
		Sequence:    taskSequence,
		ExecAction:  executeAPIPost,
		ExecAddress: "{system.api_url}/core/admin/schemas",
		ExecPayload: (json.RawMessage)([]byte(fmt.Sprintf(`{
			"code": "%s",
			"content_code": "%s",
			"name": "%s",
			"description": "%s"
		}`, elmCode, job.ContentCode, elmName, elmDescription))),
	}

	job.Tasks = append(job.Tasks, task)

	if err := job.addTask(element.ChildElements(), taskSequence); err != nil {
		return err
	}
	return nil
}

func createField(job *Job, element *etree.Element, taskSequence int) error {
	elmSchemaCode := element.SelectAttrValue("schemaCode", "")
	elmType := element.SelectAttrValue("type", "")
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")
	payload := fmt.Sprintf(`
		"code": "%s",
		"content_code": "%s",
		"schema_code": "%s",
		"field_type": "%s",
		"name": "%s",
		"description": "%s",
		"active": true
	`, elmCode, job.ContentCode, elmSchemaCode, elmType, elmName, elmDescription)

	switch elmType {
	case fieldText:
		elmDisplay := element.SelectAttrValue("display", "single_line")
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"display": "%s"
			}
		}`, payload, elmDisplay)
		break
	case fieldNumber:
		elmDisplay := element.SelectAttrValue("display", "number")
		elmDecimals := element.SelectAttrValue("decimals", "0")
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"display": "%s",
				"decimals": %s
			}
		}`, payload, elmDisplay, elmDecimals)
		break
	case fieldDate:
		elmDisplay := element.SelectAttrValue("display", "date_time")
		elmFormat := element.SelectAttrValue("format", "DD/MM/YYYY HH:MM")
		elmAggRates := element.SelectAttrValue("aggRates", "null")
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"display": "%s",
				"format": "%s",
				"aggr_rates": %s
			}
		}`, payload, elmDisplay, elmFormat, elmAggRates)
		break
	case fieldLookup:
		payload = fmt.Sprintf(`{
			%s
		}`, payload)
		break
	}

	task := Task{
		Sequence:    taskSequence,
		ExecAction:  executeAPIPost,
		ExecAddress: fmt.Sprintf("{system.api_url}/core/admin/schemas/%s/fields", elmSchemaCode),
		ExecPayload: (json.RawMessage)([]byte(payload)),
	}

	job.Tasks = append(job.Tasks, task)

	if err := job.addTask(element.ChildElements(), taskSequence); err != nil {
		return err
	}
	return nil
}

func createColumn(job *Job, element *etree.Element, taskSequence int) error {
	elmTable := element.SelectAttrValue("table", "")
	elmType := element.SelectAttrValue("type", "")
	elmCode := element.SelectAttrValue("code", "")

	task := Task{
		Sequence:    taskSequence,
		ExecAction:  executeQuery,
		ExecAddress: "local",
		ExecPayload: fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, elmTable, elmCode, elmType),
	}

	job.Tasks = append(job.Tasks, task)

	if err := job.addTask(element.ChildElements(), taskSequence); err != nil {
		return err
	}
	return nil
}

func createFeature(job *Job, element *etree.Element, taskSequence int) error {
	elmModuleCode := element.SelectAttrValue("moduleCode", "")
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")
	permissions := make(map[string]string)

	for _, p := range element.SelectElements("permission") {
		permissions[p.SelectAttrValue("code", "")] = p.SelectAttrValue("name", "")
	}

	permissionsByte, err := json.MarshalIndent(permissions, "", "  ")
	if err != nil {
		return err
	}

	task := Task{
		Sequence:    taskSequence,
		ExecAction:  executeAPIPost,
		ExecAddress: fmt.Sprintf("{system.api_url}/core/admin/modules/%s/features", elmModuleCode),
		ExecPayload: (json.RawMessage)([]byte(fmt.Sprintf(`{
			"%s": {
				"name": "%s",
				"description": "%s",
				"permissions": %s
			}
		}`, elmCode, elmName, elmDescription, string(permissionsByte)))),
	}

	job.Tasks = append(job.Tasks, task)

	if err := job.addTask(element.ChildElements(), taskSequence); err != nil {
		return err
	}
	return nil
}
