package xml

import (
	"fmt"

	"github.com/agile-work/srv-shared/constants"
	"github.com/beevik/etree"
)

func createColumn(x *xml, element *etree.Element, taskSequence int, path string) error {
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

	if err := x.processTask(element.ChildElements(), taskSequence, path); err != nil {
		return err
	}
	return nil
}
