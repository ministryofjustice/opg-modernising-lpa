package donordata

import "time"

const (
	witnessCodeExpireAfter  = 15 * time.Minute
	witnessCodeIgnoreAfter  = 2 * time.Hour
	witnessCodeRequestAfter = time.Minute
)

type WitnessCode struct {
	Code    string
	Created time.Time
}

func (w WitnessCode) HasExpired() bool {
	return w.Created.Add(witnessCodeExpireAfter).Before(time.Now())
}

type WitnessCodes []WitnessCode

func (ws WitnessCodes) Find(code string) (WitnessCode, bool) {
	for _, w := range ws {
		if w.Code == code {
			if w.Created.Add(witnessCodeIgnoreAfter).Before(time.Now()) {
				break
			}

			return w, true
		}
	}

	return WitnessCode{}, false
}

func (ws WitnessCodes) CanRequest(now time.Time) bool {
	if len(ws) == 0 {
		return true
	}

	lastCode := ws[len(ws)-1]
	return lastCode.Created.Add(witnessCodeRequestAfter).Before(now)
}
