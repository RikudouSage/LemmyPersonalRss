package dto

import "time"

type DateTime struct {
	time.Time
}

func (receiver *DateTime) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02T15:04:05.999999Z"`, string(b))
	if err != nil {
		return err
	}
	receiver.Time = date
	return
}
