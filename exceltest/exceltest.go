package main

import (
	"fmt"
	"github.com/tealeg/xlsx/v3"
)

/*
18 Oct 2024 -- I'm going to first test out here the code to write an Excel file in xlsx format.  When I understand this well enough,
               I'll add this code to fromfx.go so I'll end up writing a text file and an xlsx file, at least for a while.
               This will use https://github.com/tealeg/xlsx.
               https://github.com/tealeg/xlsx/blob/master/tutorial/tutorial.adoc
*/

const lastAltered = "18 Oct 24"
const excelFormat = "$#,##0.0_);[Red](-$#,##0.0.00)"

func main() {
	fmt.Printf(" Excel Test, last altered %s\n", lastAltered)
	workbook := xlsx.NewFile()
	sheet, err := workbook.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(" Error from workbook.AddSheet: %s\n", err)
		return
	}

	row1 := sheet.AddRow()
	cell1 := row1.AddCell()
	cell1.SetFloat(123.456)
	cell2 := row1.AddCell()
	cell2.SetFloatWithFormat(456.78, excelFormat)
	cell3 := row1.AddCell()
	cell3.SetFloatWithFormat(-12456.78, excelFormat)

	row2 := sheet.AddRow()
	cell4 := row2.AddCell()
	cell4.SetFloatWithFormat(12456.7, excelFormat)
	cell5 := row2.AddCell()
	cell5.SetFloatWithFormat(-2312456.8, excelFormat)

	err = workbook.Save("test.xlsx")
	if err != nil {
		fmt.Printf(" Error from workbook.Save: %s\n", err)
		return
	}
}
