package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	url_base   string
)

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

		userToken, err := decodeToken(agentfiledata)
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
		agentJSON, status, err := getAgentDetails(token_GET)
		if err != nil {
			log.Fatal("Trying to load agent details.\n", err)
		}
		if !strings.Contains(status, "200") {
			fmt.Println(status)
		}

		fmt.Println("Agent Details:\n", agentJSON)
	}
}
