package xml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/beevik/etree"
)

type xml struct {
	Version      string                 `json:"version"`
	LanguageCode string                 `json:"language_code"`
	ContentCode  string                 `json:"content_code"`
	Params       map[string]interface{} `json:"params"`
	Tasks        []task                 `json:"tasks"`
	Translations *translation           `json:"-"`
}
type task struct {
	Sequence    int         `json:"sequence"`
	ExecAction  string      `json:"exec_action"`
	ExecAddress string      `json:"exec_address"`
	ExecPayload interface{} `json:"exec_payload"`
}

func (x *xml) load(element *etree.Element) {
	definition := element.SelectElement("definition")
	x.Version = element.SelectAttrValue("version", "1.0")
	x.LanguageCode = definition.SelectAttrValue("languageCode", "en-us")
	x.ContentCode = definition.SelectAttrValue("contentPackage", "")
	x.Translations = &translation{
		Structure: translationStructure{
			CSVHeader:       []string{"valid", "path", "code", x.LanguageCode},
			CSVTranslations: make(map[string]csvTranslation),
		},
	}
}

// Process start xml parse
func Process(xmlFile, translationFile, jsonFile string) error {
	fmt.Println("Starting xml parse")
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(xmlFile); err != nil {
		return err
	}

	root := doc.Root()
	tasks := root.SelectElement("tasks")

	x := &xml{}
	x.load(root)

	if err := x.Translations.loadCSV(translationFile); err != nil {
		return err
	}

	if err := x.processTask(tasks.ChildElements(), -1, tasks.GetPath()); err != nil {
		return err
	}

	if translationFile != "" {
		if err := x.createTranslation(translationFile); err != nil {
			return err
		}
	}

	if jsonFile != "" {
		jobByte, err := json.MarshalIndent(x, "", "  ")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(jsonFile, jobByte, 0644); err != nil {
			return err
		}
	}

	fmt.Println("Finished xml parse")

	return nil
}

func (x *xml) processTask(childElements []*etree.Element, taskSequence int, path string) error {
	taskSequence++
	for _, element := range childElements {
		switch element.Tag {
		case "createContent":
			if err := createContent(x, element, taskSequence, path); err != nil {
				return err
			}
			break
		case "createSchema":
			if err := createSchema(x, element, taskSequence, path); err != nil {
				return err
			}
			break
		case "createField":
			if err := createField(x, element, taskSequence, path); err != nil {
				return err
			}
			break
		case "createColumn":
			if err := createColumn(x, element, taskSequence, path); err != nil {
				return err
			}
			break
		case "createFeature":
			if err := createFeature(x, element, taskSequence, path); err != nil {
				return err
			}
			break
		case "createDataset":
			if err := createDataset(x, element, taskSequence, path); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
