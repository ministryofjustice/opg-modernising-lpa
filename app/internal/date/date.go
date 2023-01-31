package date

import "time"

type Date struct {
	Day   string
	Month string
	Year  string
	T     time.Time
	TErr  error
}

func Read(t time.Time) Date {
	return Date{
		Day:   t.Format("2"),
		Month: t.Format("1"),
		Year:  t.Format("2006"),
		T:     t,
	}
}

func FromParts(year, month, day string) Date {
	t, err := time.Parse("2006-1-2", year+"-"+month+"-"+day)

	return Date{
		Day:   day,
		Month: month,
		Year:  year,
		T:     t,
		TErr:  err,
	}
}
