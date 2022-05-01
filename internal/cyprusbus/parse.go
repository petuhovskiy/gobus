package cyprusbus

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//<td style="width:85%;">
//<a href="https://www.cyprusbybus.com/RouteDetails.aspx?id=393">
//<b><i>7A</i></b>, My Mall- Ipsonas Industrial Area - Leontiou EMEL Station - Old Market -TEPAK
//</a>
//</td>
func ParseRouteShort(node *html.Node) (*RouteShort, error) {
	doc := goquery.NewDocumentFromNode(node)
	link := doc.Find("a")
	if link.Length() == 0 {
		return nil, fmt.Errorf("no route link found")
	}

	route := &RouteShort{
		DetailsLink: "",
		ID:          "",
		Number:      "",
		Name:        "",
	}
	route.DetailsLink, _ = link.Attr("href")
	href, err := url.Parse(route.DetailsLink)
	if err != nil {
		return nil, err
	}
	route.ID = href.Query().Get("id")
	route.Number = link.Find("b").Text()

	data := strings.SplitN(link.Text(), ",", 2)
	if len(data) != 2 {
		return nil, fmt.Errorf("no name found")
	}
	route.Name = strings.TrimSpace(data[1])
	return route, nil
}

func ParseWay(docOrig *goquery.Document) (*Way, error) {
	doc := docOrig.Clone()
	selectedTab := strings.TrimSpace(doc.Find(".innermenucell-sel").Text())

	var stops []StopRow
	rows := doc.Find("tr.routedisplay-row")
	for _, row := range rows.Nodes {
		stop, err := ParseStopRow(row)
		if err != nil {
			return nil, err
		}

		stops = append(stops, *stop)
	}

	return &Way{
		SelectedTab: selectedTab,
		Stops:       stops,
	}, nil
}

func ParseStopRow(node *html.Node) (*StopRow, error) {
	doc := goquery.NewDocumentFromNode(node)
	districtName := strings.TrimSpace(doc.Find("tr > td > span > center > b > i").Text())

	link := doc.Find("tr.routedisplay-row > td:nth-child(3) > a:nth-child(1)")
	if link.Length() == 0 {
		return nil, fmt.Errorf("no stop link found")
	}

	stop := &StopRow{
		District:    districtName,
		DetailsLink: "",
		PathID:      "",
		Name:        "",
	}

	stop.DetailsLink, _ = link.Attr("href")
	if stop.DetailsLink == "" {
		return nil, fmt.Errorf("stop link empty")
	}
	tmp := strings.Split(stop.DetailsLink, "/")
	stop.PathID = tmp[len(tmp)-1]
	stop.Name = strings.TrimSpace(link.Text())

	return stop, nil
}

func ParseTimetable(sel *goquery.Selection) (*Timetable, error) {
	if sel.Length() != 1 {
		return nil, fmt.Errorf("invalid selection for timetable")
	}

	inner := sel.Find("table > tbody > tr:nth-child(1) > td > table > tbody > tr:nth-child(3) > td > table > tbody > tr > td > table > tbody > tr:nth-child(2) > td")
	if inner.Length() != 1 {
		return nil, fmt.Errorf("cannot find inner table with times")
	}

	durationSel := inner.Find("span > b")
	if durationSel.Length() == 0 {
		return nil, fmt.Errorf("cannot find duration span")
	}

	if durationSel.Text() != "Journey Duration from start to finish:" {
		return nil, fmt.Errorf("expected duration, got %s", durationSel.Text())
	}

	durationNode := durationSel.Nodes[0]
	for durationNode.NextSibling != nil {
		durationNode = durationNode.NextSibling
	}

	if durationNode.Type != html.TextNode {
		return nil, fmt.Errorf("expected duration text, got type=%v", durationNode.Type)
	}

	res := &Timetable{
		Name:     strings.TrimSpace(sel.Find("h3").Text()),
		Duration: durationNode.Data,
	}

	dowNodes := inner.Find("td > table > tbody > tr > td > table > tbody > tr > td")
	for _, dow := range dowNodes.Nodes {
		sel := goquery.NewDocumentFromNode(dow).Find("tbody")
		if sel.Length() != 1 {
			continue
		}
		daysText := sel.Find("b").Text()
		days := daysTextToWeekdays(daysText)

		timesStr := strings.Split(sel.Find("tbody > tr:nth-child(2) > td:nth-child(1)").Text(), ",")
		if len(timesStr) == 1 && timesStr[0] == "" {
			timesStr = nil
		}
		departures := []Time{}
		for _, t := range timesStr {
			tt, err := ParseTime(t)
			if err != nil {
				return nil, err
			}
			departures = append(departures, tt)
		}

		for _, day := range days {
			if res.Weekdays[day].DepartureTimes != nil {
				return nil, fmt.Errorf("duplicate day %v", day)
			}
			res.Weekdays[day].DepartureTimes = departures
		}
	}

	return res, nil
}

func daysTextToWeekdays(text string) []time.Weekday {
	text = strings.TrimSpace(text)
	switch text {
	case "Monday to Friday":
		return []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}
	case "Saturday":
		return []time.Weekday{time.Saturday}
	case "Sunday":
		return []time.Weekday{time.Sunday}
	case "Sunday, Saturday":
		return []time.Weekday{time.Saturday, time.Sunday}
	case "Monday to Tuesday":
		return []time.Weekday{time.Monday, time.Tuesday}
	case "Thursday to Friday":
		return []time.Weekday{time.Thursday, time.Friday}
	case "Everyday":
		return []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}
	case "Monday":
		return []time.Weekday{time.Monday}
	case "Tuesday":
		return []time.Weekday{time.Tuesday}
	case "Wednesday":
		return []time.Weekday{time.Wednesday}
	case "Friday":
		return []time.Weekday{time.Friday}
	case "Thursday":
		return []time.Weekday{time.Thursday}
	}

	panic("unknown weekdays: " + text)
}

func ParseTime(text string) (Time, error) {
	text = strings.TrimSpace(text)
	var res Time
	_, err := fmt.Sscanf(text, "%d:%d", &res.Hours, &res.Minutes)
	if err != nil {
		return Time{}, fmt.Errorf("failed to parse time %s: %w", text, err)
	}
	return res, nil
}

func ParseStopInfo(ext ExtStopInfo) (StopInfo, error) {
	long, err := strconv.ParseFloat(ext.Longitude, 64)
	if err != nil {
		return StopInfo{}, err
	}
	lat, err := strconv.ParseFloat(ext.Latitude, 64)
	if err != nil {
		return StopInfo{}, err
	}

	return StopInfo{
		ID:        ext.ID,
		Title:     ext.Title,
		Longitude: long,
		Latitude:  lat,
	}, nil
}
