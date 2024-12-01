package main

// Operation represents the parameters for a CDC operation.
type Operation struct {
	JobProjectID string `json:"jobProjectID"` // The Google Cloud project ID where the BigQuery job will run.
	Query        string `json:"query"`        // The SQL query to execute against BigQuery.
	PubSubTopic  string `json:"pubsubTopic"`  // The full path to the Pub/Sub topic (e.g., "projects/my-project/topics/my-topic").
	Operation    string `json:"operation"`    // The DML operation type: "UPSERT" or "DELETE".
}
