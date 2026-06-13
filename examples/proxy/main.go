// Command proxy demonstrates the proxy management endpoints.
package main

import (
	"context"
	"fmt"
	"log"

	novada "github.com/NovadaLabs/novada-go"
	"github.com/NovadaLabs/novada-go/proxy"
)

func main() {
	// API key is read from the NOVADA_API_KEY environment variable.
	client, err := novada.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Add an IP to the residential whitelist.
	if err := client.Proxy.Whitelist.Add(ctx, proxy.AddWhitelistParams{
		Product: proxy.ProductResidential,
		IP:      "10.10.10.1",
		Remark:  "office",
	}); err != nil {
		log.Fatalf("add whitelist: %v", err)
	}

	// List whitelist entries.
	list, err := client.Proxy.Whitelist.List(ctx, proxy.ListWhitelistParams{
		Product: proxy.ProductResidential,
	})
	if err != nil {
		// Classify the error.
		if novada.IsAuthError(err) {
			log.Fatal("invalid API key")
		}
		if code, ok := novada.CodeOf(err); ok {
			log.Fatalf("business error code=%d: %v", code, err)
		}
		log.Fatal(err)
	}
	fmt.Printf("whitelist total=%d\n", list.Total)
	for _, ip := range list.List {
		fmt.Printf("  %s (%s)\n", ip.MarkIP, ip.Mark)
	}

	// Create a sub-account.
	if err := client.Proxy.Account.Create(ctx, proxy.CreateAccountParams{
		Product:  proxy.ProductResidential,
		Account:  "account11",
		Password: "pass11",
		Status:   1,
	}); err != nil {
		log.Fatalf("create account: %v", err)
	}

	// Residential remaining traffic.
	bal, err := client.Proxy.Residential.Balance(ctx)
	if err != nil {
		log.Fatalf("balance: %v", err)
	}
	fmt.Printf("residential balance=%d bytes\n", bal.Balance)
}
