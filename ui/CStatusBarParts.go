//go:build windows

package ui

import (
	"fmt"
	"unsafe"

	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

type _StatusBarPartData struct {
	sizePixels   int
	resizeWeight int
}

func (me *_StatusBarPartData) IsFixedWidth() bool {
	return me.resizeWeight == 0
}

//------------------------------------------------------------------------------

type _StatusBarParts struct {
	sb              StatusBar
	partsData       []_StatusBarPartData
	rightEdges      []int32 // buffer to speed up ResizeToFitParent() calls
	initialParentCx int     // cache used when adding parts
}

func (me *_StatusBarParts) new(ctrl StatusBar) {
	me.sb = ctrl
	me.partsData = make([]_StatusBarPartData, 1, 5)
	me.rightEdges = make([]int32, 1, 5)
	me.initialParentCx = 0
}

func (me *_StatusBarParts) cacheInitialParentCx() {
	if me.initialParentCx == 0 { // not cached yet?
		rc := me.sb.Hwnd().GetClientRect()
		me.initialParentCx = int(rc.Right) // initial width of parent's client area
	}
}

func (me *_StatusBarParts) resizeToFitParent(parm wm.Size) {
	if parm.Request() == co.SIZE_REQ_MINIMIZED || me.sb.Hwnd() == 0 {
		return
	}

	cx := int(parm.ClientAreaSize().Cx)        // available width
	me.sb.Hwnd().SendMessage(co.WM_SIZE, 0, 0) // tell status bar to fit parent

	// Find the space to be divided among variable-width parts,
	// and total weight of variable-width parts.
	totalWeight := 0
	cxVariable := cx
	for i := range me.partsData {
		if me.partsData[i].IsFixedWidth() {
			cxVariable -= me.partsData[i].sizePixels
		} else {
			totalWeight += me.partsData[i].resizeWeight
		}
	}

	// Fill right edges array with the right edge of each part.
	cxTotal := cx
	for i := len(me.partsData) - 1; i >= 0; i-- {
		me.rightEdges[i] = int32(cxTotal)
		if me.partsData[i].IsFixedWidth() {
			cxTotal -= me.partsData[i].sizePixels
		} else {
			cxTotal -= (cxVariable / totalWeight) * me.partsData[i].resizeWeight
		}
	}
	me.sb.Hwnd().SendMessage(co.SB_SETPARTS,
		win.WPARAM(len(me.rightEdges)),
		win.LPARAM(unsafe.Pointer(&me.rightEdges[0])))
}

// Adds one or more fixed-width parts.
//
// Widths will be adjusted to the current system DPI.
func (me *_StatusBarParts) AddFixed(widths ...int) {
	me.cacheInitialParentCx()

	for _, width := range widths {
		if width < 0 {
			panic(fmt.Sprintf("Width of part can't be negative: %d.", width))
		}

		size := win.SIZE{Cx: int32(width), Cy: 0}
		_MultiplyDpi(nil, &size)

		me.partsData = append(me.partsData, _StatusBarPartData{
			sizePixels: int(size.Cx),
		})
		me.rightEdges = append(me.rightEdges, 0)
	}

	me.resizeToFitParent(wm.Size{
		Msg: wm.Any{
			WParam: win.WPARAM(co.SIZE_REQ_RESTORED),
			LParam: win.MAKELPARAM(uint16(me.initialParentCx), 0),
		},
	})
}

// Adds one or more resizable parts.
//
// How resizeWeight works:
//
// - Suppose you have 3 parts, respectively with weights of 1, 1 and 2.
//
// - If available client area is 400px, respective part widths will be 100, 100 and 200px.
func (me *_StatusBarParts) AddResizable(resizeWeights ...int) {
	me.cacheInitialParentCx()

	for _, resizeWeight := range resizeWeights {
		if resizeWeight <= 0 {
			panic(fmt.Sprintf("Resize weight must be equal or greater than 1: %d.", resizeWeight))
		}

		me.partsData = append(me.partsData, _StatusBarPartData{
			resizeWeight: resizeWeight,
		})
		me.rightEdges = append(me.rightEdges, 0)
	}

	me.resizeToFitParent(wm.Size{
		Msg: wm.Any{
			WParam: win.WPARAM(co.SIZE_REQ_RESTORED),
			LParam: win.MAKELPARAM(uint16(me.initialParentCx), 0),
		},
	})
}

// Retrieves the texts of all parts at once.
func (me *_StatusBarParts) AllTexts() []string {
	texts := make([]string, 0, me.Count())
	for i := 0; i < me.Count(); i++ {
		texts = append(texts, me.Get(i).Text())
	}
	return texts
}

// Returns the number of parts.
func (me *_StatusBarParts) Count() int {
	return len(me.partsData)
}

// Returns the part at the given index.
//
// Note that this method is dumb: no validation is made, the given index is
// simply kept. If the index is invalid (or becomes invalid), subsequent
// operations on the StatusBarPart will fail.
func (me *_StatusBarParts) Get(index int) StatusBarPart {
	return StatusBarPart{sb: me.sb, index: uint32(index)}
}

// Sets the texts of all parts at once.
//
// Panics if the number of texts is greater than the number of parts.
func (me *_StatusBarParts) SetAllTexts(texts ...string) {
	if len(texts) > len(me.partsData) {
		panic(
			fmt.Sprintf("Number of texts (%d) is greater than the number of parts (%d).",
				len(texts), len(me.partsData)))
	}

	for i, text := range texts {
		me.Get(i).SetText(text)
	}
}
