package main

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iterator"
	"strings"
)

func processQuery(query string, topic string, jobProjectID string, operation string) (err error) {
	ctx := context.Background()

	if jobProjectID == "" {
		return fmt.Errorf("jobProjectID must be set")
	}

	if operation == "" {
		return fmt.Errorf("operation must be set")
	} else if strings.ToUpper(operation) != "DELETE" && strings.ToUpper(operation) != "UPSERT" {
		return fmt.Errorf("operation must be UPSERT or DELETE")
	}

	if query == "" {
		return fmt.Errorf("query must be set")
	}

	if topic == "" {
		return fmt.Errorf("topic must be set")
	} else if len(strings.Split(topic, "/")) != 4 {
		return fmt.Errorf("topic must be have the format projects/<projectID>/topics/<topicName>")
	}

	// Create the BiQuery client
	client, err := bigquery.NewClient(ctx, jobProjectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %w\n", err)
	}
	defer client.Close()

	// Create a query
	q := client.Query(query)

	// Run the query
	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("q.Run: %w\n", err)
	}

	// Wait for the query to complete
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("job.Wait: %w\n", err)
	}

	if status.Err() != nil {
		return fmt.Errorf("job.Wait: %w\n", status.Err())
	}

	// for each row, create the corresponding JSON message
	rows, err := job.Read(ctx)
	if err != nil {
		return fmt.Errorf("job.Rows: %w\n", err)
	}

	// get the schema of the table
	schema := rows.Schema

	// Prepare the Pub/Sub client to publish all rows to the topic
	topicClient, err := preparePubSubTopicClient(ctx, topic)
	if err != nil {
		return err
	}

	for {
		var row []bigquery.Value
		if err := rows.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return fmt.Errorf("rows.Next: %w\n", err)
		}

		// Create the JSON row and add the CDC change type information to delete the row
		var jsonRow map[string]interface{}
		jsonRow = make(map[string]interface{})
		for i, value := range row {
			jsonRow[schema[i].Name] = value
		}
		jsonRow["_CHANGE_TYPE"] = operation

		// Convert the row to JSON with the schema
		jsonData, err := json.Marshal(jsonRow)
		if err != nil {
			return fmt.Errorf("json.Marshal: %w\n", err)
		}

		err = publishMessage(topicClient, jsonData)
		if err != nil {
			return fmt.Errorf("publishMessage: %w\n", err)
		}

	}

	return nil
}

func preparePubSubTopicClient(ctx context.Context, topic string) (topicClient *pubsub.Topic, err error) {
	// Extract projectID from the topic string
	projectID := strings.Split(topic, "/")[1]

	// extract the topic name
	topicName := strings.Split(topic, "/")[3]

	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %w\n", err)
	}

	topicClient = pubSubClient.Topic(topicName)
	return

}

func publishMessage(topic *pubsub.Topic, message []byte) error {
	ctx := context.Background()

	result := topic.Publish(ctx, &pubsub.Message{
		Data: message,
	})

	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("result.Get: %w\n", err)
	}
	fmt.Printf("Published message with id %s\n", id)
	return nil

}
