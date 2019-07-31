package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type (

	//ParamLocation the location of the param in the request
	ParamLocation string

	//ParamFormat the format of the param will be used to get the friendly strings from siac "10TB" "100SC"
	ParamFormat string

	//CommandParam a known parameter of a command. Used only if the parameter needs special formatting or needs to be part of the help text
	CommandParam struct {
		Key       string
		HelpText  string
		Location  ParamLocation
		Formatter ParamFormat
	}

	//CommandEndpoint a known Sia API endpoint. Describes how the endpoint should be accessed, any help text and any parameters that are required
	CommandEndpoint struct {
		Path               string
		AlternativeMatches []string
		Method             string
		HelpText           string
		Params             []CommandParam
	}

	//Command the command parsed from the input
	Command struct {
		Endpoint    CommandEndpoint
		RequestPath string
		Method      string
		UserAgent   string
		APIAddress  string
		APIPassword string
		Params      map[string][]string
	}
)

const (
	//URLParam the parameter should go in the url as part of the path
	URLParam ParamLocation = "url"

	//QueryParam the parameter should go in the query
	QueryParam ParamLocation = "query"

	//BodyParam the parameter should go in the body
	BodyParam ParamLocation = "body"

	//DefaultFormat an unformatted parameter
	DefaultFormat ParamFormat = ""

	//DataFormat a parameter formatted in the friendly data size format "10TB"
	DataFormat ParamFormat = "data"

	//PriceFormat a parameter formatted in the Siacoin price format "100SC"
	PriceFormat ParamFormat = "price"

	//MonthlyPriceFormat a parameter formatted in the Siacoin monthly price format "100SC"
	MonthlyPriceFormat ParamFormat = "monthlyprice"

	//BlockTimeFormat a parameter formatted in the 10 minutes per block format "10w"
	BlockTimeFormat ParamFormat = "monthlyprice"
)

var (
	//DefaultAPIPassword the default Sia API Password
	DefaultAPIPassword string
)

//SiaAPIEndpoints all current endpoints listed in https://sia.tech/docs as of v1.4.1
var SiaAPIEndpoints = []CommandEndpoint{
	CommandEndpoint{
		Path:   "/consensus",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/consensus/blocks",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/consensus/validate/transactionset",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/daemon/constants",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/daemon/settings",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/daemon/settings",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/daemon/stop",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/daemon/update",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/daemon/update",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/daemon/version",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/gateway",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/gateway",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/gateway/connect/:netaddress",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/gateway/disconnect/:netaddress",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/host",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host/announce",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host/contracts",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/host/storage",
		Method: "GET",
		AlternativeMatches: []string{
			"/host/folders",
		},
	},
	CommandEndpoint{
		Path:   "/host/storage/folders/add",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host/storage/folders/remove",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host/storage/folders/resize",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host/storage/sectors/delete/:merkleroot",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/host/estimatescore",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/hostdb",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/hostdb/active",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/hostdb/all",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/hostdb/hosts/:pubkey",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/hostdb/filtermode",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/hostdb/filtermode",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/miner",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/miner/start",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/miner/stop",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/miner/header",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/miner/header",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/contract/cancel",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/backup",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/recoverbackup",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/uploadedbackups",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/contracts",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/dir/*siapath",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/dir/*siapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/downloads",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/downloads/clear",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/prices",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/files",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/file/*siapath",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/file/*siapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/delete/s*iapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/download/*siapath",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/download/cancel",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/downloadsync/*siapath",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/recoveryscan",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/recoveryscan",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/rename/*siapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/stream/*siapath",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/renter/upload/*siapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/uploadstream/*siapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/renter/validate/*siapath",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/tpool/confirmed/:id",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/tpool/fee",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/tpool/raw/:id",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/tpool/raw",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/tpool/confirmed/:id",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/033x",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/address",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/addresses",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/seedaddrs",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/backup",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/changepassword",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/init",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/init/seed",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/seed",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/seeds",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/siacoins",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/siafunds",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/siagkey",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/sign",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/sweep/seed",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/lock",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/transaction/:id",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/transactions",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/transactions/:addr",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/unlock",
		Method: "POST",
	},
	CommandEndpoint{
		Path:   "/wallet/unlockconditions/:addr",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/unspent",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/verify/address/:addr",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/watch",
		Method: "GET",
	},
	CommandEndpoint{
		Path:   "/wallet/watch",
		Method: "POST",
	},
}

// DefaultSiaDir returns the default data directory of siad. The values for
// supported operating systems are:
//
// Linux:   $HOME/.sia
// MacOS:   $HOME/Library/Application Support/Sia
// Windows: %LOCALAPPDATA%\Sia
func DefaultSiaDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Sia")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Sia")
	default:
		return filepath.Join(os.Getenv("HOME"), ".sia")
	}
}

//LoadDefaultAPIPassword loads the default Sia API password from the environment variable or the apipassword file
func LoadDefaultAPIPassword() (password string, err error) {
	if password = os.Getenv("SIA_API_PASSWORD"); len(password) > 0 {
		return
	}

	passBuf, err := ioutil.ReadFile(filepath.Join(DefaultSiaDir(), "apipassword"))

	if err != nil {
		return
	}

	password = strings.TrimSpace(string(passBuf))

	return
}

func matchPaths(path, template string) bool {
	pathSegments := strings.Split(path, "/")
	segments := strings.Split(template, "/")

	if len(segments) == 0 || len(pathSegments) == 0 {
		return false
	}

	if len(pathSegments) < len(segments) {
		return false
	}

	for i, pathSeg := range pathSegments {
		if len(segments) <= i {
			return false
		}

		seg := segments[i]

		if strings.HasPrefix(seg, ":") {
			continue
		}

		if strings.HasPrefix(seg, "*") {
			return true
		}

		if seg != pathSeg {
			return false
		}
	}

	return true
}

func matchEndpoints(cmd Command) (endpoints []CommandEndpoint) {
	for _, endpoint := range SiaAPIEndpoints {
		if !matchPaths(cmd.RequestPath, endpoint.Path) {
			continue
		}

		if len(cmd.Method) > 0 && cmd.Method != endpoint.Method {
			continue
		}

		endpoints = append(endpoints, endpoint)
	}

	return
}

func parseInputs(args []string) (apiCommand Command) {
	apiCommand = Command{
		APIAddress:  "localhost:9980",
		APIPassword: DefaultAPIPassword,
		UserAgent:   "Sia-Agent",
		Params:      make(map[string][]string),
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if len(arg) == 0 {
			continue
		}

		if strings.HasPrefix(arg, "--") {
			key := strings.ToLower(arg[2:])
			value := ""

			if len(args) > i+1 && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			}

			if key == "method" {
				apiCommand.Method = strings.ToUpper(value)
				continue
			} else if key == "addr" {
				apiCommand.APIAddress = value
				continue
			} else if key == "useragent" {
				apiCommand.UserAgent = value
				continue
			} else if key == "apipassword" {
				apiCommand.APIPassword = value
				continue
			}

			apiCommand.Params[key] = append(apiCommand.Params[key], value)
			continue
		}

		apiCommand.RequestPath += "/" + arg
	}

	return
}

func makeRequest(cmd Command, body io.Reader) (req *http.Request, err error) {
	urlStr := "http://" + cmd.APIAddress + cmd.RequestPath

	if err != nil {
		return
	}

	if cmd.Method == "GET" && len(cmd.Params) > 0 {
		urlStr += "?" + url.Values(cmd.Params).Encode()
	} else if cmd.Method == "POST" && body == nil && len(cmd.Params) > 0 {
		body = strings.NewReader(url.Values(cmd.Params).Encode())
	}

	req, err = http.NewRequest(cmd.Method, urlStr, body)

	if err != nil {
		return
	}

	req.SetBasicAuth("", cmd.APIPassword)
	req.Header.Add("User-Agent", cmd.UserAgent)

	if cmd.Method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	return
}

func main() {
	var err error

	DefaultAPIPassword, err = LoadDefaultAPIPassword()

	if err != nil {
		os.Stderr.WriteString("unable to load API password")
		os.Exit(1)
	}

	command := parseInputs(os.Args[1:])

	endpoints := matchEndpoints(command)

	if len(endpoints) == 0 && len(command.Method) == 0 {
		os.Stderr.WriteString("No matching endpoints. Try specifying the request method or checking http://sia.tech/docs")
		os.Exit(127)
	}

	if len(endpoints) > 1 && len(command.Method) == 0 {
		os.Stderr.WriteString("More than one matching endpoint. Try specifying the request method or checking http://sia.tech/docs")
		os.Exit(127)
	}

	if len(endpoints) > 0 {
		command.Endpoint = endpoints[0]

		if len(command.Method) == 0 {
			command.Method = command.Endpoint.Method
		}
	}

	req, err := makeRequest(command, nil)

	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	_, err = io.Copy(os.Stdout, resp.Body)

	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	return
}
