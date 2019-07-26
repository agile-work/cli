package xml

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/beevik/etree"
)

type xml struct {
	Version      string                 `json:"version"`
	LanguageCode string                 `json:"language_code"`
	ContentCode  string                 `json:"content_code"`
	Params       map[string]interface{} `json:"params"`
	Tasks        []task                 `json:"tasks"`
	Translations translation            `json:"-"`
}

type task struct {
	Sequence    int         `json:"sequence"`
	ExecAction  string      `json:"exec_action"`
	ExecAddress string      `json:"exec_address"`
	ExecPayload interface{} `json:"exec_payload"`
}

type translation struct {
	CSVStructure [][]string
	Availables   map[string]map[string]string
}

// GenerateJobTasks receives a xml file to transform in a module json job tasks
func GenerateJobTasks(xmlPath, translationPath string, onlyTranslation bool) error {
	if xmlPath == "" {
		return errors.New("The xml file path is required")
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(xmlPath); err != nil {
		return err
	}
	x := &xml{}

	fmt.Println("Starting xml parse")

	root := doc.Root()
	definition := root.SelectElement("definition")
	x.Version = root.SelectAttrValue("version", "1.0")
	x.LanguageCode = definition.SelectAttrValue("languageCode", "en-us")
	x.ContentCode = definition.SelectAttrValue("contentPackage", "")
	tasks := root.SelectElement("tasks")

	if translationPath != "" {
		csvFile, err := os.Open(translationPath)
		if err != nil {
			return err
		}
		reader := csv.NewReader(bufio.NewReader(csvFile))
		headers := []string{}
		translations := make(map[string]map[string]string)
		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if len(headers) == 0 {
				headers = line
			} else {
				languages := make(map[string]string)
				for index := 2; index < len(line); index++ {
					languages[headers[index]] = line[index]
				}
				translations[line[0]+line[1]] = languages
			}
		}
		x.Translations.Availables = translations
	}

	if err := x.addTask(tasks.ChildElements(), -1, tasks.GetPath()); err != nil {
		return err
	}

	if !onlyTranslation {
		jobByte, err := json.MarshalIndent(x, "", "  ")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(fmt.Sprintf("%s.json", strings.TrimSuffix(xmlPath, path.Ext(xmlPath))), jobByte, 0644); err != nil {
			return err
		}
	} else {
		if err := x.createTranslation(fmt.Sprintf("%s.csv", strings.TrimSuffix(xmlPath, path.Ext(xmlPath)))); err != nil {
			return err
		}
	}

	fmt.Println("Finished xml parse")

	return nil
}

func (x *xml) addTranslation(data []string) {
	x.Translations.CSVStructure = append(x.Translations.CSVStructure, data)
}

func (x *xml) loadTranslation(path, code string, value *string) error {
	translation := x.Translations.Availables[path+code]
	if len(translation) > 0 {
		translationByte, err := json.MarshalIndent(translation, "", "  ")
		if err != nil {
			return err
		}
		*value = string(translationByte)
	} else {
		*value = fmt.Sprintf(`{ "%s": "%s" }`, x.LanguageCode, *value)
	}

	return nil
}

func (x *xml) createTranslation(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	w := csv.NewWriter(file)
	if err := w.Write([]string{"path", "code", x.LanguageCode}); err != nil {
		return err
	}

	for _, translation := range x.Translations.CSVStructure {
		if err := w.Write(translation); err != nil {
			return err
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (x *xml) addTask(childElements []*etree.Element, taskSequence int, path string) error {
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
