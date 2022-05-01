package cyprusbus

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

const limassolSid = "1"

func TestClient_CityRoutesPage(t *testing.T) {
	c := NewClient()
	routes, err := c.CityRoutes(limassolSid)
	assert.NoError(t, err)
	spew.Dump(routes)
	assert.Equal(t, "Limassol Buses", routes.Name)
	if len(routes.Routes) == 0 {
		t.Error("no routes found")
	}
}

func TestClient_RouteInfo(t *testing.T) {
	c := NewClient()
	route, err := c.RouteInfo("41")
	assert.NoError(t, err)
	spew.Dump(route)
	assert.Len(t, route.Ways, 2)
}

func TestClient_RouteInfo_OnlyReverse(t *testing.T) {
	c := NewClient()
	route, err := c.RouteInfo("398")
	assert.NoError(t, err)
	spew.Dump(route)
	assert.Len(t, route.Ways, 1)
	assert.Equal(t, "Reverse Details", route.Ways[0].SelectedTab)
}

func TestClient_RouteInfo_3Ways(t *testing.T) {
	c := NewClient()
	route, err := c.RouteInfo("257")
	assert.NoError(t, err)
	spew.Dump(route)
	assert.Len(t, route.Ways, 2)
}

func TestClient_AllStops(t *testing.T) {
	c := NewClient()
	stops, err := c.AllStops()
	assert.NoError(t, err)
	assert.Zero(t, stops.ErrorMessage)
	assert.Zero(t, stops.ErrorCode)
}
