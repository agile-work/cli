package xml

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

type translation struct {
	Structure  translationStructure
	Availables map[string]map[string]string
}

type translationStructure struct {
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
	if _, err := os.Stat(csvPath); err == nil {
		csvFile, err := os.Open(csvPath)
		if err != nil {
			return err
		}
		reader := csv.NewReader(bufio.NewReader(csvFile))
		isHeader := true
		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if isHeader {
				t.Structure.CSVHeader = line
				isHeader = false
				continue
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
	}
	return nil
}

func (x *xml) processTranslation(path, code string, text *string) error {
	x.addTranslation(path, code, *text)
	if err := x.loadTranslation(path, code, text); err != nil {
		return err
	}
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

func (x *xml) loadTranslation(path, code string, text *string) error {
	if csvTranslation, ok := x.Translations.Structure.CSVTranslations[path+code]; ok {
		languages := make(map[string]string)
		for _, language := range csvTranslation.Languages {
			if language.Text != "" {
				languages[language.Code] = language.Text
			}
		}
		languagesByte, err := json.MarshalIndent(languages, "", "  ")
		if err != nil {
			return err
		}
		*text = string(languagesByte)
	} else {
		*text = fmt.Sprintf(`{ "%s": "%s" }`, x.LanguageCode, *text)
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
