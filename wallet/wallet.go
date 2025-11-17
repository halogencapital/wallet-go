package wallet

import (
	"context"
	"log"
	"net/http"
	"time"
)

const (
	AccountTypeSingle string = "single"
	AccountTypeJoint  string = "joint"

	AccountExperienceFundManagement string = "fundmanagement"
	AccountExperienceMandate        string = "mandate"
	AccountExperienceDim            string = "dim"
)

type Client struct {
	options     *Options
	credentials *credentials
}

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

	// MaxReadRetry reports how many times to retry a query request when fails.
	//
	// Optional, defaulted to 5 times.
	MaxReadRetry int

	// RetryInterval reports how long to wait before retrying a query request when fails.
	//
	// Optional, defaulted to 50 milliseconds.
	RetryInterval time.Duration

	// Debug reports whether the client is running in debug mode which enables logging.
	//
	// Optional, defaulted to false.
	Debug bool
}

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

// ClientAccount is ...
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

type ListClientAccountsInput struct {
	AccountIDs []string `json:"accountIds,omitempty"`
}

type ListClientAccountsOutput struct {
	Amount           float64         `json:"amount"`
	Asset            string          `json:"asset,omitempty"`
	CanCreateAccount bool            `json:"canCreateAccount"`
	Accounts         []ClientAccount `json:"accounts"`
}

// ListClientAccounts lists all the accounts associated with the provided client ID
func (c *Client) ListClientAccounts(ctx context.Context, input *ListClientAccountsInput) (*ListClientAccountsOutput, error) {
	output := ListClientAccountsOutput{}
	err := c.query(ctx, "list_client_accounts", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type Address struct {
	// permanent, correspondence
	Type     string  `json:"type,omitempty"`
	Line1    string  `json:"line1,omitempty"`
	Line2    *string `json:"line2,omitempty"`
	City     string  `json:"city,omitempty"`
	Postcode string  `json:"postcode,omitempty"`
	State    *string `json:"state,omitempty"`
	Country  string  `json:"country,omitempty"`
}

type GetClientProfileInput struct {
	ClientID string `json:"clientId,omitempty"`
}

type GetClientProfileOutput struct {
	Name                  string   `json:"name,omitempty"`
	Nationality           *string  `json:"nationality,omitempty"`
	NricNo                *string  `json:"nricNo,omitempty"`
	PassportNo            *string  `json:"passportNo,omitempty"`
	Msisdn                *string  `json:"msisdn,omitempty"`
	Email                 *string  `json:"email,omitempty"`
	PermanentAddress      *Address `json:"permanentAddress,omitempty"`
	CorrespondenceAddress *Address `json:"correspondenceAddress,omitempty"`

	// individual, corporate
	Type                     string  `json:"type,omitempty"`
	InvestorCategory         string  `json:"investorCategory,omitempty"`
	CountryOfIncoporation    *string `json:"countryOfIncorporation,omitempty"`
	AuthorisedPersonName     *string `json:"authorisedPersonName,omitempty"`
	AuthorisedPersonEmail    *string `json:"authorisedPersonEmail,omitempty"`
	AuthorisedPersonMsisdn   *string `json:"authorisedPersonMsisdn,omitempty"`
	AuthorisedPersonOfficeNo *string `json:"authorisedPersonOfficeNo,omitempty"`
	CompanyRegistrationNo    *string `json:"companyRegistrationNo,omitempty"`
	OldCompanyRegistrationNo *string `json:"oldCompanyRegistrationNo,omitempty"`

	Ethnicity                *string `json:"ethnicity,omitempty"`
	DomesticRinggitBorrowing *string `json:"domesticRinggitBorrowing,omitempty"`
	TaxResidency             *string `json:"taxResidency,omitempty"`
	CountryTax               *string `json:"countryTax,omitempty"`
	TaxIdentificationNo      *string `json:"taxIdentificationNo,omitempty"`
	IncompleteProfile        bool    `json:"incompleteProfile"`

	IsAccountOwner               bool   `json:"isAccountOwner"`
	CanInvestInUnitTrust         bool   `json:"canInvestInUnitTrust"`
	CanInvestInDim               bool   `json:"canInvestInDim"`
	CanUpdateProfile             bool   `json:"canUpdateProfile"`
	CanSubscribePushNotification bool   `json:"canSubscribePushNotification"`
	Status                       string `json:"status,omitempty"`
}

func (c *Client) GetClientProfile(ctx context.Context, input *GetClientProfileInput) (*GetClientProfileOutput, error) {
	output := GetClientProfileOutput{}
	err := c.query(ctx, "get_client_profile", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type Fund struct {
	ID                   string                 `json:"id,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	ShortName            string                 `json:"shortName,omitempty"`
	BaseCurrency         string                 `json:"baseCurrency,omitempty"`
	Category             string                 `json:"category,omitempty"`
	Code                 string                 `json:"code,omitempty"`
	InvestmentObjective  string                 `json:"investmentObjective,omitempty"`
	InvestorType         string                 `json:"investorType,omitempty"`
	RiskRating           string                 `json:"riskRating,omitempty"`
	RiskScore            int                    `json:"riskScore,omitempty"`
	PrimaryFundManager   string                 `json:"primaryFundManager,omitempty"`
	SecondaryFundManager string                 `json:"secondaryFundManager,omitempty"`
	ShariahCompliant     bool                   `json:"shariahCompliant,omitempty"`
	Status               string                 `json:"status,omitempty"`
	TagLine              string                 `json:"tagLine,omitempty"`
	Trustee              string                 `json:"trustee,omitempty"`
	ImageUrl             string                 `json:"imageUrl,omitempty"`
	CreatedAt            string                 `json:"createdAt,omitempty"`
	Classes              []FundClass            `json:"classes,omitempty"`
	IsOutOfService       bool                   `json:"isOutOfService"`
	OutOfServiceMessage  string                 `json:"outOfServiceMessage,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

type FundClass struct {
	Sequence                    int                    `json:"sequence,omitempty"`
	Label                       string                 `json:"label,omitempty"`
	BaseCurrency                string                 `json:"baseCurrency,omitempty"`
	ManagementFee               float64                `json:"managementFee,omitempty"`
	TrusteeFee                  float64                `json:"trusteeFee,omitempty"`
	CustodianFee                float64                `json:"custodianFee,omitempty"`
	TransferFee                 float64                `json:"transferFee,omitempty"`
	TrusteeFeeAnnualMinimum     float64                `json:"trusteeFeeAnnualMinimum,omitempty"`
	SwitchingFee                float64                `json:"switchingFee,omitempty"`
	SubscriptionFee             float64                `json:"subscriptionFee,omitempty"`
	RedemptionFee               float64                `json:"redemptionFee,omitempty"`
	PerformanceFee              float64                `json:"performanceFee,omitempty"`
	TaxRate                     float64                `json:"taxRate,omitempty"`
	MinimumInitialInvestment    float64                `json:"minimumInitialInvestment,omitempty"`
	MinimumAdditionalInvestment float64                `json:"minimumAdditionalInvestment,omitempty"`
	MinimumUnitsHeld            float64                `json:"minimumUnitsHeld,omitempty"`
	MinimumRedemptionAmount     float64                `json:"minimumRedemptionAmount,omitempty"`
	CanDistribute               bool                   `json:"canDistribute,omitempty"`
	LaunchPrice                 float64                `json:"launchPrice,omitempty"`
	HexColor                    string                 `json:"hexColor,omitempty"`
	CommencementAt              string                 `json:"commencementAt,omitempty"`
	InitialOfferingPeriodFrom   string                 `json:"initialOfferingPeriodFrom,omitempty"`
	InitialOfferingPeriodTo     string                 `json:"initialOfferingPeriodTo,omitempty"`
	CreatedAt                   string                 `json:"createdAt,omitempty"`
	DistributionFrequency       string                 `json:"distributionFrequency,omitempty"`
	TagLine                     string                 `json:"tagLine,omitempty"`
	Metadata                    map[string]interface{} `json:"metadata,omitempty"`
}

type GetFundInput struct {
	FundID string `json:"fundId,omitempty"`
}

type GetFundOutput struct {
	Fund *Fund `json:"fund,omitempty"`
}

func (c *Client) GetFund(ctx context.Context, input *GetFundInput) (*GetFundOutput, error) {
	output := GetFundOutput{}
	err := c.query(ctx, "get_fund", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type GetRequestByDuitNowEndToEndIDInput struct {
	AccountID  string `json:"accountId,omitempty"`
	EndToEndID string `json:"endToEndId,omitempty"`
}

type GetRequestByDuitNowEndToEndIDOutput struct {
	RequestID string `json:"requestId,omitempty"`
}

func (c *Client) GetRequestByDuitNowEndToEndID(ctx context.Context, input *GetRequestByDuitNowEndToEndIDInput) (*GetRequestByDuitNowEndToEndIDOutput, error) {
	output := GetRequestByDuitNowEndToEndIDOutput{}
	err := c.query(ctx, "get_request_by_duitnow_endToEndId", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type AllocationPerformance struct {
	Date                 string  `json:"date,omitempty"`
	Units                float64 `json:"units,omitempty"`
	Asset                string  `json:"asset,omitempty"`
	NetAssetValuePerUnit float64 `json:"netAssetValuePerUnit,omitempty"`
	Value                float64 `json:"value,omitempty"`
	PostFeeAmount        float64 `json:"postFeeAmount,omitempty"`
}

type GetClientAccountAllocationPerformanceInput struct {
	AccountID         string `json:"accountId,omitempty"`
	AllocationID      string `json:"allocationId,omitempty"`
	Type              string `json:"type,omitempty"`
	FundClassSequence int    `json:"fundClassSequence,omitempty"`
	Timeframe         string `json:"timeframe,omitempty"`
	Interval          string `json:"interval,omitempty"`
}

type GetClientAccountAllocationPerformanceOutput struct {
	Performance []AllocationPerformance `json:"performance"`
}

func (c *Client) GetClientAccountAllocationPerformance(ctx context.Context, input *GetClientAccountAllocationPerformanceInput) (*GetClientAccountAllocationPerformanceOutput, error) {
	output := GetClientAccountAllocationPerformanceOutput{}
	err := c.query(ctx, "get_client_account_allocation_performance", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type GetClientAccountStatementInput struct {
	AccountID string `json:"accountId,omitempty"`
	FromDate  string `json:"fromDate,omitempty"`
	ToDate    string `json:"toDate,omitempty"`
	Format    string `json:"format"`
}

type GetClientAccountStatementOutput struct {
	FromDate string `json:"fromDate,omitempty"`
	ToDate   string `json:"toDate,omitempty"`
	Format   string `json:"format,omitempty"`
	Filename string `json:"filename,omitempty"`
	Bytes    []byte `json:"bytes,omitempty"`
}

func (c *Client) GetClientAccountStatement(ctx context.Context, input *GetClientAccountStatementInput) (*GetClientAccountStatementOutput, error) {
	output := GetClientAccountStatementOutput{}
	err := c.query(ctx, "get_client_account_statement", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

type GetClientAccountRequestConfirmationInput struct {
	AccountID string `json:"accountId,omitempty"`
	RequestID string `json:"requestId,omitempty"`
	Format    string `json:"format,omitempty"`
}

type GetClientAccountRequestConfirmationOutput struct {
	Format   string `json:"format,omitempty"`
	Filename string `json:"filename,omitempty"`
	Bytes    []byte `json:"bytes,omitempty"`
}

func (c *Client) GetClientAccountRequestConfirmation(ctx context.Context, input *GetClientAccountRequestConfirmationInput) (*GetClientAccountRequestConfirmationOutput, error) {
	output := GetClientAccountRequestConfirmationOutput{}
	err := c.query(ctx, "get_client_account_request_confirmation", input, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
