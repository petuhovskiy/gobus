package cyprusbus

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
)

type opts struct {
	form url.Values
}

type Client struct {
	baseURL        string
	dataserviceURL string
	client         *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:        "https://www.cyprusbybus.com",
		dataserviceURL: "https://api.cyprusbybus.com",
		client:         http.DefaultClient,
	}
}

func (c *Client) fetch(opts opts, formatPath string, args ...interface{}) (*goquery.Document, error) {
	path := c.baseURL + fmt.Sprintf(formatPath, args...)

	var (
		req *http.Request
		err error
	)

	if opts.form == nil {
		req, err = http.NewRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest("POST", path, strings.NewReader(opts.form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return goquery.NewDocumentFromReader(res.Body)
}

func (c *Client) fetchDataservice(path string, dst interface{}) error {
	path = c.dataserviceURL + path

	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequest("GET", path, nil)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(dst)
}

func (c *Client) CityRoutes(sid string) (*CityRoutesPage, error) {
	doc, err := c.fetch(opts{}, "/routes.aspx?sid=%v", sid)
	if err != nil {
		return nil, err
	}

	var routes []*RouteShort

	sel := doc.Find("#ctl00_ContentPlaceHolder1_gvRoutes > tbody > tr > :first-child")
	for _, route := range sel.Nodes {
		//infoNode := route.FirstChild
		//if infoNode == nil {
		//	log.WithField("route", route).Error("route has no info node")
		//	continue
		//}

		res, err := ParseRouteShort(route)
		if err != nil {
			log.WithField("route", route).WithError(err).Error("failed to parse route")
			continue
		}
		routes = append(routes, res)
	}

	return &CityRoutesPage{
		Sid:    sid,
		Routes: routes,
		Name:   strings.TrimSpace(doc.Find(".route-paths-heading > table > tbody > tr > :nth-child(3)").Text()),
	}, nil
}

func (c *Client) RouteInfo(routeID string) (*RouteFull, error) {
	const reverseTarget = "ctl00$ContentPlaceHolder1$lbReverse"
	const timetableTarget = "ctl00$ContentPlaceHolder1$lbTimes"

	doc, err := c.fetch(opts{}, "/RouteDetails.aspx?id=%v", routeID)
	if err != nil {
		return nil, fmt.Errorf("route %s: %w", routeID, err)
	}

	hasDirect := !strings.Contains(doc.Text(), "No Details Found. The bus route maybe circular.")

	reverseBtn := doc.Find("#ctl00_ContentPlaceHolder1_lbReverse")
	hasReverse := reverseBtn.Length() != 0

	var (
		dir1          *Way
		dir2          *Way
		timetableDir1 *Timetable
		timetableDir2 *Timetable
	)

	if hasDirect {
		dir1, err = ParseWay(doc)
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", routeID, err)
		}
	}

	reverseForm, err := c.aspnet(doc, reverseTarget)
	if err != nil {
		return nil, fmt.Errorf("route %s: %w", routeID, err)
	}

	timetableForm, err := c.aspnet(doc, timetableTarget)
	if err != nil {
		return nil, fmt.Errorf("route %s: %w", routeID, err)
	}

	if hasReverse {
		doc, err = c.fetch(opts{
			form: reverseForm,
		}, "/RouteDetails.aspx?id=%v", routeID)
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", routeID, err)
		}

		dir2, err = ParseWay(doc)
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", routeID, err)
		}
	}

	doc, err = c.fetch(opts{
		form: timetableForm,
	}, "/RouteDetails.aspx?id=%v", routeID)
	if err != nil {
		return nil, fmt.Errorf("route %s: %w", routeID, err)
	}

	var ways []Way

	if hasDirect {
		timetableDir1, err = ParseTimetable(doc.Find("#ctl00_ContentPlaceHolder1_RouteScheduleTimes_TopGroupsList1"))
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", routeID, err)
		}
		dir1.Timetable = timetableDir1
		ways = append(ways, *dir1)
	}

	if hasReverse {
		timetableDir2, err = ParseTimetable(doc.Find("#ctl00_ContentPlaceHolder1_RouteReversedScheduleTimes_TopGroupsList1"))
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", routeID, err)
		}
		dir2.Timetable = timetableDir2
		ways = append(ways, *dir2)
	}

	return &RouteFull{
		ID:   routeID,
		Ways: ways,
	}, nil
}

func (c *Client) aspnet(docOrig *goquery.Document, customEventTarget string) (url.Values, error) {
	doc := docOrig.Clone()

	list := []string{
		"__EVENTARGUMENT",
		"__VIEWSTATE",
		"__VIEWSTATEGENERATOR",
		"__EVENTVALIDATION",
	}

	input := doc.Find("input")
	found := map[string]string{}
	for _, node := range input.Nodes {
		attrs := map[string]string{}
		for _, attr := range node.Attr {
			attrs[attr.Key] = attr.Val
		}

		name, ok := attrs["name"]
		if !ok {
			continue
		}

		value, ok := attrs["value"]
		if !ok {
			continue
		}

		found[name] = value
	}

	res := url.Values{}

	const usePredefinedKeys = false
	if usePredefinedKeys {
		for _, key := range list {
			attr, ok := found[key]
			if !ok {
				return nil, fmt.Errorf("aspnet value doesn't exist")
			}

			res.Set(key, attr)
		}
	} else {
		for k, v := range found {
			res.Set(k, v)
		}
	}

	res.Set("__EVENTTARGET", customEventTarget)
	res.Set("__EVENTARGUMENT", "")
	return res, nil
}

func (c *Client) AllStops() (*ExtAllStops, error) {
	var res ExtAllStops
	err := c.fetchDataservice("/dataservice/api/v1/stops/getallstops", &res)
	return &res, err
}
