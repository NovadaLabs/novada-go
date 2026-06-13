// Command wallet demonstrates the wallet endpoints.
package main

import (
	"context"
	"fmt"
	"log"

	novada "github.com/NovadaLabs/novada-go"
	"github.com/NovadaLabs/novada-go/wallet"
)

func main() {
	client, err := novada.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	bal, err := client.Wallet.Balance(ctx)
	if err != nil {
		log.Fatalf("balance: %v", err)
	}
	fmt.Printf("wallet balance=%d\n", bal.Balance)

	records, err := client.Wallet.UsageRecord(ctx, wallet.UsageRecordParams{
		Page:  1,
		Limit: 20,
	})
	if err != nil {
		log.Fatalf("usage record: %v", err)
	}
	fmt.Printf("usage records: count=%d\n", records.Count)
	for _, r := range records.List {
		fmt.Printf("  #%d %s %s\n", r.ID, r.OrderType, r.Description)
	}
}
