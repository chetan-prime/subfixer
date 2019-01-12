package main

import (
	"./astisub"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
	
	limitToPtr := flag.String(	"limit_to",
								"",
								"Limit to range or list of subtitle id's (1-2,4-10,14-16,18)")
	
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
	
	res := astisub.CommandParams{	*filePtr,
									*speedPtr,
									*speedEpsilonPtr,
									*minLengthPtr,
									*trimSpacesPtr,
									*joinShorterThanPtr,
									*expandCloserThanPtr,
									*splitLongerThanPtr,
									*shrinkLongerThanPtr,
									limitRanges,	}
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
				stop, _  := strconv.Atoi(limitrec.Stop)
				for i:=start; i < len(s.Items) && i+1<=stop; i++ {
					s.Items[i].Process = true
					fmt.Printf("Marking subtitle id #%d for processing\n", i+1)
				}
			}
		}
		
		
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
	} else {
		avail := fmt.Sprintf("%s: Available parameters are below", os.Args[0])
		avail += "\n" + strings.Repeat("-", len(avail)) + "\n"
		os.Stderr.WriteString(avail)
		flag.PrintDefaults()
	}
}


