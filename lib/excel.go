package lib

import (
	"fmt"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type Excel struct {
	App       *ole.IDispatch
	WorkBooks *ole.IDispatch
}

// https://github.com/tanaton/go-ole-msoffice/blob/master/excel/excel.go
func NewExcel() (e *Excel, err error) {
	unknown, err := oleutil.GetActiveObject("Excel.Application")
	if err != nil {
		return nil, fmt.Errorf("no active excel found: %w", err)
	}
	app, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, err
	}
	workBooks_, err := app.GetProperty("Workbooks")
	if err != nil {
		return nil, err
	}
	workBooks := workBooks_.ToIDispatch()

	return &Excel{App: app, WorkBooks: workBooks}, nil
}

func (e *Excel) WorkBookCount() (count int, err error) {
	wbCount_, err := e.WorkBooks.GetProperty("Count")
	if err != nil {
		return 0, err
	}
	return (int)(wbCount_.Val), nil
}

func (e *Excel) WorkBookByIndex(index int) (wb *ole.IDispatch, err error) {
	wb_, err := e.WorkBooks.GetProperty("Item", index)
	if err != nil {
		return nil, err
	}
	return wb_.ToIDispatch(), nil
}

func (e *Excel) WorkBookByName(name string) (wb *ole.IDispatch, err error) {
	wb_, err := e.WorkBooks.GetProperty("Item", name)
	if err != nil {
		return nil, err
	}
	return wb_.ToIDispatch(), nil
}

func (e *Excel) OpenedWookBookNames() (names []string, err error) {
	count, err := e.WorkBookCount()
	if err != nil {
		return nil, err
	}
	wbNames := make([]string, count)
	for i := 0; i < count; i++ {
		wb, err := e.WorkBookByIndex(i + 1) // 1origin
		if err != nil {
			return nil, err
		}

		name_, err := wb.GetProperty("Name")
		if err != nil {
			return nil, err
		}
		name := name_.ToString()
		wbNames[i] = name
	}
	return wbNames, nil
}

func (e *Excel) ActivateWorkBook(name string) (err error) {
	wb, err := e.WorkBookByName(name)
	if err != nil {
		return err
	}
	_, err = wb.CallMethod("Activate")
	if err != nil {
		return err
	}
	// https://stackoverflow.com/questions/56841745/need-to-maximize-the-excel-window-after-opening-it-using-win32com-client
	_, err = e.App.PutProperty("WindowState", -4137)
	if err != nil {
		return err
	}
	return nil
}
