package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

var (
	url_base   string
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

// Not implemented
func (self *Ship) DownloadDetailsBySymbol() (err error) {
	return
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

func createAgent(agent string, faction string) (string, error) {
	responseBuffer := make([]byte, 20000)
	agentMap := make(map[string]string)
	agentMap["symbol"] = agent
	agentMap["faction"] = faction

	agent_json_bytes, err := json.Marshal(agentMap)
	if err != nil {
		return "", errors.New("Trying to create JSON for new agent request.\t" + err.Error())
	}

	resp, err := http.Post(url_base+"/v2/register", "application/json", strings.NewReader(string(agent_json_bytes)))
	if err != nil {
		return "", errors.New("Trying to POST new agent request: \n\t" + string(agent_json_bytes) + "\n\t" + err.Error())
	}
	defer resp.Body.Close()

	_, err = resp.Body.Read(responseBuffer)
	if err != nil {
		return string(responseBuffer), errors.New("Trying to read body of new agent request: " + err.Error())
	}

	responseBody := bytes.TrimRight(responseBuffer, "\x00")
	responseBody = append(responseBody, byte('\n'))

	err = os.WriteFile("savefiles/agent_"+agent+".json", responseBody, fs.ModePerm)
	if err != nil {
		return string(responseBody), errors.New("Trying to write new agent file: " + err.Error())
	}

	return string(responseBody), err
}

// Not fully implemented. Returns the raw JSON and the request status.
// Attempts to preserve the body of the template but if it fails gives it an empty one. We don't guard this. Even if you give this an extremely long body it will copy the whole thing twice.
// Give it a template with the "Authorization: Bearer [token] header already added.
func getAgentDetails(requestTemplate *http.Request) (agent string, status string, err error) {
	bodyBuf := make([]byte, 10000) //This is the buffer for the response body. It's probably a good idea not to let the responder give us as much data as they want.
	var jsonBuf []byte

	oldBody, err := io.ReadAll(requestTemplate.Body)
	if err != nil {
		oldBody = []byte("")
	}

	err = token_GET.Body.Close()
	if err != nil {
		return "", "", errors.New("While trying getAgentDetails. While trying to close token_GET.Body after cloning.\n" + err.Error())
	}

	req := token_GET.Clone(token_GET.Context())
	token_GET.Body = io.NopCloser(bytes.NewReader(oldBody))
	req.Body = io.NopCloser(strings.NewReader(""))
	defer req.Body.Close()

	req.URL.Path = "/v2/my/agent"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", resp.Status, errors.New("While trying to send GET request for agent:\t" + err.Error())
	}
	defer resp.Body.Close()

	status = resp.Status

	bodySize, err := resp.Body.Read(bodyBuf)
	if len(bodyBuf) >= bodySize {
		jsonBuf = make([]byte, bodySize)
		copy(jsonBuf, bodyBuf)
	} else {
		jsonBuf = bodyBuf
	}
	agent = string(jsonBuf)

	return
}

// Not implemented yet
// TODO: Get ship location.
// TODO: Get symbols and locations of waypoint with traits.
// TODO: Remove lines that assign test values.
func findNearestWaypointWithTraits(shipSymbol string, traits []string) (waypointSymbol string, err error) {
	minDistance := math.MaxFloat64
	waypoints := make(map[string]Vector2)
	responseBuffer := make([]byte, 20000)

	//Convert traits to a format we can use in the request.

	//Get the symbols and locations of all waypoints in the system

	//Get the ship location (where you want to find nearest to) https://api.spacetraders.io/v2/my/ships/{shipSymbol}/nav
	req := token_GET.Clone(token_GET.Context())
	token_GET.Body.Close()

	token_GET.Body = io.NopCloser(strings.NewReader(""))
	req.Body = io.NopCloser(strings.NewReader(""))
	defer req.Body.Close()
	req.URL.Path = fmt.Sprintf("/v2/my/ships/%s/nav", shipSymbol)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		//TODO: Handle failed request
	}
	defer resp.Body.Close()

	shipRequestlen, _ := resp.Body.Read(responseBuffer)
	shipRequestBody := make([]byte, shipRequestlen)

	copy(shipRequestBody, responseBuffer)

	var rawShipObject map[string]any
	err = json.Unmarshal(shipRequestBody, &rawShipObject)
	if err != nil {
		//TODO: Handle json unmarshal error
	}

	shipLocation := Vector2{rawShipObject["data"].(map[string]any)["route"].(map[string]any)["destination"].(map[string]int)["x"], rawShipObject["data"].(map[string]any)["route"].(map[string]any)["destination"].(map[string]int)["y"]}

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

func getToken(data []byte) ([]byte, error) {
	var jsonMap map[string]any

	err := json.Unmarshal(data, &jsonMap)
	if err != nil {
		return []byte{}, errors.New("While trying to get token, while trying to decode JSON.\n" + err.Error())
	}

	if token, ok := jsonMap["data"].(map[string]any)["token"].(string); ok {
		return []byte(token), err
	}

	return []byte{}, errors.New("Could not get token. Here's the JSON.\n" + string(data))
}

func main() {
	//Flag Variables
	var new_agent string
	var faction string
	var get_agent bool
	var save_file string

	//Base options
	flag.StringVar(
		&url_base,
		"base-url",
		"https://api.spacetraders.io",
		"Specifies the base url. Paths to resources will be appended to this. default is https://api.spacetraders.io/v2",
	)

	// Options for creating a new agent
	flag.StringVar(
		&new_agent,
		"create-user",
		"",
		"Creates a new Space Traders agent account.\nSave files are of the form savefiles/[agent].json",
	)
	flag.StringVar(
		&faction,
		"faction",
		"COSMIC",
		"Specifies the faction. Default is COSMIC",
	)

	//Options for getting agent data
	flag.BoolVar(
		&get_agent,
		"get-agent",
		false,
		"View the agent details.",
	)

	//Misc options
	flag.StringVar(&save_file,
		"savefile",
		"",
		"Specify the save file\nIf you are creating a new agent, the new agent will get their own save file. This file will be used for any other operations.",
	)

	flag.Parse()

	if new_agent != "" {
		fmt.Println("Creating Agent: ", new_agent)
		agentJson, err := createAgent(new_agent, faction)
		if err != nil {
			log.Fatal("Failed to create agent.\n\t", err)
		}
		if save_file == "" {
			save_file = new_agent + ".json"
		}
		fmt.Println("Dumping details for new agent.\n", agentJson)
	}

	if save_file != "" {
		if _, err := os.Stat(save_file); os.IsNotExist(err) {
			save_file = "savefiles/" + save_file
			if _, err := os.Stat(save_file); os.IsNotExist(err) {
				log.Fatal("Unable to find save file: " +
					save_file +
					"\n" + err.Error(),
				)
			} else if err != nil {
				log.Fatal("Save file might exist but there was a problem: " +
					save_file +
					"\n" + err.Error(),
				)
			}

		} else if err != nil {
			log.Fatal("Save file might exist but there was a problem: " +
				save_file +
				"\n" + err.Error(),
			)
		}

		agentfiledata, err := os.ReadFile(save_file)
		if err != nil {
			log.Fatal(
				"While trying to read save file: " +
					save_file +
					err.Error(),
			)
		}
		agentfiledata = bytes.TrimRight(agentfiledata, "\x00")

		userToken, err := getToken(agentfiledata)
		if err != nil {
			log.Fatal(
				"While trying to get token from save file: savefiles/" +
					save_file +
					".json\n" +
					err.Error(),
			)
		}
		token_GET, err = http.NewRequest(
			"GET",
			url_base,
			strings.NewReader(""),
		)
		if err != nil {
			log.Fatal(
				"While trying to create GET request template with agent token.\n",
				err.Error())
		}
		token_GET.Header.Add(
			"Authorization",
			"Bearer "+string(userToken),
		)

		token_POST, err = http.NewRequest(
			"POST",
			url_base,
			strings.NewReader(""),
		)
		if err != nil {
			log.Fatal(
				"While trying to create POST request template with agent token.\n",
				err.Error(),
			)
		}
		token_POST.Header.Add(
			"Authorization",
			"Bearer "+string(userToken),
		)
	}

	if get_agent {
		agentJSON, _, err := getAgentDetails(token_GET)
		if err != nil {
			log.Fatal("Trying to load agent details.\n", err)
		}

		fmt.Println("Agent Details:\n", agentJSON)
	}
}
