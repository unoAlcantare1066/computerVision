package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/nvevg/gopusenc/gopusenc"
	"io"
	"os"
)

const SamplesPerChannel = 256

var rateArg = flag.Int("rate", 44100, "rate (hz)")
var nChannelsArg = flag.Int("nchannels", 2, "number of channels")
var filenameArg = flag.String("input", "", "int16 PCM file")
var outputArg = flag.String("output", "out.ogg", "OGG file")

func main() {
	flag.Parse()

	if *filenameArg == "" {
		fmt.Fprintf(os.Stderr, "no input file provided")
		os.Exit(1)
	}

	if *nChannelsArg != 1 && *nChannelsArg != 2 {
		fmt.Fprintf(os.Stderr, "unconventional channel count")
		os.Exit(1)
	}

	pcmFile, err := os.Open(*filenameArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open input file: %s\n", err.Error())
		os.Exit(1)
	}

	oggFile, err := os.Create(*outputArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create file: %s\n", err.Error())
		os.Exit(1)
	}

	encoder := gopusenc.NewEncoder(*rateArg, *nChannelsArg, gopusenc.MonoStereo, oggFile)
	defer encoder.Finish()

	if e := encoder.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "error while initializing encoder: %s\n", e.Error())
		os.Exit(1)
	}

	pcms := make([]int16, *nChannelsArg*SamplesPerChannel)

	for true {
		err = binary.Read(pcmFile, binary.LittleEndian, &pcms)
		if err != nil {
			break
		}

		if e := encoder.Encode(SamplesPerChannel, pcms); e != nil {
			fmt.Fprintf(os.Stderr, "encoding error: %s\n", e.Error())
			os.Exit(3)
		}
	}

	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "i/o error: %s\n", err.Error())
		os.Exit(2)
	}

	encoder.Drain()
}
