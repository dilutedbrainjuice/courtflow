package api

import (
	"context"
	"time"
)

type court struct {
	Price, Open, Close int
	Zone               string
}

func (a *API) getCourt(ctx context.Context, id int) (court, error) {
	var c court
	err := a.DB.QueryRow(ctx, `
		SELECT c.price_sen,f.opens,f.closes,f.timezone
		FROM courts c
		JOIN facilities f ON f.id=c.facility_id
		WHERE c.id=$1
	`, id).Scan(&c.Price, &c.Open, &c.Close, &c.Zone)
	return c, err
}

// validSlot checks a future, whole-hour start against the venue clock and booking horizon.
func validSlot(t, now time.Time, c court) bool {
	loc, err := time.LoadLocation(c.Zone)
	if err != nil {
		return false
	}
	v := t.In(loc)
	withinHorizon := t.After(now) && !t.After(now.AddDate(0, 0, 30))
	onHour := v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0
	withinOpeningHours := v.Hour() >= c.Open && v.Hour() < c.Close
	return withinHorizon && onHour && withinOpeningHours
}
