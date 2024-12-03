# Pub/Sub CDC to BigQuery DML Operations

This application faciliates performing Data Manipulation Language (DML) operations (UPSERT and DELETE) on a BigQuery 
table based on messages received from a Pub/Sub topic, effectively enabling Change Data Capture (CDC). It leverages a 
specific message format on the Pub/Sub topic to identify rows for modification and push them in the PubSub CDC topic 
for application to BigQuery

## Concept

This application addresses the challenge of applying DML operations to BigQuery based on real-time changes captured via
Pub/Sub.  While Pub/Sub CDC streaming to BigQuery create limitation in DML usage, it's impossible to use DML operations 
in BigQuery with SQL jobs. This application bridges that gap. More detail in [this article](https://medium.com/google-cloud/bigquery-cdc-with-pubsub-overcoming-limitations-ceae431acfec)

When a row needs to be updated or deleted in BigQuery, a message is published to a designated Pub/Sub topic.  This 
message contains information identifying the row (e.g., primary key values) and the changes to apply. The application 
subscribes to this topic, processes each message, and constructs the corresponding `UPDATE`, `INSERT`, or `DELETE` 
query for BigQuery.

## Use Cases

This approach is essential for several scenarios where standard Pub/Sub to BigQuery streaming falls short:

* **Row Deletion:** BigQuery's CDC streaming ingestion don't support deleting rows based on SQL queries. This
application explicitly handles row deletion by constructing and publishing a deletion CDC message into PubSub.

* **Schema Updates with Default Values:** When adding a new column to a BigQuery table, you might need to update
existing rows with a default value for that column.  This application can handle such updates by constructing and 
publishing a upsert CDC message into PubSub with the new column's default value.
***Important note**: If you use a `_CHANGE_SEQUENCE_NUMBER`, the upsert will lose the latest sequence reference in 
cache. This can lead to inconsistency in case of message duplication and/or late arrival*

It is important to note that direct DML operations via BigQuery SQL is not supported in CDC mode.
This tool addresses this limitation.

## Quick Start

### Prerequisites

* **Google Cloud Project:** An active Google Cloud Project with BigQuery and Pub/Sub APIs enabled.
* **BigQuery Table:** The target BigQuery table where operations will be applied.
* **Pub/Sub Topic:** A Pub/Sub topic for change messages.
* **Service Account:**  A Cloud Run runtime service account with the following roles
  * **BigQuery Job User** in the `JobProjectId` project ID
  * **BigQuery Data Viewer** in the the tables/views use in the `query` parameter
  * **PubSub Publisher** on the `topic` to publish CDCPub/Sub messages into the CDC topic.

### Cloud Run Deployment

1. Use the container image.

Use the existing built container: `us-central1-docker.pkg.dev/gblaquiere-dev/public/pubsub-cdc-bq-dml`

Or build your own
   ```bash
   gcloud builds submit --tag gcr.io/<YOUR_PROJECT_ID>/pubsub-cdc-bq-dml
   ```

2. **Deploy to Cloud Run:**

   ```bash
   gcloud run deploy pubsub-cdc-bq-dml \
     --image us-central1-docker.pkg.dev/gblaquiere-dev/public/pubsub-cdc-bq-dml \
     --platform managed \
     --region <YOUR_REGION> \
     --allow-unauthenticated
   ```
   Replace placeholders with your region.  
   `--allow-unauthenticated` allows public access; adjust security settings as needed.

### Usage - HTTP Web Server Mode (Cloud Run)

Send a POST request to the deployed Cloud Run service with a JSON payload:

```bash
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $(gcloud auth print-access-token)" \
-d '{"jobProjectID":"<YOUR_PROJECT_ID>","query":"<YOUR_QUERY>","pubsubTopic":"projects/<YOUR_ PROJECT_ ID>/topics/<YOUR_TOPIC_NAME>","operation":"[UPSERT|DELETE]"}' \
  <YOUR_CLOUD_RUN_ENDPOINT> 
```
Replace placeholders `<YOUR_PROJECT_ID>`, `<YOUR_QUERY>`, `<YOUR_TOPIC_NAME>` and  `<YOUR_CLOUD_RUN_ENDPOINT>`.
Choose between "UPSERT" or "DELETE" operation.

### Usage - Command-Line Interface (CLI)

1. **Download the binary (choose your os):

* [Linux (AMD64)](https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-linux) - `https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-linux`
* [Windows (AMD64)](https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-windows.exe) - `https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-windows.exe`
* [MacOS (Darwin ARM64)](https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-darwin) - `https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-darwin`

You can get binaries for multiple OS and architectures in Cloud Storage. For example for a linux architecture you could run
```bash
wget https://storage.googleapis.com/pubsub-cdc-bq-dml/cdc-dml-linux && chmod +x cdc-dml-linux
```

2. **Execute the CLI:**

   ```bash
   ./cdc-dml-linux -query="<YOUR_QUERY>" -topic="projects/<YOUR_PROJECT_ID>/topics/<YOUR_TOPIC_NAME>" -job_project_id="<YOUR_PROJECT_ID>" -operation="[UPSERT|DELETE]"
   ```
   Replace placeholders `<YOUR_PROJECT_ID>`, `<YOUR_QUERY>` and `<YOUR_TOPIC_NAME>`.
      Choose between "UPSERT" or "DELETE" operation.

# Licence

This library is licensed under Apache 2.0. Full license text is available in
[LICENSE](https://github.com/guillaumeblaquiere/cdc-dml/tree/main/LICENSE).