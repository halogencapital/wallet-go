// # Authentication
//
// The Halogen Wallet API uses stateless JWT bearer token authentication. The client automatically generates and signs JWT tokens
// for each request using your credentials, which are added to the Authorization header in the format: "Bearer <JWT>".
//
// # Token Generation
//
// The JWT token is constructed with the following Payload fields
//
//   - `kid`: Key identifier (the Key ID returned by Halogen Wallet settings)
//   - `sub`: Subject — currently set to the fixed value `"wallet"`
//   - `iat`: Issued At (Unix timestamp)
//   - `exp`: Expiration (Unix timestamp) — set to `iat + ttl` where `ttl` is the token lifetime the client uses
//   - `nonce`: A random hex string used to prevent replay attacks
//   - `bodyHash`: Hex-encoded SHA-256 hash of the request body
//   - `uri`: The request URI (for example `/query` or `/command`)
//
// The JWT header contains `alg` (set to `ES256` or `RS256` depending on the private key)
// and `typ: "JWT"`.
//
// The token is then signed using either:
//
//   - ES256 (ECDSA with P-256 curve) when an EC private key is provided
//   - RS256 (RSA) when an RSA private key is provided
//
// You do not need to manually generate or sign tokens. The client handles this automatically
// when you provide credentials via [Client.SetCredentials] or [Client.Options.CredentialsLoaderFunc].
//
// # Rate Limiting
//
// The Halogen Wallet API implements rate limiting to ensure fair usage and system stability.
// The server allows a maximum of 10 requests per second with a burst capacity of 10 requests.
// This means you can make up to 10 requests immediately (burst), but sustained traffic is limited to 10 requests per second.
//
// If you exceed the rate limit, the server will respond with an HTTP 429 (Too Many Requests) error.
// The client automatically retries requests when receiving a 429 response. This ensures that rate limit
// errors are handled transparently without manual intervention.
//
// Note: The retry configuration in [Client.Options] (MaxReadRetry and RetryInterval) only applies to
// read operations when the server responds with HTTP status codes >= 500. Rate limit retries (429 errors)
// are handled separately and automatically.
//
// # Example
//
// Here's a complete example showing how to list accounts, available funds, get the projected price, and create an investment:
//
//	package main
//
//	import (
//		"context"
//		"log"
//		"os"
//
//		"github.com/halogencapital/wallet-go"
//	)
//
//	func main() {
//		// Initialize the client
//		client := wallet.New()
//		client.SetCredentials(
//			os.Getenv("HALOGEN_WALLET_KEY_ID"),
//			[]byte(os.Getenv("HALOGEN_WALLET_PRIVATE_KEY_PEM")),
//		)
//		ctx := context.Background()
//
//		// List all client accounts
//		accounts, err := client.ListClientAccounts(ctx, &wallet.ListClientAccountsInput{})
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("Found %d accounts\n", len(accounts.Accounts))
//
//		// Use the first account
//		accountID := accounts.Accounts[0].ID
//
//		// List available funds for subscription in this account
//		funds, err := client.ListFundsForSubscription(ctx, &wallet.ListFundsForSubscriptionInput{
//			AccountID: accountID,
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("Found %d available funds\n", len(funds.Funds))
//
//		// Use the first fund
//		fundID := funds.Funds[0].ID
//		fundClassSequence := funds.Funds[0].Classes[0].Sequence
//
//		// Get the projected price for the fund
//		price, err := client.GetProjectedFundPrice(ctx, &wallet.GetProjectedFundPriceInput{
//			FundID:            fundID,
//			FundClassSequence: fundClassSequence,
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("Fund NAV: %f %s\n", price.NetAssetValuePerUnit, price.Asset)
//
//		// Create an investment request
//		investmentAmount := 10000.0
//		investReq, err := client.CreateInvestmentRequest(ctx, &wallet.CreateInvestmentRequestInput{
//			AccountID:         accountID,
//			FundID:            fundID,
//			FundClassSequence: fundClassSequence,
//			Amount:            investmentAmount,
//			Consents: map[string]bool{
//				"IM": true,
//			},
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("Investment request created with ID: %s\n", investReq.RequestID)
//	}
//
// # Query APIs
//
// - [Client.ListClientAccounts]
//
// - [Client.GetClientProfile]
//
// - [Client.GetFund]
//
// - [Client.GetClientAccountAllocationPerformance]
//
// - [Client.GetClientAccountStatement]
//
// - [Client.GetClientAccountRequestConfirmation]
//
// - [Client.GetClientReferral]
//
// - [Client.GetClientAccountRequestPolicy]
//
// - [Client.ListFundsForSubscription]
//
// - [Client.ListClientAccountBalance]
//
// - [Client.ListClientAccountRequests]
//
// - [Client.ListClientBankAccounts]
//
// - [Client.ListDisplayCurrencies]
//
// - [Client.ListClientSuitabilityAssessments]
//
// - [Client.ListInvestConsents]
//
// - [Client.ListBanks]
//
// - [Client.ListClientPromos]
//
// - [Client.ListClientAccountPerformance]
//
// - [Client.ListPaymentMethods]
//
// - [Client.GetVoucher]
//
// - [Client.GetPreviewInvest]
//
// - [Client.GetProjectedFundPrice]
//
// # Command APIs
//
// - [Client.CreateInvestmentRequest]
//
// - [Client.CreateRedemptionRequest]
//
// - [Client.CreateSwitchRequest]
//
// - [Client.CreateRequestCancellation]
//
// - [Client.CreateSuitabilityAssessment]
//
// - [Client.CreateClientBankAccount]
//
// - [Client.UpdateDisplayCurrency]
//
// - [Client.UpdateAccountName]
//
// - [Client.UpdateClientProfile]
package wallet
