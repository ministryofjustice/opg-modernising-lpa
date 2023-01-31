package date

import "time"

type Date struct {
	T     time.Time
	Err   error
	Day   string
	Month string
	Year  string
}

func Read(t time.Time) Date {
	return Date{
		T:     t,
		Day:   t.Format("2"),
		Month: t.Format("1"),
		Year:  t.Format("2006"),
	}
}

func FromParts(year, month, day string) Date {
	t, err := time.Parse("2006-1-2", year+"-"+month+"-"+day)

	return Date{
		T:     t,
		Err:   err,
		Day:   day,
		Month: month,
		Year:  year,
	}
}
