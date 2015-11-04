package main

import (
	"fmt"
	"math"
	"os"
)

type bit byte
type bitInfo struct {
	v                       bit
	l1, l2                  int
	firstSample, lastSample int
	unclear                 bool
}

type byteInfo struct {
	v                 byte
	firstBit, lastBit int
	unclear, chkErr   bool
}

type lineInfo struct {
	v                   string
	elements            []string
	firstByte, lastByte int
	expectedLastByte    int
	lenErr              bool
}

type bitStream struct {
	bits                    []bitInfo
	samples                 []int16
	firstSample, lastSample int
	minVal, maxVal          int16
}

type program struct {
	stream bitStream
	bytes  []byteInfo
	lines  []lineInfo
	name   string
}

const (
	ShortThreshold    int = 20
	LongThreshold     int = 24
	NoSignalThreshold int = 46
)

const CLR_0 = "\x1b[30;1m"
const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_Y = "\x1b[33;1m"
const CLR_B = "\x1b[34;1m"
const CLR_M = "\x1b[35;1m"
const CLR_C = "\x1b[36;1m"
const CLR_W = "\x1b[37;1m"
const CLR_N = "\x1b[0m"

var keywords []string = []string{"END", "EDIT", "STORE", "RECALL", "TRON", "TROFF", "POP", "PLOT",
	"PULL", "LORES", "DOKE", "REPEAT", "UNTIL", "FOR", "LLIST", "LPRINT", "NEXT", "DATA",
	"INPUT", "DIM", "CLS", "READ", "LET", "GOTO", "RUN", "IF", "RESTORE", "GOSUB", "RETURN",
	"REM", "HIMEM", "GRAB", "RELEASE", "TEXT", "HIRES", "SHOOT", "EXPLODE", "ZAP", "PING",
	"SOUND", "MUSIC", "PLAY", "CURSET", "CURMOV", "DRAW", "CIRCLE", "PATTERN", "FILL",
	"CHAR", "PAPER", "INK", "STOP", "ON", "WAIT", "CLOAD", "CSAVE", "DEF", "POKE", "PRINT",
	"CONT", "LIST", "CLEAR", "GET", "CALL", "!", "NEW", "TAB(", "TO", "FN", "SPC(", "@",
	"AUTO", "ELSE", "THEN", "NOT", "STEP", "+", "-", "*", "/", "^", "AND", "OR", ">", "=", "<",
	"SGN", "INT", "ABS", "USR", "FRE", "POS", "HEX$", "&", "SQR", "RND", "LN", "EXP", "COS",
	"SIN", "TAN", "ATN", "PEEK", "DEEK", "LOG", "LEN", "STR$", "VAL", "ASC", "CHR$", "PI",
	"TRUE", "FALSE", "KEY$", "SCRN", "POINT", "LEFT$", "RIGHT$", "MID$"}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: orictape <input wav file>")
		return
	}

	left, _, err := readWavFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	streams := readBitStreams(left)
	fmt.Printf("Read %d streams\n", len(streams))

	programs := readPrograms(streams)
	fmt.Printf("Read %d programs\n", len(programs))

	for _, prog := range programs {
		fmt.Printf("[%s]\n", prog.name)
		for _, line := range prog.lines {
			if line.lenErr {
				fmt.Printf("%d %d %s%s%s\n", line.expectedLastByte-line.lastByte, line.lastByte-line.firstByte+1, CLR_R, line.v, CLR_0)
			} else {
				fmt.Println(line.v)
			}
		}
	}

	fmt.Println("\n**done**")

	displayUI(programs[0])
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func min16(a, b int16) int16 {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func max16(a, b int16) int16 {
	if a > b {
		return a
	} else {
		return b
	}
}

func abs(i int) int {
	if i >= 0 {
		return i
	} else {
		return -i
	}
}

func readBitStreams(samples []int16) (streams []bitStream) {
	startSample := 0
	for stream, samplesRead := readBitStream(samples, startSample); samplesRead > 0; stream, samplesRead = readBitStream(samples, startSample) {
		streams = append(streams, stream)
		startSample += samplesRead
	}

	fmt.Printf("Found %d streams:\n", len(streams))
	for i, stream := range streams {
		fmt.Printf(" %d) Starting at %ds found stream of length %ds (%d bits)\n", i, stream.firstSample/44100, (stream.lastSample-stream.firstSample)/44100, len(stream.bits))
	}
	return
}

func readBitStream(samples []int16, startSample int) (stream bitStream, samplesRead int) {
	var minVal, maxVal, threshold int16
	var minIndex, maxIndex, belowIndex, aboveIndex, searchWindowIndex int
	var searchWindow []int16
	var lengthBelow, lengthAbove, length int

	readCycle := func() (noSignal bool) {
		// Search for the next min.
		minVal = math.MaxInt16
		searchWindow = samples[maxIndex+1 : min(maxIndex+20, len(samples))]
		for i, v := range searchWindow {
			if v < minVal {
				minVal = v
				searchWindowIndex = i
			}
		}
		minIndex = maxIndex + 1 + searchWindowIndex
		stream.minVal = min16(stream.minVal, minVal)

		// Now find the cross over point where we fall below the threshold.
		threshold = (maxVal + minVal) / 2
		for i, v := range searchWindow {
			if v <= threshold {
				searchWindowIndex = i
				break
			}
		}
		belowIndex = maxIndex + 1 + searchWindowIndex
		lengthBelow = belowIndex - aboveIndex

		// Search for the next max.
		maxVal = math.MinInt16
		searchWindow = samples[minIndex+1 : min(minIndex+20, len(samples))]
		for i, v := range searchWindow {
			if v > maxVal {
				maxVal = v
				searchWindowIndex = i
			}
		}
		maxIndex = minIndex + 1 + searchWindowIndex
		stream.maxVal = max16(stream.maxVal, maxVal)

		// Now find the cross over point where we fall below the threshold.
		threshold = (maxVal + minVal) / 2
		for i, v := range searchWindow {
			if v >= threshold {
				searchWindowIndex = i
				break
			}
		}
		aboveIndex = minIndex + 1 + searchWindowIndex
		lengthAbove = aboveIndex - belowIndex
		length = lengthBelow + lengthAbove

		switch {
		case length > NoSignalThreshold:
			noSignal = true
		case length >= LongThreshold:
			stream.bits = append(stream.bits, bitInfo{v: 0, l1: lengthBelow, l2: lengthAbove,
				firstSample: aboveIndex - length, lastSample: aboveIndex - 1})
		case length <= ShortThreshold:
			stream.bits = append(stream.bits, bitInfo{v: 1, l1: lengthBelow, l2: lengthAbove,
				firstSample: aboveIndex - length, lastSample: aboveIndex - 1})
		case abs(lengthBelow-lengthAbove) <= (LongThreshold-ShortThreshold)/2:
			// Unclear long
			stream.bits = append(stream.bits, bitInfo{v: 0, l1: lengthBelow, l2: lengthAbove, unclear: true,
				firstSample: aboveIndex - length, lastSample: aboveIndex - 1})
		default:
			// Unclear short
			stream.bits = append(stream.bits, bitInfo{v: 1, l1: lengthBelow, l2: lengthAbove, unclear: true,
				firstSample: aboveIndex - length, lastSample: aboveIndex - 1})
		}
		return noSignal
	}

	stream.samples = samples
	// Search for a stream until we find one at least 0.2 seconds long.
	maxIndex = startSample
	aboveIndex = startSample
	for maxIndex < len(samples) && len(stream.bits) < 8820 {
		stream.bits = nil
		stream.firstSample = aboveIndex

		// Read stream until we hit no signal.
		for maxIndex < len(samples) {
			if noSig := readCycle(); noSig {
				break
			}
		}
	}
	samplesRead = aboveIndex - startSample
	stream.lastSample = aboveIndex
	return
}

func readPrograms(streams []bitStream) (programs []program) {
	for _, stream := range streams {
		prog := readProgramBytes(stream)
		if len(prog.bytes) > 0 {
			readProgramLines(&prog)

			programs = append(programs, prog)

			fmt.Println("Program:")
			for _, bti := range prog.bytes {
				switch {
				case bti.chkErr:
					fmt.Printf(" %s%02x%s", CLR_R, bti.v, CLR_0)
				case bti.unclear:
					fmt.Printf(" %s%02x%s", CLR_Y, bti.v, CLR_0)
				default:
					fmt.Printf(" %02x", bti.v)
				}
			}
			fmt.Println("")
		}
	}
	return
}

func readProgramLines(prog *program) {
	var nextByte int
	var lineStart int
	var b byte
	var ok bool

	getByte := func() (b byte) {
		if nextByte < len(prog.bytes) {
			bi := prog.bytes[nextByte]
			b = bi.v
			ok = true
			nextByte++
		} else {
			b = 0
			ok = false
		}
		return
	}

	syncCount := 0
findSync:
	for {
		b = getByte()
		switch {
		case !ok:
			return
		case b == 0x16:
			syncCount++
		case b == 0x24 && syncCount > 3:
			break findSync
		default:
			syncCount = 0
		}
	}
	fmt.Printf("\n%s**** synchronized ****%s\n", CLR_G, CLR_0)

	// Read the file header.
	header := make([]byte, 9)
	for i := 0; i < len(header); i++ {
		header[i] = getByte()
	}
	if header[2] != 0 {
		fmt.Printf("\n%s**** not a basic file ****%s\n", CLR_R, CLR_0)
		return
	}

	// Strip the program name.
	fmt.Printf("%sLoading ", CLR_G)
	for b = getByte(); b > 0; b = getByte() {
		prog.name = prog.name + string(b)
	}
	fmt.Printf("%s%s\n", prog.name, CLR_0)

	// Read the program lines.
	correctionOffset := 0
	for {
		lineStart = nextByte
		nextLineStart := int(uint(getByte()) + 256*uint(getByte()))
		if nextLineStart == 0 {
			// Reached end of program.
			break
		}
		nextLineStart = nextLineStart - correctionOffset

		elements := make([]string, 0, 40)

		// Get the line number.
		element := fmt.Sprintf("%d ", uint(getByte())+256*uint(getByte()))
		elements = append(elements, element)
		line := element

	readLine:
		for {
			b = getByte()
			switch {
			case b == 0:
				break readLine
			case b < 128:
				element = string(b)
			case int(b-byte(128)) < len(keywords):
				element = keywords[b-128]
			default:
				element = CLR_R + "[UNKOWN_KEYWORD]" + CLR_0
			}
			elements = append(elements, element)
			line = line + element
		}
		prog.lines = append(prog.lines,
			lineInfo{v: line,
				elements:  elements,
				firstByte: lineStart, lastByte: nextByte - 1,
				expectedLastByte: nextLineStart - 1,
				lenErr:           nextLineStart != nextByte})
		correctionOffset = correctionOffset + nextLineStart - nextByte
	}

	// We can't deduce line length error for the first line because we didn't yet know the offset.
	// So fix that up now.
	if len(prog.lines) > 0 {
		prog.lines[0].lenErr = false
		prog.lines[0].expectedLastByte = prog.lines[0].lastByte
	}

	fmt.Println(len(prog.lines))
}

func readProgramBytes(stream bitStream) (prog program) {
	var bt bit
	var by, chk byte
	var ok bool
	var byteStart int
	var currentBit int
	var byteUnclear bool

	prog.stream = stream

	getBit := func() (bt bit, ok bool) {
		if currentBit < len(stream.bits) {
			bti := stream.bits[currentBit]
			bt = bti.v
			byteUnclear = byteUnclear || bti.unclear
			ok = true
			currentBit++
		} else {
			bt = 0
			ok = false
		}
		return
	}

	// Search for beginning of sync.
	by = 0
	for by != 0x16 {
		if bt, ok = getBit(); !ok {
			return
		}
		by = by>>1 | byte(bt<<7)
	}

	// Read bytes.
	for {
		byteUnclear = false
		byteStart = currentBit
		// Skip first bit.
		if bt, ok = getBit(); !ok {
			return
		}

		// Skip until 0.
		if bt, ok = getBit(); !ok {
			return
		}
		for bt != 0 {
			if bt, ok = getBit(); !ok {
				return
			}
		}

		// Read the data byte.
		by = 0
		chk = 0
		for i := 0; i < 8; i++ {
			if bt, ok = getBit(); !ok {
				return
			}
			by = by>>1 | byte(bt<<7)
			chk = chk + byte(bt)
		}

		// Read the checksum.
		if bt, ok = getBit(); !ok {
			return
		}
		//		if bi == bit(chk & 1) {
		//			fmt.Printf(" %s%02x%s", CLR_R, by, CLR_0)
		//		} else {
		//			fmt.Printf(" %02x", by)
		//		}
		prog.bytes = append(prog.bytes, byteInfo{v: by, firstBit: byteStart, lastBit: currentBit - 1,
			unclear: byteUnclear, chkErr: bt == bit(chk&1)})
	}
}
