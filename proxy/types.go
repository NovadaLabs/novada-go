package proxy

// --- Shared ----------------------------------------------------------------

// TimeRange is a start/end window used by the traffic-consumption endpoints.
// Both bounds use the "2006-01-02 15:04:05" layout and are required.
type TimeRange struct {
	Start string // required; "2006-01-02 15:04:05"
	End   string // required; "2006-01-02 15:04:05"
}

// FlowBalance is the remaining-traffic payload shared by the residential,
// rotating ISP and rotating datacenter balance endpoints. ExpireTime is a Unix
// timestamp (seconds) and is zero for products that do not report it.
type FlowBalance struct {
	Balance    int64 `json:"balance"`
	ExpireTime int64 `json:"expire_time"`
}

// FlowConsumeLog is a single traffic-consumption record shared by the
// residential, rotating ISP, rotating datacenter and mobile consumption
// endpoints. All byte counts are in bytes.
type FlowConsumeLog struct {
	ID         int   `json:"id"`
	UID        int   `json:"uid"`
	Balance    int64 `json:"balance"`
	AllBuy     int64 `json:"all_buy"`
	Use        int64 `json:"use"`
	Day        int64 `json:"day"`
	ExpireFlow int64 `json:"expire_flow"`
}

// FlowConsumeLogList is the data payload of the traffic-consumption endpoints.
type FlowConsumeLogList struct {
	List []FlowConsumeLog `json:"list"`
}

// --- Account ---------------------------------------------------------------

// CreateAccountParams are the parameters for AccountService.Create. Product,
// Account, Password and Status are required.
type CreateAccountParams struct {
	Product  Product // required; e.g. ProductResidential, ProductBrowserAPI
	Account  string  // required; account name
	Password string  // required; account password
	Status   int     // required; 1=normal, -3=disabled
	Remark   string  // optional remark
	// LimitFlow is the dynamic residential data cap in GB, sent as a string.
	LimitFlow string
}

// ListAccountParams are the parameters for AccountService.List. Product is
// required; Page and Limit default to 1 and 10 when left at zero.
type ListAccountParams struct {
	Product Product // required
	Status  int     // optional filter; 1=normal, -3=disabled
	Account string  // optional account-name filter
	Page    int     // page number (default 1)
	Limit   int     // entries per page (default 10)
}

// UpdateAccountParams are the parameters for AccountService.Update. ID, Account
// and Password are required.
type UpdateAccountParams struct {
	ID        int    // required; account ID
	Account   string // required; account name
	Password  string // required; account password
	Status    int    // optional; 1=normal, -3=disabled
	Remark    string // optional remark
	LimitFlow string // optional residential data cap in GB
}

// ConsumeLogParams are the parameters for AccountService.ConsumeLog. AccountID
// is required; Page and Limit default to 1 and 10 when left at zero. Times use
// the "2006-01-02 15:04:05" layout.
type ConsumeLogParams struct {
	AccountID int    // required
	StartTime string // optional; "2006-01-02 15:04:05"
	EndTime   string // optional; "2006-01-02 15:04:05"
	Page      int    // page number (default 1)
	Limit     int    // entries per page (default 10)
}

// Account is a proxy sub-account as returned by AccountService.List.
type Account struct {
	CreatedAt               int64  `json:"created_at"`
	UpdatedAt               int64  `json:"updated_at"`
	ID                      int    `json:"id"`
	UID                     int    `json:"uid"`
	Account                 string `json:"account"`
	Password                string `json:"password"`
	Status                  int    `json:"status"`
	ResidentialBalance      int64  `json:"residential_balance"`
	ResidentialAllBuy       int64  `json:"residential_all_buy"`
	ResidentialStatus       int    `json:"residential_status"`
	DCBalance               int64  `json:"dc_balance"`
	DCAllBuy                int64  `json:"dc_all_buy"`
	DCStatus                int    `json:"dc_status"`
	ISPBalance              int64  `json:"isp_balance"`
	ISPAllBuy               int64  `json:"isp_all_buy"`
	ISPStatus               int    `json:"isp_status"`
	FlowType                string `json:"flow_type"`
	AccountType             string `json:"account_type"`
	CheckWhiteList          int    `json:"check_white_list"`
	Remark                  string `json:"remark"`
	ConsumedResidentialFlow int64  `json:"consumed_residential_flow"`
	LimitResidentialFlow    int64  `json:"limit_residential_flow"`
	ConsumedDCFlow          int64  `json:"consumed_dc_flow"`
	LimitDCFlow             int64  `json:"limit_dc_flow"`
	ConsumedISPFlow         int64  `json:"consumed_isp_flow"`
	LimitISPFlow            int64  `json:"limit_isp_flow"`
}

// AccountList is the data payload of AccountService.List.
type AccountList struct {
	List  []Account `json:"list"`
	Page  int       `json:"page"`
	Total int       `json:"total"`
}

// AccountConsumeLog is a single daily traffic-consumption record.
type AccountConsumeLog struct {
	ID                      int   `json:"id"`
	AccountID               int   `json:"account_id"`
	UID                     int   `json:"uid"`
	ResidentialBalance      int64 `json:"residential_balance"`
	ResidentialAllBuy       int64 `json:"residential_all_buy"`
	ConsumedResidentialFlow int64 `json:"consumed_residential_flow"`
	DCBalance               int64 `json:"dc_balance"`
	DCAllBuy                int64 `json:"dc_all_buy"`
	ConsumedDCFlow          int64 `json:"consumed_dc_flow"`
	ISPBalance              int64 `json:"isp_balance"`
	ISPAllBuy               int64 `json:"isp_all_buy"`
	ConsumedISPFlow         int64 `json:"consumed_isp_flow"`
	// Day is the Unix timestamp (seconds) of the record's day.
	Day int64 `json:"day"`
}

// AccountConsumeLogList is the data payload of AccountService.ConsumeLog.
type AccountConsumeLogList struct {
	List []AccountConsumeLog `json:"list"`
}

// --- Whitelist -------------------------------------------------------------

// AddWhitelistParams are the parameters for WhitelistService.Add. Product and
// IP are required. Valid products: 1=Residential, 5=Static ISP, 4=Unlimited.
type AddWhitelistParams struct {
	Product Product // required
	IP      string  // required
	Remark  string  // optional
}

// ListWhitelistParams are the parameters for WhitelistService.List. Product is
// required; the remaining fields are optional filters.
type ListWhitelistParams struct {
	Product   Product // required
	IP        string  // optional IP filter
	StartTime string  // optional; "2006-01-02 15:04:05"
	EndTime   string  // optional; "2006-01-02 15:04:05"
	// Lock filters by lock state: 0=unlocked, 1=locked. Nil means no filter.
	Lock *int
}

// DeleteWhitelistParams are the parameters for WhitelistService.Delete. Product
// and IPs are required; pass multiple IPs comma-separated.
type DeleteWhitelistParams struct {
	Product Product // required
	IPs     string  // required; comma-separated list, e.g. "1.1.1.1,2.2.2.2"
}

// RemarkWhitelistParams are the parameters for WhitelistService.Remark. Product
// and ID are required.
type RemarkWhitelistParams struct {
	Product Product // required
	ID      string  // required; the whitelist entry ID
	Remark  string  // optional new remark
}

// WhitelistIP is a whitelist entry as returned by WhitelistService.List.
type WhitelistIP struct {
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
	ID        int   `json:"id"`
	UID       int   `json:"uid"`
	// MarkIP is the human-readable IP string (e.g. "10.10.10.1").
	MarkIP string `json:"mark_ip"`
	// IP is the integer form of the address as stored by the API.
	IP      int64  `json:"ip"`
	Status  int    `json:"status"`
	Lock    bool   `json:"lock"`
	LockUID int    `json:"lock_uid"`
	Flag    int    `json:"flag"`
	IsBiger int    `json:"is_biger"`
	IsLimit int    `json:"is_limit"`
	Mark    string `json:"mark"`
}

// WhitelistList is the data payload of WhitelistService.List.
type WhitelistList struct {
	List  []WhitelistIP `json:"list"`
	Total int           `json:"total"`
}
