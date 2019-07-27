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
	"strconv"

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

type translation struct {
	Structure  translationStructure
	Availables map[string]map[string]string
}

type translationStructure struct {
	TotalLanguages  int
	CSVHeader       []string
	CSVTranslations map[string]csvTranslation
}

type csvTranslation struct {
	Valid     bool
	Path      string
	Code      string
	Languages []language
}

type language struct {
	Code string
	Text string
}

func (t csvTranslation) toSlice() []string {
	slice := []string{
		strconv.FormatBool(t.Valid),
		t.Path,
		t.Code,
	}

	for _, language := range t.Languages {
		slice = append(slice, language.Text)
	}
	return slice
}

func (t *translation) loadCSV(csvPath string) error {
	t.Structure.CSVTranslations = make(map[string]csvTranslation)
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(t.Structure.CSVHeader) == 0 {
			t.Structure.CSVHeader = line
		} else {
			csvStructure := csvTranslation{
				Valid: false,
				Path:  line[1],
				Code:  line[2],
			}
			for index := 3; index < len(line); index++ {
				csvStructure.Languages = append(csvStructure.Languages, language{
					Code: t.Structure.CSVHeader[index],
					Text: line[index],
				})
			}
			t.Structure.CSVTranslations[csvStructure.Path+csvStructure.Code] = csvStructure
		}
	}
	t.Structure.TotalLanguages = len(t.Structure.CSVHeader) - 3
	return nil
}

func CreateTranslation(from, to string) error {
	if from == "" {
		return errors.New("The xml file path is required")
	}
	if to == "" {
		return errors.New("The csv file path is required")
	}

	fmt.Println("Starting xml parse to translation")

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(from); err != nil {
		return err
	}
	x := &xml{}
	x.Translations = &translation{}
	root := doc.Root()
	definition := root.SelectElement("definition")
	x.Version = root.SelectAttrValue("version", "1.0")
	x.LanguageCode = definition.SelectAttrValue("languageCode", "en-us")
	x.ContentCode = definition.SelectAttrValue("contentPackage", "")
	tasks := root.SelectElement("tasks")

	if err := x.Translations.loadCSV(to); err != nil {
		return err
	}

	if err := x.addTask(tasks.ChildElements(), -1, tasks.GetPath(), true); err != nil {
		return err
	}

	if err := x.createTranslation(to); err != nil {
		return err
	}

	fmt.Println("Finished xml parse to translation")

	return nil
}

// GenerateJobTasks receives a xml file to transform in a module json job tasks
func GenerateJobTasks(from, to, translationPath string) error {
	if from == "" {
		return errors.New("The xml file path is required")
	}
	if to == "" {
		return errors.New("The json file path is required")
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(from); err != nil {
		return err
	}
	x := &xml{}
	x.Translations = &translation{}
	fmt.Println("Starting xml parse")

	root := doc.Root()
	definition := root.SelectElement("definition")
	x.Version = root.SelectAttrValue("version", "1.0")
	x.LanguageCode = definition.SelectAttrValue("languageCode", "en-us")
	x.ContentCode = definition.SelectAttrValue("contentPackage", "")
	tasks := root.SelectElement("tasks")

	if translationPath != "" {
		if err := x.Translations.loadCSV(translationPath); err != nil {
			return err
		}
	}

	if err := x.addTask(tasks.ChildElements(), -1, tasks.GetPath(), false); err != nil {
		return err
	}

	jobByte, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(to, jobByte, 0644); err != nil {
		return err
	}

	fmt.Println("Finished xml parse")

	return nil
}

func (x *xml) addTranslation(path, code, text string) {
	key := path + code
	if value, ok := x.Translations.Structure.CSVTranslations[key]; ok {
		value.Valid = true
		x.Translations.Structure.CSVTranslations[key] = value
	} else {
		languages := []language{}
		for index := 3; index < len(x.Translations.Structure.CSVHeader); index++ {
			languageCode := x.Translations.Structure.CSVHeader[index]
			language := language{
				Code: languageCode,
			}
			if languageCode == x.LanguageCode {
				language.Text = text
			}
			languages = append(languages, language)
		}
		x.Translations.Structure.CSVTranslations[key] = csvTranslation{
			Code:      code,
			Path:      path,
			Valid:     true,
			Languages: languages,
		}
	}
}

func (x *xml) loadTranslation(path, code string, value *string) error {
	if csvTranslation, ok := x.Translations.Structure.CSVTranslations[path+code]; ok {
		languages := make(map[string]string)
		for _, language := range csvTranslation.Languages {
			languages[language.Code] = language.Text
		}
		languagesByte, err := json.MarshalIndent(languages, "", "  ")
		if err != nil {
			return err
		}
		*value = string(languagesByte)
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
	if err := w.Write(x.Translations.Structure.CSVHeader); err != nil {
		return err
	}

	for _, csvTranslation := range x.Translations.Structure.CSVTranslations {
		if err := w.Write(csvTranslation.toSlice()); err != nil {
			return err
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (x *xml) addTask(childElements []*etree.Element, taskSequence int, path string, createTranslation bool) error {
	taskSequence++
	for _, element := range childElements {
		switch element.Tag {
		case "createContent":
			if err := createContent(x, element, taskSequence, path, createTranslation); err != nil {
				return err
			}
			break
		case "createSchema":
			if err := createSchema(x, element, taskSequence, path, createTranslation); err != nil {
				return err
			}
			break
		case "createField":
			if err := createField(x, element, taskSequence, path, createTranslation); err != nil {
				return err
			}
			break
		case "createColumn":
			if err := createColumn(x, element, taskSequence, path, createTranslation); err != nil {
				return err
			}
			break
		case "createFeature":
			if err := createFeature(x, element, taskSequence, path, createTranslation); err != nil {
				return err
			}
			break
		case "createDataset":
			if err := createDataset(x, element, taskSequence, path, createTranslation); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
