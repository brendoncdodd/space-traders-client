package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	st "github.com/brendoncdodd/space-traders-api"
)

func main() {
	var err error

	//Flag Variables
	var new_agent string
	var faction string
	var get_agent bool
	var save_file string

	//Base options
	if u := *flag.String(
		"base-url",
		st.URL_base.String(),
		"Specifies the base url. Paths to resources will be appended to this. default is https://api.spacetraders.io/v2",
	); u != st.URL_base.String() {
		st.SetBaseURL(u)
	}

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
		agentJson, err := st.CreateAgent(new_agent, faction)
		if err != nil {
			log.Fatal("Failed to create agent.\n\t", err)
		}
		if save_file == "" {
			save_file = new_agent + ".json"
		}
		fmt.Println("Dumping details for new agent.\n", agentJson)
	}

	if save_file != "" {
		if _, err = os.Stat(save_file); os.IsNotExist(err) {
			save_file = "savefiles/" + save_file
			err = nil
			if _, err = os.Stat(save_file); os.IsNotExist(err) {
				log.Fatal("Unable to find save file: " +
					save_file +
					"\n" + err.Error(),
				)
			}
			if err != nil {
				log.Fatal("Save file might exist but there was a problem: " +
					save_file +
					"\n" + err.Error(),
				)
			}

		}
		if err != nil {
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
		//Some old save files got the whole, untrimmed buffer.
		agentfiledata = bytes.TrimRight(agentfiledata, "\x00")

		userToken, err := st.DecodeToken(agentfiledata)
		if err != nil {
			log.Fatal(
				"While trying to get token from save file: " +
					save_file +
					".json\n" +
					err.Error(),
			)
		}

		err = st.LoadToken(string(userToken))
		if err != nil {
			log.Fatal(
				"While trying to load token:",
				err.Error(),
			)
		}
	}

	if get_agent {
		agentJSON, status, err := st.GetAgentDetails(nil)
		if err != nil {
			log.Fatal("Trying to load agent details.\n", err)
		}
		if !strings.Contains(status, "200") {
			fmt.Println(status)
		}

		fmt.Println("Agent Details:\n", agentJSON)
	}
}
