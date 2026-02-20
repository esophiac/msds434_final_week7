package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// define a struct to hold the query results
type Item struct {
	Row       int64   `bigquery:"row"`
	Precision float64 `bigquery:"precision"`
	Recall    float64 `bigquery:"recall"`
	Accuracy  float64 `bigquery:"accuracy"`
	F1score   float64 `bigquery:"f1_score"`
	Logloss   float64 `bigquery:"log_loss"`
	Rocauc    float64 `bigquery:"roc_auc"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT") // Cloud Run provides this env var

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create BigQuery client: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	query := client.Query("SELECT * FROM ML.EVALUATE(MODEL final434.imdbmodel.logistic_reg_classifier)")

	// Run the query
	it, err := query.Read(ctx)
	if err != nil {
		log.Printf("Failed to run query: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Model performance:\n")
	for {
		var item Item
		err := it.Next(&item)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate results: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Precision: %v\nRecall: %v\nAccuracy: %v\nF1_score: %v, Log_loss: %v\nROC AUC: %v", item.Precision, item.Recall, item.Accuracy, item.F1score, item.Logloss, item.Rocauc)
	}
}

func main() {
	http.HandleFunc("/", handler)
	// Use the PORT environment variable provided by Cloud Run
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
