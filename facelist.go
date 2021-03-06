/*
Copyright 2018 Tink AB

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	//"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	//"time"
	"github.com/open-networks/go-msgraph"
	//"fmt"

	"gopkg.in/yaml.v2"
)

var (
	cfg           config
	userlist 	  msgraph.Users
	IndexTemplate = template.Must(template.ParseFiles("templates/index.html"))
)
type (
	config struct {
		EmailFilter   string `yaml:"emailFilter"`
		GraphAPIToken string `yaml:"graphAPIToken"`
		ApplicationID string `yaml:"applicationID"`
		TenantID      string `yaml:"tenantID"`
		GroupID		  string `yaml:"groupID"`
	}
)



func init() {
	log.Println("Starting facelist")

	configFile := flag.String("config", "scouterna.yaml", "Configuration file to load")
	flag.Parse()
	b, err := ioutil.ReadFile(*configFile)

	if err != nil {
		log.Fatalf("Unable to read config: %v\n", err)
	}

	err = yaml.Unmarshal(b, &cfg)

	if err != nil {
		log.Fatalf("Unable to decode config: %v\n", err)
	}

	if cfg.ApplicationID == "" {
		log.Fatalf("appID is not set!")
		os.Exit(1)
	}
	if cfg.TenantID == "" {
		log.Fatalf("tenantID is not set!")
		os.Exit(1)
	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//client := http.Client{Timeout: time.Duration(5 * time.Second)}

	// Use mocked data for local dev
	if cfg.GraphAPIToken == "" {
		userlist = getMockedUsers()
	} else {
		graphClient, err := msgraph.NewGraphClient(cfg.TenantID, cfg.ApplicationID, cfg.GraphAPIToken)
		if err != nil {
    		log.Println("Credentials are probably wrong or system time is not synced: ", err)
		}


		var g msgraph.Group
		g, err = graphClient.GetGroup(cfg.GroupID)
		userlist, err = g.ListMembers()



		if err != nil {
			log.Printf(err.Error())
		}
	}

	// Filter out deleted accounts, bots and users without email addresses according to cfg.EmailFilter

	filteredUsers := []msgraph.User{}

	for _, user := range userlist {
		if strings.HasSuffix(user.Mail, cfg.EmailFilter) {
			filteredUsers = append(filteredUsers, user)
		}
	}


	//Sort users on first name

	sort.SliceStable(filteredUsers, func(i, j int) bool {
		return strings.ToLower(filteredUsers[i].DisplayName) < strings.ToLower(filteredUsers[j].DisplayName)
	})


	userlist = filteredUsers

	if err := IndexTemplate.Execute(w, userlist); err != nil {
		log.Printf("Failed to execute index template: %v\n", err)
		http.Error(w, "Oops. That's embarrassing. Please try again later.", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}
