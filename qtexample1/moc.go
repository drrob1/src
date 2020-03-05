package main

//#include <stdint.h>
//#include <stdlib.h>
//#include <string.h>
//#include "moc.h"
import "C"
import (
	"runtime"
	"strings"
	"unsafe"

	"github.com/therecipe/qt"
	std_core "github.com/therecipe/qt/core"
	std_gui "github.com/therecipe/qt/gui"
	std_widgets "github.com/therecipe/qt/widgets"
)

func cGoUnpackString(s C.struct_Moc_PackedString) string {
	if int(s.len) == -1 {
		return C.GoString(s.data)
	}
	return C.GoStringN(s.data, C.int(s.len))
}
func cGoUnpackBytes(s C.struct_Moc_PackedString) []byte {
	if int(s.len) == -1 {
		gs := C.GoString(s.data)
		return *(*[]byte)(unsafe.Pointer(&gs))
	}
	return C.GoBytes(unsafe.Pointer(s.data), C.int(s.len))
}
func unpackStringList(s string) []string {
	if len(s) == 0 {
		return make([]string, 0)
	}
	return strings.Split(s, "¡¦!")
}

type TextEdit_ITF interface {
	std_widgets.QMainWindow_ITF
	TextEdit_PTR() *TextEdit
}

func (ptr *TextEdit) TextEdit_PTR() *TextEdit {
	return ptr
}

func (ptr *TextEdit) Pointer() unsafe.Pointer {
	if ptr != nil {
		return ptr.QMainWindow_PTR().Pointer()
	}
	return nil
}

func (ptr *TextEdit) SetPointer(p unsafe.Pointer) {
	if ptr != nil {
		ptr.QMainWindow_PTR().SetPointer(p)
	}
}

func PointerFromTextEdit(ptr TextEdit_ITF) unsafe.Pointer {
	if ptr != nil {
		return ptr.TextEdit_PTR().Pointer()
	}
	return nil
}

func NewTextEditFromPointer(ptr unsafe.Pointer) (n *TextEdit) {
	if gPtr, ok := qt.Receive(ptr); !ok {
		n = new(TextEdit)
		n.SetPointer(ptr)
	} else {
		switch deduced := gPtr.(type) {
		case *TextEdit:
			n = deduced

		case *std_widgets.QMainWindow:
			n = &TextEdit{QMainWindow: *deduced}

		default:
			n = new(TextEdit)
			n.SetPointer(ptr)
		}
	}
	return
}

//export callbackTextEdit3352e3_Constructor
func callbackTextEdit3352e3_Constructor(ptr unsafe.Pointer) {
	this := NewTextEditFromPointer(ptr)
	qt.Register(ptr, this)
}

func TextEdit_QRegisterMetaType() int {
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QRegisterMetaType()))
}

func (ptr *TextEdit) QRegisterMetaType() int {
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QRegisterMetaType()))
}

func TextEdit_QRegisterMetaType2(typeName string) int {
	var typeNameC *C.char
	if typeName != "" {
		typeNameC = C.CString(typeName)
		defer C.free(unsafe.Pointer(typeNameC))
	}
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QRegisterMetaType2(typeNameC)))
}

func (ptr *TextEdit) QRegisterMetaType2(typeName string) int {
	var typeNameC *C.char
	if typeName != "" {
		typeNameC = C.CString(typeName)
		defer C.free(unsafe.Pointer(typeNameC))
	}
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QRegisterMetaType2(typeNameC)))
}

func TextEdit_QmlRegisterType() int {
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QmlRegisterType()))
}

func (ptr *TextEdit) QmlRegisterType() int {
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QmlRegisterType()))
}

func TextEdit_QmlRegisterType2(uri string, versionMajor int, versionMinor int, qmlName string) int {
	var uriC *C.char
	if uri != "" {
		uriC = C.CString(uri)
		defer C.free(unsafe.Pointer(uriC))
	}
	var qmlNameC *C.char
	if qmlName != "" {
		qmlNameC = C.CString(qmlName)
		defer C.free(unsafe.Pointer(qmlNameC))
	}
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QmlRegisterType2(uriC, C.int(int32(versionMajor)), C.int(int32(versionMinor)), qmlNameC)))
}

func (ptr *TextEdit) QmlRegisterType2(uri string, versionMajor int, versionMinor int, qmlName string) int {
	var uriC *C.char
	if uri != "" {
		uriC = C.CString(uri)
		defer C.free(unsafe.Pointer(uriC))
	}
	var qmlNameC *C.char
	if qmlName != "" {
		qmlNameC = C.CString(qmlName)
		defer C.free(unsafe.Pointer(qmlNameC))
	}
	return int(int32(C.TextEdit3352e3_TextEdit3352e3_QmlRegisterType2(uriC, C.int(int32(versionMajor)), C.int(int32(versionMinor)), qmlNameC)))
}

func (ptr *TextEdit) __resizeDocks_docks_atList(i int) *std_widgets.QDockWidget {
	if ptr.Pointer() != nil {
		tmpValue := std_widgets.NewQDockWidgetFromPointer(C.TextEdit3352e3___resizeDocks_docks_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __resizeDocks_docks_setList(i std_widgets.QDockWidget_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___resizeDocks_docks_setList(ptr.Pointer(), std_widgets.PointerFromQDockWidget(i))
	}
}

func (ptr *TextEdit) __resizeDocks_docks_newList() unsafe.Pointer {
	return C.TextEdit3352e3___resizeDocks_docks_newList(ptr.Pointer())
}

func (ptr *TextEdit) __resizeDocks_sizes_atList(i int) int {
	if ptr.Pointer() != nil {
		return int(int32(C.TextEdit3352e3___resizeDocks_sizes_atList(ptr.Pointer(), C.int(int32(i)))))
	}
	return 0
}

func (ptr *TextEdit) __resizeDocks_sizes_setList(i int) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___resizeDocks_sizes_setList(ptr.Pointer(), C.int(int32(i)))
	}
}

func (ptr *TextEdit) __resizeDocks_sizes_newList() unsafe.Pointer {
	return C.TextEdit3352e3___resizeDocks_sizes_newList(ptr.Pointer())
}

func (ptr *TextEdit) __tabifiedDockWidgets_atList(i int) *std_widgets.QDockWidget {
	if ptr.Pointer() != nil {
		tmpValue := std_widgets.NewQDockWidgetFromPointer(C.TextEdit3352e3___tabifiedDockWidgets_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __tabifiedDockWidgets_setList(i std_widgets.QDockWidget_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___tabifiedDockWidgets_setList(ptr.Pointer(), std_widgets.PointerFromQDockWidget(i))
	}
}

func (ptr *TextEdit) __tabifiedDockWidgets_newList() unsafe.Pointer {
	return C.TextEdit3352e3___tabifiedDockWidgets_newList(ptr.Pointer())
}

func (ptr *TextEdit) __actions_atList(i int) *std_widgets.QAction {
	if ptr.Pointer() != nil {
		tmpValue := std_widgets.NewQActionFromPointer(C.TextEdit3352e3___actions_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __actions_setList(i std_widgets.QAction_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___actions_setList(ptr.Pointer(), std_widgets.PointerFromQAction(i))
	}
}

func (ptr *TextEdit) __actions_newList() unsafe.Pointer {
	return C.TextEdit3352e3___actions_newList(ptr.Pointer())
}

func (ptr *TextEdit) __addActions_actions_atList(i int) *std_widgets.QAction {
	if ptr.Pointer() != nil {
		tmpValue := std_widgets.NewQActionFromPointer(C.TextEdit3352e3___addActions_actions_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __addActions_actions_setList(i std_widgets.QAction_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___addActions_actions_setList(ptr.Pointer(), std_widgets.PointerFromQAction(i))
	}
}

func (ptr *TextEdit) __addActions_actions_newList() unsafe.Pointer {
	return C.TextEdit3352e3___addActions_actions_newList(ptr.Pointer())
}

func (ptr *TextEdit) __insertActions_actions_atList(i int) *std_widgets.QAction {
	if ptr.Pointer() != nil {
		tmpValue := std_widgets.NewQActionFromPointer(C.TextEdit3352e3___insertActions_actions_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __insertActions_actions_setList(i std_widgets.QAction_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___insertActions_actions_setList(ptr.Pointer(), std_widgets.PointerFromQAction(i))
	}
}

func (ptr *TextEdit) __insertActions_actions_newList() unsafe.Pointer {
	return C.TextEdit3352e3___insertActions_actions_newList(ptr.Pointer())
}

func (ptr *TextEdit) __children_atList(i int) *std_core.QObject {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQObjectFromPointer(C.TextEdit3352e3___children_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __children_setList(i std_core.QObject_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___children_setList(ptr.Pointer(), std_core.PointerFromQObject(i))
	}
}

func (ptr *TextEdit) __children_newList() unsafe.Pointer {
	return C.TextEdit3352e3___children_newList(ptr.Pointer())
}

func (ptr *TextEdit) __dynamicPropertyNames_atList(i int) *std_core.QByteArray {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQByteArrayFromPointer(C.TextEdit3352e3___dynamicPropertyNames_atList(ptr.Pointer(), C.int(int32(i))))
		runtime.SetFinalizer(tmpValue, (*std_core.QByteArray).DestroyQByteArray)
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __dynamicPropertyNames_setList(i std_core.QByteArray_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___dynamicPropertyNames_setList(ptr.Pointer(), std_core.PointerFromQByteArray(i))
	}
}

func (ptr *TextEdit) __dynamicPropertyNames_newList() unsafe.Pointer {
	return C.TextEdit3352e3___dynamicPropertyNames_newList(ptr.Pointer())
}

func (ptr *TextEdit) __findChildren_atList(i int) *std_core.QObject {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQObjectFromPointer(C.TextEdit3352e3___findChildren_atList(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __findChildren_setList(i std_core.QObject_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___findChildren_setList(ptr.Pointer(), std_core.PointerFromQObject(i))
	}
}

func (ptr *TextEdit) __findChildren_newList() unsafe.Pointer {
	return C.TextEdit3352e3___findChildren_newList(ptr.Pointer())
}

func (ptr *TextEdit) __findChildren_atList3(i int) *std_core.QObject {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQObjectFromPointer(C.TextEdit3352e3___findChildren_atList3(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __findChildren_setList3(i std_core.QObject_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___findChildren_setList3(ptr.Pointer(), std_core.PointerFromQObject(i))
	}
}

func (ptr *TextEdit) __findChildren_newList3() unsafe.Pointer {
	return C.TextEdit3352e3___findChildren_newList3(ptr.Pointer())
}

func (ptr *TextEdit) __qFindChildren_atList2(i int) *std_core.QObject {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQObjectFromPointer(C.TextEdit3352e3___qFindChildren_atList2(ptr.Pointer(), C.int(int32(i))))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

func (ptr *TextEdit) __qFindChildren_setList2(i std_core.QObject_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3___qFindChildren_setList2(ptr.Pointer(), std_core.PointerFromQObject(i))
	}
}

func (ptr *TextEdit) __qFindChildren_newList2() unsafe.Pointer {
	return C.TextEdit3352e3___qFindChildren_newList2(ptr.Pointer())
}

func NewTextEdit(parent std_widgets.QWidget_ITF, flags std_core.Qt__WindowType) *TextEdit {
	TextEdit_QRegisterMetaType()
	tmpValue := NewTextEditFromPointer(C.TextEdit3352e3_NewTextEdit(std_widgets.PointerFromQWidget(parent), C.longlong(flags)))
	if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
		tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
	}
	return tmpValue
}

//export callbackTextEdit3352e3_DestroyTextEdit
func callbackTextEdit3352e3_DestroyTextEdit(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "~TextEdit"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).DestroyTextEditDefault()
	}
}

func (ptr *TextEdit) ConnectDestroyTextEdit(f func()) {
	if ptr.Pointer() != nil {

		if signal := qt.LendSignal(ptr.Pointer(), "~TextEdit"); signal != nil {
			f := func() {
				(*(*func())(signal))()
				f()
			}
			qt.ConnectSignal(ptr.Pointer(), "~TextEdit", unsafe.Pointer(&f))
		} else {
			qt.ConnectSignal(ptr.Pointer(), "~TextEdit", unsafe.Pointer(&f))
		}
	}
}

func (ptr *TextEdit) DisconnectDestroyTextEdit() {
	if ptr.Pointer() != nil {

		qt.DisconnectSignal(ptr.Pointer(), "~TextEdit")
	}
}

func (ptr *TextEdit) DestroyTextEdit() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DestroyTextEdit(ptr.Pointer())
		ptr.SetPointer(nil)
		runtime.SetFinalizer(ptr, nil)
	}
}

func (ptr *TextEdit) DestroyTextEditDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DestroyTextEditDefault(ptr.Pointer())
		ptr.SetPointer(nil)
		runtime.SetFinalizer(ptr, nil)
	}
}

//export callbackTextEdit3352e3_ContextMenuEvent
func callbackTextEdit3352e3_ContextMenuEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "contextMenuEvent"); signal != nil {
		(*(*func(*std_gui.QContextMenuEvent))(signal))(std_gui.NewQContextMenuEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).ContextMenuEventDefault(std_gui.NewQContextMenuEventFromPointer(event))
	}
}

func (ptr *TextEdit) ContextMenuEventDefault(event std_gui.QContextMenuEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ContextMenuEventDefault(ptr.Pointer(), std_gui.PointerFromQContextMenuEvent(event))
	}
}

//export callbackTextEdit3352e3_CreatePopupMenu
func callbackTextEdit3352e3_CreatePopupMenu(ptr unsafe.Pointer) unsafe.Pointer {
	if signal := qt.GetSignal(ptr, "createPopupMenu"); signal != nil {
		return std_widgets.PointerFromQMenu((*(*func() *std_widgets.QMenu)(signal))())
	}

	return std_widgets.PointerFromQMenu(NewTextEditFromPointer(ptr).CreatePopupMenuDefault())
}

func (ptr *TextEdit) CreatePopupMenuDefault() *std_widgets.QMenu {
	if ptr.Pointer() != nil {
		tmpValue := std_widgets.NewQMenuFromPointer(C.TextEdit3352e3_CreatePopupMenuDefault(ptr.Pointer()))
		if !qt.ExistsSignal(tmpValue.Pointer(), "destroyed") {
			tmpValue.ConnectDestroyed(func(*std_core.QObject) { tmpValue.SetPointer(nil) })
		}
		return tmpValue
	}
	return nil
}

//export callbackTextEdit3352e3_Event
func callbackTextEdit3352e3_Event(ptr unsafe.Pointer, event unsafe.Pointer) C.char {
	if signal := qt.GetSignal(ptr, "event"); signal != nil {
		return C.char(int8(qt.GoBoolToInt((*(*func(*std_core.QEvent) bool)(signal))(std_core.NewQEventFromPointer(event)))))
	}

	return C.char(int8(qt.GoBoolToInt(NewTextEditFromPointer(ptr).EventDefault(std_core.NewQEventFromPointer(event)))))
}

func (ptr *TextEdit) EventDefault(event std_core.QEvent_ITF) bool {
	if ptr.Pointer() != nil {
		return int8(C.TextEdit3352e3_EventDefault(ptr.Pointer(), std_core.PointerFromQEvent(event))) != 0
	}
	return false
}

//export callbackTextEdit3352e3_IconSizeChanged
func callbackTextEdit3352e3_IconSizeChanged(ptr unsafe.Pointer, iconSize unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "iconSizeChanged"); signal != nil {
		(*(*func(*std_core.QSize))(signal))(std_core.NewQSizeFromPointer(iconSize))
	}

}

//export callbackTextEdit3352e3_SetAnimated
func callbackTextEdit3352e3_SetAnimated(ptr unsafe.Pointer, enabled C.char) {
	if signal := qt.GetSignal(ptr, "setAnimated"); signal != nil {
		(*(*func(bool))(signal))(int8(enabled) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetAnimatedDefault(int8(enabled) != 0)
	}
}

func (ptr *TextEdit) SetAnimatedDefault(enabled bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetAnimatedDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(enabled))))
	}
}

//export callbackTextEdit3352e3_SetDockNestingEnabled
func callbackTextEdit3352e3_SetDockNestingEnabled(ptr unsafe.Pointer, enabled C.char) {
	if signal := qt.GetSignal(ptr, "setDockNestingEnabled"); signal != nil {
		(*(*func(bool))(signal))(int8(enabled) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetDockNestingEnabledDefault(int8(enabled) != 0)
	}
}

func (ptr *TextEdit) SetDockNestingEnabledDefault(enabled bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetDockNestingEnabledDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(enabled))))
	}
}

//export callbackTextEdit3352e3_SetUnifiedTitleAndToolBarOnMac
func callbackTextEdit3352e3_SetUnifiedTitleAndToolBarOnMac(ptr unsafe.Pointer, set C.char) {
	if signal := qt.GetSignal(ptr, "setUnifiedTitleAndToolBarOnMac"); signal != nil {
		(*(*func(bool))(signal))(int8(set) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetUnifiedTitleAndToolBarOnMacDefault(int8(set) != 0)
	}
}

func (ptr *TextEdit) SetUnifiedTitleAndToolBarOnMacDefault(set bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetUnifiedTitleAndToolBarOnMacDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(set))))
	}
}

//export callbackTextEdit3352e3_TabifiedDockWidgetActivated
func callbackTextEdit3352e3_TabifiedDockWidgetActivated(ptr unsafe.Pointer, dockWidget unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "tabifiedDockWidgetActivated"); signal != nil {
		(*(*func(*std_widgets.QDockWidget))(signal))(std_widgets.NewQDockWidgetFromPointer(dockWidget))
	}

}

//export callbackTextEdit3352e3_ToolButtonStyleChanged
func callbackTextEdit3352e3_ToolButtonStyleChanged(ptr unsafe.Pointer, toolButtonStyle C.longlong) {
	if signal := qt.GetSignal(ptr, "toolButtonStyleChanged"); signal != nil {
		(*(*func(std_core.Qt__ToolButtonStyle))(signal))(std_core.Qt__ToolButtonStyle(toolButtonStyle))
	}

}

//export callbackTextEdit3352e3_ActionEvent
func callbackTextEdit3352e3_ActionEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "actionEvent"); signal != nil {
		(*(*func(*std_gui.QActionEvent))(signal))(std_gui.NewQActionEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).ActionEventDefault(std_gui.NewQActionEventFromPointer(event))
	}
}

func (ptr *TextEdit) ActionEventDefault(event std_gui.QActionEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ActionEventDefault(ptr.Pointer(), std_gui.PointerFromQActionEvent(event))
	}
}

//export callbackTextEdit3352e3_ChangeEvent
func callbackTextEdit3352e3_ChangeEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "changeEvent"); signal != nil {
		(*(*func(*std_core.QEvent))(signal))(std_core.NewQEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).ChangeEventDefault(std_core.NewQEventFromPointer(event))
	}
}

func (ptr *TextEdit) ChangeEventDefault(event std_core.QEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ChangeEventDefault(ptr.Pointer(), std_core.PointerFromQEvent(event))
	}
}

//export callbackTextEdit3352e3_Close
func callbackTextEdit3352e3_Close(ptr unsafe.Pointer) C.char {
	if signal := qt.GetSignal(ptr, "close"); signal != nil {
		return C.char(int8(qt.GoBoolToInt((*(*func() bool)(signal))())))
	}

	return C.char(int8(qt.GoBoolToInt(NewTextEditFromPointer(ptr).CloseDefault())))
}

func (ptr *TextEdit) CloseDefault() bool {
	if ptr.Pointer() != nil {
		return int8(C.TextEdit3352e3_CloseDefault(ptr.Pointer())) != 0
	}
	return false
}

//export callbackTextEdit3352e3_CloseEvent
func callbackTextEdit3352e3_CloseEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "closeEvent"); signal != nil {
		(*(*func(*std_gui.QCloseEvent))(signal))(std_gui.NewQCloseEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).CloseEventDefault(std_gui.NewQCloseEventFromPointer(event))
	}
}

func (ptr *TextEdit) CloseEventDefault(event std_gui.QCloseEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_CloseEventDefault(ptr.Pointer(), std_gui.PointerFromQCloseEvent(event))
	}
}

//export callbackTextEdit3352e3_CustomContextMenuRequested
func callbackTextEdit3352e3_CustomContextMenuRequested(ptr unsafe.Pointer, pos unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "customContextMenuRequested"); signal != nil {
		(*(*func(*std_core.QPoint))(signal))(std_core.NewQPointFromPointer(pos))
	}

}

//export callbackTextEdit3352e3_DragEnterEvent
func callbackTextEdit3352e3_DragEnterEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "dragEnterEvent"); signal != nil {
		(*(*func(*std_gui.QDragEnterEvent))(signal))(std_gui.NewQDragEnterEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).DragEnterEventDefault(std_gui.NewQDragEnterEventFromPointer(event))
	}
}

func (ptr *TextEdit) DragEnterEventDefault(event std_gui.QDragEnterEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DragEnterEventDefault(ptr.Pointer(), std_gui.PointerFromQDragEnterEvent(event))
	}
}

//export callbackTextEdit3352e3_DragLeaveEvent
func callbackTextEdit3352e3_DragLeaveEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "dragLeaveEvent"); signal != nil {
		(*(*func(*std_gui.QDragLeaveEvent))(signal))(std_gui.NewQDragLeaveEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).DragLeaveEventDefault(std_gui.NewQDragLeaveEventFromPointer(event))
	}
}

func (ptr *TextEdit) DragLeaveEventDefault(event std_gui.QDragLeaveEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DragLeaveEventDefault(ptr.Pointer(), std_gui.PointerFromQDragLeaveEvent(event))
	}
}

//export callbackTextEdit3352e3_DragMoveEvent
func callbackTextEdit3352e3_DragMoveEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "dragMoveEvent"); signal != nil {
		(*(*func(*std_gui.QDragMoveEvent))(signal))(std_gui.NewQDragMoveEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).DragMoveEventDefault(std_gui.NewQDragMoveEventFromPointer(event))
	}
}

func (ptr *TextEdit) DragMoveEventDefault(event std_gui.QDragMoveEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DragMoveEventDefault(ptr.Pointer(), std_gui.PointerFromQDragMoveEvent(event))
	}
}

//export callbackTextEdit3352e3_DropEvent
func callbackTextEdit3352e3_DropEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "dropEvent"); signal != nil {
		(*(*func(*std_gui.QDropEvent))(signal))(std_gui.NewQDropEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).DropEventDefault(std_gui.NewQDropEventFromPointer(event))
	}
}

func (ptr *TextEdit) DropEventDefault(event std_gui.QDropEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DropEventDefault(ptr.Pointer(), std_gui.PointerFromQDropEvent(event))
	}
}

//export callbackTextEdit3352e3_EnterEvent
func callbackTextEdit3352e3_EnterEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "enterEvent"); signal != nil {
		(*(*func(*std_core.QEvent))(signal))(std_core.NewQEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).EnterEventDefault(std_core.NewQEventFromPointer(event))
	}
}

func (ptr *TextEdit) EnterEventDefault(event std_core.QEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_EnterEventDefault(ptr.Pointer(), std_core.PointerFromQEvent(event))
	}
}

//export callbackTextEdit3352e3_FocusInEvent
func callbackTextEdit3352e3_FocusInEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "focusInEvent"); signal != nil {
		(*(*func(*std_gui.QFocusEvent))(signal))(std_gui.NewQFocusEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).FocusInEventDefault(std_gui.NewQFocusEventFromPointer(event))
	}
}

func (ptr *TextEdit) FocusInEventDefault(event std_gui.QFocusEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_FocusInEventDefault(ptr.Pointer(), std_gui.PointerFromQFocusEvent(event))
	}
}

//export callbackTextEdit3352e3_FocusNextPrevChild
func callbackTextEdit3352e3_FocusNextPrevChild(ptr unsafe.Pointer, next C.char) C.char {
	if signal := qt.GetSignal(ptr, "focusNextPrevChild"); signal != nil {
		return C.char(int8(qt.GoBoolToInt((*(*func(bool) bool)(signal))(int8(next) != 0))))
	}

	return C.char(int8(qt.GoBoolToInt(NewTextEditFromPointer(ptr).FocusNextPrevChildDefault(int8(next) != 0))))
}

func (ptr *TextEdit) FocusNextPrevChildDefault(next bool) bool {
	if ptr.Pointer() != nil {
		return int8(C.TextEdit3352e3_FocusNextPrevChildDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(next))))) != 0
	}
	return false
}

//export callbackTextEdit3352e3_FocusOutEvent
func callbackTextEdit3352e3_FocusOutEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "focusOutEvent"); signal != nil {
		(*(*func(*std_gui.QFocusEvent))(signal))(std_gui.NewQFocusEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).FocusOutEventDefault(std_gui.NewQFocusEventFromPointer(event))
	}
}

func (ptr *TextEdit) FocusOutEventDefault(event std_gui.QFocusEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_FocusOutEventDefault(ptr.Pointer(), std_gui.PointerFromQFocusEvent(event))
	}
}

//export callbackTextEdit3352e3_HasHeightForWidth
func callbackTextEdit3352e3_HasHeightForWidth(ptr unsafe.Pointer) C.char {
	if signal := qt.GetSignal(ptr, "hasHeightForWidth"); signal != nil {
		return C.char(int8(qt.GoBoolToInt((*(*func() bool)(signal))())))
	}

	return C.char(int8(qt.GoBoolToInt(NewTextEditFromPointer(ptr).HasHeightForWidthDefault())))
}

func (ptr *TextEdit) HasHeightForWidthDefault() bool {
	if ptr.Pointer() != nil {
		return int8(C.TextEdit3352e3_HasHeightForWidthDefault(ptr.Pointer())) != 0
	}
	return false
}

//export callbackTextEdit3352e3_HeightForWidth
func callbackTextEdit3352e3_HeightForWidth(ptr unsafe.Pointer, w C.int) C.int {
	if signal := qt.GetSignal(ptr, "heightForWidth"); signal != nil {
		return C.int(int32((*(*func(int) int)(signal))(int(int32(w)))))
	}

	return C.int(int32(NewTextEditFromPointer(ptr).HeightForWidthDefault(int(int32(w)))))
}

func (ptr *TextEdit) HeightForWidthDefault(w int) int {
	if ptr.Pointer() != nil {
		return int(int32(C.TextEdit3352e3_HeightForWidthDefault(ptr.Pointer(), C.int(int32(w)))))
	}
	return 0
}

//export callbackTextEdit3352e3_Hide
func callbackTextEdit3352e3_Hide(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "hide"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).HideDefault()
	}
}

func (ptr *TextEdit) HideDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_HideDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_HideEvent
func callbackTextEdit3352e3_HideEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "hideEvent"); signal != nil {
		(*(*func(*std_gui.QHideEvent))(signal))(std_gui.NewQHideEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).HideEventDefault(std_gui.NewQHideEventFromPointer(event))
	}
}

func (ptr *TextEdit) HideEventDefault(event std_gui.QHideEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_HideEventDefault(ptr.Pointer(), std_gui.PointerFromQHideEvent(event))
	}
}

//export callbackTextEdit3352e3_InitPainter
func callbackTextEdit3352e3_InitPainter(ptr unsafe.Pointer, painter unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "initPainter"); signal != nil {
		(*(*func(*std_gui.QPainter))(signal))(std_gui.NewQPainterFromPointer(painter))
	} else {
		NewTextEditFromPointer(ptr).InitPainterDefault(std_gui.NewQPainterFromPointer(painter))
	}
}

func (ptr *TextEdit) InitPainterDefault(painter std_gui.QPainter_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_InitPainterDefault(ptr.Pointer(), std_gui.PointerFromQPainter(painter))
	}
}

//export callbackTextEdit3352e3_InputMethodEvent
func callbackTextEdit3352e3_InputMethodEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "inputMethodEvent"); signal != nil {
		(*(*func(*std_gui.QInputMethodEvent))(signal))(std_gui.NewQInputMethodEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).InputMethodEventDefault(std_gui.NewQInputMethodEventFromPointer(event))
	}
}

func (ptr *TextEdit) InputMethodEventDefault(event std_gui.QInputMethodEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_InputMethodEventDefault(ptr.Pointer(), std_gui.PointerFromQInputMethodEvent(event))
	}
}

//export callbackTextEdit3352e3_InputMethodQuery
func callbackTextEdit3352e3_InputMethodQuery(ptr unsafe.Pointer, query C.longlong) unsafe.Pointer {
	if signal := qt.GetSignal(ptr, "inputMethodQuery"); signal != nil {
		return std_core.PointerFromQVariant((*(*func(std_core.Qt__InputMethodQuery) *std_core.QVariant)(signal))(std_core.Qt__InputMethodQuery(query)))
	}

	return std_core.PointerFromQVariant(NewTextEditFromPointer(ptr).InputMethodQueryDefault(std_core.Qt__InputMethodQuery(query)))
}

func (ptr *TextEdit) InputMethodQueryDefault(query std_core.Qt__InputMethodQuery) *std_core.QVariant {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQVariantFromPointer(C.TextEdit3352e3_InputMethodQueryDefault(ptr.Pointer(), C.longlong(query)))
		runtime.SetFinalizer(tmpValue, (*std_core.QVariant).DestroyQVariant)
		return tmpValue
	}
	return nil
}

//export callbackTextEdit3352e3_KeyPressEvent
func callbackTextEdit3352e3_KeyPressEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "keyPressEvent"); signal != nil {
		(*(*func(*std_gui.QKeyEvent))(signal))(std_gui.NewQKeyEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).KeyPressEventDefault(std_gui.NewQKeyEventFromPointer(event))
	}
}

func (ptr *TextEdit) KeyPressEventDefault(event std_gui.QKeyEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_KeyPressEventDefault(ptr.Pointer(), std_gui.PointerFromQKeyEvent(event))
	}
}

//export callbackTextEdit3352e3_KeyReleaseEvent
func callbackTextEdit3352e3_KeyReleaseEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "keyReleaseEvent"); signal != nil {
		(*(*func(*std_gui.QKeyEvent))(signal))(std_gui.NewQKeyEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).KeyReleaseEventDefault(std_gui.NewQKeyEventFromPointer(event))
	}
}

func (ptr *TextEdit) KeyReleaseEventDefault(event std_gui.QKeyEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_KeyReleaseEventDefault(ptr.Pointer(), std_gui.PointerFromQKeyEvent(event))
	}
}

//export callbackTextEdit3352e3_LeaveEvent
func callbackTextEdit3352e3_LeaveEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "leaveEvent"); signal != nil {
		(*(*func(*std_core.QEvent))(signal))(std_core.NewQEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).LeaveEventDefault(std_core.NewQEventFromPointer(event))
	}
}

func (ptr *TextEdit) LeaveEventDefault(event std_core.QEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_LeaveEventDefault(ptr.Pointer(), std_core.PointerFromQEvent(event))
	}
}

//export callbackTextEdit3352e3_Lower
func callbackTextEdit3352e3_Lower(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "lower"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).LowerDefault()
	}
}

func (ptr *TextEdit) LowerDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_LowerDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_Metric
func callbackTextEdit3352e3_Metric(ptr unsafe.Pointer, m C.longlong) C.int {
	if signal := qt.GetSignal(ptr, "metric"); signal != nil {
		return C.int(int32((*(*func(std_gui.QPaintDevice__PaintDeviceMetric) int)(signal))(std_gui.QPaintDevice__PaintDeviceMetric(m))))
	}

	return C.int(int32(NewTextEditFromPointer(ptr).MetricDefault(std_gui.QPaintDevice__PaintDeviceMetric(m))))
}

func (ptr *TextEdit) MetricDefault(m std_gui.QPaintDevice__PaintDeviceMetric) int {
	if ptr.Pointer() != nil {
		return int(int32(C.TextEdit3352e3_MetricDefault(ptr.Pointer(), C.longlong(m))))
	}
	return 0
}

//export callbackTextEdit3352e3_MinimumSizeHint
func callbackTextEdit3352e3_MinimumSizeHint(ptr unsafe.Pointer) unsafe.Pointer {
	if signal := qt.GetSignal(ptr, "minimumSizeHint"); signal != nil {
		return std_core.PointerFromQSize((*(*func() *std_core.QSize)(signal))())
	}

	return std_core.PointerFromQSize(NewTextEditFromPointer(ptr).MinimumSizeHintDefault())
}

func (ptr *TextEdit) MinimumSizeHintDefault() *std_core.QSize {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQSizeFromPointer(C.TextEdit3352e3_MinimumSizeHintDefault(ptr.Pointer()))
		runtime.SetFinalizer(tmpValue, (*std_core.QSize).DestroyQSize)
		return tmpValue
	}
	return nil
}

//export callbackTextEdit3352e3_MouseDoubleClickEvent
func callbackTextEdit3352e3_MouseDoubleClickEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "mouseDoubleClickEvent"); signal != nil {
		(*(*func(*std_gui.QMouseEvent))(signal))(std_gui.NewQMouseEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).MouseDoubleClickEventDefault(std_gui.NewQMouseEventFromPointer(event))
	}
}

func (ptr *TextEdit) MouseDoubleClickEventDefault(event std_gui.QMouseEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_MouseDoubleClickEventDefault(ptr.Pointer(), std_gui.PointerFromQMouseEvent(event))
	}
}

//export callbackTextEdit3352e3_MouseMoveEvent
func callbackTextEdit3352e3_MouseMoveEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "mouseMoveEvent"); signal != nil {
		(*(*func(*std_gui.QMouseEvent))(signal))(std_gui.NewQMouseEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).MouseMoveEventDefault(std_gui.NewQMouseEventFromPointer(event))
	}
}

func (ptr *TextEdit) MouseMoveEventDefault(event std_gui.QMouseEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_MouseMoveEventDefault(ptr.Pointer(), std_gui.PointerFromQMouseEvent(event))
	}
}

//export callbackTextEdit3352e3_MousePressEvent
func callbackTextEdit3352e3_MousePressEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "mousePressEvent"); signal != nil {
		(*(*func(*std_gui.QMouseEvent))(signal))(std_gui.NewQMouseEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).MousePressEventDefault(std_gui.NewQMouseEventFromPointer(event))
	}
}

func (ptr *TextEdit) MousePressEventDefault(event std_gui.QMouseEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_MousePressEventDefault(ptr.Pointer(), std_gui.PointerFromQMouseEvent(event))
	}
}

//export callbackTextEdit3352e3_MouseReleaseEvent
func callbackTextEdit3352e3_MouseReleaseEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "mouseReleaseEvent"); signal != nil {
		(*(*func(*std_gui.QMouseEvent))(signal))(std_gui.NewQMouseEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).MouseReleaseEventDefault(std_gui.NewQMouseEventFromPointer(event))
	}
}

func (ptr *TextEdit) MouseReleaseEventDefault(event std_gui.QMouseEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_MouseReleaseEventDefault(ptr.Pointer(), std_gui.PointerFromQMouseEvent(event))
	}
}

//export callbackTextEdit3352e3_MoveEvent
func callbackTextEdit3352e3_MoveEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "moveEvent"); signal != nil {
		(*(*func(*std_gui.QMoveEvent))(signal))(std_gui.NewQMoveEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).MoveEventDefault(std_gui.NewQMoveEventFromPointer(event))
	}
}

func (ptr *TextEdit) MoveEventDefault(event std_gui.QMoveEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_MoveEventDefault(ptr.Pointer(), std_gui.PointerFromQMoveEvent(event))
	}
}

//export callbackTextEdit3352e3_NativeEvent
func callbackTextEdit3352e3_NativeEvent(ptr unsafe.Pointer, eventType unsafe.Pointer, message unsafe.Pointer, result *C.long) C.char {
	var resultR int
	if result != nil {
		resultR = int(int32(*result))
		defer func() { *result = C.long(int32(resultR)) }()
	}
	if signal := qt.GetSignal(ptr, "nativeEvent"); signal != nil {
		return C.char(int8(qt.GoBoolToInt((*(*func(*std_core.QByteArray, unsafe.Pointer, *int) bool)(signal))(std_core.NewQByteArrayFromPointer(eventType), message, &resultR))))
	}

	return C.char(int8(qt.GoBoolToInt(NewTextEditFromPointer(ptr).NativeEventDefault(std_core.NewQByteArrayFromPointer(eventType), message, &resultR))))
}

func (ptr *TextEdit) NativeEventDefault(eventType std_core.QByteArray_ITF, message unsafe.Pointer, result *int) bool {
	if ptr.Pointer() != nil {
		var resultC C.long
		if result != nil {
			resultC = C.long(int32(*result))
			defer func() { *result = int(int32(resultC)) }()
		}
		return int8(C.TextEdit3352e3_NativeEventDefault(ptr.Pointer(), std_core.PointerFromQByteArray(eventType), message, &resultC)) != 0
	}
	return false
}

//export callbackTextEdit3352e3_PaintEngine
func callbackTextEdit3352e3_PaintEngine(ptr unsafe.Pointer) unsafe.Pointer {
	if signal := qt.GetSignal(ptr, "paintEngine"); signal != nil {
		return std_gui.PointerFromQPaintEngine((*(*func() *std_gui.QPaintEngine)(signal))())
	}

	return std_gui.PointerFromQPaintEngine(NewTextEditFromPointer(ptr).PaintEngineDefault())
}

func (ptr *TextEdit) PaintEngineDefault() *std_gui.QPaintEngine {
	if ptr.Pointer() != nil {
		return std_gui.NewQPaintEngineFromPointer(C.TextEdit3352e3_PaintEngineDefault(ptr.Pointer()))
	}
	return nil
}

//export callbackTextEdit3352e3_PaintEvent
func callbackTextEdit3352e3_PaintEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "paintEvent"); signal != nil {
		(*(*func(*std_gui.QPaintEvent))(signal))(std_gui.NewQPaintEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).PaintEventDefault(std_gui.NewQPaintEventFromPointer(event))
	}
}

func (ptr *TextEdit) PaintEventDefault(event std_gui.QPaintEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_PaintEventDefault(ptr.Pointer(), std_gui.PointerFromQPaintEvent(event))
	}
}

//export callbackTextEdit3352e3_Raise
func callbackTextEdit3352e3_Raise(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "raise"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).RaiseDefault()
	}
}

func (ptr *TextEdit) RaiseDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_RaiseDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_Repaint
func callbackTextEdit3352e3_Repaint(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "repaint"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).RepaintDefault()
	}
}

func (ptr *TextEdit) RepaintDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_RepaintDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_ResizeEvent
func callbackTextEdit3352e3_ResizeEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "resizeEvent"); signal != nil {
		(*(*func(*std_gui.QResizeEvent))(signal))(std_gui.NewQResizeEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).ResizeEventDefault(std_gui.NewQResizeEventFromPointer(event))
	}
}

func (ptr *TextEdit) ResizeEventDefault(event std_gui.QResizeEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ResizeEventDefault(ptr.Pointer(), std_gui.PointerFromQResizeEvent(event))
	}
}

//export callbackTextEdit3352e3_SetDisabled
func callbackTextEdit3352e3_SetDisabled(ptr unsafe.Pointer, disable C.char) {
	if signal := qt.GetSignal(ptr, "setDisabled"); signal != nil {
		(*(*func(bool))(signal))(int8(disable) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetDisabledDefault(int8(disable) != 0)
	}
}

func (ptr *TextEdit) SetDisabledDefault(disable bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetDisabledDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(disable))))
	}
}

//export callbackTextEdit3352e3_SetEnabled
func callbackTextEdit3352e3_SetEnabled(ptr unsafe.Pointer, vbo C.char) {
	if signal := qt.GetSignal(ptr, "setEnabled"); signal != nil {
		(*(*func(bool))(signal))(int8(vbo) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetEnabledDefault(int8(vbo) != 0)
	}
}

func (ptr *TextEdit) SetEnabledDefault(vbo bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetEnabledDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(vbo))))
	}
}

//export callbackTextEdit3352e3_SetFocus2
func callbackTextEdit3352e3_SetFocus2(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "setFocus2"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).SetFocus2Default()
	}
}

func (ptr *TextEdit) SetFocus2Default() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetFocus2Default(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_SetHidden
func callbackTextEdit3352e3_SetHidden(ptr unsafe.Pointer, hidden C.char) {
	if signal := qt.GetSignal(ptr, "setHidden"); signal != nil {
		(*(*func(bool))(signal))(int8(hidden) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetHiddenDefault(int8(hidden) != 0)
	}
}

func (ptr *TextEdit) SetHiddenDefault(hidden bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetHiddenDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(hidden))))
	}
}

//export callbackTextEdit3352e3_SetStyleSheet
func callbackTextEdit3352e3_SetStyleSheet(ptr unsafe.Pointer, styleSheet C.struct_Moc_PackedString) {
	if signal := qt.GetSignal(ptr, "setStyleSheet"); signal != nil {
		(*(*func(string))(signal))(cGoUnpackString(styleSheet))
	} else {
		NewTextEditFromPointer(ptr).SetStyleSheetDefault(cGoUnpackString(styleSheet))
	}
}

func (ptr *TextEdit) SetStyleSheetDefault(styleSheet string) {
	if ptr.Pointer() != nil {
		var styleSheetC *C.char
		if styleSheet != "" {
			styleSheetC = C.CString(styleSheet)
			defer C.free(unsafe.Pointer(styleSheetC))
		}
		C.TextEdit3352e3_SetStyleSheetDefault(ptr.Pointer(), C.struct_Moc_PackedString{data: styleSheetC, len: C.longlong(len(styleSheet))})
	}
}

//export callbackTextEdit3352e3_SetVisible
func callbackTextEdit3352e3_SetVisible(ptr unsafe.Pointer, visible C.char) {
	if signal := qt.GetSignal(ptr, "setVisible"); signal != nil {
		(*(*func(bool))(signal))(int8(visible) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetVisibleDefault(int8(visible) != 0)
	}
}

func (ptr *TextEdit) SetVisibleDefault(visible bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetVisibleDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(visible))))
	}
}

//export callbackTextEdit3352e3_SetWindowModified
func callbackTextEdit3352e3_SetWindowModified(ptr unsafe.Pointer, vbo C.char) {
	if signal := qt.GetSignal(ptr, "setWindowModified"); signal != nil {
		(*(*func(bool))(signal))(int8(vbo) != 0)
	} else {
		NewTextEditFromPointer(ptr).SetWindowModifiedDefault(int8(vbo) != 0)
	}
}

func (ptr *TextEdit) SetWindowModifiedDefault(vbo bool) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_SetWindowModifiedDefault(ptr.Pointer(), C.char(int8(qt.GoBoolToInt(vbo))))
	}
}

//export callbackTextEdit3352e3_SetWindowTitle
func callbackTextEdit3352e3_SetWindowTitle(ptr unsafe.Pointer, vqs C.struct_Moc_PackedString) {
	if signal := qt.GetSignal(ptr, "setWindowTitle"); signal != nil {
		(*(*func(string))(signal))(cGoUnpackString(vqs))
	} else {
		NewTextEditFromPointer(ptr).SetWindowTitleDefault(cGoUnpackString(vqs))
	}
}

func (ptr *TextEdit) SetWindowTitleDefault(vqs string) {
	if ptr.Pointer() != nil {
		var vqsC *C.char
		if vqs != "" {
			vqsC = C.CString(vqs)
			defer C.free(unsafe.Pointer(vqsC))
		}
		C.TextEdit3352e3_SetWindowTitleDefault(ptr.Pointer(), C.struct_Moc_PackedString{data: vqsC, len: C.longlong(len(vqs))})
	}
}

//export callbackTextEdit3352e3_Show
func callbackTextEdit3352e3_Show(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "show"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).ShowDefault()
	}
}

func (ptr *TextEdit) ShowDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ShowDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_ShowEvent
func callbackTextEdit3352e3_ShowEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "showEvent"); signal != nil {
		(*(*func(*std_gui.QShowEvent))(signal))(std_gui.NewQShowEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).ShowEventDefault(std_gui.NewQShowEventFromPointer(event))
	}
}

func (ptr *TextEdit) ShowEventDefault(event std_gui.QShowEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ShowEventDefault(ptr.Pointer(), std_gui.PointerFromQShowEvent(event))
	}
}

//export callbackTextEdit3352e3_ShowFullScreen
func callbackTextEdit3352e3_ShowFullScreen(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "showFullScreen"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).ShowFullScreenDefault()
	}
}

func (ptr *TextEdit) ShowFullScreenDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ShowFullScreenDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_ShowMaximized
func callbackTextEdit3352e3_ShowMaximized(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "showMaximized"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).ShowMaximizedDefault()
	}
}

func (ptr *TextEdit) ShowMaximizedDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ShowMaximizedDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_ShowMinimized
func callbackTextEdit3352e3_ShowMinimized(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "showMinimized"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).ShowMinimizedDefault()
	}
}

func (ptr *TextEdit) ShowMinimizedDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ShowMinimizedDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_ShowNormal
func callbackTextEdit3352e3_ShowNormal(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "showNormal"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).ShowNormalDefault()
	}
}

func (ptr *TextEdit) ShowNormalDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ShowNormalDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_SizeHint
func callbackTextEdit3352e3_SizeHint(ptr unsafe.Pointer) unsafe.Pointer {
	if signal := qt.GetSignal(ptr, "sizeHint"); signal != nil {
		return std_core.PointerFromQSize((*(*func() *std_core.QSize)(signal))())
	}

	return std_core.PointerFromQSize(NewTextEditFromPointer(ptr).SizeHintDefault())
}

func (ptr *TextEdit) SizeHintDefault() *std_core.QSize {
	if ptr.Pointer() != nil {
		tmpValue := std_core.NewQSizeFromPointer(C.TextEdit3352e3_SizeHintDefault(ptr.Pointer()))
		runtime.SetFinalizer(tmpValue, (*std_core.QSize).DestroyQSize)
		return tmpValue
	}
	return nil
}

//export callbackTextEdit3352e3_TabletEvent
func callbackTextEdit3352e3_TabletEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "tabletEvent"); signal != nil {
		(*(*func(*std_gui.QTabletEvent))(signal))(std_gui.NewQTabletEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).TabletEventDefault(std_gui.NewQTabletEventFromPointer(event))
	}
}

func (ptr *TextEdit) TabletEventDefault(event std_gui.QTabletEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_TabletEventDefault(ptr.Pointer(), std_gui.PointerFromQTabletEvent(event))
	}
}

//export callbackTextEdit3352e3_Update
func callbackTextEdit3352e3_Update(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "update"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).UpdateDefault()
	}
}

func (ptr *TextEdit) UpdateDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_UpdateDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_UpdateMicroFocus
func callbackTextEdit3352e3_UpdateMicroFocus(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "updateMicroFocus"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).UpdateMicroFocusDefault()
	}
}

func (ptr *TextEdit) UpdateMicroFocusDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_UpdateMicroFocusDefault(ptr.Pointer())
	}
}

//export callbackTextEdit3352e3_WheelEvent
func callbackTextEdit3352e3_WheelEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "wheelEvent"); signal != nil {
		(*(*func(*std_gui.QWheelEvent))(signal))(std_gui.NewQWheelEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).WheelEventDefault(std_gui.NewQWheelEventFromPointer(event))
	}
}

func (ptr *TextEdit) WheelEventDefault(event std_gui.QWheelEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_WheelEventDefault(ptr.Pointer(), std_gui.PointerFromQWheelEvent(event))
	}
}

//export callbackTextEdit3352e3_WindowIconChanged
func callbackTextEdit3352e3_WindowIconChanged(ptr unsafe.Pointer, icon unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "windowIconChanged"); signal != nil {
		(*(*func(*std_gui.QIcon))(signal))(std_gui.NewQIconFromPointer(icon))
	}

}

//export callbackTextEdit3352e3_WindowTitleChanged
func callbackTextEdit3352e3_WindowTitleChanged(ptr unsafe.Pointer, title C.struct_Moc_PackedString) {
	if signal := qt.GetSignal(ptr, "windowTitleChanged"); signal != nil {
		(*(*func(string))(signal))(cGoUnpackString(title))
	}

}

//export callbackTextEdit3352e3_ChildEvent
func callbackTextEdit3352e3_ChildEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "childEvent"); signal != nil {
		(*(*func(*std_core.QChildEvent))(signal))(std_core.NewQChildEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).ChildEventDefault(std_core.NewQChildEventFromPointer(event))
	}
}

func (ptr *TextEdit) ChildEventDefault(event std_core.QChildEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ChildEventDefault(ptr.Pointer(), std_core.PointerFromQChildEvent(event))
	}
}

//export callbackTextEdit3352e3_ConnectNotify
func callbackTextEdit3352e3_ConnectNotify(ptr unsafe.Pointer, sign unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "connectNotify"); signal != nil {
		(*(*func(*std_core.QMetaMethod))(signal))(std_core.NewQMetaMethodFromPointer(sign))
	} else {
		NewTextEditFromPointer(ptr).ConnectNotifyDefault(std_core.NewQMetaMethodFromPointer(sign))
	}
}

func (ptr *TextEdit) ConnectNotifyDefault(sign std_core.QMetaMethod_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_ConnectNotifyDefault(ptr.Pointer(), std_core.PointerFromQMetaMethod(sign))
	}
}

//export callbackTextEdit3352e3_CustomEvent
func callbackTextEdit3352e3_CustomEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "customEvent"); signal != nil {
		(*(*func(*std_core.QEvent))(signal))(std_core.NewQEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).CustomEventDefault(std_core.NewQEventFromPointer(event))
	}
}

func (ptr *TextEdit) CustomEventDefault(event std_core.QEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_CustomEventDefault(ptr.Pointer(), std_core.PointerFromQEvent(event))
	}
}

//export callbackTextEdit3352e3_DeleteLater
func callbackTextEdit3352e3_DeleteLater(ptr unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "deleteLater"); signal != nil {
		(*(*func())(signal))()
	} else {
		NewTextEditFromPointer(ptr).DeleteLaterDefault()
	}
}

func (ptr *TextEdit) DeleteLaterDefault() {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DeleteLaterDefault(ptr.Pointer())
		runtime.SetFinalizer(ptr, nil)
	}
}

//export callbackTextEdit3352e3_Destroyed
func callbackTextEdit3352e3_Destroyed(ptr unsafe.Pointer, obj unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "destroyed"); signal != nil {
		(*(*func(*std_core.QObject))(signal))(std_core.NewQObjectFromPointer(obj))
	}
	qt.Unregister(ptr)

}

//export callbackTextEdit3352e3_DisconnectNotify
func callbackTextEdit3352e3_DisconnectNotify(ptr unsafe.Pointer, sign unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "disconnectNotify"); signal != nil {
		(*(*func(*std_core.QMetaMethod))(signal))(std_core.NewQMetaMethodFromPointer(sign))
	} else {
		NewTextEditFromPointer(ptr).DisconnectNotifyDefault(std_core.NewQMetaMethodFromPointer(sign))
	}
}

func (ptr *TextEdit) DisconnectNotifyDefault(sign std_core.QMetaMethod_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_DisconnectNotifyDefault(ptr.Pointer(), std_core.PointerFromQMetaMethod(sign))
	}
}

//export callbackTextEdit3352e3_EventFilter
func callbackTextEdit3352e3_EventFilter(ptr unsafe.Pointer, watched unsafe.Pointer, event unsafe.Pointer) C.char {
	if signal := qt.GetSignal(ptr, "eventFilter"); signal != nil {
		return C.char(int8(qt.GoBoolToInt((*(*func(*std_core.QObject, *std_core.QEvent) bool)(signal))(std_core.NewQObjectFromPointer(watched), std_core.NewQEventFromPointer(event)))))
	}

	return C.char(int8(qt.GoBoolToInt(NewTextEditFromPointer(ptr).EventFilterDefault(std_core.NewQObjectFromPointer(watched), std_core.NewQEventFromPointer(event)))))
}

func (ptr *TextEdit) EventFilterDefault(watched std_core.QObject_ITF, event std_core.QEvent_ITF) bool {
	if ptr.Pointer() != nil {
		return int8(C.TextEdit3352e3_EventFilterDefault(ptr.Pointer(), std_core.PointerFromQObject(watched), std_core.PointerFromQEvent(event))) != 0
	}
	return false
}

//export callbackTextEdit3352e3_ObjectNameChanged
func callbackTextEdit3352e3_ObjectNameChanged(ptr unsafe.Pointer, objectName C.struct_Moc_PackedString) {
	if signal := qt.GetSignal(ptr, "objectNameChanged"); signal != nil {
		(*(*func(string))(signal))(cGoUnpackString(objectName))
	}

}

//export callbackTextEdit3352e3_TimerEvent
func callbackTextEdit3352e3_TimerEvent(ptr unsafe.Pointer, event unsafe.Pointer) {
	if signal := qt.GetSignal(ptr, "timerEvent"); signal != nil {
		(*(*func(*std_core.QTimerEvent))(signal))(std_core.NewQTimerEventFromPointer(event))
	} else {
		NewTextEditFromPointer(ptr).TimerEventDefault(std_core.NewQTimerEventFromPointer(event))
	}
}

func (ptr *TextEdit) TimerEventDefault(event std_core.QTimerEvent_ITF) {
	if ptr.Pointer() != nil {
		C.TextEdit3352e3_TimerEventDefault(ptr.Pointer(), std_core.PointerFromQTimerEvent(event))
	}
}
