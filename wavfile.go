// Copyright © 2015 The Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the LICENSE file for the specific language governing permissions and
// limitations under the License in the main package.

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

type riff struct {
	Sig            [4]byte
	RiffSize       uint32
	DataSig        [4]byte
	FmtSig         [4]byte
	FmtSize        uint32
	Tag            uint16
	Channels       uint16
	Freq           uint32
	BytesPerSec    uint32
	BytesPerSample uint16
	BitsPerSample  uint16
	SamplesSig     [4]byte
	Length         uint32
}

func readWavFile(fileName string) (left, right []int16, err error) {
	var r riff

	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return
	}

	if err = binary.Read(file, binary.LittleEndian, &r); err != nil {
		return
	}

	if string(r.Sig[:]) != "RIFF" ||
		r.Tag != 1 ||
		r.Channels != 2 ||
		r.Freq != 44100 ||
		r.BitsPerSample != 16 {
		err = errors.New("Only 44.1khz 16 bit stereo wav files supported")
		return
	}

	samplesToRead := r.Length / 4
	left = make([]int16, samplesToRead)
	right = make([]int16, samplesToRead)

	var c int
	bytes := make([]byte, r.Length)
	if c, err = file.Read(bytes); err != nil {
		return
	}
	if uint32(c) != r.Length {
		err = errors.New("Wrong number of bytes read")
		return
	}

	for i := uint32(0); i < samplesToRead; i++ {
		bi := i * 4
		left[i] = int16(binary.LittleEndian.Uint16(bytes[bi : bi+2]))
		right[i] = int16(binary.LittleEndian.Uint16(bytes[bi+2 : bi+4]))
	}

	fmt.Printf("Found %d seconds of audio (%d samples)\n", samplesToRead/44100, samplesToRead)

	return left, right, err
}
