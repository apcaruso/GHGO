package vocab

import (
	"strconv"
	"strings"
)

type Month int

const (
	MonthJanuary   Month = 1
	MonthFebruary  Month = 2
	MonthMarch     Month = 3
	MonthApril     Month = 4
	MonthMay       Month = 5
	MonthJune      Month = 6
	MonthJuly      Month = 7
	MonthAugust    Month = 8
	MonthSeptember Month = 9
	MonthOctober   Month = 10
	MonthNovember  Month = 11
	MonthDecember  Month = 12
)

var monthAliases = map[string]Month{
	"january":   MonthJanuary,
	"jan":       MonthJanuary,
	"gennaio":   MonthJanuary,
	"gen":       MonthJanuary,
	"february":  MonthFebruary,
	"feb":       MonthFebruary,
	"febbraio":  MonthFebruary,
	"march":     MonthMarch,
	"mar":       MonthMarch,
	"marzo":     MonthMarch,
	"april":     MonthApril,
	"apr":       MonthApril,
	"aprile":    MonthApril,
	"may":       MonthMay,
	"maggio":    MonthMay,
	"mag":       MonthMay,
	"june":      MonthJune,
	"jun":       MonthJune,
	"giugno":    MonthJune,
	"giu":       MonthJune,
	"july":      MonthJuly,
	"jul":       MonthJuly,
	"luglio":    MonthJuly,
	"lug":       MonthJuly,
	"august":    MonthAugust,
	"aug":       MonthAugust,
	"agosto":    MonthAugust,
	"ago":       MonthAugust,
	"september": MonthSeptember,
	"sep":       MonthSeptember,
	"sept":      MonthSeptember,
	"settembre": MonthSeptember,
	"set":       MonthSeptember,
	"sett":      MonthSeptember,
	"october":   MonthOctober,
	"oct":       MonthOctober,
	"ottobre":   MonthOctober,
	"ott":       MonthOctober,
	"november":  MonthNovember,
	"nov":       MonthNovember,
	"novembre":  MonthNovember,
	"december":  MonthDecember,
	"dec":       MonthDecember,
	"dicembre":  MonthDecember,
	"dic":       MonthDecember,
}

func ParseMonth(s string) (Month, bool) {
	s = NormalizeSpace(s)
	if s == "" {
		return 0, false
	}

	if m, ok := parseYearMonth(s); ok {
		return m, true
	}

	if allDigits(s) {
		return parseMonthNumber(s)
	}

	m, ok := monthAliases[Key(s)]
	return m, ok
}

func (m Month) Valid() bool {
	return m >= MonthJanuary && m <= MonthDecember
}

func (m Month) EnglishName() string {
	switch m {
	case MonthJanuary:
		return "January"
	case MonthFebruary:
		return "February"
	case MonthMarch:
		return "March"
	case MonthApril:
		return "April"
	case MonthMay:
		return "May"
	case MonthJune:
		return "June"
	case MonthJuly:
		return "July"
	case MonthAugust:
		return "August"
	case MonthSeptember:
		return "September"
	case MonthOctober:
		return "October"
	case MonthNovember:
		return "November"
	case MonthDecember:
		return "December"
	}
	return ""
}

func (m Month) Number() int {
	if !m.Valid() {
		return 0
	}
	return int(m)
}

func parseYearMonth(s string) (Month, bool) {
	sep := ""
	if strings.Count(s, "-") == 1 {
		sep = "-"
	} else if strings.Count(s, "/") == 1 {
		sep = "/"
	}
	if sep == "" {
		return 0, false
	}

	parts := strings.Split(s, sep)
	if len(parts[0]) != 4 || len(parts[1]) != 2 || !allDigits(parts[0]) || !allDigits(parts[1]) {
		return 0, false
	}

	return parseMonthNumber(parts[1])
}

func parseMonthNumber(s string) (Month, bool) {
	if len(s) < 1 || len(s) > 2 {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	m := Month(n)
	return m, m.Valid()
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
