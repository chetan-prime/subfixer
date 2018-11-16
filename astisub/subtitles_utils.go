package astisub

import (
	"fmt"
	"math"
	"time"
)

type CommandParams struct {
	File		string
	Speed		float64
	MinLength	float64
}

func (i *Item) Add(d time.Duration) {
	i.EndAt += d
	i.StartAt += d
	if i.StartAt <= 0 {
		i.StartAt = time.Duration(0)
	}
}

func (item *Item) GetSpeed() float64 {
	text := item.String()
	return float64(len(text)) / item.GetLength()
}

func (item *Item) GetLength() float64 {
	return float64(item.EndAt - item.StartAt) / float64(time.Second)
}

func (item *Item) GetExtendBy(params CommandParams) float64 {
	line_speed := item.GetSpeed()
	line_time := item.GetLength()
	
	extend_by := (line_speed / params.Speed - 1) * line_time
	if line_time + extend_by < params.MinLength {
		extend_by = params.MinLength - line_time
	}
	
	return extend_by
}

func (s *Subtitles) AdjustStart(i int, params CommandParams, reduce bool) (float64, float64) {
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
	extend_by := item.GetExtendBy(params)
	
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
	
	if reduce && extend_by < 0 && item.StartAt < item.EndAt {
		last_diff := float64(item.EndAt - item.StartAt) / float64(time.Second)
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

func (s *Subtitles) AdjustEnd(i int, params CommandParams, reduce bool) (float64, float64) {
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
	extend_by := item.GetExtendBy(params)
	
	if !reduce && extend_by > 0 && next_start > item.EndAt {
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
	
	if reduce && extend_by < 0 && item.StartAt < item.EndAt {
		last_diff := float64(item.EndAt - item.StartAt) / float64(time.Second)
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

func (s *Subtitles) AdjustDuration(i int, params CommandParams) {
	var item *Item = s.Items[i]
	line_speed := item.GetSpeed()
	line_time := item.GetLength()
	
	if line_speed > params.Speed || line_time < params.MinLength {
		extend_by := item.GetExtendBy(params)
		adjusted_by := 0.0
		
		//fmt.Printf("#%d/line_speed=%g/reading_speed=%g/last_stop=%d/line_time=%g/extend_by=%g/next_start=%d\n", i+1, line_speed, reading_speed, last_stop, line_time, extend_by, next_start)
		//fmt.Printf("#%d/item=%#v\n", i+1, s.Items[i])
		if extend_by > 0 {
			adjusted_by, line_time = s.AdjustStart(i, params, false)
			extend_by -= adjusted_by
		}
		if extend_by > 0 {
			adjusted_by, line_time = s.AdjustEnd(i, params, false)
			extend_by -= adjusted_by
		}
		
		// More Advanced adjustments
		if extend_by > 0 {
			if i+1 < len(s.Items) {
				// Reduction algorithm will be minus figure
				adjusted_by, line_time = s.AdjustStart(i+1, params, true)
				
				if adjusted_by < 0 {
					adjusted_by, line_time = s.AdjustEnd(i, params, false)
					extend_by -= adjusted_by
				}
			}
			
			if extend_by > 0 && i > 0 {
				// Reduction algorithm will be minus figure
				adjusted_by, line_time = s.AdjustEnd(i-1, params, true)
				
				if adjusted_by < 0 {
					adjusted_by, line_time = s.AdjustStart(i, params, false)
					extend_by -= adjusted_by
				}
			}
		}
	}
}


