package templates

import (
	"fmt"
	"github.com/bakape/meguca/common"
	"html"
	"strconv"
	"time"
)

// CalculateOmit returns the omitted post and image counts for a thread
func CalculateOmit(t common.Thread) (int, int) {
	// There might still be posts missing due to deletions even in complete
	// thread queries. Ensure we are actually retrieving an abbreviated thread
	// before calculating.
	if !t.Abbrev {
		return 0, 0
	}

	var (
		omit    = int(t.PostCount) - (len(t.Posts) + 1)
		imgOmit uint32
	)
	if omit != 0 {
		imgOmit = t.ImageCount
		if t.Image != nil {
			imgOmit--
		}
		for _, p := range t.Posts {
			if p.Image != nil {
				imgOmit--
			}
		}
	}
	return omit, int(imgOmit)
}

func bold(s string) string {
	s = html.EscapeString(s)
	b := make([]byte, 3, len(s)+7)
	copy(b, "<b>")
	b = append(b, s...)
	b = append(b, "</b>"...)
	return string(b)
}

func getTokID(filename string) *string {
	//Max TokID is a week from now
	//Min TokID is 2016-08-01
	now := time.Now()
	maxTokID := now.Add(time.Hour*24*7).Unix() << 32
	var minTokID int64 = 6313705004335104000
	//This function validates tok IDs
	isValidTokID := func(digits string) bool {
		numDigits := len(digits)
		if numDigits > 20 || numDigits < 19 {
			return false
		}
		tokID, _ := strconv.ParseInt(digits, 10, 64)
		return tokID > minTokID && tokID < maxTokID
	}
	digits := ""
	for _, c := range filename {
		if c >= '0' && c <= '9' {
			digits += string(c)
		} else {
			if isValidTokID(digits) {
				return &digits
			}
			digits = ""
		}
	}
	if isValidTokID(digits) {
		return &digits
	}
	return nil
}

func relativeTime(id string) string {
	//convert id to int
	then, _ := strconv.ParseInt(id, 10, 64)
	then = then >> 32
	now := time.Now().Unix()
	timeElapsed := (now - then) / 60
	isFuture := false

	if timeElapsed < 1 {
		if timeElapsed > -5 { // Assume to be client clock imprecision
			return "just now"
		}
		isFuture = true
		timeElapsed = -timeElapsed
	}

	divide := []int64{60, 24, 30, 12}
	threshold := []int64{120, 48, 90, 24}
	units := [][]string{
		{"minute", "minutes"},
		{"hour", "hours"},
		{"day", "days"},
		{"month", "months"},
		{"year", "years"},
	}

	for i, d := range divide {
		if timeElapsed < threshold[i] {
			return ago(timeElapsed, units[i][0], units[i][1], isFuture)
		}
		timeElapsed = timeElapsed / d
	}

	return ago(timeElapsed, units[4][0], units[4][1], isFuture)
}

// Renders "56 minutes ago" or "in 56 minutes" like relative time text
func ago(timeElapsed int64, singular, plural string, isFuture bool) string {
	count := pluralizeTime(timeElapsed, singular, plural)
	if isFuture {
		return fmt.Sprintf("in %s", count)
	}
	return fmt.Sprintf("%s ago", count)
}

func pluralizeTime(num int64, singular, plural string) string {
	if num == 1 || num == -1 {
		return fmt.Sprintf("%d %s", num, singular)
	}
	return fmt.Sprintf("%d %s", num, plural)
}
