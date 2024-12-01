package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	webserverFlag := flag.Bool("webserver", false, "Start the webserver. Other parameters are ignored")
	queryFlag := flag.String("query", "", "Query to run into BigQuery. Required if -webserver is not set")
	topicFlag := flag.String("topic", "", "Topic to publish the row to delete. Required if -webserver is not set")
	operationFlag := flag.String("operation", "", "DML operation. Must be UPSERT for update/insert or DELETE for deletion. Required if -webserver is not set")
	jobProjectIdFlag := flag.String("job_project_id", "", "Project ID of the job. Required if -webserver is not set")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Printf("No parameters provided\n")
		cliHelp()
	} else {
		if *webserverFlag {
			startWebserver()
		} else {
			err := processQuery(*queryFlag, *topicFlag, *jobProjectIdFlag, *operationFlag)
			if err != nil {
				fmt.Printf("processQuery: %v\n", err)
				cliHelp()
			}
		}
	}

}

func cliHelp() {
	fmt.Print("\nUsage of cdc-delete [-webserver|(-query=\"<query>\" -job_project_id=\"<job_project_id>\" -topic=\"<topic>\")]\n")
	flag.PrintDefaults()
}

// Start the webserver
func startWebserver() {
	http.HandleFunc("/", dmlEndpoint)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dmlEndpoint(w http.ResponseWriter, r *http.Request) {
	// Extract the body from the request
	body := r.Body
	defer body.Close()

	// Parse the body into an Operation struct
	var operation Operation
	err := json.NewDecoder(body).Decode(&operation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Printf("json parsing error %s\n", err)
		return
	}

	err = processQuery(operation.Query, operation.PubSubTopic, operation.JobProjectID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Printf("processing query error %s\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)

}
