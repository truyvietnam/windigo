/**
 * Part of Windigo - Win32 API layer for Go
 * https://github.com/rodrigocfd/windigo
 * This library is released under the MIT license.
 */

package ui

import (
	"windigo/co"
	"windigo/win"
)

// Modal popup window.
//
// Allows message and notification handling.
type WindowModal struct {
	_WindowBase
	prevFocusParent win.HWND // child control last focused on parent
	setup           _WindowSetupModal
}

// Parameters that will be used to create the window.
func (me *WindowModal) Setup() *_WindowSetupModal {
	if me.Hwnd() != 0 {
		panic("Cannot change setup after the window was created.")
	}
	me.setup.initOnce() // guard
	return &me.setup
}

// Creates the modal window and disables the parent.
//
// This function will return only after the modal is closed.
func (me *WindowModal) Show(parent Window) {
	me.setup.initOnce() // guard
	hInst := parent.Hwnd().GetInstance()
	me._WindowBase.registerClass(me.setup.genWndclassex(hInst))

	me.defaultMessageHandling()

	me.prevFocusParent = win.GetFocus() // currently focused control
	parent.Hwnd().EnableWindow(false)   // https://devblogs.microsoft.com/oldnewthing/20040227-00/?p=40463

	x, y, cx, cy := me.setup.calcCoords(parent)
	me._WindowBase.createWindow("WindowModal", me.setup.ExStyle,
		me.setup.ClassName, me.setup.Title, me.setup.Style,
		x, y, cx, cy, parent, win.HMENU(0), hInst)

	me.runModalLoop()
}

func (me *WindowModal) defaultMessageHandling() {
	me.OnMsg().WmSetFocus(func(hwndLosingFocus win.HWND) {
		if me._WindowBase.Hwnd() == win.GetFocus() {
			// If window receive focus, delegate to first child.
			// This also happens right after the modal is created.
			me._WindowBase.Hwnd().
				GetNextDlgTabItem(win.HWND(0), false).
				SetFocus()
		}
	})

	me.OnMsg().WmClose(func() {
		me.Hwnd().GetWindow(co.GW_OWNER).EnableWindow(true) // re-enable parent
		me.Hwnd().DestroyWindow()                           // then destroy modal
		me.prevFocusParent.SetFocus()                       // could be on WM_DESTROY too
	})
}

func (me *WindowModal) runModalLoop() {
	msg := win.MSG{}
	for {
		if win.GetMessage(&msg, win.HWND(0), 0, 0) == 0 {
			// WM_QUIT was sent, exit modal loop now and signal parent.
			// If it returned -1, it will simply panic.
			// https://devblogs.microsoft.com/oldnewthing/20050222-00/?p=36393
			win.PostQuitMessage(int32(msg.WParam))
			break
		}

		// If a child window, will retrieve its top-level parent.
		// If a top-level, use itself.
		if msg.HWnd.GetAncestor(co.GA_ROOT).IsDialogMessage(&msg) {
			// Processed all keyboard actions for child controls.
			if me.hwnd == win.HWND(0) {
				break // our modal was destroyed, terminate loop
			} else {
				continue
			}
		}

		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)

		if me.hwnd == win.HWND(0) {
			break // our modal was destroyed, terminate loop
		}
	}
}
