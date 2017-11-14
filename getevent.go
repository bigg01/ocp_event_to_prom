package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	//"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"net"

	"strings"

	"golang.org/x/build/kubernetes/api"
)

// Stream : structure for holding the stream of data coming from OpenShift
type Stream struct {
	Type  string    `json:"type,omitempty"`
	Event api.Event `json:"object"`
}

func main() {
	//apiAddr := os.Getenv("OPENSHIFT_API_URL")
	apiAddr := "https://127.0.0.1:8443"
	apiToken := "nhmqaQNQxAG7O3PemhClSBxMwt3CFCyYGtM-hxgyLaU"
	//apiToken := os.Getenv("OPENSHIFT_TOKEN")
	//syslogServer := os.Getenv("SYSLOG_SERVER")
	//syslogProto := strings.ToLower(os.Getenv("SYSLOG_PROTO"))
	syslogTag := strings.ToUpper(os.Getenv("SYSLOG_TAG"))
	//ignoreSSL := strings.ToUpper(os.Getenv("IGNORE_SSL"))
	ignoreSSL := "TRUE"
	debugFlag := strings.ToUpper(os.Getenv("DEBUG"))

	// enable signal trapping
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point
		sig := <-c
		log.Fatalf("Signal (%v) Detected, Shutting Down", sig)
	}()

	// check and make sure we have the minimum config information before continuing
	if apiAddr == "" {
		// use the default internal cluster URL if not defined
		apiAddr = "https://openshift.default.svc.cluster.local"
		ignoreSSL = "TRUE"
		log.Print("Missing environment variable OPENSHIFT_API_URL. Using default API URL")
	}
	if apiToken == "" {
		// if we dont set it in the environment variable, read it out of
		// /var/run/secrets/kubernetes.io/serviceaccount/token
		log.Print("Missing environment variable OPENSHIFT_TOKEN. Leveraging serviceaccount token")
		fileData, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err != nil {
			log.Fatal("Service Account token does not exist.")
		}
		apiToken = string(fileData)
	}
	if syslogTag == "" {
		// we dont need to error out here, but we do need to set a default if the variable isnt defined
		syslogTag = "OSE"
	}
	if ignoreSSL == "" {
		// we dont need to error out here, but we do need to set a default if the variable isnt defined
		ignoreSSL = "FALSE"
	}
	if debugFlag == "" {
		// we dont need to error out here, but we do need to set a default if the variable isnt defined
		debugFlag = "FALSE"
	}

	// setup ose connection
	var client http.Client
	if ignoreSSL == "TRUE" {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client = http.Client{Transport: tr}
	} else {
		client = http.Client{}
	}
	req, err := http.NewRequest("GET", apiAddr+"/api/v1/events?watch=true", nil)
	if err != nil {
		log.Fatal("## Error while opening connection to openshift api", err)
	}
	req.Header.Add("Authorization", "Bearer "+apiToken)

	fmt.Printf("echo curl %v", apiAddr+"/api/v1/events?watch=true")

	for {
		resp, err := client.Do(req)

		if err != nil {
			log.Println("## Error while connecting to:", apiAddr, err)
			time.Sleep(5 * time.Second)
			continue
		}

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				log.Println("## Error reading from response stream.", err)
				resp.Body.Close()
				break
			}

			event := Stream{}
			decErr := json.Unmarshal(line, &event)
			if decErr != nil {
				log.Println("## Error decoding json.", err)
				resp.Body.Close()
				break
			}
			if "Node" == event.Event.InvolvedObject.Kind {
				//volumes_work_duration{quantile="0.5"} 20

				fmt.Printf("nodevent{project=%v,name=%v,kind=%v,reason=%v} 1\n",

					event.Event.Namespace, event.Event.InvolvedObject.Name,
					event.Event.Kind, event.Event.Reason)

				/*
					fmt.Printf("%v | Project: %v | Name: %v | Kind: %v | Reason: %v | Message: %v\n",
						event.Event.LastTimestamp,
						event.Event.Namespace, event.Event.Name,
						event.Event.Kind, event.Event.Reason, event.Event.Message)
				*/
			}

			//fmt.Println(event.Event.InvolvedObject.Kind)
			/*
				fmt.Printf("%v | Project: %v | Name: %v | Kind: %v | Reason: %v | Message: %v\n",
					event.Event.LastTimestamp,
					event.Event.Namespace, event.Event.Name,
					event.Event.Kind, event.Event.Reason, event.Event.Message)
			*/
		}
	}

}
