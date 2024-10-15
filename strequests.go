package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"strings"
)

var (
	token_GET  *http.Request
	token_POST *http.Request
)

type Ship struct {
	symbol string
	nav    struct {
		systemSymbol   string
		waypointSymbol string
	}
}

type Vector2 struct {
	x int
	y int
}

func (self *Vector2) Distance(other Vector2) float64 {
	d := Vector2{0, 0}
	d.x = self.x - other.x
	d.y = self.y - other.y

	if d.x < 0 {
		d.x *= -1
	}
	if d.y < 0 {
		d.y *= -1
	}

	return math.Sqrt(float64(d.x ^ 2 + d.y ^ 2))
}

func init() {
}

// Get a spacetraders.io agent token from some JSON.
// Give this some JSON that follows the pattern:
// { "data": { "token": [TOKEN] } }
func decodeToken(JSONdata []byte) ([]byte, error) {
	var jsonMap map[string]any
	error_prefix := "While trying to decode JSON and get token."

	err := json.Unmarshal(JSONdata, &jsonMap)
	if err != nil {
		return []byte{}, fmt.Errorf("%s %w", error_prefix, err)
	}

	if data, ok := jsonMap["data"].(map[string]any); !ok {
		return nil, fmt.Errorf(
			"%s data is not a map of strings.",
			error_prefix,
		)
	} else if token, ok := data["token"].(string); !ok {
		return nil, fmt.Errorf(
			"%s data.token is not string.",
			error_prefix,
		)
	} else {
		return []byte(token), err
	}
}

// Creates a spacetraders.io agent.
// Writes the (JSON) response body to a file: savefiles/[agent].json
// Returns the entire response body (including the token)
func createAgent(agent string, faction string) (string, error) {
	const BUFFER_SIZE = 20000
	error_prefix := "Trying to create new agent."
	responseBody := make([]byte, BUFFER_SIZE)
	agentMap := make(map[string]string)
	agentMap["symbol"] = agent
	agentMap["faction"] = faction

	agent_json_bytes, err := json.Marshal(agentMap)
	if err != nil {
		return "", fmt.Errorf("%s Marshalling JSON. %w",
			error_prefix,
			err,
		)
	}

	resp, err := http.Post(
		url_base+"/v2/register",
		"application/json",
		strings.NewReader(string(agent_json_bytes)),
	)
	if err != nil {
		return "", fmt.Errorf(
			"%s Trying to POST new agent request.\n%s\n%w",
			error_prefix,
			string(agent_json_bytes),
			err,
		)
	}
	defer resp.Body.Close()

	bodySize, err := resp.Body.Read(responseBody)
	if err != nil {
		return string(responseBody), fmt.Errorf(
			"%s Trying to read body of new agent request. %w",
			error_prefix,
			err,
		)
	}
	responseBody = bytes.TrimRight(responseBody, "\x00")
	responseBody = append(responseBody, byte('\n'))

	err = os.WriteFile("savefiles/"+agent+".json", responseBody, fs.ModePerm)
	if err != nil {
		return string(responseBody), fmt.Errorf(
			"%s Trying to write new agent file. %w",
			error_prefix,
			err,
		)
	}

	if bodySize >= BUFFER_SIZE {
		return string(responseBody), fmt.Errorf(
			"%s Response body too big for buffer (%d bytes).",
			error_prefix,
			BUFFER_SIZE,
		)
	}

	return string(responseBody), err
}

// Not fully implemented. Returns the raw JSON and the request status, in that order.
// Give it an http.Request with the GET method and "Authorization: Bearer [token] header already added.
//
//	or nil to use token_GET
//
// Attempts to preserve the body of the template but if it fails gives it an empty one.
//
//	We don't guard this. Even if you give this an extremely long body it will copy the whole thing twice.
func getAgentDetails(requestTemplate *http.Request) (string, string, error) {
	const BUFFER_SIZE = 10000
	error_prefix := "Trying to get agent details."
	responseBody := make([]byte, BUFFER_SIZE)

	if requestTemplate == nil {
		requestTemplate = token_GET
	}

	req := requestTemplate.Clone(requestTemplate.Context())
	req.Body = io.NopCloser(strings.NewReader(""))
	req.URL.Path = "/v2/my/agent"
	defer req.Body.Close()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", resp.Status, fmt.Errorf(
			"%s Trying to send GET request: %s %w",
			req.URL.String(),
			error_prefix,
			err,
		)
	}
	defer resp.Body.Close()

	bodySize, err := resp.Body.Read(responseBody)
	responseBody = bytes.TrimRight(responseBody, "\x00")
	if err != nil {
		return string(responseBody), resp.Status, fmt.Errorf(
			"%s Trying to read response body. %w",
			error_prefix,
			err,
		)
	}
	if bodySize >= BUFFER_SIZE {
		return string(responseBody), resp.Status, fmt.Errorf(
			"%s Response body too big for buffer (%d bytes).",
			error_prefix,
			BUFFER_SIZE,
		)
	}

	return string(responseBody), resp.Status, err
}

// Not implemented yet
// TODO: Get ship location.
// TODO: Get symbols and locations of waypoint with traits.
// TODO: Remove lines that assign test values.
func findNearestWaypointWithTraits(
	requestTemplate *http.Request,
	shipSymbol string,
	traits []string,
) (
	waypointSymbol string,
	err error,
) {
	const BUFFER_SIZE = 20000
	var shipLocation Vector2
	var rawShipObject map[string]any
	error_prefix := "Trying to find nearest waypoint with traits."
	minDistance := math.Inf(1)
	waypoints := make(map[string]Vector2)
	responseBody := make([]byte, BUFFER_SIZE)

	if requestTemplate == nil {
		requestTemplate = token_GET
	}



	//Get the symbols and locations of all waypoints https://api.spacetraders.io/v2/systems/{systemSymbol}/waypoints
	//Query Parameters
	//	limit
	//	integer
	//	How many entries to return per page
	//	
	//	>= 1
	//	<= 20
	//	Default:
	//	10
	//	page
	//	integer
	//	What entry offset to request
	//	
	//	>= 1
	//	Default:
	//	1
	//	traits
	//	stringarray[string]
	//	
	//	one of: string
	//	The unique identifier of the trait.
	//	
	//	type
	//	string
	//	Filter waypoints by type.
	//	
	//	Allowed values:
	//	PLANET
	//	GAS_GIANT
	//	MOON
	//	ORBITAL_STATION
	//	JUMP_GATE
	//	ASTEROID_FIELD
	//	ASTEROID
	//	ENGINEERED_ASTEROID
	//	ASTEROID_BASE
	//	NEBULA
	//	DEBRIS_FIELD
	//	GRAVITY_WELL
	//	ARTIFICIAL_GRAVITY_WELL
	//	FUEL_STATION


	//Get the ship location https://api.spacetraders.io/v2/my/ships/{shipSymbol}/nav
	req := requestTemplate.Clone(requestTemplate.Context())
	requestTemplate.Body.Close()

	requestTemplate.Body = io.NopCloser(strings.NewReader(""))
	req.Body = io.NopCloser(strings.NewReader(""))
	defer req.Body.Close()

	req.URL.Path = fmt.Sprintf("/v2/my/ships/%s/nav", shipSymbol)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf(
			"%s Request failed. %w",
			error_prefix,
			err,
		)
		if resp.Body != nil {
			resp.Body.Close()
		}
		return
	}
	defer resp.Body.Close()

	responseBodySize, err := resp.Body.Read(responseBody)
	if err != nil {
		err = fmt.Errorf(
			"%s Trying to read response body. %w",
			error_prefix,
			err,
		)
		return
	}
	if responseBodySize >= BUFFER_SIZE {
		err = fmt.Errorf(
			"%s Response too big for buffer (%d bytes)\n%s",
			error_prefix,
			BUFFER_SIZE,
			responseBody,
		)
		return
	}
	responseBody = bytes.TrimRight(responseBody, "\x00")

	err = json.Unmarshal(responseBody, &rawShipObject)
	if err != nil {
		err = fmt.Errorf(
			"%s Failed to unmarshal JSON:\n %s\n%w",
			error_prefix,
			string(responseBody),
			err,
		)
	}

	if data, ok := rawShipObject["data"].(map[string]any); !ok {
		// rawShipObject["data"] not map[string]any
	} else if route, ok := data["route"].(map[string]any); !ok {
		// rawShipObject["data"]["route"] not map[string]any
	} else if destination, ok := route["destination"].(map[string]int); !ok {
		// rawShipObject["data"]["route"]["destination"] not map[string]int
	} else {
		shipLocation = Vector2{destination["x"], destination["y"]}
	}

	//Delete this and create the request for GET https://api.spacetraders.io/v2/systems/{systemSymbol}/waypoints
	req = new(http.Request)
	defer req.Body.Close()

	//Add parameters to the waypoints request
	q := req.URL.Query()
	for _, v := range traits {
		q.Add("traits", v)
	}
	req.URL.RawQuery = q.Encode()

	waypoints["DERP"] = Vector2{3, 4}     //Delete when we actually have waypoints
	waypoints["DERP2"] = Vector2{7, 1024} //Delete when we actually have waypoints
	shipSymbol = "derp"                   //Delete when we actually use shipSymbol
	traits[0] = "derp"                    //Delete when we actually use traits

	for waypoint, location := range waypoints {
		d := shipLocation.Distance(location)
		if d < minDistance {
			waypointSymbol = waypoint
			minDistance = d
		}
	}

	return
}
