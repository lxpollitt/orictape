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

var brailleRunesL = [4]rune{'⠁', '⠂', '⠄', '⡀'}

var brailleRunesR = [4]rune{'⠈', '⠐', '⠠', '⢀'}

var brailleRunesLR = [4][4]rune{
	{'⠉', '⠑', '⠡', '⢁'},
	{'⠊', '⠒', '⠢', '⢂'},
	{'⠌', '⠔', '⠤', '⢄'},
	{'⡈', '⡐', '⡠', '⣀'}}

const bgCol = termbox.ColorDefault
const fgCol = termbox.ColorDefault
const selCol = termbox.ColorWhite
const curCol = termbox.ColorCyan

var currentHeight, currentWidth int
var wavY, wavHeight int
var hexHeaderY int
var hexY, hexHeight, hexCols int
var basicHeaderY int
var basicY, basicHeight int
var statusY int

func resetSize(w, h int) {
	currentWidth, currentHeight = w, h

	usedHeight := 0

	// Wav
	wavY = usedHeight
	wavHeight = 4
	usedHeight = usedHeight + wavHeight

	// Hex header
	hexHeaderY = usedHeight
	usedHeight++

	// Hex
	hexCols = w / 3
	hexY = usedHeight
	hexHeight = (h - usedHeight - 1) / 2
	usedHeight = usedHeight + hexHeight

	// Basic header
	basicHeaderY = usedHeight
	usedHeight++

	// Basic
	basicY = usedHeight
	basicHeight = h - usedHeight - 1
	usedHeight = usedHeight + basicHeight

	// Status
	statusY = usedHeight
	usedHeight = usedHeight + 1

	termbox.Clear(fgCol, bgCol)
}

func redrawWav() {
	bytei := prog.bytes[hexCursor]
	bits := prog.stream.bits
	samples := prog.stream.samples

	// Clear existing wav.
	cells := termbox.CellBuffer()
	s := wavY * currentWidth
	e := s + wavHeight*currentWidth
	for i := s; i < e; i++ {
		cells[i].Ch = ' '
	}

	// Draw new wav.
	yOffset := int(prog.stream.minVal)
	yScale := 1 + (int(prog.stream.maxVal)-int(prog.stream.minVal))/(4*wavHeight)
	xOffset := bits[bytei.firstBit].firstSample
	xScale := (100 * (bits[bytei.lastBit].lastSample - xOffset + 1)) / (currentWidth - 4)

	fgLabel := fgCol
	fgWav := fgCol | termbox.AttrBold
	i := bytei.firstBit
	bt := bits[i]
	label := bit(255)
	y := 4*(wavY+wavHeight) - 1
	for x := 0; x < currentWidth-2; x++ {
		j := xOffset + (xScale * x / 100)
		if j > bt.lastSample {
			label = bt.v
			if bt.unclear {
				fgLabel = termbox.ColorYellow
			} else {
				fgLabel = fgCol
			}
			i++
			bt = bits[i]
			if bt.unclear {
				fgWav = termbox.ColorYellow
			} else {
				fgWav = fgCol | termbox.AttrBold
			}
		}

		y1 := y - (int(samples[j])-yOffset)/yScale
		y2 := y - (int(samples[j+xScale/200])-yOffset)/yScale
		if y1/4 == y2/4 {
			termbox.SetCell(x+1, y1/4, brailleRunesLR[y1%4][y2%4], fgWav, bgCol)
		} else {
			termbox.SetCell(x+1, y1/4, brailleRunesL[y1%4], fgWav, bgCol)
			termbox.SetCell(x+1, y2/4, brailleRunesR[y2%4], fgWav, bgCol)
		}

		if label != 255 {
			switch {
			case label == 1:
				termbox.SetCell(x+1, wavY+wavHeight-1, '1', fgLabel, bgCol)
			case label == 0:
				termbox.SetCell(x+1, wavY+wavHeight-1, '0', fgLabel, bgCol)
			default:
				// Should never happen:
				termbox.SetCell(x+1, wavY+wavHeight-1, '?', fgLabel, bgCol)
			}
			label = 255
		}

	}
}

var hexStart, hexEnd int

func redrawHex() {
	i := hexStart
	for row := 0; row < hexHeight && i < len(prog.bytes); row++ {
		for col := 0; col < hexCols; col++ {
			if i < len(prog.bytes) {
				bti := prog.bytes[i]
				v := fmt.Sprintf("%02x", bti.v)
				switch {
				case bti.chkErr:
					tbPrint(col*3+1, hexY+row, termbox.ColorRed, bgCol, v)
				case bti.unclear:
					tbPrint(col*3+1, hexY+row, termbox.ColorYellow, bgCol, v)
				default:
					tbPrint(col*3+1, hexY+row, fgCol, bgCol, v)
				}
				i++
			} else {
				tbPrint(col*3+1, hexY+row, fgCol, bgCol, "  ")
			}

		}
	}
	hexEnd = i - 1
}

var basicStart int

func redrawBasic() {
	i := basicStart
	for row := 0; row < basicHeight && i < len(prog.lines); row++ {
		l := prog.lines[basicStart+row]
		v := l.v
		fg := fgCol
		if l.lenErr {
			fg = termbox.ColorRed
		}
		for col := 0; col < currentWidth-1; col++ {
			if col < len(v) {
				termbox.SetCell(1+col, row+basicY, rune(v[col]), fg, bgCol)
			} else {
				termbox.SetCell(1+col, row+basicY, ' ', fg, bgCol)
			}
		}
		i++
	}
}

var hexErrStatus string
var hexWarnStatus string
var basicErrStatus string

type headerText struct {
	fg   termbox.Attribute
	text string
}

func drawHeader(y int, hts ...headerText) {
	termbox.SetCell(0, y, horizontalLine, termbox.ColorBlue, bgCol)
	x := 1
	for _, ht := range hts {
		for _, c := range ht.text {
			termbox.SetCell(x, y, c, ht.fg, bgCol)
			x++
		}
		termbox.SetCell(x, y, horizontalLine, termbox.ColorBlue, bgCol)
		x++
	}
	for ; x < currentWidth; x++ {
		termbox.SetCell(x, y, horizontalLine, termbox.ColorBlue, bgCol)
	}

}

func redrawHeaders() {
	switch {
	case hexErrStatus != "" && hexWarnStatus != "":
		drawHeader(hexHeaderY, headerText{termbox.ColorYellow, hexWarnStatus}, headerText{termbox.ColorRed, hexErrStatus})
	case hexErrStatus != "":
		drawHeader(hexHeaderY, headerText{termbox.ColorRed, hexErrStatus})
	case hexWarnStatus != "":
		drawHeader(hexHeaderY, headerText{termbox.ColorYellow, hexWarnStatus})
	default:
		drawHeader(hexHeaderY)
	}

	switch {
	case basicErrStatus != "":
		drawHeader(basicHeaderY, headerText{termbox.ColorRed, basicErrStatus})
	default:
		drawHeader(basicHeaderY)
	}
}

func redrawStatus() {
	status := " Instructions go here....  Press Esc to quit"
	x := 0
	for _, c := range status {
		termbox.SetCell(x, statusY, c, termbox.ColorWhite, termbox.ColorBlue)
		x++
	}
	for ; x < currentWidth; x++ {
		termbox.SetCell(x, statusY, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}
}

func redrawAll() {
	termbox.SetOutputMode(termbox.Output256)
	w, h := termbox.Size()
	if currentWidth != w || currentHeight != h {
		resetSize(w, h)
	}

	redrawWav()
	redrawHex()
	redrawBasic()
	redrawSelection(true)
	redrawHeaders()
	redrawStatus()

	termbox.Flush()

	// Size has sometimes changed already (during the Flush?).
	w, h = termbox.Size()
	if currentWidth != w || currentHeight != h {
		redrawAll()
	}
}

var hexSelStart, hexSelEnd int = 0, 20
var hexCursor = 0
var basicCursorLine int = -1
var basicCursorL, basicCursorR = 0, 10

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
		si := sc*3 + (hexY+sr)*currentWidth
		ei := min(ec*3, currentWidth-1) + (hexY+er)*currentWidth

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
		ci := cc*3 + (hexY+cr)*currentWidth
		cells[ci+1].Bg = bg
		cells[ci+2].Bg = bg
	}

	if basicCursorLine >= basicStart && basicCursorLine < basicStart+basicHeight {
		if visible {
			bg = selCol
		}
		si := ((basicCursorLine - basicStart) + basicY) * currentWidth
		ei := si + currentWidth
		for i := si; i < ei; i++ {
			cells[i].Bg = bg
		}

		if basicCursorL <= basicCursorR {
			if visible {
				bg = curCol
			}
			for i := si + basicCursorL; i <= si+basicCursorR; i++ {
				cells[i].Bg = bg
			}
		}
	}
}

func moveHexCursor(newHexCur int) {
	if newHexCur >= 0 && newHexCur < len(prog.bytes) {
		redrawSelection(false)
		hexCursor = newHexCur

		if prog.bytes[hexCursor].chkErr {
			hexErrStatus = "Byte checksum error"
		} else {
			hexErrStatus = ""
		}
		if prog.bytes[hexCursor].unclear {
			hexWarnStatus = "Byte unclear"
		} else {
			hexWarnStatus = ""
		}

		// Scroll so that hex cursor is visible.
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

		// Update the basic cursor to track new hex cursor location.
		switch {
		case hexCursor > hexSelEnd:
			for hexCursor > hexSelEnd {
				moveBasicCursor(basicCursorLine + 1)
			}
		case hexCursor < hexSelStart:
			for hexCursor < hexSelStart {
				moveBasicCursor(basicCursorLine - 1)
			}
		default:
			moveBasicCursor(basicCursorLine)
		}

		redrawWav()
		redrawSelection(true)
		redrawHeaders()
		redrawStatus()
		termbox.Flush()
	}
}

func moveBasicCursor(newBasicCursLine int) {
	var line lineInfo
	if newBasicCursLine >= -1 && newBasicCursLine <= len(prog.lines) {
		if basicCursorLine != newBasicCursLine {
			// Move basic cursor to correct line and update hex selection.
			basicCursorLine = newBasicCursLine
			if basicCursorLine < 0 {
				hexSelStart = 0
				hexSelEnd = prog.lines[0].firstByte - 1
			} else if basicCursorLine >= len(prog.lines) {
				hexSelStart = prog.lines[len(prog.lines)-1].lastByte + 1
				hexSelEnd = len(prog.bytes) - 1
			} else {
				line = prog.lines[basicCursorLine]
				hexSelStart = line.firstByte
				hexSelEnd = line.lastByte
			}

			// Scroll so basic cursor is visible.
			if basicCursorLine > basicStart+basicHeight-2 {
				basicStart = min(basicCursorLine-basicHeight+2, len(prog.lines)-basicHeight)
				redrawBasic()
			} else if basicCursorLine < basicStart {
				basicStart = max(basicCursorLine-1, 0)
				redrawBasic()
			}

			// Update the basic header status.
			if line.lenErr {
				basicErrStatus = fmt.Sprintf("Line length error (expected %d bytes, found %d bytes)",
					line.expectedLastByte-line.firstByte+1,
					line.lastByte-line.firstByte+1)
			} else {
				basicErrStatus = ""
			}

		}

		// Move basic cursor to correct element based hex cursor location.
		if basicCursorLine >= 0 && basicCursorLine < len(prog.lines) {
			hc := hexCursor - hexSelStart
			l := prog.lines[basicCursorLine]
			switch {
			case hc == 0, hc == 1:
				basicCursorL = 0
				basicCursorR = 0
			case hc == 2, hc == 3:
				basicCursorL = 1
				basicCursorR = len(l.elements[0]) - 1
			case hc-3 < len(l.elements):
				basicCursorL = 1
				i := 0
				for ; i < hc-3; i++ {
					basicCursorL = basicCursorL + len(l.elements[i])
				}
				basicCursorR = basicCursorL + len(l.elements[i]) - 1
			default:
				basicCursorL = len(l.v) + 1
				basicCursorR = basicCursorL
			}
		}
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
			//			tbPrint(0, currentHeight - 1, fgCol, bgCol,
			//				fmt.Sprintf("EventKey: k: %d, c: %c, mod: %d", ev.Key, ev.Ch, ev.Mod))
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
