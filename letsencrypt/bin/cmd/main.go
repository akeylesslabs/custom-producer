package main

import (
	"log"
	"net/http"
	"os"

	"github.com/akeylesslabs/custom-producer/letsencrypt/internal/producer"
	"github.com/akeylesslabs/custom-producer/letsencrypt/internal/webhook"
)

func main() {
	p, err := producer.New(
		producer.WithDryRunEmail(os.Getenv("LE_DRY_RUN_EMAIL")),
		producer.WithDryRunDomain(os.Getenv("LE_DRY_RUN_DOMAIN")),
	)
	if err != nil {
		log.Fatalln(err)
	}

	h, err := webhook.New(
		p,
		webhook.WithAllowedAccessID(os.Getenv("AKEYLESS_ACCESS_ID")),
		webhook.WithAllowedItemName(os.Getenv("AKEYLESS_ITEM_NAME")),
	)
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatalln(http.ListenAndServe(":80", h))
}
