package xml

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/agile-work/srv-shared/constants"

	"github.com/beevik/etree"
)

func createField(x *xml, element *etree.Element, taskSequence int, path string) error {
	elmSchemaCode := element.SelectAttrValue("schemaCode", "")
	elmType := element.SelectAttrValue("type", "")
	elmCode := element.SelectAttrValue("code", "")
	elmName := element.SelectAttrValue("name", "")
	elmDescription := element.SelectAttrValue("desc", "")

	path = fmt.Sprintf("%s/createField[@schemaCode='%s'][@code='%s']", path, elmSchemaCode, elmCode)

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
		"content_code": "%s",
		"schema_code": "%s",
		"field_type": "%s",
		"name": %s,
		"description": %s,
		"active": true
	`, elmCode, x.ContentCode, elmSchemaCode, elmType, elmName, elmDescription)

	switch elmType {
	case constants.FieldText:
		elmDisplay := element.SelectAttrValue("display", "single_line")
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"display": "%s"
			}
		}`, payload, elmDisplay)
		break
	case constants.FieldNumber:
		elmDisplay := element.SelectAttrValue("display", "number")
		elmDecimals := element.SelectAttrValue("decimals", "0")
		elmScale := element.SelectAttrValue("scale", "")
		elmScaleItems := element.ChildElements()
		payloadScale := ""
		payloadDataset := ""
		payloadAggRates := ""
		if elmScale != "" {
			payloadDataset = fmt.Sprintf(`"dataset_code": "%s"`, elmScale)
		}
		if len(elmScaleItems) > 0 {
			items := make(map[string]interface{})
			for _, elmScaleItem := range elmScaleItems {
				elmScaleItemValues := elmScaleItem.ChildElements()
				values := make(map[string]interface{})
				for _, elmScaleItemValue := range elmScaleItemValues {
					values[elmScaleItemValue.Tag] = elmScaleItemValue.SelectAttrValue("value", "0")
				}
				items[elmScaleItem.Tag] = values
			}
			itemsByte, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				return err
			}
			payloadAggRates = fmt.Sprintf(`,
					"aggr_rates": %s
				`, string(itemsByte))
		}
		if payloadDataset != "" {
			payloadScale = fmt.Sprintf(`,
				"scale": {
					%s
					%s
				}`, payloadDataset, payloadAggRates)
		}
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"display": "%s",
				"decimals": %s
				%s
			}
		}`, payload, elmDisplay, elmDecimals, payloadScale)
		break
	case constants.FieldDate:
		elmDisplay := element.SelectAttrValue("display", "date_time")
		elmFormat := element.SelectAttrValue("format", "DD/MM/YYYY HH:MM")
		payload = fmt.Sprintf(`{
			%s,
			"definitions": {
				"display": "%s",
				"format": "%s"
			}
		}`, payload, elmDisplay, elmFormat)
		break
	case constants.FieldLookup:
		elmDisplay := element.SelectAttrValue("display", "select_single")
		elemDataset := element.SelectElement("dataset")
		elmDatasetCode := elemDataset.SelectAttrValue("code", "")
		elmLookupType := elemDataset.SelectAttrValue("type", "")
		if elmLookupType == constants.FieldLookupStatic {
			payload = fmt.Sprintf(`{
				%s,
				"definitions": {
					"display": "%s",
					"dataset_code": "%s",
					"lookup_type": "%s"     
				}
			}`, payload, elmDisplay, elmDatasetCode, elmLookupType)
		} else {
			elmLookupLabel := elemDataset.SelectAttrValue("lookup_label", "name")
			elmLookupValue := elemDataset.SelectAttrValue("lookup_value", "code")
			elmFields := elemDataset.SelectElement("fields").SelectElements("field")
			elmGroups := elemDataset.SelectElement("groups")
			fields := []map[string]interface{}{}
			params := []map[string]interface{}{}
			payloadSecurityGroups := ""
			if elmGroups != nil {
				payloadSecurityGroups = fmt.Sprintf(`, "security_groups": ["%s"]`, strings.Join(strings.Split(strings.Trim(elmGroups.Text(), " \n\r"), ","), `","`))
			}
			for _, elmField := range elmFields {
				field := make(map[string]interface{})
				code := elmField.SelectAttrValue("code", "")
				name := elmField.SelectAttrValue("name", "")

				pathField := fmt.Sprintf("%s/fields/field[@code='%s']", path, code)

				x.addTranslation([]string{pathField, "name", name})

				if err := x.loadTranslation(pathField, "name", &name); err != nil {
					return err
				}

				field["code"] = code
				field["label"] = (json.RawMessage)([]byte(name))
				elmFilter := elmField.SelectElement("filter")
				if elmFilter != nil {
					filter := make(map[string]interface{})
					filter["value_type"] = elmFilter.SelectAttrValue("type", "")
					filter["value"] = elmFilter.SelectAttrValue("value", "")
					filter["operator"] = elmFilter.SelectAttrValue("operator", "")
					filter["readonly"], _ = strconv.ParseBool(elmFilter.SelectAttrValue("readonly", "false"))
					field["filter"] = filter
				}
				fields = append(fields, field)
			}
			fieldsByte, err := json.MarshalIndent(fields, "", "  ")
			if err != nil {
				return err
			}
			payloadParams := ""
			elmParamsAgg := elemDataset.SelectElement("params")
			if elmParamsAgg != nil {
				elmParams := elmParamsAgg.SelectElements("param")
				for _, elmParam := range elmParams {
					param := make(map[string]interface{})
					param["code"] = elmParam.SelectAttrValue("code", "")
					param["value_type"] = elmParam.SelectAttrValue("type", "")
					param["value"] = elmParam.SelectAttrValue("value", "")
					params = append(params, param)
				}
				paramsByte, err := json.MarshalIndent(params, "", "  ")
				if err != nil {
					return err
				}
				payloadParams = fmt.Sprintf(`, "lookup_params": %s`, string(paramsByte))
			}
			payload = fmt.Sprintf(`{
				%s,
				"definitions": {
					"display": "%s",
					"dataset_code": "%s",
					"lookup_type": "%s",
					"lookup_label": "%s",
					"lookup_value": "%s",
					"lookup_fields": %s
					%s
					%s
				}
			}`, payload, elmDisplay, elmDatasetCode, elmLookupType, elmLookupLabel, elmLookupValue, string(fieldsByte), payloadParams, payloadSecurityGroups)
		}
		break
	}

	task := task{
		Sequence:    taskSequence,
		ExecAction:  constants.ExecuteAPIPost,
		ExecAddress: fmt.Sprintf("{system.api_host}/api/v1/core/admin/schemas/%s/fields", elmSchemaCode),
		ExecPayload: (json.RawMessage)([]byte(payload)),
	}

	x.Tasks = append(x.Tasks, task)

	if err := x.addTask(element.ChildElements(), taskSequence, path); err != nil {
		return err
	}
	return nil
}
