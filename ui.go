package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

const horizontalLine = '─'

func tbHLine(y int, fg, bg termbox.Attribute) {
	for x := 0; x < currentWidth; x++ {
		termbox.SetCell(x, y, horizontalLine, fg, bg)
	}
}

var current string
var prog program

func mouse_button_num(k termbox.Key) int {
	switch k {
	case termbox.MouseLeft:
		return 0
	case termbox.MouseMiddle:
		return 1
	case termbox.MouseRight:
		return 2
	}
	return 0
}

const bgCol = termbox.ColorDefault
const fgCol = termbox.ColorDefault
const selCol = termbox.ColorWhite
const curCol = termbox.ColorCyan

//const bgCol = termbox.ColorBlack
//const fgCol = termbox.ColorWhite

var currentHeight, currentWidth int
var hexWidth, hexHeight, hexCols int
var basicWidth, basicHeight int
var mainHeight int

func resetSize(w, h int) {
	currentWidth, currentHeight = w, h
	mainHeight = h - 1

	hexCols = w / 3
	hexWidth = hexCols*3 + 1
	hexHeight = h / 2

	basicWidth = w
	basicHeight = mainHeight - hexHeight - 1

	termbox.Clear(fgCol, bgCol)
	tbPrint(0, mainHeight, termbox.ColorWhite, termbox.ColorBlue, "Instructions go here....  Press Esc to quit")
	tbHLine(hexHeight, fgCol, bgCol)
}

var hexStart, hexEnd int

func redrawHex() {
	i := hexStart
	for h := 0; h < hexHeight && i < len(prog.bytes); h++ {
		for w := 0; w < hexCols; w++ {
			if i < len(prog.bytes) {
				bti := prog.bytes[i]
				v := fmt.Sprintf("%02x", bti.v)
				switch {
				case bti.chkErr:
					tbPrint(w*3+1, h, termbox.ColorRed, bgCol, v)
				case bti.unclear:
					tbPrint(w*3+1, h, termbox.ColorYellow, bgCol, v)
				default:
					tbPrint(w*3+1, h, fgCol, bgCol, v)
				}
				i++
			} else {
				tbPrint(w*3+1, h, fgCol, bgCol, "  ")
			}

		}
	}
	hexEnd = i - 1
}

var hexSelStart, hexSelEnd int = 0, 20
var hexCursor = 10
var basicSel int = -1

func redrawSelection(visible bool) {
	cells := termbox.CellBuffer()
	bg := bgCol
	if hexEnd >= hexSelStart && hexStart <= hexSelEnd {
		if visible {
			bg = selCol
		}
		// Calc the hex number start and end offsets.
		sh := max(hexSelStart, hexStart) - hexStart
		eh := min(hexSelEnd, hexEnd) - hexStart

		// Calc the start row and column.
		sr := sh / hexCols
		sc := (sh % hexCols)

		// Calc the end row and column.
		er := eh / hexCols
		ec := sc + (eh - sh + 1) - (er-sr)*hexCols

		// Calc the start and end cell indexes.
		si := sc*3 + sr*currentWidth
		ei := min(ec*3, currentWidth-1) + er*currentWidth

		for i := si; i <= ei; i++ {
			cells[i].Bg = bg
		}
	}

	if hexEnd >= hexCursor && hexStart <= hexCursor {
		if visible {
			bg = termbox.ColorCyan
		}

		// Now do the equivalent for the hex cursor.
		ch := hexCursor - hexStart
		cr := ch / hexCols
		cc := (ch % hexCols)
		ci := cc*3 + cr*currentWidth
		cells[ci+1].Bg = bg
		cells[ci+2].Bg = bg
	}

	if basicSel >= basicStart && basicSel < basicStart+basicHeight {
		if visible {
			bg = selCol
		}
		si := ((basicSel - basicStart) + hexHeight + 1) * currentWidth
		ei := si + currentWidth
		for i := si; i < ei; i++ {
			cells[i].Bg = bg
		}
	}
}

var basicStart int

func redrawBasic() {
	i := basicStart
	for h := 0; h < basicHeight && i < len(prog.lines); h++ {
		tbPrint(1, h+hexHeight+1, fgCol, bgCol, prog.lines[i].v)
		i++
	}
}

func redrawAll() {
	termbox.SetOutputMode(termbox.Output256)
	w, h := termbox.Size()
	if currentWidth != w || currentHeight != h {
		resetSize(w, h)
	}

	redrawHex()
	redrawBasic()
	redrawSelection(true)

	termbox.Flush()

	// Size has sometimes changed already (during the Flush?).
	w, h = termbox.Size()
	if currentWidth != w || currentHeight != h {
		redrawAll()
	}
}

func moveHexCursor(newHexCur int) {
	if newHexCur >= 0 && newHexCur < len(prog.bytes) {
		redrawSelection(false)
		hexCursor = newHexCur
		if hexCursor > hexStart+(hexHeight-2)*hexCols {
			hexStart = max(0,
				min((len(prog.bytes)/hexCols-hexHeight+1)*hexCols,
					((hexCursor-(hexHeight-2)*hexCols)/hexCols)*hexCols))
			hexEnd = min(len(prog.bytes)-1, hexStart+(hexHeight-1)*hexCols)
			redrawHex()
		} else if hexCursor < hexStart+hexCols {
			hexStart = max(0, ((hexCursor/hexCols)-1)*hexCols)
			hexEnd = min(len(prog.bytes)-1, hexStart+hexHeight*hexCols)
			redrawHex()
		}

		for hexCursor > hexSelEnd {
			basicSel++
			if basicSel < len(prog.lines) {
				hexSelStart = prog.lines[basicSel].firstByte
				hexSelEnd = prog.lines[basicSel].lastByte
			} else {
				hexSelStart = prog.lines[len(prog.lines)-1].lastByte + 1
				hexSelEnd = len(prog.bytes) - 1
			}
		}

		for hexCursor < hexSelStart {
			basicSel--
			if basicSel >= 0 {
				hexSelStart = prog.lines[basicSel].firstByte
				hexSelEnd = prog.lines[basicSel].lastByte
			} else {
				hexSelStart = 0
				hexSelEnd = prog.lines[0].firstByte - 1
			}
		}

		redrawSelection(true)
		termbox.Flush()
	}
}

func displayUI(p program) {
	err := termbox.Init()
	if err != nil {
		fmt.Printf("%s**** %s ****%s", CLR_R, err, CLR_0)
		return
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	prog = p
	hexSelStart = 0
	hexSelEnd = prog.lines[0].firstByte - 1

	redrawAll()

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			tbPrint(0, currentHeight-1, fgCol, bgCol,
				fmt.Sprintf("EventKey: k: %d, c: %c, mod: %d", ev.Key, ev.Ch, ev.Mod))
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyArrowLeft, termbox.KeyCtrlB:
				moveHexCursor(hexCursor - 1)
			case termbox.KeyArrowRight, termbox.KeyCtrlF:
				moveHexCursor(hexCursor + 1)
			case termbox.KeyArrowUp:
				moveHexCursor(hexCursor - hexCols)
			case termbox.KeyArrowDown:
				moveHexCursor(hexCursor + hexCols)
			}
		case termbox.EventMouse:
			tbPrint(0, currentHeight-1, fgCol, bgCol,
				fmt.Sprintf("EventMouse: x: %d, y: %d, b: %d", ev.MouseX, ev.MouseY, mouse_button_num(ev.Key)))
		case termbox.EventNone:
			tbPrint(0, currentHeight-1, fgCol, bgCol, "EventNone")
		case termbox.EventResize:
			redrawAll()
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
