package wallet

import (
	"context"
	"log"
	"net/http"
	"time"
)

const (
	// AccountTypeSingle indicates an account owned by one individual.
	AccountTypeSingle string = "single"
	// AccountTypeJoint indicates an account owned by two individuals.
	AccountTypeJoint string = "joint"

	// AccountExperienceFundManagement indicates the account is used for wholesale fund investments.
	AccountExperienceFundManagement string = "fundmanagement"
	// AccountExperienceMandate indicates the account is used for private mandates.
	AccountExperienceMandate string = "mandate"
	// AccountExperienceDim indicates the account is used for Diversified Investment Mandates.
	AccountExperienceDim string = "dim"
)

// Client is the main entry point for the Wallet SDK. It holds configuration
// and credentials required to interact with the API.
type Client struct {
	options     *Options
	credentials *credentials
}

// Options contains configuration settings for the Client.
type Options struct {
	// CredentialsLoaderFunc is responsible for retreiving credentials to enable the client
	// sending authenticated requests. This is recommended over [wallet.Client.SetCredentials] which
	// lets credentials live in memory along with Client instance.
	//
	// Optional, if set, credentials will be retrieved for every request, and
	// at best-effort cleared from the memory post call.
	CredentialsLoaderFunc func() (keyID string, privateKeyPEM []byte, err error)

	// HTTPClient specifies an HTTP client used to call the server
	//
	// Optional.
	HTTPClient *http.Client

	// MaxReadRetry specifies how many times to retry a query request when fails.
	//
	// Optional, defaulted to 5 times.
	MaxReadRetry int

	// RetryInterval specifies how long to wait before retrying a query request when fails.
	//
	// Optional, defaulted to 50 milliseconds.
	RetryInterval time.Duration

	// Debug reports whether the client is running in debug mode which enables logging.
	//
	// Optional, defaulted to false.
	Debug bool
}

// New creates a new instance of the Client with the provided options.
func New(opts ...*Options) *Client {
	defaultOptions := Options{
		HTTPClient:    &http.Client{Timeout: 10 * time.Second},
		MaxReadRetry:  5,
		RetryInterval: 50 * time.Millisecond,
	}
	if len(opts) == 0 {
		return &Client{
			options: &defaultOptions,
		}
	}
	o := opts[0]
	// HTTP options
	if o.HTTPClient == nil {
		o.HTTPClient = defaultOptions.HTTPClient
	}
	// force timeout in HTTP client
	if o.HTTPClient.Timeout <= 0 {
		o.HTTPClient.Timeout = 10 * time.Second
	}

	// retry options
	if o.MaxReadRetry <= 0 {
		o.MaxReadRetry = defaultOptions.MaxReadRetry
	}
	if o.RetryInterval <= 0 {
		o.RetryInterval = defaultOptions.RetryInterval
	}

	return &Client{
		options: o,
	}
}

// credentials holds the authentication details for signing requests.
type credentials struct {
	keyID         string
	privateKeyPEM []byte
}

// SetCredentials sets credentials to the client instance. If [wallet.Options.CredentialsLoaderFunc] is set
// upon client's initialization then this is ignored.
func (c *Client) SetCredentials(keyID string, privateKeyPEM []byte) {
	if c.options.CredentialsLoaderFunc != nil {
		if c.options.Debug {
			log.Println("INFO: ignoring SetCredentials call as CredentialsLoaderFunc was set to the client.")
		}
		return
	}
	c.credentials = &credentials{
		keyID:         keyID,
		privateKeyPEM: privateKeyPEM,
	}
}

// ClientAccount represents Halogen investment account. One client may have many accounts.
//
// Halogen offers two Types of accounts, "single" and "joint".
//
//   - Single: Owned by one client, who is the the primary.
//   - Joint: Owned by two clients, the first is the "primary", and the second is the "secondary". Secondary client may or may not have
//     control over the account, however, funds can only be withdrawn to a bank account in the primary name.
//
// Halogen offers three Experiences "fundmanagement", "mandate" and "dim".
//
//   - Fund Management experience is meant for sophesticated investors who are eligible to invest in Public Wholesale Funds.
//   - Mandate experience is meant for clients who have private mandates with Halogen.
//   - DIM experience is meant for retail-investors who have diversifed automated portfolios with Halogen.
type ClientAccount struct {
	// ID specifies the identifier of the account.
	ID string `json:"id,omitempty"`

	// Type specifies the type of the account.
	//
	// Value can be one of "single" or "joint".
	Type string `json:"type,omitempty"`

	// Name specifies the name of the account.
	Name string `json:"name,omitempty"`

	// Experience specifies the investing experience this account has.
	//
	// Value can be one of "fundmanagement", "mandate" or "dim".
	Experience string `json:"experience,omitempty"`

	// ExperienceLabel specifies a friendly name of the experience to
	// be shown on the UI.
	ExperienceLabel string `json:"experienceLabel,omitempty"`

	// Asset specifies the quote asset of the portfolio value and other related
	// fields such PnlAmount, NetInflow.
	Asset string `json:"asset,omitempty"`

	// PortfolioValue specifies the value of this account in Asset terms
	PortfolioValue float64 `json:"portfolioValue"`

	// ExposurePercentage specifies the exposure of this account relatively to the total
	// value of other accounts
	ExposurePercentage float64 `json:"exposurePercentage"`

	// PnlAmount specifies the profit or loss amount in Asset terms.
	//
	// The value will be negative when it is a loss.
	PnlAmount float64 `json:"pnlAmount"`

	// PnlAmount specifies the percentage of profit or loss relative
	// to the invested amount.
	//
	// The value will be negative when it is a loss.
	PnlPercentage float64 `json:"pnlPercentage"`

	// NetInflow specifies the net total traded in this account
	NetInflow float64 `json:"netInflow"`

	// TotalInflow specifies the total amount that has been injected
	// into this account.
	TotalInflow float64 `json:"totalInflow"`

	// TotalOutflow specifies the total amount that has been redeemed
	// from this account.
	TotalOutflow float64 `json:"totalOutflow"`

	// PendingSwitchInAmount specifies the total switching amount that is pending
	// confirmation.
	PendingSwitchInAmount float64 `json:"pendingSwitchInAmount"`

	RiskLabel       string `json:"riskLabel"`
	RiskDescription string `json:"riskDescription"`

	// CanInvest reports whether the requester can create investment request
	//
	// It is only available for "fundmanagement" experience
	CanInvest bool `json:"canInvest"`

	// CanRedeem reports whether the requester can create redemption request
	//
	// It is only available for "fundmanagement" experience
	CanRedeem bool `json:"canRedeem"`

	// CanSwitch reports whether the requester can create switch request
	//
	// It is only available for "fundmanagement" experience
	CanSwitch bool `json:"canSwitch"`

	// CanDeposit reports whether the requester can create deposit request
	//
	// It is only available for "dim" experience
	CanDeposit bool `json:"canDeposit"`

	// CanWithdraw reports whether the requester can create withdrawal request
	//
	// It is only available for "dim" experience
	CanWithdraw bool `json:"canWithdraw"`

	// CanUpdateAccountName reports whether the requester can update the account name
	CanUpdateAccountName bool `json:"canUpdateAccountName"`
}

// ListClientAccountsInput contains parameters for filtering the list of client accounts.
type ListClientAccountsInput struct {
	// AccountIDs filters the list of returned accounts.
	//
	// Optional. If not set, all accounts associated with the client are returned.
	AccountIDs []string `json:"accountIds,omitempty"`
}

// ListClientAccountsOutput contains the list of client accounts and summary data.
type ListClientAccountsOutput struct {
	// Amount is the total value of all returned accounts.
	Amount float64 `json:"amount"`
	// Asset specifies the Amount's asset.
	//
	// In case the display currency is updated then the amount will be
	// converted to the display currency and this will hold the display currency value.
	Asset string `json:"asset,omitempty"`
	// CanCreateAccount reports whether the requester can create a new account under the client.
	CanCreateAccount bool `json:"canCreateAccount"`
	// Accounts is the list of accounts the client has access to. Filter may apply
	// using AccountIDs in the input.
	Accounts []ClientAccount `json:"accounts"`
}

// ListClientAccounts lists all the accounts associated with the client.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_accounts",
//	  "payload": {
//	    "accountIds": ["<accountId>"]
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListClientAccounts(ctx context.Context, input *ListClientAccountsInput) (output *ListClientAccountsOutput, err error) {
	err = c.query(ctx, "list_client_accounts", input, &output)
	return output, err
}

// Address represents a physical mailing or residential address.
type Address struct {
	// Type specifies whether the address is "permanent" or "correspondence".
	Type string `json:"type,omitempty"`
	// Line1 is the first line of the address in text format.
	Line1 string `json:"line1,omitempty"`
	// Line2 is the second line of the address in text format.
	//
	// Optional.
	Line2 *string `json:"line2,omitempty"`
	// City is the city of the address.
	City string `json:"city,omitempty"`
	// Postcode is the postcode of the address.
	Postcode string `json:"postcode,omitempty"`
	// Status is the status of the address.
	//
	// Optional.
	State *string `json:"state,omitempty"`
	// Country is the country of the address.
	Country string `json:"country,omitempty"`
}

// GetClientProfileInput is the input for retrieving the client profile.
type GetClientProfileInput struct {
}

// GetClientProfileOutput contains detailed personal and demographic information about the client.
type GetClientProfileOutput struct {
	// Name is the full name of the client as per official documents.
	//
	// Expect '/' to be part of the name.
	Name string `json:"name,omitempty"`

	// Nationality is the nationality of the client.
	Nationality *string `json:"nationality,omitempty"`

	// NricNo is the Malaysian NRIC number of the client.
	//
	// Only exists for Malaysian clients.
	NricNo *string `json:"nricNo,omitempty"`

	// PassportNp is the Passport number of the client.
	//
	// Only exists for Non-Malaysian clients.
	PassportNo *string `json:"passportNo,omitempty"`

	// Msisdn is the phone number of the client.
	Msisdn *string `json:"msisdn,omitempty"`

	// Email is the email of the client. Value is NULL for corporates.
	//
	// AuthorisedPersonEmail is used for corporates.
	Email *string `json:"email,omitempty"`

	// PermanentAddress is the permanent address of the client.
	PermanentAddress *Address `json:"permanentAddress,omitempty"`

	// CorrespondenceAddress is the correspondence address of the client.
	CorrespondenceAddress *Address `json:"correspondenceAddress,omitempty"`

	// Type specifies the client's type "individual" or "corporate".
	Type string `json:"type,omitempty"`

	// InvestorCategory specifies the investor category that can be one of "accreditedInvestor", "highNetworthInvestor",
	// "sophisticatedInvestor250k", "retailInvestor".
	//
	// 	- Accredited investor:
	// 		- Licensed CMSRL or Registered persons, including CEOs/Directors of CMSLs.
	// 		- Allowed to invest in Public Wholesale Fund, Private Mandate, DIM.
	//	- High net-worth investor:
	// 		- Annual Income of > RM300,000 (Individual)
	// 		- or > RM400,000 (Households)
	// 		- or > RM1,000,000 (investment portfolio)
	// 		- or > RM3,000,000 (net personal assets)
	// 		- Allowed to invest in Public Wholesale Fund, Private Mandate, DIM.
	//	- Sophisticated investor - RM 250k:
	// 		- Any investor who can invest RM250k or more for each transaction.
	// 		- Allowed to invest in Public Wholesale Fund, Private Mandate, DIM.
	//	- Retail investor:
	// 		- None of the above.
	// 		- Allowed to invest in Private Mandate, DIM.
	//
	InvestorCategory string `json:"investorCategory,omitempty"`

	// CountryOfIncoporation specifies the origin country of a corporate.
	//
	// This field is NULL for an individual client.
	CountryOfIncoporation *string `json:"countryOfIncorporation,omitempty"`

	// AuthorisedPersonName specifies the authorised person's name of a corporate.
	//
	// This field is NULL for an individual client.
	AuthorisedPersonName *string `json:"authorisedPersonName,omitempty"`

	// AuthorisedPersonEmail specifies the authorised person's email of a corporate.
	//
	// This field is NULL for an individual client.
	AuthorisedPersonEmail *string `json:"authorisedPersonEmail,omitempty"`

	// AuthorisedPersonMsisdn specifies the authorised person's phone number of a corporate.
	//
	// This field is NULL for an individual client.
	AuthorisedPersonMsisdn *string `json:"authorisedPersonMsisdn,omitempty"`

	// AuthorisedPersonOfficeNo specifies the authorised person's office phone number of a corporate.
	//
	// This field is NULL for an individual client.
	AuthorisedPersonOfficeNo *string `json:"authorisedPersonOfficeNo,omitempty"`

	// CompanyRegistrationNo specifies the registration number of a corporate.
	//
	// This field is NULL for an individual client.
	CompanyRegistrationNo *string `json:"companyRegistrationNo,omitempty"`

	// OldCompanyRegistrationNo specifies the old format of the registration number of a corporate.
	//
	// This field is NULL for an individual client.
	OldCompanyRegistrationNo *string `json:"oldCompanyRegistrationNo,omitempty"`

	// Ethnicity specifies the ethnicity of the client. Value is one of "bumiputera", "chinese",
	// "indian" or "other".
	//
	// Field is filled post registration.
	Ethnicity *string `json:"ethnicity,omitempty"`

	// DomesticRinggitBorrowing specifies the domestic ringgit borrowing status of the client. Value
	// is one of "residentWithoutDrbOrNonResidentOfMalaysia", "residentWithDRBButNotExceeded1M" or "residentWithDRBExceeded1M".
	//
	// Field is filled post registration.
	DomesticRinggitBorrowing *string `json:"domesticRinggitBorrowing,omitempty"`

	// TaxResidency specifies the tax residency of the client. Value is one of "onlyMalaysia", "multiple"
	// or "nonMalaysia".
	//
	// Field is filled post registration.
	TaxResidency *string `json:"taxResidency,omitempty"`

	// CountryTax specifies the country which the client pays tax to. Value is free-text country name.
	//
	// Field is filled post registration.
	CountryTax *string `json:"countryTax,omitempty"`

	// TaxIdentificationNo specifies the tax account number of the client. Value is free-text.
	//
	// Field is filled post registration.
	TaxIdentificationNo *string `json:"taxIdentificationNo,omitempty"`

	// IncompleteProfile reports whether the client's profile incomplete.
	//
	// False when all fields filled.
	IncompleteProfile bool `json:"incompleteProfile"`

	// IsAccountOwner reports whether the requester is the account owner. False when the requester
	// is acting on behalf the account's owner.
	IsAccountOwner bool `json:"isAccountOwner"`

	// CanInvestInUnitTrust reports whether the requester can invest in unit trust funds.
	CanInvestInUnitTrust bool `json:"canInvestInUnitTrust"`

	// CanInvestInDim reports whether the requester can invest in dim.
	CanInvestInDim bool `json:"canInvestInDim"`

	// CanUpdateProfile reports whether the requester can update and complete the profile.
	CanUpdateProfile bool `json:"canUpdateProfile"`

	// CanSubscribePushNotification reports whether the requester can subscribe to push notifications.
	CanSubscribePushNotification bool `json:"canSubscribePushNotification"`

	// Status specifies the status of the client's profile. Value is one of "pending", "rejected",
	// "active" or "withdrawn".
	Status string `json:"status,omitempty"`
}

// GetClientProfile retrieves detailed profile information including personal and demographic data for the authenticated client.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_client_profile",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) GetClientProfile(ctx context.Context, input *GetClientProfileInput) (output *GetClientProfileOutput, err error) {
	err = c.query(ctx, "get_client_profile", input, &output)
	return output, err
}

// Fund represents an investment product available on the platform.
type Fund struct {
	// ID specifies the hexadecimal representation of the Fund identifier. 20 bytes
	// in length (40 hexadecimal characters).
	ID string `json:"id,omitempty"`

	// Type specifies the Fund type. Value is one of "income", "growth".
	Type string `json:"type,omitempty"`

	// Name specifies the legal name of the Fund.
	Name string `json:"name,omitempty"`

	// ShortName specifies the short name of the fund for better user experience.
	ShortName string `json:"shortName,omitempty"`

	// BaseCurrency specifies the fund's denmoinated currency.
	BaseCurrency string `json:"baseCurrency,omitempty"`

	// Category specifies the fund's category.
	Category string `json:"category,omitempty"`

	// Code specifies the fund's code. (e.g HSBTCF, HSETHF, HSCTF, HSRIF)
	Code string `json:"code,omitempty"`

	// InvestmentObjective specifies the objective of the fund as per the
	// information memorandum.
	InvestmentObjective string `json:"investmentObjective,omitempty"`
	// InvestorType specifies the type of the investor who can invest in this fund.
	InvestorType string `json:"investorType,omitempty"`

	// RiskRating specifies the amount of risk this fund has in text format. Value is one of
	// "low", "moderate", "high".
	RiskRating string `json:"riskRating,omitempty"`
	// RiskScore specifies the amount of risk this fund has in numeric format. Value is range
	// from 5 (low) to 16 (high).
	RiskScore int `json:"riskScore,omitempty"`

	// PrimaryFundManager specifies the name of the primary fund manager who
	// is managing the fund.
	PrimaryFundManager string `json:"primaryFundManager,omitempty"`
	// SecondaryFundManager specifies the name of the secondary fund manager who
	// is managing the fund.
	SecondaryFundManager string `json:"secondaryFundManager,omitempty"`

	// ShariahCompliant reports whether the fund is shariah compliant. True when it is shariah compliant.
	ShariahCompliant bool `json:"shariahCompliant,omitempty"`

	// Status specifies the status of the fund. Value is one of "pending", "active" or "archived".
	Status string `json:"status,omitempty"`

	// TagLine specifies the marketing line where it outlines the feature of the fund.
	TagLine string `json:"tagLine,omitempty"`

	// Trustee specifies the name of the trustee of the fund.
	Trustee string `json:"trustee,omitempty"`

	// ImageUrl specifies the Web URL that leads to the logo of the fund. For instance,
	// "https://media.halogen.my/fund/hsbtcf/logo.svg".
	ImageUrl string `json:"imageUrl,omitempty"`

	// CreatedAt specifies the date-time of which the fund was created on.
	CreatedAt string `json:"createdAt,omitempty"`

	// Classes specifies the fund's classes.
	Classes []FundClass `json:"classes,omitempty"`

	// IsOutOfService reports whether the fund is out of service. True when the fund is out of service.
	IsOutOfService bool `json:"isOutOfService"`
	// OutOfServiceMessage specifies the reason of which the fund is out of service.
	OutOfServiceMessage string `json:"outOfServiceMessage,omitempty"`

	// Metadata includes extra attributes related to the requester. For instance,
	// the minimum investment amount the requester must specify upon creating an investment request.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FundClass represents a specific class or tranche within a fund, usually differing by fee structure or currency.
type FundClass struct {
	// Sequence is the unique identifier index for this class within the fund.
	Sequence int `json:"sequence,omitempty"`
	// Label is the display name of the class (e.g., "Class A").
	Label string `json:"label,omitempty"`
	// BaseCurrency is the currency in which this class is denominated.
	BaseCurrency string `json:"baseCurrency,omitempty"`
	// ManagementFee is the annual percentage fee charged by the manager.
	ManagementFee float64 `json:"managementFee,omitempty"`
	// TrusteeFee is the annual percentage fee charged by the trustee.
	TrusteeFee float64 `json:"trusteeFee,omitempty"`
	// CustodianFee is the annual percentage fee charged for asset custody.
	CustodianFee float64 `json:"custodianFee,omitempty"`
	// TransferFee is the fee charged for transferring units.
	TransferFee float64 `json:"transferFee,omitempty"`
	// TrusteeFeeAnnualMinimum is the minimum monetary amount charged by the trustee annually.
	TrusteeFeeAnnualMinimum float64 `json:"trusteeFeeAnnualMinimum,omitempty"`
	// SwitchingFee is the percentage fee charged when switching between funds.
	SwitchingFee float64 `json:"switchingFee,omitempty"`
	// SubscriptionFee is the percentage fee charged upon initial investment.
	SubscriptionFee float64 `json:"subscriptionFee,omitempty"`
	// RedemptionFee is the percentage fee charged upon withdrawal.
	RedemptionFee float64 `json:"redemptionFee,omitempty"`
	// PerformanceFee is the percentage of profits charged if the fund outperforms its benchmark.
	PerformanceFee float64 `json:"performanceFee,omitempty"`
	// TaxRate is the applicable tax rate for this fund class.
	TaxRate float64 `json:"taxRate,omitempty"`
	// MinimumInitialInvestment is the minimum amount required to open an investment in this class.
	MinimumInitialInvestment float64 `json:"minimumInitialInvestment,omitempty"`
	// MinimumAdditionalInvestment is the minimum amount required for subsequent investments.
	MinimumAdditionalInvestment float64 `json:"minimumAdditionalInvestment,omitempty"`
	// MinimumUnitsHeld is the minimum number of units that must be maintained in the account.
	MinimumUnitsHeld float64 `json:"minimumUnitsHeld,omitempty"`
	// MinimumRedemptionAmount is the minimum value allowed for a withdrawal request.
	MinimumRedemptionAmount float64 `json:"minimumRedemptionAmount,omitempty"`
	// CanDistribute reports whether this class pays out distributions (dividends).
	CanDistribute bool `json:"canDistribute,omitempty"`
	// LaunchPrice is the initial NAV per unit when the class was launched.
	LaunchPrice float64 `json:"launchPrice,omitempty"`
	// HexColor is the visual color code associated with this class.
	HexColor string `json:"hexColor,omitempty"`
	// CommencementAt is the date the fund class started operations.
	CommencementAt string `json:"commencementAt,omitempty"`
	// InitialOfferingPeriodFrom is the start date of the initial offering period.
	InitialOfferingPeriodFrom string `json:"initialOfferingPeriodFrom,omitempty"`
	// InitialOfferingPeriodTo is the end date of the initial offering period.
	InitialOfferingPeriodTo string `json:"initialOfferingPeriodTo,omitempty"`
	// CreatedAt is the timestamp when this record was created.
	CreatedAt string `json:"createdAt,omitempty"`
	// DistributionFrequency describes how often distributions are paid (e.g., "Quarterly").
	DistributionFrequency string `json:"distributionFrequency,omitempty"`
	// TagLine is a short marketing descriptor for this class.
	TagLine string `json:"tagLine,omitempty"`
	// Metadata contains arbitrary additional data for this class.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetFundInput is the input for retrieving specific fund details.
type GetFundInput struct {
	// FundID is the unique identifier of the fund to retrieve.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
}

// GetFundOutput contains the full details of the requested fund.
type GetFundOutput struct {
	// Fund contains the fund details.
	Fund *Fund `json:"fund,omitempty"`
}

// GetFund retrieves detailed information about a specific fund including its characteristics, fees, and investment requirements.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_fund",
//	  "payload": {
//	    "fundId": "<fundId>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) GetFund(ctx context.Context, input *GetFundInput) (output *GetFundOutput, err error) {
	err = c.query(ctx, "get_fund", input, &output)
	return output, err
}

// AllocationPerformance represents a data point for fund performance at a specific time.
type AllocationPerformance struct {
	// Date is the timestamp of the performance data point.
	Date string `json:"date,omitempty"`
	// Units is the number of units held at this time.
	Units float64 `json:"units,omitempty"`
	// Asset is the currency or asset symbol.
	Asset string `json:"asset,omitempty"`
	// NetAssetValuePerUnit is the price per unit at this time.
	NetAssetValuePerUnit float64 `json:"netAssetValuePerUnit,omitempty"`
	// Value is the total value of the holding.
	Value float64 `json:"value,omitempty"`
	// PostFeeAmount is the value after accounting for accrued fees.
	PostFeeAmount float64 `json:"postFeeAmount,omitempty"`
}

// GetClientAccountAllocationPerformanceInput is the input for retrieving performance data.
type GetClientAccountAllocationPerformanceInput struct {
	// AccountID is the ID of the client account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// AllocationID is the ID of the specific fund allocation.
	//
	// Required.
	AllocationID string `json:"allocationId,omitempty"`
	// Type specifies the data type to retrieve. (e.g., "fund", "spot")
	//
	// Required.
	Type string `json:"type,omitempty"`
	// FundClassSequence identifies the specific class of the fund.
	//
	// Optional. Required only if Type is "fund".
	FundClassSequence int `json:"fundClassSequence,omitempty"`
	//
	// Required.
	// Timeframe specifies the duration of data to retrieve (e.g., "3M", "YTD").
	Timeframe string `json:"timeframe,omitempty"`
	// Interval specifies the granularity of data points (e.g., "day", "week").
	//
	// Required.
	Interval string `json:"interval,omitempty"`
}

// GetClientAccountAllocationPerformanceOutput contains the list of performance data points.
type GetClientAccountAllocationPerformanceOutput struct {
	// Performance is a slice of historical performance data points.
	Performance []AllocationPerformance `json:"performance"`
}

// GetClientAccountAllocationPerformance retrieves historical performance metrics for a specific fund allocation within an account over a defined timeframe.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_client_account_allocation_performance",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "allocationId": "<allocationId>",
//	    "type": "<type>",
//	    "fundClassSequence": "<fundClassSequence>",
//	    "timeframe": "<timeframe>",
//	    "interval": "<interval>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) GetClientAccountAllocationPerformance(ctx context.Context, input *GetClientAccountAllocationPerformanceInput) (output *GetClientAccountAllocationPerformanceOutput, err error) {
	err = c.query(ctx, "get_client_account_allocation_performance", input, &output)
	return output, err
}

// GetClientAccountStatementInput is the input for generating account statements.
type GetClientAccountStatementInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// FromDate is the start date for the statement range.
	//
	// Required.
	FromDate string `json:"fromDate,omitempty"`
	// ToDate is the end date for the statement range.
	//
	// Required.
	ToDate string `json:"toDate,omitempty"`
	// Format specifies the output format.
	//
	// Optional. Defaults to "pdf" if not specified.
	Format string `json:"format"`
}

// GetClientAccountStatementOutput contains the generated statement file.
type GetClientAccountStatementOutput struct {
	// FromDate is the actual start date used in the statement.
	FromDate string `json:"fromDate,omitempty"`
	// ToDate is the actual end date used in the statement.
	ToDate string `json:"toDate,omitempty"`
	// Format is the format of the generated file.
	Format string `json:"format,omitempty"`
	// Filename is the suggested name for the downloaded file.
	Filename string `json:"filename,omitempty"`
	// Bytes contains the raw binary data of the statement file.
	Bytes []byte `json:"bytes,omitempty"`
}

// GetClientAccountStatement retrieves the account statement as a document (PDF or HTML) for transactions within a specified date range.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_client_account_statement",
//	  "payload": {
//	    "accountId": <accountId>,
//	    "fromDate": "<fromDate>",
//	    "toDate": "<toDate<",
//	    "format": "<format>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInvalidDateRange]
//   - [ErrInternal]
func (c *Client) GetClientAccountStatement(ctx context.Context, input *GetClientAccountStatementInput) (output *GetClientAccountStatementOutput, err error) {
	err = c.query(ctx, "get_client_account_statement", input, &output)
	return output, err
}

// GetClientAccountRequestConfirmationInput is the input for retrieving a transaction confirmation note.
type GetClientAccountRequestConfirmationInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// RequestID is the ID of the transaction request.
	//
	// Optional.
	RequestID string `json:"requestId,omitempty"`
	// Format specifies the output format (e.g., "pdf").
	//
	// Optional. Defaults to "pdf" if not specified.
	Format string `json:"format,omitempty"`
}

// GetClientAccountRequestConfirmationOutput contains the generated confirmation note.
type GetClientAccountRequestConfirmationOutput struct {
	// Format is the format of the generated file.
	Format string `json:"format,omitempty"`
	// Filename is the suggested name for the downloaded file.
	Filename string `json:"filename,omitempty"`
	// Bytes contains the raw binary data of the confirmation note.
	Bytes []byte `json:"bytes,omitempty"`
}

// GetClientAccountRequestConfirmation retrieves the confirmation document for a specific investment, redemption, or switch request.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_client_account_request_confirmation",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "requestId": "<requestId>",
//	    "format": "<format>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) GetClientAccountRequestConfirmation(ctx context.Context, input *GetClientAccountRequestConfirmationInput) (output *GetClientAccountRequestConfirmationOutput, err error) {
	err = c.query(ctx, "get_client_account_request_confirmation", input, &output)
	return output, err
}

// GetClientReferralInput is the input for retrieving referral details.
type GetClientReferralInput struct {
}

// GetClientReferralOutput contains the client's referral code and statistics.
type GetClientReferralOutput struct {
	// ReferralCode is the unique code used to refer new clients.
	ReferralCode string `json:"referralCode,omitempty"`
	// ReferredClientsCount is the number of clients successfully referred.
	ReferredClientsCount int `json:"referredClientsCount"`
}

// GetClientReferral retrieves the client's referral code and the count of successfully referred clients.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_client_referral",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) GetClientReferral(ctx context.Context, input *GetClientReferralInput) (output *GetClientReferralOutput, err error) {
	err = c.query(ctx, "get_client_referral", input, &output)
	return output, err
}

// PolicyGroup represents a group of approvers or rules within a policy.
type PolicyGroup struct {
	// Label is the name of the policy group.
	Label string `json:"label,omitempty"`
	// Min is the minimum number of approvals required from this group.
	Min int `json:"min,omitempty"`
	// Max is the maximum number of approvers in this group.
	Max int `json:"max,omitempty"`
}

// PolicyParticipant represents a user involved in the approval policy.
type PolicyParticipant struct {
	// Email is the email address of the participant.
	Email string `json:"email,omitempty"`
	// GroupLabel is the label of the group this participant belongs to.
	GroupLabel string `json:"groupLabel,omitempty"`
	// Name is the display name of the participant.
	Name string `json:"name,omitempty"`
	// Signed reports whether this participant has approved/signed the request.
	Signed bool `json:"signed,omitempty"`
	// SignedAt is the timestamp when the participant signed the request.
	SignedAt string `json:"signedAt,omitempty"`
}

// GetClientAccountRequestPolicyInput is the input for retrieving policy details for a request.
type GetClientAccountRequestPolicyInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId"`
	// RequestID is the ID of the request to check the policy for.
	//
	// Required.
	RequestID string `json:"requestId"`
}

// GetClientAccountRequestPolicyOutput contains the policy groups and participants for a request.
type GetClientAccountRequestPolicyOutput struct {
	// Groups defines the approval requirements.
	Groups []PolicyGroup `json:"groups"`
	// Participants lists the users involved in the approval process.
	Participants []PolicyParticipant `json:"participants"`
}

// GetClientAccountRequestPolicy retrieves the approval policy and participant information for a specific account request.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_client_account_request_policy",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "requestId": "<requestId>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidRequestPolicy]
//   - [ErrInternal]
func (c *Client) GetClientAccountRequestPolicy(ctx context.Context, input *GetClientAccountRequestPolicyInput) (output *GetClientAccountRequestPolicyOutput, err error) {
	err = c.query(ctx, "get_client_account_request_policy", input, &output)
	return output, err
}

// ListFundsForSubscriptionInput is the input for listing purchasable funds.
type ListFundsForSubscriptionInput struct {
	// AccountID is the ID of the account for which to list eligible funds.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
}

// ListFundsForSubscriptionOutput contains the list of funds available for subscription.
type ListFundsForSubscriptionOutput struct {
	// Funds is the list of eligible funds.
	Funds []Fund `json:"funds"`
}

// ListFundsForSubscription lists all funds available for investment within a specific account based on the account's investor category and experience.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_funds_for_subscription",
//	  "payload": {
//	    "accountId": "<accountId>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListFundsForSubscription(ctx context.Context, input *ListFundsForSubscriptionInput) (output *ListFundsForSubscriptionOutput, err error) {
	err = c.query(ctx, "list_funds_for_subscription", input, &output)
	return output, err
}

// Balance represents the holding of a specific fund within an account.
type Balance struct {
	// FundID is the unique identifier of the fund.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence identifies the specific class of the fund.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
	// FundName is the name of the fund.
	FundName string `json:"fundName,omitempty"`
	// FundShortName is the short name of the fund.
	FundShortName string `json:"fundShortName,omitempty"`
	// FundClassLabel is the label of the fund class.
	FundClassLabel string `json:"fundClassLabel,omitempty"`
	// FundCode is the code of the fund.
	FundCode string `json:"fundCode,omitempty"`
	// FundImageUrl is the URL to the fund's logo.
	FundImageUrl string `json:"fundImageUrl,omitempty"`
	// Units is the quantity of units held.
	Units float64 `json:"units,omitempty"`
	// Asset is the currency of the holding.
	Asset string `json:"asset,omitempty"`
	// Value is the current market value of the holding.
	Value float64 `json:"value,omitempty"`
	// ValuedAt is the date of the last valuation.
	ValuedAt string `json:"valuedAt,omitempty"`
	// MinimumRedemptionAmount is the minimum allowed withdrawal amount for this holding.
	MinimumRedemptionAmount float64 `json:"minimumRedemptionAmount,omitempty"`
	// MinimumRedemptionUnits is the minimum allowed withdrawal units.
	MinimumRedemptionUnits float64 `json:"minimumRedemptionUnits,omitempty"`
	// MinimumSubscriptionAmount is the minimum allowed amount for additional subscriptions.
	MinimumSubscriptionAmount float64 `json:"minimumSubscriptionAmount,omitempty"`
	// MinimumSubscriptionUnits is the minimum allowed units for additional subscriptions.
	MinimumSubscriptionUnits float64 `json:"minimumSubscriptionUnits,omitempty"`
	// RedemptionFeePercentage is the fee charged on redemption.
	RedemptionFeePercentage float64 `json:"redemptionFeePercentage,omitempty"`
	// SwitchFeePercentage is the fee charged on switching.
	SwitchFeePercentage float64 `json:"switchFeePercentage,omitempty"`
	// AvailableModes lists the actions available for this holding (e.g., "redeem", "switch").
	AvailableModes []string `json:"availableModes"`
	// IsOutOfService reports whether the fund is currently unavailable.
	IsOutOfService bool `json:"isOutOfService"`
	// OutOfServiceMessage provides the reason if the fund is out of service.
	OutOfServiceMessage string `json:"outOfServiceMessage,omitempty"`
}

// ListClientAccountBalanceInput is the input for listing account balances.
type ListClientAccountBalanceInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
}

// ListClientAccountBalanceOutput contains the list of holdings in the account.
type ListClientAccountBalanceOutput struct {
	// Balance is a list of fund holdings.
	Balance []*Balance `json:"balance,omitempty"`
}

// ListClientAccountBalance lists the current holdings and balances for each fund allocation in a specific account.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_account_balance",
//	  "payload": {
//	    "accountId": "<accountId>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) ListClientAccountBalance(ctx context.Context, input *ListClientAccountBalanceInput) (output *ListClientAccountBalanceOutput, err error) {
	err = c.query(ctx, "list_client_account_balance", input, &output)
	return output, err
}

// BankAccount represents a registered bank account for the client.
type BankAccount struct {
	// AccountNumber is the bank account number.
	AccountNumber string `json:"accountNumber,omitempty"`
	// AccountName is the name associated with the bank account.
	AccountName string `json:"accountName,omitempty"`
	// AccountCurrency is the currency of the bank account.
	AccountCurrency string `json:"accountCurrency,omitempty"`
	// AccountType is the type of the bank account (e.g., "savings", "current").
	AccountType string `json:"accountType,omitempty"`
	// BankName is the name of the bank.
	BankName string `json:"bankName,omitempty"`
	// BankBic is the BIC/SWIFT code of the bank.
	BankBic string `json:"bankBic,omitempty"`
	// ReferenceNumber is an optional reference for the account registration.
	ReferenceNumber string `json:"referenceNumber,omitempty"`
	// ImageUrl is a URL to the bank's logo.
	ImageUrl string `json:"imageUrl,omitempty"`
	// Status is the verification status of the bank account.
	Status string `json:"status,omitempty"`
	// Source indicates how the account was added (e.g., "user").
	Source string `json:"source,omitempty"`
	// CreatedAt is the timestamp when the account was added.
	CreatedAt string `json:"createdAt,omitempty"`
	// CreatedBy indicates who added the account.
	CreatedBy string `json:"createdBy,omitempty"`
}

// ClientAccountRequest represents a transaction request (Investment, Redemption, Switch).
type ClientAccountRequest struct {
	// ID is the unique identifier of the request.
	ID string `json:"id,omitempty"`
	// fundmanagement: investment, redemption, switch out, switch in
	// dim: deposit, withdrawal
	Type string `json:"type,omitempty"`

	// FundID is the ID of the fund involved.
	FundID string `json:"fundId,omitempty"`
	// FundName is the name of the fund involved.
	FundName string `json:"fundName,omitempty"`
	// FundShortName is the short name of the fund involved.
	FundShortName string `json:"fundShortName,omitempty"`
	// FundClassLabel is the label of the fund class involved.
	FundClassLabel string `json:"fundClassLabel,omitempty"`

	// Asset is the currency of the transaction.
	Asset string `json:"asset,omitempty"`
	// Amount is the gross amount of the transaction.
	Amount float64 `json:"amount,omitempty"`
	// PostFeeAmount is the amount after fees are deducted.
	PostFeeAmount float64 `json:"postFeeAmount,omitempty"`
	// Units is the number of units involved in the transaction.
	Units float64 `json:"units,omitempty"`
	// UnitPrice is the price per unit applied to the transaction.
	UnitPrice *float64 `json:"unitPrice,omitempty"`
	// FeePercentage is the fee percentage applied.
	FeePercentage float64 `json:"feePercentage,omitempty"`
	// StrokedFeePercentage is the original fee percentage before any discounts.
	StrokedFeePercentage float64 `json:"strokedFeePercentage,omitempty"`
	// FeeAmount is the total monetary value of fees charged.
	FeeAmount float64 `json:"feeAmount,omitempty"`
	// RebateFromDate is the start date for rebate calculations (if applicable).
	RebateFromDate string `json:"rebateFromDate,omitempty"`
	// RebateToDate is the end date for rebate calculations (if applicable).
	RebateToDate string `json:"rebateToDate,omitempty"`
	// Status is the current status of the request (e.g., "pending", "completed").
	Status string `json:"status,omitempty"`

	// VoucherCode is the promo code applied to this request.
	VoucherCode *string `json:"voucherCode,omitempty"`
	// ConsentType indicates the type of consent given.
	ConsentType *string `json:"consentType,omitempty"`
	// ConsentStatus indicates the status of the consent.
	ConsentStatus *string `json:"consentStatus,omitempty"`

	// CollectionBankAccount is the bank account used for the transaction (if applicable).
	CollectionBankAccount *BankAccount `json:"collectionBankAccount,omitempty"`

	// CreatedAt is the timestamp when the request was created.
	CreatedAt string `json:"createdAt,omitempty"`
}

// ListClientAccountRequestsInput is the input for listing transaction requests.
type ListClientAccountRequestsInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// RequestID is an optional specific request ID to retrieve.
	//
	// Optional.
	RequestID *string `json:"requestId,omitempty"`
	// Deprecated: Use FundIDs instead.
	//
	// Optional.
	FundID *string `json:"fundId,omitempty"`
	// FundIDs filters requests by specific funds.
	//
	// Optional.
	FundIDs []*string `json:"fundIds,omitempty"`
	// FromDate filters requests starting from this date.
	//
	// Optional.
	FromDate *string `json:"fromDate,omitempty"`
	// ToDate filters requests up to this date.
	//
	// Optional.
	ToDate *string `json:"toDate,omitempty"`
	// Types filters requests by type (e.g., "investment", "redemption").
	//
	// Optional.
	Types []*string `json:"types,omitempty"`
	// Statuses filters requests by status.
	//
	// Optional.
	Statuses []*string `json:"statuses,omitempty"`
	// Limit restricts the number of returned records.
	//
	// Optional.
	Limit *int `json:"limit,omitempty"`
	// Offset determines the starting point for pagination.
	//
	// Optional.
	Offset *int `json:"offset,omitempty"`
	// CompletedOnly filters for only completed requests.
	//
	// Optional. Defaults to false.
	CompletedOnly bool `json:"completedOnly,omitempty"`
}

// ListClientAccountRequestsOutput contains the list of transaction requests.
type ListClientAccountRequestsOutput struct {
	// Requests is the list of client account requests matching the filter.
	Requests []ClientAccountRequest `json:"requests"`
}

// ListClientAccountRequests lists all transaction requests (investments, redemptions, switches) for a specific account with optional filtering and pagination.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_account_requests",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "requestId": "<requestId>",
//	    "fundIds": ["<fundIds>"],
//	    "fromDate": "<fromDate>",
//	    "toDate": "<toDate>",
//	    "types": "<types>",
//	    "statuses": "<statuses>",
//	    "limit": <limit>,
//	    "offset": <offset>,
//	    "completedOnly": <completedOnly>,
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) ListClientAccountRequests(ctx context.Context, input *ListClientAccountRequestsInput) (output *ListClientAccountRequestsOutput, err error) {
	err = c.query(ctx, "list_client_account_requests", input, &output)
	return output, err
}

// ListClientBankAccountsInput is the input for listing bank accounts.
type ListClientBankAccountsInput struct {
}

// ListClientBankAccountsOutput contains the list of registered bank accounts.
type ListClientBankAccountsOutput struct {
	// BankAccounts is the list of bank accounts.
	BankAccounts []BankAccount `json:"bankAccounts"`
}

// ListClientBankAccounts lists all bank accounts registered to the client that can be used for fund transfers and redemptions.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_bank_accounts",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListClientBankAccounts(ctx context.Context, input *ListClientBankAccountsInput) (output *ListClientBankAccountsOutput, err error) {
	err = c.query(ctx, "list_client_bank_accounts", input, &output)
	return output, err
}

// DisplayCurrency represents a currency option for displaying portfolio values.
type DisplayCurrency struct {
	// ID is the unique identifier of the currency.
	ID string `json:"id,omitempty"`
	// Label is the display name of the currency.
	Label string `json:"label,omitempty"`
	// ImageUrl is a URL to the flag or icon of the currency.
	ImageUrl string `json:"imageUrl,omitempty"`
}

// ListDisplayCurrenciesInput is the input for listing display currencies.
type ListDisplayCurrenciesInput struct {
}

// ListDisplayCurrenciesOutput contains the current setting and available options for display currency.
type ListDisplayCurrenciesOutput struct {
	// DisplayCurrency is the currently selected display currency for the client.
	DisplayCurrency string `json:"displayCurrency,omitempty"`
	// Currencies is the list of all available display currencies.
	Currencies []DisplayCurrency `json:"currencies"`
}

// ListDisplayCurrencies lists all available currencies that can be used to display portfolio values and transactions.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_display_currencies",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListDisplayCurrencies(ctx context.Context, input *ListDisplayCurrenciesInput) (output *ListDisplayCurrenciesOutput, err error) {
	err = c.query(ctx, "list_display_currencies", input, &output)
	return output, err
}

// SuitabilityAssessment represents a risk profile assessment record.
type SuitabilityAssessment struct {
	// ID is the unique identifier of the assessment.
	ID string `json:"id,omitempty"`
	// ClientID is the ID of the client who took the assessment.
	ClientID string `json:"clientId,omitempty"`
	// Source indicates where the assessment was taken.
	Source string `json:"source,omitempty"`
	// InvestmentExperience describes the client's prior experience.
	InvestmentExperience string `json:"investmentExperience,omitempty"`
	// InvestmentObjective describes the client's goals.
	InvestmentObjective string `json:"investmentObjective,omitempty"`
	// InvestmentHorizon describes how long the client plans to invest.
	InvestmentHorizon string `json:"investmentHorizon,omitempty"`
	// CurrentInvestment describes the client's current portfolio status.
	CurrentInvestment string `json:"currentInvestment,omitempty"`
	// ReturnExpectations describes what returns the client expects.
	ReturnExpectations string `json:"returnExpectations,omitempty"`
	// Attachment refers to any supporting documents uploaded.
	Attachment string `json:"attachment,omitempty"`
	// TotalScore is the calculated risk score based on answers.
	TotalScore int `json:"totalScore,omitempty"`
	// RiskTolerance is the resulting risk category (e.g., "Aggressive").
	RiskTolerance string `json:"riskTolerance,omitempty"`
	// CreatedBy indicates who created the assessment record.
	CreatedBy string `json:"createdBy,omitempty"`
	// CreatedAt is the timestamp when the assessment was created.
	CreatedAt string `json:"createdAt,omitempty"`
}

// ListClientSuitabilityAssessmentsInput is the input for listing assessments.
type ListClientSuitabilityAssessmentsInput struct {
}

// ListClientSuitabilityAssessmentsOutput contains the list of past assessments and status flags.
type ListClientSuitabilityAssessmentsOutput struct {
	// ShouldAskSuitabilityAssessment reports whether the client needs to take a new assessment.
	ShouldAskSuitabilityAssessment bool `json:"shouldAskSuitabilityAssessment"`
	// CanIgnoreSuitabilityAssessment reports whether the client is allowed to bypass the assessment.
	CanIgnoreSuitabilityAssessment bool `json:"canIgnoreSuitabilityAssessment"`
	// Assessments is the list of historical suitability assessments.
	Assessments []SuitabilityAssessment `json:"assessments"`
}

// ListClientSuitabilityAssessments lists all suitability assessments completed by the client, including risk tolerance evaluations.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_suitability_assessments",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListClientSuitabilityAssessments(ctx context.Context, input *ListClientSuitabilityAssessmentsInput) (output *ListClientSuitabilityAssessmentsOutput, err error) {
	err = c.query(ctx, "list_client_suitability_assessments", input, &output)
	return output, err
}

// Consent represents a specific agreement or acknowledgment required from the user.
type Consent struct {
	// Name is the internal identifier of the consent.
	Name string `json:"name,omitempty"`
	// Label is the display text for the consent checkbox.
	Label string `json:"label,omitempty"`
}

// ListInvestConsentsInput is the input for listing consents required for an investment.
type ListInvestConsentsInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// FundID is the ID of the fund being invested in.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence is the class of the fund.
	//
	// Optional. Defaults to 0 if not specified.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
}

// ListInvestConsentsOutput contains the list of required consents.
type ListInvestConsentsOutput struct {
	// Consents is the list of consent items the user must agree to.
	Consents []Consent `json:"consents"`
	// ConsentFundIM is a legacy flag indicating if Information Memorandum consent is needed.
	ConsentFundIM bool `json:"consentFundIM,omitempty"`
	// ConsentHighRisk is a legacy flag indicating if High Risk consent is needed.
	ConsentHighRisk bool `json:"consentHighRisk,omitempty"`
}

// ListInvestConsents lists the required consent types that must be obtained before making an investment in a specific fund.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_invest_consents",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "fundId": "<fundId>",
//	    "fundClassSequence": <fundClassSequence>
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) ListInvestConsents(ctx context.Context, input *ListInvestConsentsInput) (output *ListInvestConsentsOutput, err error) {
	err = c.query(ctx, "list_invest_consents", input, &output)
	return output, err
}

// Bank represents a supported banking institution.
type Bank struct {
	// Name is the name of the bank.
	Name string `json:"name,omitempty"`
	// Bic is the BIC/SWIFT code of the bank.
	Bic string `json:"bic,omitempty"`
	// ImageUrl is a URL to the bank's logo.
	ImageUrl string `json:"imageUrl,omitempty"`
	// Rank is used for sorting the bank list order.
	Rank int `json:"rank,omitempty"`
}

// ListBanksInput is the input for listing supported banks.
type ListBanksInput struct {
}

// ListBanksOutput contains the list of supported banks.
type ListBanksOutput struct {
	// Banks is the list of supported banks.
	Banks []Bank `json:"banks"`
}

// ListBanks lists all banks supported by the platform for withdrawing funds.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_banks",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListBanks(ctx context.Context, input *ListBanksInput) (output *ListBanksOutput, err error) {
	err = c.query(ctx, "list_banks", input, &output)
	return output, err
}

// Promo represents a promotional offer available to the client.
type Promo struct {
	// AccountID is the ID of the account eligible for the promo.
	AccountID string `json:"accountId,omitempty"`
	// AccountName is the name of the eligible account.
	AccountName string `json:"accountName,omitempty"`
	// Code is the promo code string.
	Code string `json:"code,omitempty"`
	// Label is the display title of the promo.
	Label string `json:"label,omitempty"`
	// Description details the benefits of the promo.
	Description string `json:"description,omitempty"`
	// DiscountPercentage is the percentage discount applied by this promo.
	DiscountPercentage float64 `json:"discountPercentage,omitempty"`
	// DiscountFrom specifies what the discount applies to.
	DiscountFrom string `json:"discountFrom,omitempty"`
	// ValidFromDate is the start date of the promo validity.
	ValidFromDate *string `json:"validFromDate,omitempty"`
	// ValidToDate is the expiration date of the promo.
	ValidToDate *string `json:"validToDate,omitempty"`
	// CreatedAt is the timestamp when the promo was created.
	CreatedAt string `json:"createdAt,omitempty"`
}

// ListClientPromosInput is the input for listing client promos.
type ListClientPromosInput struct {
}

// ListClientPromosOutput contains the list of available promos.
type ListClientPromosOutput struct {
	// Promos is the list of available promotions.
	Promos []Promo `json:"promos"`
}

// ListClientPromos Lists available promotional offers that are applied to client investments.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_promos",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListClientPromos(ctx context.Context, input *ListClientPromosInput) (output *ListClientPromosOutput, err error) {
	err = c.query(ctx, "list_client_promos", input, &output)
	return output, err
}

// ClientAccountPerformance represents a performance metric for an account at a specific time.
type ClientAccountPerformance struct {
	// Date is the timestamp of the performance record.
	Date string `json:"date,omitempty"`
	// AccountID is the ID of the account.
	AccountID string `json:"accountId,omitempty"`
	// Value is the total value of the account at this date.
	Value float64 `json:"value,omitempty"`
}

// ListClientAccountPerformanceInput is the input for listing account performance.
type ListClientAccountPerformanceInput struct {
	// AccountIDs filters for specific accounts.
	//
	// Required.
	AccountIDs []string `json:"accountIds,omitempty"`
	// Timeframe specifies the duration.
	//
	// Required.
	Timeframe string `json:"timeframe,omitempty"`
	// Interval specifies the data granularity.
	//
	// Required.
	Interval string `json:"interval,omitempty"`
}

// ListClientAccountPerformanceOutput contains the performance data.
type ListClientAccountPerformanceOutput struct {
	// Performance is the list of performance data points.
	Performance []ClientAccountPerformance `json:"performance,omitempty"`
}

// ListClientAccountPerformance lists historical performance data for one or more client accounts over a specified timeframe.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_client_account_performance",
//	  "payload": {
//	    "accountIds": ["<accountId>", "<accountId>"],
//	    "timeframe": "<timeframe>",
//	    "interval": "<interval>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) ListClientAccountPerformance(ctx context.Context, input *ListClientAccountPerformanceInput) (output *ListClientAccountPerformanceOutput, err error) {
	err = c.query(ctx, "list_client_account_performance", input, &output)
	return output, err
}

// ListPaymentMethodsInput is the input for listing payment methods.
type ListPaymentMethodsInput struct {
}

// ListPaymentMethodsOutput contains flags for available payment methods.
type ListPaymentMethodsOutput struct {
	// Duitnow reports whether DuitNow payment is available.
	Duitnow bool `json:"duitnow"`
	// BankTransfer reports whether manual bank transfer is available.
	BankTransfer bool `json:"bankTransfer"`
}

// ListPaymentMethods lists the available payment methods for fund transfers, such as DuitNow and bank transfers.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "list_payment_methods",
//	  "payload": {}
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInternal]
func (c *Client) ListPaymentMethods(ctx context.Context, input *ListPaymentMethodsInput) (output *ListPaymentMethodsOutput, err error) {
	err = c.query(ctx, "list_payment_methods", input, &output)
	return output, err
}

// GetVoucherInput is the input for checking a voucher code.
type GetVoucherInput struct {
	// AccountID is the ID of the account used for the investment.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// FundID is the ID of the fund.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence is the class of the fund.
	//
	// Required. Must be greater than 0.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
	// Amount is the investment amount.
	//
	// Required. Must be greater than 0.
	Amount float64 `json:"amount,omitempty"`
	// VoucherCode is the code to validate.
	//
	// Optional.
	VoucherCode *string `json:"voucherCode,omitempty"`
}

// GetVoucherOutput contains the validation result and fee impact of a voucher.
type GetVoucherOutput struct {
	// Valid reports whether the voucher code is valid for this transaction.
	Valid bool `json:"valid"`
	// Code is the validated voucher code.
	Code string `json:"code"`
	// StrokedSubscriptionFeePercentage is the original fee before discount.
	StrokedSubscriptionFeePercentage float64 `json:"strokedSubscriptionFeePercentage"`
	// AppliedSubscriptionFeePercentage is the fee after discount.
	AppliedSubscriptionFeePercentage float64 `json:"appliedSubscriptionFeePercentage"`
	// VoucherDiscountPercentage is the percentage shaved off the fee.
	VoucherDiscountPercentage float64 `json:"voucherDiscountPercentage"`
	// FeeAmount is the monetary value of the final fee.
	FeeAmount float64 `json:"feeAmount"`
	// PostFeeAmount is the investment amount minus the fee.
	PostFeeAmount float64 `json:"postFeeAmount"`
}

// GetVoucher retrieves details and validates a specific voucher code, calculating the discounted fees for an investment.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_voucher",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "fundId": "<fundId>",
//	    "fundClassSequence": <fundClassSequence>,
//	    "amount": <amount>,
//	    "voucherCode": "<voucherCode>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) GetVoucher(ctx context.Context, input *GetVoucherInput) (output *GetVoucherOutput, err error) {
	err = c.query(ctx, "get_voucher", input, &output)
	return output, err
}

// GetPreviewInvestInput is the input for calculating investment fees.
type GetPreviewInvestInput struct {
	// AccountID is the ID of the account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// FundID is the ID of the fund.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence is the class of the fund.
	//
	// Required. Must be greater than 0.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
	// Amount is the intended investment amount.
	//
	// Required. Must be greater than 0.
	Amount float64 `json:"amount,omitempty"`
}

// GetPreviewInvestOutput contains the fee calculations for a potential investment.
type GetPreviewInvestOutput struct {
	// StrokedSubscriptionFeePercentage is the standard fee percentage.
	StrokedSubscriptionFeePercentage float64 `json:"strokedSubscriptionFeePercentage"`
	// AppliedSubscriptionFeePercentage is the actual fee percentage applied (after defaults).
	AppliedSubscriptionFeePercentage float64 `json:"appliedSubscriptionFeePercentage"`
	// PostFeeAmount is the amount that will actually buy units.
	PostFeeAmount float64 `json:"postFeeAmount"`
	// FeeAmount is the monetary value of the fees.
	FeeAmount float64 `json:"feeAmount"`
	// DefaultVoucher contains details if a voucher was automatically applied.
	DefaultVoucher *GetVoucherOutput `json:"defaultVoucher,omitempty"`
}

// GetPreviewInvest calculates a preview of an investment transaction, including applicable fees and any default voucher discounts.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_preview_invest",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "fundId": "<fundId>",
//	    "fundClassSequence": <fundClassSequence>,
//	    "amount": <amount>
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) GetPreviewInvest(ctx context.Context, input *GetPreviewInvestInput) (output *GetPreviewInvestOutput, err error) {
	err = c.query(ctx, "get_preview_invest", input, &output)
	return output, err
}

// GetProjectedFundPriceInput is the input for retrieving projected NAV.
type GetProjectedFundPriceInput struct {
	// FundID is the ID of the fund.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence is the class of the fund.
	//
	// Required. Must be greater than 0.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
}

// GetProjectedFundPriceOutput contains the projected NAV data.
type GetProjectedFundPriceOutput struct {
	// Asset is the currency of the fund.
	Asset string `json:"asset"`
	// NetAssetValuePerUnit is the projected price per unit.
	NetAssetValuePerUnit float64 `json:"netAssetValuePerUnit"`
}

// GetProjectedFundPrice retrieves the projected unit net asset value per unit (NAV per unit) for a specific fund class.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/query" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "get_projected_fund_price",
//	  "payload": {
//	    "fundId": "<fundId>",
//	    "fundClassSequence": <fundClassSequence>
//	  }
//	}'
func (c *Client) GetProjectedFundPrice(ctx context.Context, input *GetProjectedFundPriceInput) (output *GetProjectedFundPriceOutput, err error) {
	err = c.query(ctx, "get_projected_fund_price", input, &output)
	return output, err
}

//
// Commands
//

// CreateInvestmentRequestInput represents the payload for creating a new investment request.
type CreateInvestmentRequestInput struct {
	// AccountID specifies the identifier of the client account for the investment.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// FundID specifies the identifier of the fund to invest in.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence specifies the class of the fund to invest in.
	//
	// Required. Must be greater than 0.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
	// Amount specifies the amount to be invested.
	//
	// Required. Must be greater than 0.
	Amount float64 `json:"amount,omitempty"`

	// ConsentFundIM is deprecated, use Consents instead.
	//
	// Required. Value must be true.
	ConsentFundIM bool `json:"consentFundIM,omitempty"`
	// ConsentHighRisk is deprecated, use Consents instead.
	//
	// Optional.
	ConsentHighRisk bool `json:"consentHighRisk,omitempty"`

	// Consents specifies a map of consent names to boolean values (true if consented).
	//
	// Required.
	Consents map[string]bool `json:"consents,omitempty"`

	// VoucherCode specifies an optional voucher code to apply to the investment.
	//
	// Optional.
	VoucherCode string `json:"voucherCode,omitempty"`
}

// CreateInvestmentRequestOutput represents the response for an investment request.
type CreateInvestmentRequestOutput struct {
	// RequestID specifies the identifier of the created investment request.
	RequestID string `json:"requestId,omitempty"`
}

// CreateInvestmentRequest submits a new investment request to purchase units in a specified fund class with the provided amount.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "create_investment_request",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "fundId": "<fundId>",
//	    "fundClassSequence": <fundClassSequence>
//	    "amount": <amount>,
//	    "consentFundIM": <consentFundIM>,
//	    "consentHighRisk": <consentHighRisk>,
//	    "consents": {
//	      "IM": true
//	    },
//	    "voucherCode": <voucherCode>
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInvalidParameter]
//   - [ErrActionOutsideFundHours]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) CreateInvestmentRequest(ctx context.Context, input *CreateInvestmentRequestInput) (output *CreateInvestmentRequestOutput, err error) {
	err = c.command(ctx, "create_investment_request", input, &output)
	return output, err
}

// CreateRedeemRequestInput represents the payload for creating a new redemption (withdrawal) request.
type CreateRedemptionRequestInput struct {
	// AccountID specifies the identifier of the client account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// FundID specifies the identifier of the fund to redeem from.
	//
	// Required.
	FundID string `json:"fundId,omitempty"`
	// FundClassSequence specifies the class of the fund to redeem from.
	//
	// Required. Must be greater than 0.
	FundClassSequence int `json:"fundClassSequence,omitempty"`
	// RequestedAmount specifies the amount to redeem.
	//
	// Optional. Must be greater than 0.
	// Mutually exclusive with Units (exactly one must be provided).
	RequestedAmount float64 `json:"requestedAmount,omitempty"`
	// Units specifies the number of units to redeem.
	//
	// Optional. Must be greater than 0.
	// Mutually exclusive with RequestedAmount (exactly one must be provided).
	Units float64 `json:"units,omitempty"`
	// ToBankAccountNumber specifies the bank account number for the redemption proceeds.
	//
	// Required.
	ToBankAccountNumber string `json:"toBankAccountNumber,omitempty"`
}

// CreateRedeemRequestOutput represents the response for a redemption request.
type CreateRedemptionRequestOutput struct {
	// RequestID specifies the identifier of the created redemption request.
	RequestID string `json:"requestId,omitempty"`
}

// CreateRedemptionRequest submits a new redemption request to sell fund units or withdraw an amount from an account.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "create_redemption_request",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "fundId": "<fundId>",
//	    "fundClassSequence": <fundClassSequence>
//	    "requestedAmount": <amount>,
//	    "units": <units>,
//	    "toBankAccountNumber": "<toBankAccountNumber>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInvalidParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInsufficientBalance]
//   - [ErrActionOutsideFundHours]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) CreateRedemptionRequest(ctx context.Context, input *CreateRedemptionRequestInput) (output *CreateRedemptionRequestOutput, err error) {
	err = c.command(ctx, "create_redemption_request", input, &output)
	return output, err
}

// CreateSwitchRequestInput represents the payload for creating a new fund switch request.
type CreateSwitchRequestInput struct {
	// AccountID specifies the identifier of the client account.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`

	// SwitchFromFundID specifies the fund ID to switch units *from*.
	//
	// Required.
	SwitchFromFundID string `json:"switchFromFundId,omitempty"`
	// SwitchFromFundClassSequence specifies the fund class sequence to switch units *from*.
	//
	// Required. Must be greater than 0.
	SwitchFromFundClassSequence int `json:"switchFromFundClassSequence,omitempty"`
	// SwitchToFundID specifies the fund ID to switch units *to*.
	//
	// Required.
	SwitchToFundID string `json:"switchToFundId,omitempty"`
	// SwitchToFundClassSequence specifies the fund class sequence to switch units *to*.
	//
	// Required. Must be greater than 0.
	SwitchToFundClassSequence int `json:"switchToFundClassSequence,omitempty"`

	// RequestedAmount specifies the amount to switch.
	//
	// Optional. Must be greater than 0.
	// Mutually exclusive with Units (exactly one must be provided).
	RequestedAmount float64 `json:"requestedAmount,omitempty"`
	// Units specifies the number of units to switch.
	//
	// Optional. Must be greater than 0.
	// Mutually exclusive with RequestedAmount (exactly one must be provided).
	Units float64 `json:"units,omitempty"`
}

// CreateSwitchRequestOutput represents the response for a switch request.
type CreateSwitchRequestOutput struct {
	// RequestID specifies the identifier of the created switch request.
	RequestID string `json:"requestId,omitempty"`
}

// CreateSwitchRequest submits a request to transfer units from one fund to another within the same account.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "create_switch_request",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "switchFromFundId": "<switchFromFundId>",
//	    "switchFromFundClassSequence": <fundClassSequence>,
//	    "switchToFundId": "<switchToFundId>",
//	    "switchToFundClassSequence": <switchToFundClassSequence>,
//	    "requestedAmount": <amount>,
//	    "units": <units>
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInvalidParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInsufficientBalance]
//   - [ErrActionOutsideFundHours]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) CreateSwitchRequest(ctx context.Context, input *CreateSwitchRequestInput) (output *CreateSwitchRequestOutput, err error) {
	err = c.command(ctx, "create_switch_request", input, &output)
	return output, err
}

// CreateRequestCancellationInput represents the payload for canceling an existing request.
type CreateRequestCancellationInput struct {
	// AccountID specifies the identifier of the client account associated with the request.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// RequestID specifies the identifier of the request to cancel.
	//
	// Required.
	RequestID string `json:"requestId,omitempty"`
}

// CreateRequestCancellationOutput represents the response for a cancel request command (empty upon success).
type CreateRequestCancellationOutput struct {
}

// CreateRequestCancellation cancels a pending transaction request (investment, redemption, or switch) before it is executed.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "create_request_cancellation",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "requestId": "<requestId>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrRequestCannotBeCancelled]
//   - [ErrInsufficientAccess]
//   - [ErrMissingResource]
//   - [ErrInternal]
func (c *Client) CreateRequestCancellation(ctx context.Context, input *CreateRequestCancellationInput) (output *CreateRequestCancellationOutput, err error) {
	err = c.command(ctx, "create_request_cancellation", input, &output)
	return output, err
}

// CreateSuitabilityAssessmentInput represents the payload for submitting a new suitability assessment.
type CreateSuitabilityAssessmentInput struct {
	// SuitabilityAssessment contains the details of the assessment being submitted.
	//
	// Required.
	SuitabilityAssessment *SuitabilityAssessment `json:"suitabilityAssessment,omitempty"`
}

// CreateSuitabilityAssessmentOutput represents the response for creating a suitability assessment.
type CreateSuitabilityAssessmentOutput struct {
	// SuitabilityAssessmentID specifies the identifier of the created assessment.
	SuitabilityAssessmentID string `json:"suitabilityAssessmentId,omitempty"`
}

// CreateSuitabilityAssessment submits a new risk suitability assessment for the client, evaluating investment risk tolerance.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "create_suitability_assessment",
//	  "payload": {
//	    "suitabilityAssessment": {
//	      "id": "<id>",
//	      "clientId": "<clientId>",
//	      "source": "<source>",
//	      "investmentExperience": "<investmentExperience>",
//	      "investmentObjective": "<investmentObjective>",
//	      "investmentHorizon": "<investmentHorizon>",
//	      "currentInvestment": "<currentInvestment>",
//	      "returnExpectations": "<returnExpectations>",
//	      "attachment": "<attachment>",
//	      "totalScore": <totalScore>,
//	      "riskTolerance": "<riskTolerance>"
//	    }
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) CreateSuitabilityAssessment(ctx context.Context, input *CreateSuitabilityAssessmentInput) (output *CreateSuitabilityAssessmentOutput, err error) {
	err = c.command(ctx, "create_suitability_assessment", input, &output)
	return output, err
}

// CreateClientBankAccountInput represents the payload for adding a new bank account.
type CreateClientBankAccountInput struct {
	// BankAccount contains the details of the bank account to be created.
	//
	// Required.
	BankAccount *BankAccount `json:"bankAccount,omitempty"`
}

// CreateClientBankAccountOutput represents the response for adding a bank account (empty upon success).
type CreateClientBankAccountOutput struct {
}

// CreateClientBankAccount registers a new bank account with the client's profile for receiving redemption proceeds.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "create_client_bank_account",
//	  "payload": {
//	    "bankAccount": {
//	      "accountNumber": "<accountNumber>",
//	      "accountName": "<accountName>",
//	      "accountCurrency": "<accountCurrency>",
//	      "accountType": "<accountType>",
//	      "accountType": "<accountType>",
//	      "bankName": "<bankName>",
//	      "bankBic": "<bankBic>",
//	      "referenceNumber": "<referenceNumber>",
//	      "imageUrl": "<imageUrl>",
//	      "status": "<status>",
//	      "source": "<source>"
//	    }
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrAlreadyExists]
//   - [ErrInternal]
func (c *Client) CreateClientBankAccount(ctx context.Context, input *CreateClientBankAccountInput) (output *CreateClientBankAccountOutput, err error) {
	err = c.command(ctx, "create_client_bank_account", input, &output)
	return output, err
}

// UpdateDisplayCurrencyInput represents the payload for changing the client's display currency.
type UpdateDisplayCurrencyInput struct {
	// DisplayCurrency specifies the new currency ID to be used for display.
	//
	// Required.
	DisplayCurrency string `json:"displayCurrency,omitempty"`
}

// UpdateDisplayCurrencyOutput represents the response for updating the display currency (empty upon success).
type UpdateDisplayCurrencyOutput struct {
}

// UpdateDisplayCurrency changes the currency in which the client's portfolio values and transactions are displayed.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "update_display_currency",
//	  "payload": {
//	    "displayCurrency": "<displayCurrency>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) UpdateDisplayCurrency(ctx context.Context, input *UpdateDisplayCurrencyInput) (output *UpdateDisplayCurrencyOutput, err error) {
	err = c.command(ctx, "update_display_currency", input, &output)
	return output, err
}

// UpdateAccountNameInput represents the payload for changing a client account's name.
type UpdateAccountNameInput struct {
	// AccountID specifies the ID of the account to update.
	//
	// Required.
	AccountID string `json:"accountId,omitempty"`
	// AccountName specifies the new name for the account.
	//
	// Required. Must be at least 3 characters.
	AccountName string `json:"accountName,omitempty"`
}

// UpdateAccountNameOutput represents the response for updating an account name (empty upon success).
type UpdateAccountNameOutput struct {
}

// UpdateAccountName changes the friendly name or label of a specific client investment account.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "update_account_name",
//	  "payload": {
//	    "accountId": "<accountId>",
//	    "accountName": "<accountName>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) UpdateAccountName(ctx context.Context, input *UpdateAccountNameInput) (output *UpdateAccountNameOutput, err error) {
	err = c.command(ctx, "update_account_name", input, &output)
	return output, err
}

// UpdateClientProfileInput represents the payload for updating specific fields on the client's profile.
type UpdateClientProfileInput struct {
	// Ethnicity specifies the client's ethnicity. Value is one of "bumiputera", "chinese", "indian" or "other".
	//
	// Optional. If set to "other", OtherEthnicity becomes required.
	Ethnicity string `json:"ethnicity,omitempty"`
	// OtherEthnicity is used if Ethnicity is "other" to specify the exact ethnicity.
	//
	// Optional. Required if Ethnicity is "other".
	OtherEthnicity string `json:"otherEthnicity,omitempty"`

	// DomesticRinggitBorrowing specifies the client's domestic ringgit borrowing status.
	//
	// Optional.
	DomesticRinggitBorrowing string `json:"domesticRinggitBorrowing,omitempty"`
	// TaxResidency specifies the client's tax residency status. Value is one of "onlyMalaysia", "multiple" or "nonMalaysia".
	//
	// Optional. Required if TaxResidency is "nonMalaysia" or "multiple".
	TaxResidency string `json:"taxResidency,omitempty"`
	// CountryTax specifies the country where the client pays tax.
	//
	// Optional. Required if TaxResidency is "nonMalaysia" or "multiple".
	CountryTax string `json:"countryTax,omitempty"`
	// TaxIdentificationNo specifies the client's tax account number.
	//
	// Optional. Required if TaxResidency is "nonMalaysia" or "multiple".
	TaxIdentificationNo string `json:"taxIdentificationNo,omitempty"`
}

// UpdateClientProfileOutput represents the response for updating the client profile (empty upon success).
type UpdateClientProfileOutput struct {
}

// UpdateClientProfile updates the client's demographic information, ethnicity, and tax residency details.
//
// cURL:
//
//	curl -X "POST" "https://external-api.wallet.halogen.my/command" \
//	  -H 'Authorization: Bearer <JWT>' \
//	  -H 'Content-Type: application/json; charset=utf-8' \
//	  -d $'{
//	  "name": "update_client_profile",
//	  "payload": {
//	    "ethnicity": "<ethnicity>",
//	    "otherEthnicity": "<otherEthnicity>",
//	    "domesticRinggitBorrowing": "<domesticRinggitBorrowing>",
//	    "taxResidency": "<taxResidency>",
//	    "countryTax": "<countryTax>",
//	    "taxIdentificationNo": "<taxIdentificationNo>"
//	  }
//	}'
//
// Errors:
//   - [ErrMissingParameter]
//   - [ErrInsufficientAccess]
//   - [ErrInvalidParameter]
//   - [ErrInternal]
func (c *Client) UpdateClientProfile(ctx context.Context, input *UpdateClientProfileInput) (output *UpdateClientProfileOutput, err error) {
	err = c.command(ctx, "update_client_profile", input, &output)
	return output, err
}
