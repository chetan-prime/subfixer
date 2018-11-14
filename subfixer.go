package main

import (
	"./astisub"
	"fmt"
	"os"
	"strconv"
	"time"
)


func main() {
	last_stop := 0.0
	if len(os.Args) > 4 {
		fname := os.Args[1]
		i, _ := strconv.Atoi(os.Args[2])
		i -= 1
		
		start_adj, _ := strconv.ParseInt(os.Args[3], 10, 64)
		stop_adj, _  := strconv.ParseInt(os.Args[4], 10, 64)
		
		last_stop =
		
		// Open
		s, err := astisub.OpenFile( fname )
		
		if err!=nil {
			fmt.Printf("Error opening %s: %s\n", fname, err)
			return
		}
		
		s.Items[i].StartAt += time.Duration(start_adj) * time.Millisecond
		s.Items[i].EndAt += time.Duration(stop_adj) * time.Millisecond
		
		fmt.Printf("Now saving changes to file %s: ", fname)
		s.Write(fname)
		fmt.Printf("[DONE]\n")
	} else {
		fmt.Printf("Format is below :\n%s <filename.srt> <index> <start_adjust> <stop_adjust>\n All adjust entries are in MilliSeconds\n", os.Args[0])
	}
}


