package services

import (
	"github.com/xuri/excelize/v2"
)

func writeDownloadWorkbook(filePath, sheetName string, headers []interface{}, writeRows func(*excelize.StreamWriter) error) (resultErr error) {
	return writeDownloadWorkbookWithHeader(filePath, sheetName, headers, true, writeRows)
}

func writeDownloadWorkbookWithHeader(filePath, sheetName string, headers []interface{}, includeHeader bool, writeRows func(*excelize.StreamWriter) error) (resultErr error) {
	return writeDownloadWorkbookWithWidths(filePath, sheetName, headers, includeHeader, nil, writeRows)
}

func writeDownloadWorkbookWithWidths(filePath, sheetName string, headers []interface{}, includeHeader bool, widths []int, writeRows func(*excelize.StreamWriter) error) (resultErr error) {
	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); resultErr == nil && err != nil {
			resultErr = err
		}
	}()
	defaultSheet := file.GetSheetName(0)
	if defaultSheet != sheetName {
		if err := file.SetSheetName(defaultSheet, sheetName); err != nil {
			return err
		}
	}
	writer, err := file.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}
	for index, width := range widths {
		if width <= 0 {
			continue
		}
		excelWidth := float64(width) / 7
		if excelWidth < 1 {
			excelWidth = 1
		}
		if err := writer.SetColWidth(index+1, index+1, excelWidth); err != nil {
			return err
		}
	}
	if includeHeader {
		if err := writer.SetRow("A1", headers); err != nil {
			return err
		}
	}
	if err := writeRows(writer); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return file.SaveAs(filePath)
}
