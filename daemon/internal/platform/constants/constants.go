package constants

const UNIX_SOCK_NAME = "backerd.sock" // The UNIX socket name
const CRONTAB_TAG_PREFIX = "#backer:" // Default suffix used in crontab
const NOF_WORKERS_DEFAULT = 4         // Number of default workers for running jobs

type SMTP_Provider struct {
	Hostname string // The hostname of the smtp domain
	Port     uint16 // The network port
	Ssl      bool   // If SSL is enabled or not
}

var SMTP_PROVIDERS = map[string]SMTP_Provider{
	"gmail.com":   {"smtp.gmail.com", 587, false},
	"outlook.com": {"smtp.office365.com", 587, false},
	"yahoo.com":   {"smtp.mail.yahoo.com", 465, true},
	"icloud.com":  {"smtp.mail.me.com", 587, false},
}

var AVAILABLE_WEBHOOKS = map[string]string{
	"discord": "https://discord.com/api/webhooks/", // The discord webhook prefix
}

var WEEKDAY_NAMES = map[string]string{
	"0": "Sunday", "1": "Monday", "2": "Tuesday", "3": "Wednesday",
	"4": "Thursday", "5": "Friday", "6": "Saturday", "7": "Sunday",
}

var MONTH_NAMES = map[string]string{
	"1": "January", "2": "February", "3": "March", "4": "April", "5": "May",
	"6": "June", "7": "July", "8": "August", "9": "September", "10": "October",
	"11": "November", "12": "December",
}

var COMMON_4XX_STATUS_CODE = map[int]string{
	400: "Webhook endpoint rejected the request (400 Bad Request)",
	401: "Webhook endpoint requires authentication (401 Unauthorized)",
	403: "Webhook endpoint denied access (403 Forbidden)",
	404: "Webhook endpoint not found (404)",
	408: "Webhook endpoint request timed out (408)",
}

var HTTP_RETRY_STATUS = []int{408, 429, 500, 502, 503, 504}
