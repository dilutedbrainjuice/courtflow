package api

import (
	"context"
	"time"
)

// court combines court pricing with the opening schedule inherited from its venue.
// Price is integer Malaysian sen; Open and Close are whole hours in Zone.
type court struct {
	Price, Open, Close int
	Zone               string
}

// getCourt loads the configuration shared by availability and booking validation.
// A missing court returns pgx.ErrNoRows; other errors are passed to the handler.
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

// validSlot checks start-time eligibility, not occupancy. A start must be strictly
// after now, no later than now.AddDate(0, 0, 30), and on a whole hour in the venue
// timezone. Close is exclusive, so the last one-hour session starts at Close - 1.
// Invalid timezone configuration makes the start ineligible.
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
