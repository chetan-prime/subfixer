package main

import (
	"./astisub"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	DefaultReadingSpeed = 21.0
	DefaultMinLength = 1.0
	DefaultSpeedEpsilon = 1.0
)

func parseFlags() (astisub.CommandParams, error) {
	filePtr  := flag.String("file", "", "Subtitle Input File (Required)")
	
	speedPtr := flag.Float64(	"speed",
								DefaultReadingSpeed,
								"Desired Characters Per Second"	)
	
	minLengthPtr := flag.Float64(	"min_length",
									DefaultMinLength,
									"Minimum Length for each subtitle"	)
	
	speedEpsilonPtr := flag.Float64("speed_epsilon",
									DefaultSpeedEpsilon,
									"Epsilon in % of Speed value"	)
	
	flag.Parse()
	
	res := astisub.CommandParams{ *filePtr, *speedPtr, *speedEpsilonPtr, *minLengthPtr}
	var err error = nil
	
	if res.File=="" {
		err = errors.New("Input Subtitle file is required")
	}
	
	return res, err
}

func main() {
	params, err := parseFlags()
	
	if len(os.Args) > 2 {
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error: %s", err))
			return
		}
		// Open
		s, err := astisub.OpenFile( params.File )
		
		if err!=nil {
			os.Stderr.WriteString(fmt.Sprintf("Error opening %s: %s\n", params.File, err))
			return
		}
		
		for i:=0; i < len(s.Items); i++ {
			s.AdjustDuration(i, params)
		}
		
		fmt.Printf("Now saving changes to file %s: ", params.File)
		s.Write(params.File)
		fmt.Printf("[DONE]\n")
	} else {
		avail := fmt.Sprintf("%s: Available parameters are below", os.Args[0])
		avail += "\n" + strings.Repeat("-", len(avail)) + "\n"
		os.Stderr.WriteString(avail)
		flag.PrintDefaults()
	}
}


