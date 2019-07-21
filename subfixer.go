// Copyright 2019 Michele Gianella & Chetan Chauhan.
// Use of this source code is governed by an AGPL
// license that can be found in the LICENSE.md file.

package main

import (
	//"github.com/chetan-prime/subfixer/astisub"
	"./astisub"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Below are the default constants for options
const (
	DefaultMode = "normal"
	DefaultReadingSpeed = 21.0
	DefaultMinLength = 1.0
	DefaultSpeedEpsilon = 1.0
	DefaultTrimSpaces = 1
	DefaultJoinShorterThan = 42
	DefaultExpandCloserThan = 0.5
	DefaultSplitLongerThan = 7.0
	DefaultShrinkLongerThan = 7.0
	DefaultForbiddenChars = "{./;/!/?/,:}"
	DefaultMaxLines = 2
	DefaultCharsPerLine = 42
	DefaultLineBalance = 50.0
	DefaultPreferCompact = true
	DefaultSpacesAsChars = true
	DefaultNewlinesAsChars = false
)

// parseFlags processes flags on command line, assigns to CommandParams structs & returns
func parseFlags() (astisub.CommandParams, error) {
	filePtr  := flag.String("file", "", "Subtitle Input File (Required)")
	
	modePtr  := flag.String("mode", DefaultMode, "Operation Mode (normal/perfection)")

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
	
	limitToPtr := flag.String(	"limit_to",
								"",
								"Limit to range or list of subtitle id's (1-2,4-10,14-16,18)")
	
	forbiddenCharsPtr := flag.String(	"forbidden_chars",
										DefaultForbiddenChars,
										"Perfection Check - Forbidden Characters")
	
	maxLinesPtr := flag.Int(	"max_lines",
								DefaultMaxLines,
								"Perfection Check - Max. lines")
	
	charsPerLinePtr := flag.Int(	"chars_per_line",
									DefaultCharsPerLine,
									"Perfection Check - No. of characters/line")
	
	readingSpeedPtr := flag.Float64("reading_speed",
									DefaultReadingSpeed,
									"Perfection Check - Reading Speed (ch/sec)"	)
	
	lineBalancePtr := flag.Float64(	"line_balance",
									DefaultLineBalance,
									"Perfection Check - Length Balance (%)")
	
	preferCompactPtr := flag.Bool(	"prefer_compact",
									DefaultPreferCompact,
									"Perfection Check - Prefer Compact Subtitles")
	
	spacesAsCharsPtr := flag.Bool(	"spaces_as_chars",
									DefaultSpacesAsChars,
									"Perfection Check - Treat Spaces as characters")
	
	newlinesAsCharsPtr := flag.Bool(	"newlines_as_chars",
										DefaultNewlinesAsChars,
										"Perfection Check - Treat newlines as characters")
	
	flag.Parse()
	
	limitTo := *limitToPtr
	var limitRanges []astisub.RangeStruct
	
	if limitTo!="" {
		limitArr := strings.Split(limitTo, ",")
		for _, limitStr := range limitArr {
			limitStr = strings.Trim(limitStr, " ")
			limit := strings.Split(limitStr, "-")
			
			if len(limit)>0 {
				if limitRanges==nil {
					limitRanges = make([]astisub.RangeStruct,0)
				}
				if start := limit[0]; start!="" {
					rangeStr := astisub.RangeStruct{start, start}
					if len(limit)>1 {
						if stop := limit[1]; stop!="" {
							rangeStr.Stop = stop
						}
					}
					
					limitRanges = append(limitRanges, rangeStr)
				}
			}
		}
	}
	
	res := astisub.CommandParams{	File: *filePtr,
									Mode: *modePtr,
									Speed: *speedPtr,
									SpeedEpsilon: *speedEpsilonPtr,
									MinLength: *minLengthPtr,
									TrimSpaces: *trimSpacesPtr,
									JoinShorterThan: *joinShorterThanPtr,
									ExpandCloserThan: *expandCloserThanPtr,
									SplitLongerThan: *splitLongerThanPtr,
									ShrinkLongerThan: *shrinkLongerThanPtr,
									LimitTo: limitRanges,
									ForbiddenChars: *forbiddenCharsPtr,
									MaxLines: 	*maxLinesPtr,
									CharsPerLine: *charsPerLinePtr,
									ReadingSpeed: *readingSpeedPtr,
									LineBalance: *lineBalancePtr,
									PreferCompact: *preferCompactPtr,
									SpacesAsChars: *spacesAsCharsPtr,
									NewlinesAsChars: *newlinesAsCharsPtr	}
	var err error = nil
	
	if res.File=="" {
		err = errors.New("Input Subtitle file is required")
	}
	
	return res, err
}

// NormalOperation runs normal operation (Read / Write).
// This function is called based on the command line parameters used
func NormalOperation(s *astisub.Subtitles, params astisub.CommandParams) int {
	incBy := 1
	
	for i:=0; i < len(s.Items); i+= incBy {
		if s.Items[i].Process {
			for p:=0; p<3; p++ {
				fmt.Printf("id #%d: Starting Pass %d..\n", i+1, p+1)
				incBy = s.AdjustDuration(i, params)
				if incBy <= 0 {
					fmt.Printf("Skipping further passes as subtitles seems to have been deleted / split")
					break
				}
			}
		}
	}
	
	fmt.Printf("Now saving changes to file %s: ", params.File)
	s.Write(params.File)
	fmt.Printf("[DONE]\n")
	
	return 0
}

// Perform Perfection check(read only).
// This function is called based on the command line parameters used
func PerfectionOperation(s *astisub.Subtitles, params astisub.CommandParams) int {
	error_code := 0
	
	errors := make(map[int][]string)
	
	for i:=0; i < len(s.Items); i+= 1 {
		if s.Items[i].Process {
			perrs := s.PerfectionCheck(i, params)
			if len(perrs)>0 {
				errors[i] = perrs
				
				serrs := strings.Join(perrs, " @@ ")
				fmt.Fprintf(os.Stderr, "Subtitle #%d failed perfection check - %s\n", i+1, serrs)
				
				error_code = 10
			}
		}	
	}

	if len(errors)>0 {
		ikeys := make([]int, 0)
		for k, _ := range errors {
			ikeys = append(ikeys, k+1)
		}
		
		sort.Ints(ikeys)
		
		keylist := ""
		
		for _, k := range ikeys {
			if len(keylist)>0 {
				keylist += ","
			}
			keylist += strconv.FormatInt(int64(k), 10)
		}
		
		fmt.Fprintf(os.Stderr, "Perfection check failed on these ids - [%s]\n", keylist)
	} else {
		fmt.Fprintf(os.Stderr, "Perfection check passed succesfully!\n")
	}

	return error_code
}

// Main entry point for the program
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
		
		if params.LimitTo==nil {
			params.LimitTo=make([]astisub.RangeStruct, 0)
			rangeStr := astisub.RangeStruct{"0", strconv.FormatInt(int64(len(s.Items)-1),10)}
	
			params.LimitTo = append(params.LimitTo, rangeStr)
		}
	
		for _, limitrec := range params.LimitTo {
			if strings.Index(limitrec.Start, ":")>-1 ||
			   strings.Index(limitrec.Start, ".")>-1 {
				start := time.Duration(0)
				stop := time.Duration(0)
			
				if strings.Index(limitrec.Start, ":")>-1 {
					start, _ = astisub.ParseDuration(limitrec.Start, ".", 3)
					stop, _  = astisub.ParseDuration(limitrec.Stop, ".", 3)
				} else {
					startf,_ := strconv.ParseFloat(limitrec.Start, 64)
					stopf,_ := strconv.ParseFloat(limitrec.Stop, 64)
				
					start = time.Duration( startf * float64( time.Second ) )
					stop  = time.Duration( stopf  * float64( time.Second ) )
				}
			
				for i:=0; i < len(s.Items); i++ {
					if (	s.Items[i].StartAt>=start &&
							s.Items[i].EndAt<=stop ) ||
					   (	s.Items[i].StartAt<=stop &&
							s.Items[i].EndAt>=start ) {
						s.Items[i].Process = true
						fmt.Printf("Marking subtitle id #%d for processing\n", i+1)
					}
				}
			} else {
				start, _ := strconv.Atoi(limitrec.Start)
				if start < 1 {
					start = 1
				}
				stop, _  := strconv.Atoi(limitrec.Stop)
				for i:=start-1; i < len(s.Items) && i<=stop-1; i++ {
					s.Items[i].Process = true
					fmt.Printf("Marking subtitle id #%d for processing\n", i+1)
				}
			}
		}
		
		error_code := 0
		
		if params.Mode=="perfection" {
			fmt.Printf("Performing in Perfection mode\n")
			error_code = PerfectionOperation(s, params)
		} else {
			fmt.Printf("Performing in Normal mode\n")
			error_code = NormalOperation(s, params)
		}
		
		os.Exit(error_code)
	} else {
		avail := fmt.Sprintf("%s: Available parameters are below", os.Args[0])
		avail += "\n" + strings.Repeat("-", len(avail)) + "\n"
		os.Stderr.WriteString(avail)
		flag.PrintDefaults()
	}
}


