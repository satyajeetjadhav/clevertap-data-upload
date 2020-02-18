package globals

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var CSVFilePath *string
var JSONFilePath *string
var SchemaFilePath *string
var MixpanelSecret *string
var LeanplumClientKey *string
var LeanplumAppID *string
var ImportService *string
var AWSSecretAccessKey *string
var AWSAccessKeyID *string
var AWSRegion *string
var S3Bucket *string
var StartDate *string
var EndDate *string
var AccountID *string
var AccountPasscode *string
var AccountToken *string
var EvtName *string
var Type *string
var Region *string
var DryRun *bool
var StartTs *float64
var LeanplumOutFilesPath *string
var LeanplumAPIEndpoint *string
var AmplitudeStart *string
var AmplitudeEnd *string

//var AutoConvert *bool

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var MPEventsFilePaths arrayFlags
var FEvents arrayFlags

func Init() bool {
	flag.Var(&MPEventsFilePaths, "mixpanelEventsFile", "Absolute path to the MixPanel events file")
	flag.Var(&FEvents, "filterEvent", "Event to be filtered (would not be uploaded)")
	CSVFilePath = flag.String("csv", "", "Absolute path to the csv file")
	JSONFilePath = flag.String("json", "", "Absolute path to the json file")
	SchemaFilePath = flag.String("schema", "", "Absolute path to the schema file")
	MixpanelSecret = flag.String("mixpanelSecret", "", "Mixpanel API secret key")
	LeanplumClientKey = flag.String("leanplumClientKey", "", "Leanplum Client Key")
	LeanplumAppID = flag.String("leanplumAppID", "", "Leanplum App ID")
	LeanplumOutFilesPath = flag.String("leanplumOutFilesPath", "", "Absolute path to file that contains names of files generated by LeanPlum")
	LeanplumAPIEndpoint = flag.String("leanplumAPIEndpoint", "", "LeanPlum API Endpoint")
	ImportService = flag.String("importService", "", "Service you want to import data from")
	AWSAccessKeyID = flag.String("awsAccessKeyID", "", "AWS access key id")
	AWSSecretAccessKey = flag.String("awsSecretAccessKey", "", "AWS secret access key")
	AWSRegion = flag.String("awsRegion", "", "AWS Region")
	S3Bucket = flag.String("s3Bucket", "", "S3 bucket")
	StartDate = flag.String("startDate", "", "Start date for exporting events "+
		"<yyyy-mm-dd>")
	EndDate = flag.String("endDate", "", "End date for exporting events "+
		"<yyyy-mm-dd>")
	AmplitudeStart = flag.String("amplitudeStart", "", "First hour included in data series, formatted YYYYMMDDTHH (e.g. '20150201T05')")
	AmplitudeEnd = flag.String("amplitudeEnd", "", "Last hour included in data series, formatted YYYYMMDDTHH (e.g. '20150203T20')")
	StartTs = flag.Float64("startTs", 0, "Start timestamp for events upload")
	AccountID = flag.String("id", "", "CleverTap Account ID")
	AccountPasscode = flag.String("p", "", "CleverTap Account Passcode")
	AccountToken = flag.String("tk", "", "CleverTap Account Token")
	EvtName = flag.String("evtName", "", "Event name")
	Type = flag.String("t", "profile", "The type of data, either profile, event, or both, defaults to profile")
	Region = flag.String("r", "eu", "The account region, either eu, in, sk, or sg, defaults to eu")
	DryRun = flag.Bool("dryrun", false, "Do a dry run, process records but do not upload")
	//AutoConvert = flag.Bool("autoConvert", false, "automatically covert property value type to number for number entries")
	flag.Parse()
	if (*JSONFilePath == "" && *CSVFilePath == "" && *MixpanelSecret == "" && MPEventsFilePaths == nil && *ImportService == "") || *AccountID == "" || (*AccountPasscode == "" && *ImportService != "leanplumToS3" && *ImportService != "leanplumToS3Throttled") {
		log.Println("Mixpanel secret or CSV file path or JSON file path or Mixpanel events file path or Import service option, account id, and passcode are mandatory")
		return false
	}
	if (*CSVFilePath != "" || *JSONFilePath != "") && *MixpanelSecret != "" {
		log.Println("Both Mixpanel secret and CSV file path detected. Only one data source is allowed")
		return false
	}
	if *Type != "profile" && *Type != "event" && *Type != "both" {
		log.Println("Type can be either profile or event")
		return false
	}
	if *CSVFilePath != "" && *EvtName == "" && *Type == "event" {
		log.Println("Event name is mandatory for event csv uploads")
		return false
	}
	if *MixpanelSecret != "" && *Type == "event" && *StartDate == "" {
		log.Println("Start date is mandatory when exporting events from Mixpanel. Format: <yyyy-mm-dd>")
		return false
	}
	if *MixpanelSecret != "" && *Type == "event" && *StartDate != "" {
		//check start date format
		_, err := time.Parse("2006-01-02", *StartDate)
		if err != nil {
			log.Println("Start date is not in correct format. Format: <yyyy-mm-dd>")
			return false
		}
	}
	if *MixpanelSecret != "" && *Type == "event" && *EndDate != "" {
		//check end date format
		_, err := time.Parse("2006-01-02", *EndDate)
		if err != nil {
			log.Println("End date is not in correct format. Format: <yyyy-mm-dd>")
			return false
		}
	}
	if *EndDate != "" && *StartDate != "" {
		//start date should be less than or equal to end date
		s, _ := time.Parse("2006-01-02", *StartDate)
		e, _ := time.Parse("2006-01-02", *EndDate)
		if s.After(e) {
			log.Println("Start date cannot be after End date")
			return false
		}
	}
	if *AmplitudeEnd != "" && *AmplitudeStart != "" {
		//start date should be less than or equal to end date
		e := strings.Split(*AmplitudeEnd, "T")
		if len(e) != 2 {
			log.Println("Amplitude end date is in invalid format")
			return false
		}
		s := strings.Split(*AmplitudeStart, "T")
		if len(s) != 2 {
			log.Println("Amplitude start date is in invalid format")
			return false
		}
		eTime, err := time.Parse("20060102", e[0])
		if err != nil {
			log.Println("Amplitude end date is in invalid format")
			return false
		}
		sTime, err := time.Parse("20060102", s[0])
		if err != nil {
			log.Println("Amplitude start date is in invalid format")
			return false
		}
		if sTime.After(eTime) {
			log.Println("Start date cannot be after End date")
			return false
		}
	}
	if MPEventsFilePaths != nil && len(MPEventsFilePaths) > 0 && *Type != "event" {
		log.Println("Mixpanel events file path is supported only with events")
		return false
	}
	if *Region != "eu" && *Region != "in" && *Region != "sk" && *Region != "sg" {
		log.Println("Region can be either eu, in, sk, or sg")
		return false
	}
	if *ImportService == "mparticle" && (*AWSSecretAccessKey == "" || *AWSAccessKeyID == "" || *S3Bucket == "" ||
		*AWSRegion == "") {
		log.Println("Importing from mparticle requires AWS access key, secret key, region, and S3 bucket")
		return false
	}

	if (*ImportService == "leanplumToS3" || *ImportService == "leanplumS3ToCT" || *ImportService == "leanplumToS3Throttled") && (*AWSSecretAccessKey == "" || *AWSAccessKeyID == "" || *S3Bucket == "" ||
		*AWSRegion == "" || *LeanplumAppID == "" || *LeanplumClientKey == "" || *StartDate == "" ||
		*EndDate == "" || *LeanplumOutFilesPath == "") {
		log.Println("Importing from Leanplum to S3 requires AWS access key, secret key, region, S3 bucket, " +
			"leanplum app ID, leanplum client ID, leanplum out files path, and start and end date")
		return false
	}

	if (*ImportService == "leanplumToS3" || *ImportService == "leanplumS3ToCT" || *ImportService == "leanplumToS3Throttled") && *EndDate != "" {
		//check end date format
		t, err := time.Parse("2006-01-02", *EndDate)
		if err != nil {
			log.Println("End date is not in correct format. Format: <yyyy-mm-dd>")
			return false
		}
		*EndDate = t.Format("20060102")
	}

	if (*ImportService == "leanplumToS3" || *ImportService == "leanplumS3ToCT" || *ImportService == "leanplumToS3Throttled") && *StartDate != "" {
		//check start date format
		t, err := time.Parse("2006-01-02", *StartDate)
		if err != nil {
			log.Println("Start date is not in correct format. Format: <yyyy-mm-dd>")
			return false
		}
		*StartDate = t.Format("20060102")
	}

	if *ImportService == "leanplumS3ToCT" && *AccountToken == "" {
		log.Println("Account token is missing")
		return false
	}

	return true
}

var Schema map[string]string

func ParseSchema(file *os.File) bool {
	/**
	{
		"key": "Float",
		"key 1": "Integer",
		"key 2": "Number",
		"key 3": "Float[]",
		"key 4": "Integer[]",
		"key 5": "String[]",
		"key 6": "Boolean[]"
	}
	*/
	err := json.NewDecoder(file).Decode(&Schema)
	if err != nil {
		log.Println(err)
		log.Println("Unable to parse schema file")
		return false
	}
	return true
}

var FilterEventsSet map[string]bool

func InitFilterEventsSet() {
	FilterEventsSet = make(map[string]bool)
	for _, v := range FEvents {
		FilterEventsSet[v] = true
	}
}
