package xml

import (
	"fmt"

	"github.com/agile-work/srv-shared/constants"
	"github.com/beevik/etree"
)

func createColumn(x *xml, element *etree.Element, taskSequence int, path string, createTranslation bool) error {
	elmTable := element.SelectAttrValue("table", "")
	elmType := element.SelectAttrValue("type", "")
	elmCode := element.SelectAttrValue("code", "")

	path = fmt.Sprintf("%s/createColumn[@table='%s'][@code='%s']", path, elmTable, elmCode)

	task := task{
		Sequence:    taskSequence,
		ExecAction:  constants.ExecuteQuery,
		ExecAddress: "local",
		ExecPayload: fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, elmTable, elmCode, elmType),
	}

	x.Tasks = append(x.Tasks, task)

	if err := x.addTask(element.ChildElements(), taskSequence, path, createTranslation); err != nil {
		return err
	}
	return nil
}
