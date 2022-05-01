package cyprusbus

type CityRoutesPage struct {
	Sid    string
	Name   string
	Routes []*RouteShort
}

type RouteShort struct {
	DetailsLink string
	ID          string
	Number      string
	Name        string
}

type RouteFull struct {
	ID   string
	Ways []Way
}

type Way struct {
	SelectedTab string
	Stops       []StopRow
	Timetable   *Timetable
}

type StopRow struct {
	District    string
	DetailsLink string
	PathID      string
	StopID      int
	Name        string
}

type Timetable struct {
	Name     string
	Duration string
	Weekdays [7]Weekday
}

type Weekday struct {
	DepartureTimes []Time
}

type Time struct {
	Hours   int
	Minutes int
}

type ExtAllStops struct {
	ErrorCode    int           `json:"errorCode"`
	ErrorMessage string        `json:"errorMessage"`
	Data         []ExtStopInfo `json:"data"`
}

type ExtStopInfo struct {
	Order     int    `json:"order"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

type StopInfo struct {
	ID        int
	Title     string
	Longitude float64
	Latitude  float64
}
