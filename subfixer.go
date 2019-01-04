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
	DefaultTrimSpaces = 1
	DefaultJoinShorterThan = 42
	DefaultExpandCloserThan = 0.5
	DefaultSplitLongerThan = 7.0
	DefaultShrinkLongerThan = 7.0
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
	
	trimSpacesPtr := flag.Int(	"trim_spaces",
								DefaultTrimSpaces,
								"Trim space to left & right of each subtitle")
	
	joinShorterThanPtr := flag.Int(	"join_shorter_than",
									DefaultJoinShorterThan,
									"Join two lines shorter in length than")
	
	expandCloserThanPtr := flag.Float64("expand_closer_than",
										DefaultExpandCloserThan,
										"Expand two subtitles closer than n seconds")
	
	splitLongerThanPtr := flag.Float64(	"split_longer_than",
										DefaultSplitLongerThan,
										"Proportionately split a two line subtitle longer than n seconds")
	
	shrinkLongerThanPtr := flag.Float64("shrink_longer_than",
										DefaultShrinkLongerThan,
										"Shrink a single line subtitle longer than n seconds")
	
	flag.Parse()
	
	res := astisub.CommandParams{	*filePtr,
									*speedPtr,
									*speedEpsilonPtr,
									*minLengthPtr,
									*trimSpacesPtr,
									*joinShorterThanPtr,
									*expandCloserThanPtr,
									*splitLongerThanPtr,
									*shrinkLongerThanPtr,	}
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
		
		incBy := 1
		for i:=0; i < len(s.Items); i+= incBy {
			for p:=0; p<3; p++ {
				fmt.Printf("id #%d: Starting Pass %d..\n", i+1, p+1)
				incBy = s.AdjustDuration(i, params)
				if incBy <= 0 {
					fmt.Printf("Skipping further passes as subtitles seems to have been deleted / split")
					break
				}
			}
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


