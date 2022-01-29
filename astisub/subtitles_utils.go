// Copyright 2019 Michele Gianella & Chetan Chauhan.
// Use of this source code is governed by an AGPL
// license that can be found in the LICENSE.md file.

package astisub

import (
	"fmt"
	"math"
	"strings"
	"time"
	//"github.com/chetan-prime/subfixer/strip"
	"../strip"
)


const (
	SUB_LEVELS = 3
)

// This file contains changes made for subfixer

// RangeStruct is used for specifying the Range of subtitles to process
type RangeStruct struct {
	Start	string
	Stop	string
}

// CommandParams contains parameters for normal operation & Perfection check
type CommandParams struct {
	File			string
	Mode			string
	Speed			float64
	SpeedEpsilon	float64
	MinLength		float64
	TrimSpaces		int
	JoinShorterThan	int
	ExpandCloserThan	float64
	SplitLongerThan		float64
	ShrinkLongerThan	float64
	LimitTo			[]RangeStruct
	MaxLines		int
	CharsPerLine	int
	ReadingSpeed	float64
	LineBalance		float64
	PreferCompact	bool
	SpacesAsChars	bool
	NewlinesAsChars	bool
	ForbiddenChars	string
}

// AddStringIfNotInArray is a helper function
// to add a string only if it doesn't already exist.
// It is case-sensitive
func AddStringIfNotInArray(arr []string, str string) []string {
	exists := false
	
	if arr == nil {
		arr = make([]string, 0)
	}
	
	for _, elem := range arr {
		if elem == str {
			exists = true
			break
		}
	}
	
	if !exists {
		arr = append(arr, str)
	}
	
	return arr
}

// ParseDuration is the public function for internal parseDuration
func ParseDuration(i string, separator string, length int) (time.Duration, error) {
	return parseDuration(i, separator, length)
}

// Add function uses arithmetic to add the Duration d
// to BOTH start & end times for an Item
func (i *Item) Add(d time.Duration) {
	i.EndAt += d
	i.StartAt += d
	if i.StartAt <= 0 {
		i.StartAt = time.Duration(0)
	}
}

// Within checks if the Duration d is between the current items
// start & end times
func (i *Item) Within(d time.Duration) bool {
	if i.StartAt<=d &&
	   i.EndAt>= d {
		return true
	}
	
	return false
}

// GetRuneCount processes the current item as an utf8 string
// and returns the length in on-screen utf8 characters
func (item *Item) GetRuneCount(params CommandParams) int {
	var s int = 0
	
	for i, l := range item.Lines {
		runeArr := []rune( strip.StripTags(l.String()) )
		s += len(runeArr)
		
		if i < len(item.Lines) - 1 &&
		   params.NewlinesAsChars {
			s += 1
		}
	}

	return s
}

// GetSpeed returns the speed in characters / second of the current
// subtitle. It bases this on utf8 length / duration
func (item *Item) GetSpeed(id int, params CommandParams) float64 {
	length := item.GetLength()
	line_speed := float64(item.GetRuneCount(params)) / length
	text := strip.StripTags(item.String())
	
	fmt.Printf("#%d/Read --> '%s'/length=%g/line_speed=%g\n", id, text, length, line_speed)
	return line_speed
}

// GetLength returns the duration(in seconds) of the current subtitle
func (item *Item) GetLength() float64 {
	return float64(item.EndAt - item.StartAt) / float64(time.Second)
}

// GetExtendBy returns the amount by which a subtitle can be extended.
// This does not consider the direction of extension
// or whether any subtitles exist to left or right of the current one
func (item *Item) GetExtendBy(id int, params CommandParams) float64 {
	line_speed := item.GetSpeed(id, params)
	line_time := item.GetLength()
	
	desired_speed := params.Speed - (params.Speed * params.SpeedEpsilon / 100.0)
	
	extend_by := (line_speed / desired_speed - 1) * line_time
	
	min_length := ( ( 100.0 +
	                  params.SpeedEpsilon ) /
	                100.0) *
	              params.MinLength
	
	if line_time + extend_by < min_length {
		extend_by = min_length - line_time
	}
	
	return extend_by
}

// AdjustStart changes the duration of the current subtitle
// by shifting the start time to the right of present one
// or to the left if the reduce flag is true.
// It takes into account any existing subtitle to the left
// of the current one, and extends only upto the end of the
// subtitle to the left
func (s *Subtitles) AdjustStart(i int, params CommandParams, reduce_by float64) (float64, float64) {
	var item *Item = s.Items[i]
	var last_item *Item = nil
	
	line_time := item.GetLength()
	
	if i > 0 {
		last_item = s.Items[i-1]
	}
	
	last_stop := time.Duration(0)
	if last_item != nil {
		last_stop = last_item.EndAt 
	}
	
	adjusted_by := 0.0
	extend_by := item.GetExtendBy(i+1, params)
	
	if extend_by > 0 && last_stop < item.StartAt {
		last_diff := float64(item.StartAt - last_stop) / float64(time.Second)
		if last_diff > extend_by {
			last_diff = extend_by
		}
		item.StartAt -= time.Duration( last_diff * float64( time.Second ) )
		
		line_time += last_diff
		extend_by -= last_diff
		
		adjusted_by += last_diff
		
		fmt.Printf("id #%d: StartAt decreased by %gs/line_time=%g\n", i+1, last_diff, line_time)
	}
	
	if reduce_by>0 && extend_by < 0 && item.StartAt < item.EndAt {
		last_diff := float64(item.EndAt - item.StartAt) / float64(time.Second)
		
		if math.Abs(extend_by) > reduce_by {
			extend_by = -reduce_by
		}
		
		if last_diff > math.Abs(extend_by) {
			last_diff = math.Abs(extend_by)
		}
		item.StartAt += time.Duration( last_diff * float64( time.Second ) )
		
		line_time -= last_diff
		extend_by += last_diff
		
		adjusted_by -= last_diff
		
		fmt.Printf("id #%d: StartAt increased by %gs/line_time=%g\n", i+1, last_diff, line_time)
	}
	
	return adjusted_by, line_time
}

// AdjustEnd changes the duration of the current subtitle
// by shifting the end time to the right of present one
// or to the left if the reduce flag is true.
// It takes into account any existing subtitle to the right
// of the current one, and extends only upto the start of the
// subtitle to the right
func (s *Subtitles) AdjustEnd(i int, params CommandParams, reduce_by float64) (float64, float64) {
	var item *Item = s.Items[i]
	var next_item *Item = nil
	
	line_time := item.GetLength()
	
	if i+1 < len(s.Items)  {
		next_item = s.Items[i+1]
	}
	
	next_start := item.EndAt
	if next_item != nil {
		next_start = next_item.StartAt
	}
	
	adjusted_by := 0.0
	extend_by := item.GetExtendBy(i+1, params)
	
	if reduce_by == 0 && extend_by > 0 && next_start > item.EndAt {
		next_diff := float64(next_start - item.EndAt) / float64(time.Second)
		
		if next_diff > extend_by {
			next_diff = extend_by
		}
		item.EndAt += time.Duration( next_diff * float64( time.Second ) )
		
		line_time += next_diff
		extend_by -= next_diff
		
		adjusted_by += next_diff
		fmt.Printf("id #%d: EndAt increased by %gs/line_time=%g\n", i+1, next_diff, line_time)
	}
	
	if reduce_by > 0 && extend_by < 0 && item.StartAt < item.EndAt {
		last_diff := float64(item.EndAt - item.StartAt) / float64(time.Second)
		
		if math.Abs(extend_by) > reduce_by {
			extend_by = -reduce_by
		}
		
		if last_diff > math.Abs(extend_by) {
			last_diff = math.Abs(extend_by)
		}
		item.EndAt -= time.Duration( last_diff * float64( time.Second ) )
		
		line_time -= last_diff
		extend_by += last_diff
		
		adjusted_by -= last_diff
		
		fmt.Printf("id #%d: EndAt decreased by %gs/line_time=%g\n", i+1, last_diff, line_time)
	}
	
	return adjusted_by, line_time
}

// AdjustDuration is exported as a function which can be
// called by the main program. It processes each subtitle
// within the Subtitles collection and depending on the
// CommandParams it is called with performs :
// 1. Trim spaces from start and end
// 2. Join short two line subtitles
// 3. Proportionally split a very long subtitle
// 4. Expand or shrink duration of a subtitle to left or right
func (s *Subtitles) AdjustDuration(i int, params CommandParams) int {
	var item *Item = s.Items[i]
	line_speed := item.GetSpeed(i+1, params)
	line_time := item.GetLength()
	
	incBy := 1
	
	min_length := ( ( 100.0 +
	                  params.SpeedEpsilon ) /
	                100.0) *
	              params.MinLength
	
	// Process each line
	for si:=0; si<len(item.Lines); si++ {
		if params.TrimSpaces>0 {
			// Trim spaces to left & right
			lastItem := len(item.Lines[si].Items) - 1
			if lastItem>=0 {
				tleft := strings.TrimLeft(item.Lines[si].Items[0].Text, " ")
				if tleft != item.Lines[si].Items[0].Text {
					fmt.Printf("id #%d: Trimming %d spaces to left of line\n",
								i+1,
								len(item.Lines[si].Items[0].Text) - len(tleft) )
					item.Lines[si].Items[0].Text = tleft
				}
				
				tright := strings.TrimRight(item.Lines[si].Items[lastItem].Text, " ")
				if tright != item.Lines[si].Items[lastItem].Text {
					fmt.Printf("id #%d: Trimming %d spaces to right of line\n",
								i+1,
								len(item.Lines[si].Items[lastItem].Text) - len(tright) )
					item.Lines[si].Items[lastItem].Text = tright
				}
				
				if lastItem>0 &&
				   item.Lines[si].Items[lastItem].Text == "" {
					fmt.Printf("id #%d: Trimming empty space to right of line\n",
								i+1)
					item.Lines[si].Items = item.Lines[si].Items[0:lastItem]
				}
			}
		}
	}
	
	if len(item.Lines)==2 && item.GetRuneCount(params)<params.JoinShorterThan {
		// Join two line subtitle shorter than n characters
		line := item.Lines[0]
		line.Items = append(line.Items, item.Lines[1].Items...)
		item.Lines = []Line{line}
		fmt.Printf(	"id #%d: Joining 2 lines into 1 as shorter than %d characters\n",
					i+1,
					params.JoinShorterThan )
	}
	
	if len(item.Lines)==2 && item.GetLength()>params.SplitLongerThan {
		c := float64(item.GetRuneCount(params))
		l := item.GetLength()
		
		firstItem := *s.Items[i]
		secondItem := *s.Items[i]
		
		firstItem.Lines = []Line{firstItem.Lines[0]}
		firstLen  := l * float64( firstItem.GetRuneCount(params)) / c
		
		firstItem.EndAt = firstItem.StartAt +
						  time.Duration( firstLen * float64(time.Second) )
		
		secondItem.Lines = []Line{s.Items[i].Lines[1]}
		secondItem.StartAt = firstItem.EndAt
		
		//secondLen := l * float64(secondItem.GetRuneCount()) / c
		
		// Split two line subtitle longer than n seconds
		newItems := make([]*Item, 0)
		if i>0 {
			newItems = append(newItems, s.Items[0:i]...)
		}
		newItems = append(newItems, &firstItem, &secondItem)
		if i+1<len(s.Items) {
			newItems = append(newItems, s.Items[i+1:]...)
		}
		s.Items = newItems
		item = s.Items[i]
		
		fmt.Printf(	"id #%d: Splitting proportionately into 2 fragments as longer than %gs\n",
					i+1,
					params.SplitLongerThan )
		
		line_speed = item.GetSpeed(i+1, params)
		line_time = item.GetLength()
	}
	
	if i+1 < len(s.Items) {
		//diff_duration := (s.Items[i+1].StartAt - item.EndAt) * time.Second
		diff_time := float64(s.Items[i+1].StartAt - item.EndAt) / float64(time.Second)
		fmt.Printf("id #%d: diff_time=%g\n", i+1, diff_time)
		
		if diff_time > 0 &&
		   diff_time < params.ExpandCloserThan {
			expand_time := (s.Items[i+1].StartAt - item.EndAt) / 2
		
			if item.GetLength() + params.ExpandCloserThan / 2 < params.ShrinkLongerThan {
				fmt.Printf("id #%d: EndAt expanded by %gs to near half point between %gs distant next subtitles\n",
							i+1,
							diff_time / 2,
							diff_time)
				item.EndAt += expand_time
			}
			if s.Items[i+1].GetLength() + params.ExpandCloserThan / 2 < params.ShrinkLongerThan {
				fmt.Printf("id #%d: StartAt reduced by %gs to near half point between %gs previous subttle\n",
							i+2,
							diff_time / 2,
							diff_time)
				s.Items[i+1].StartAt -= expand_time
			}
		}
	}
	
	if item.GetLength() > params.ShrinkLongerThan {
		item.EndAt = item.StartAt +
					 time.Duration( params.ShrinkLongerThan * float64(time.Second) )
		line_time = item.GetLength()
		fmt.Printf("id #%d: EndAt changed to reduce length/line_time=%g\n",
		           i+1,
		           line_time)
	} else
	if (	line_speed > params.Speed ||
			line_time < min_length ) &&
	   item.GetLength() < params.ShrinkLongerThan {
		extend_by := item.GetExtendBy(i+1, params)
		adjusted_by := 0.0
		
		if extend_by>0 &&
		   item.GetLength() + extend_by > params.ShrinkLongerThan {
			extend_by = item.GetLength() - params.ShrinkLongerThan
		}
		
		//fmt.Printf("#%d/line_speed=%g/reading_speed=%g/last_stop=%d/line_time=%g/extend_by=%g/next_start=%d\n", i+1, line_speed, reading_speed, last_stop, line_time, extend_by, next_start)
		//fmt.Printf("#%d/item=%#v\n", i+1, s.Items[i])
		if extend_by > 0 {
			adjusted_by, line_time = s.AdjustStart(i, params, 0.0)
			extend_by -= adjusted_by
		}
		if extend_by > 0 {
			adjusted_by, line_time = s.AdjustEnd(i, params, 0.0)
			extend_by -= adjusted_by
		}
		
		// More Advanced adjustments
		if extend_by > 0 {
			if i+1 < len(s.Items) {
				// Reduction algorithm will be minus figure
				adjusted_by, line_time = s.AdjustStart(i+1, params, extend_by)
				
				if adjusted_by < 0 {
					adjusted_by, line_time = s.AdjustEnd(i, params, 0.0)
					extend_by -= adjusted_by
				}
			}
			
			if extend_by > 0 && i > 0 {
				// Reduction algorithm will be minus figure
				adjusted_by, line_time = s.AdjustEnd(i-1, params, extend_by)
				
				if adjusted_by < 0 {
					adjusted_by, line_time = s.AdjustStart(i, params, 0.0)
					extend_by -= adjusted_by
				}
			}
		}
	}
	
	return incBy
}

func (s *Subtitles) ShiftUnfit(i int, params CommandParams) {
	var item *Item = s.Items[i]
	
	line_speed := item.GetSpeed(i+1, params)
	line_time := item.GetLength()
	
	min_length := ( ( 100.0 +
	                  params.SpeedEpsilon ) /
	                100.0) *
	              params.MinLength
	
	if line_speed > params.Speed ||
	   line_time < min_length {
		// Subtitle is still unfit
		
		target_length := math.Ceil(
				float64(item.GetRuneCount(params)) /
				float64(params.Speed) *
			1000 ) /
		1000
		
		final_space_required := target_length - line_time
		space_required := final_space_required
		
		//space_shifted := make([]float64, SUB_LEVELS)
		
		fmt.Printf(
			"ShiftUnfit/#%d/line_speed=%g/params.Speed=%g/line_time=%g/min_length=%g/target_length=%g/space_required=%g/Found subtitle which is still unfit!\n",
			i,
			line_speed,
			params.Speed,
			line_time,
			min_length,
			target_length,
			space_required,
		)
		
		for inc := -1;
			inc < 2 && space_required > 0;
			inc += 2 {
			for pass:=0; pass<3; pass++ {
				/*fmt.Printf(
					"ShiftUnfit/#%d/pass=%d/line_speed=%g/line_time=%g/target_length=%g/space_required=%g/Subtitle is still unfit!\n",
					i,
					pass,
					line_speed,
					line_time,
					target_length,
					space_required,
				)*/
			
				// Look behind
				//space_available := space_required
		
				target := i + (inc * SUB_LEVELS)
				for j := i + inc;
					j >=0 && j < len(s.Items) && j != target;
					j += inc {
					currItem := s.Items[j]
					
					move_left := true
					move_right := false
					
					if j > i {
						move_left = false
						move_right = true
					}
					
					var gap_left time.Duration = 0
					var gap_right time.Duration = 0
					var lastItem *Item
					var lastIdx int = -1
					var nextIdx int = -1
					var nextItem *Item
					
					if j - 1 < 0 {
						lastItem = &Item{
							StartAt: 0,
							EndAt: 0,
						}
					} else
					if j - 1 >= 0 {
						lastIdx = j-1
						lastItem = s.Items[lastIdx]
					}
					
					if j + 1 >= len(s.Items) {
						nextItem = &Item{
							StartAt: currItem.EndAt + (2 * time.Second),
							EndAt: currItem.EndAt + (SUB_LEVELS * time.Second),
						}
					} else {
						nextIdx = j+1
						nextItem = s.Items[nextIdx]
					}
					
					gap_left = currItem.StartAt - lastItem.EndAt
					gap_right = nextItem.StartAt - currItem.EndAt// - time.Millisecond
					
					gap_length_left := float64(gap_left) / float64(time.Second)
					gap_length_right := float64(gap_right) / float64(time.Second)
					
					if gap_length_left > space_required {
						gap_length_left = space_required
					}
					
					if move_left &&
					   gap_length_left > 0 &&
					   space_required > 0 {
						shiftBy := time.Duration(gap_length_left * float64(time.Second))
						//resizeBy := currItem.StartAt - lastItem.EndAt - shiftBy
						
						//if lastIdx > -1 &&
						//	resizeBy > 0 {
						//	s.Items[lastIdx].EndAt -= resizeBy
						//}
						
						s.Items[j].StartAt -= shiftBy
						s.Items[j].EndAt -= shiftBy
						
						fmt.Printf(
							"ShiftUnfit/#%d/pass=%d/j=%d/inc=%d/gap_length_left=%g/shiftBy=%d/Shifting subtitle left by %gs\n",
							i,
							pass,
							j,
							inc,
							gap_length_left,
							shiftBy,
							gap_length_left,
						)
						
						if j == i + inc {
							space_required -= gap_length_left
						}
					}
					
					if gap_length_right > space_required {
						gap_length_right = space_required
					}
					
					if move_right &&
					   gap_length_right > 0 &&
					   space_required > 0 {
						shiftBy := time.Duration(gap_length_right * float64(time.Second))
						
						s.Items[j].StartAt += shiftBy
						s.Items[j].EndAt += shiftBy
						
						fmt.Printf(
							"ShiftUnfit/#%d/pass=%d/j=%d/inc=%d/gap_length_right=%g/shiftBy=%d/Shifting subtitle right by %gs\n",
							i,
							pass,
							j,
							inc,
							gap_length_right,
							shiftBy,
							gap_length_right,
						)
					
						if j == i + inc {
							space_required -= gap_length_right
						}
					}
						
					fmt.Printf(
						"ShiftUnfit/#%d/pass=%d/j=%d/inc=%d/gap_length_left=%g/gap_length_right=%g/space_required=%g\n",
						i,
						pass,
						j,
						inc,
						gap_length_left,
						gap_length_right,
						space_required,
					)
					
					if space_required <=0 {
						break
					}
					
					current_length := currItem.GetLength()
					minimum_length := math.Ceil(
							float64(item.GetRuneCount(params)) /
							float64(params.Speed) *
						100.0 ) /
					100.0
					extra_length := current_length - minimum_length
					if extra_length <= 0 {
						continue
					}
					
					if extra_length >= space_required {
						extra_length = space_required
					}
					
					if move_left &&
					   space_required > 0 {
						resizeBy := time.Duration(extra_length * float64(time.Second))
						s.Items[j].EndAt -= resizeBy
						
						if j == i + inc {
							space_required -= extra_length
						}
					}
					if move_right &&
					   space_required > 0 {
						resizeBy := time.Duration(extra_length * float64(time.Second))
						s.Items[j].StartAt += resizeBy
						
						if j == i + inc {
							space_required -= extra_length
						}
					}
					
					
					fmt.Printf(
						"ShiftUnfit/#%d/pass=%d/j=%d/extra_length=%g/space_required=%g\n",
						i,
						pass,
						j,
						extra_length,
						space_required,
					)
				}
			}
		}
			
		//fmt.Printf(
		//	"ShiftUnfit/#%d/final_space_required=%gs\n",
		//	i,
		//	final_space_required,
		//)
					
		if final_space_required > 0 {
			var lastItem *Item
			if i - 1 < 0 {
				lastItem = &Item{
					StartAt: 0,
					EndAt: 0,
				}
			} else {
				lastItem = s.Items[i - 1]
			}
			
			gap_length_left := float64(item.StartAt - lastItem.EndAt) / float64(time.Second)
			fmt.Printf(
				"ShiftUnfit/#%d/final_space_required=%gs/lastItem.StartAt=%d/lastItem.EndAt=%d/item.StartAt=%d/item.EndAt=%d/gap_length_left=%g\n",
				i,
				final_space_required,
				lastItem.StartAt,
				lastItem.EndAt,
				item.StartAt,
				item.EndAt,
				gap_length_left,
			)
			
			if gap_length_left > 0 {
				if gap_length_left > final_space_required {
					gap_length_left = final_space_required
				}
				
				resizeBy := time.Duration(gap_length_left * float64(time.Second))
				s.Items[i].StartAt -= resizeBy
				final_space_required -= gap_length_left
				
				fmt.Printf(
					"ShiftUnfit/#%d/gap_length_left=%g/resizeBy=%d/final_space_required=%g\n",
					i,
					gap_length_left,
					resizeBy,
					final_space_required,
				)
			}
		}
		
		if final_space_required > 0 {
			var nextItem *Item
			if i + 1 >= len(s.Items) {
				nextItem = &Item{
					StartAt: item.EndAt + (2 * time.Second),
					EndAt: item.EndAt + (3 * time.Second),
				}
			} else {
				nextItem = s.Items[i+1]
			}
			
			gap_length_right := float64(nextItem.StartAt - item.EndAt) / float64(time.Second)
			
			fmt.Printf(
				"ShiftUnfit/#%d/final_space_required=%gs/item.StartAt=%d/item.EndAt=%d/nextItem.StartAt=%d/nextItem.EndAt=%d/gap_length_right=%g\n",
				i,
				final_space_required,
				item.StartAt,
				item.EndAt,
				nextItem.StartAt,
				nextItem.EndAt,
				gap_length_right,
			)
			
			if gap_length_right > 0 {
				if gap_length_right > final_space_required {
					gap_length_right = final_space_required
				}
				
				resizeBy := time.Duration(gap_length_right * float64(time.Second))
				s.Items[i].EndAt += resizeBy
				final_space_required -= gap_length_right
				
				fmt.Printf(
					"ShiftUnfit/#%d/gap_length_right=%g/resizeBy=%d/final_space_required=%g\n",
					i,
					gap_length_right,
					resizeBy,
					final_space_required,
				)
			}
		}
	}
}

// AdjustOverlap is exported as a function which can be
// called by the main program. It processes each subtitle
// within the Subtitles collection and depending on the
// CommandParams it is called with performs :
// 1. Expand or shrink duration of a subtitle to left or right

func (s *Subtitles) AdjustOverlap(i int, params CommandParams) int {
	var item *Item = s.Items[i]

	if i+1 < len(s.Items) {
		diff_time := float64(s.Items[i+1].StartAt - item.EndAt) / float64(time.Second)
		
		/*if diff_time > 0 &&
		   diff_time < params.ExpandCloserThan {
			expand_time := (s.Items[i+1].StartAt - item.EndAt) / 2
		
			if item.GetLength() + params.ExpandCloserThan / 2 < params.ShrinkLongerThan {
				fmt.Printf("id #%d: EndAt expanded by %gs to near half point between %gs distant next subtitles\n",
							i+1,
							diff_time / 2,
							diff_time)
				item.EndAt += expand_time
			}
			if s.Items[i+1].GetLength() + params.ExpandCloserThan / 2 < params.ShrinkLongerThan {
				fmt.Printf("id #%d: StartAt reduced by %gs to near half point between %gs previous subttle\n",
							i+2,
							diff_time / 2,
							diff_time)
				s.Items[i+1].StartAt -= expand_time
			}
		} else*/
		if diff_time < 0 {
			fmt.Printf("id #%d: diff_time=%g, < 0, adjusting EndAt\n", i+1, diff_time)
			item.EndAt = s.Items[i+1].StartAt - time.Millisecond
		}
	}

	return 0
}


// PerfectionCheck is exported as a function which can be
// called by the main program. It processes each subtitle
// within the Subtitles collection and performs "Perfection
// check" based on CommandParams it is called with.
func (s *Subtitles) PerfectionCheck(i int, params CommandParams) []string {
	var item *Item = s.Items[i]
	
	perrs := make([]string, 0)
	
	if params.MaxLines>0 && len(item.Lines)>params.MaxLines {
		perrs = AddStringIfNotInArray(perrs, "Too Many lines")
	}
	
	maxLen := 0
	minLen := -1
	plainChars := 0
	
	for i:=0; i<len(item.Lines); i++ {
		plain := strip.StripTags(item.Lines[i].String())
		plain = strings.Trim(plain, " ")
		
		if !params.SpacesAsChars {
			plain = strings.Replace(plain, " ", "", -1)
		}

		if !params.NewlinesAsChars {
			plain += "\n"
		}

		plainRune := []rune(plain)
		
		if len(params.ForbiddenChars)>0 &&
		   len(plainRune) > 0 &&
		   strings.IndexRune(params.ForbiddenChars, plainRune[0]) > -1 {
			err := fmt.Sprintf("No Subtitle line should start with this character - '%s'", plain[0])
			perrs = AddStringIfNotInArray(perrs, err);
		}
		
		if len(plainRune) > maxLen {
			maxLen = len(plainRune)
		}
		if (minLen<0 || len(plainRune) < minLen) {
			minLen = len(plainRune)
		}

		plainChars += len(plainRune)
		if params.CharsPerLine>0 &&
		   len(plainRune) > params.CharsPerLine {
			perrs = AddStringIfNotInArray(perrs, "Too many characters");
		}
	}

	if params.ReadingSpeed > 0 &&
	   item.GetSpeed(i+1, params) > params.ReadingSpeed {
		perrs = AddStringIfNotInArray(perrs, "Reading speed is too high");
	}
	
	if params.LineBalance > 0 &&
	   len(item.Lines)>1 &&
	   ( float64(minLen) / float64(maxLen) ) *
	   100.0 < params.LineBalance {
		perrs = AddStringIfNotInArray(perrs, "Lines are not in balance");
	}
	if params.PreferCompact &&
	   params.CharsPerLine > 0 &&
	   len(item.Lines) > 1 &&
	   plainChars < params.CharsPerLine {
		perrs = AddStringIfNotInArray(perrs, "Multiple lines unnecessarily used");
	}
	
	return perrs
}




