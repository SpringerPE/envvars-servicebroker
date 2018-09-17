package handler

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/kr/pretty"
	"net/http"
	"strings"
)

type BrokerCatalog struct {
	Tags                        string
	SyslogDrainUrl              string
	ServiceBinding              serviceBindingResponse
	BaseGUID                    string
	ServiceName                 string
	ServiceDescription          string
	MetadataDisplayName         string
	MetadataImageURL            string
	MetadataLongDescription     string
	MetadataProviderDisplayName string
	MetadataDocumentationUrl    string
	MetadataSupportUrl          string
	ServicePlan                 string
}
type LastOperation struct{}
type CreateServiceInstance struct {
	DashboardURL string
	FakeAsync    bool
	ServiceName  string
	ServicePlan  string
	ServiceID    string
}
type DeleteServiceInstance struct {
	ServiceName string
	ServicePlan string
	FakeAsync   bool
	ServiceID   string
}
type CreateServiceBinding struct {
	ServiceName      string
	ServicePlan      string
	SyslogDrainUrl   string
	Credentials      string
	ServiceID        string
	ServiceBindingID string
}
type DeleteServiceBinding struct {
	ServiceName      string
	ServicePlan      string
	ServiceID        string
	ServiceBindingID string
}
type ServiceDashboard struct {
	ServiceName string
	ServicePlan string
}

type serviceBindingResponse struct {
	Credentials    map[string]interface{} `json:"credentials"`
	SyslogDrainURL string                 `json:"syslog_drain_url,omitempty"`
}

type lastOperationResponse struct {
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
}

func (bc BrokerCatalog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tagArray := []string{}
	if len(bc.Tags) > 0 {
		tagArray = strings.Split(bc.Tags, ",")
	}
	var requires []string
	if bc.SyslogDrainUrl != "" {
		requires = []string{"syslog_drain"}
	}
	catalog := cf.Catalog{
		Services: []*cf.Service{
			{
				ID:          bc.BaseGUID + "-service-" + bc.ServiceName,
				Name:        bc.ServiceName,
				Description: bc.ServiceDescription,
				Bindable:    true,
				Tags:        tagArray,
				Requires:    requires,
				Metadata: &cf.ServiceMeta{
					DisplayName:         bc.MetadataDisplayName,
					ImageURL:            bc.MetadataImageURL,
					Description:         bc.MetadataLongDescription,
					ProviderDisplayName: bc.MetadataProviderDisplayName,
					DocURL:              bc.MetadataDocumentationUrl,
					SupportURL:          bc.MetadataSupportUrl,
				},
				Plans: []*cf.Plan{
					{
						ID:          bc.BaseGUID + "-plan-" + bc.ServicePlan,
						Name:        bc.ServicePlan,
						Description: bc.ServiceDescription,
						Free:        true,
					},
				},
			},
		},
	}
	json, err := json.Marshal(catalog)
	if err != nil {
		fmt.Println("Um, how did we fail to marshal this catalog:")
		fmt.Printf("%# v\n", pretty.Formatter(catalog))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (lo LastOperation) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lastOp := lastOperationResponse{
		State:       "succeeded",
		Description: "async in action",
	}
	json, err := json.Marshal(lastOp)
	if err != nil {
		fmt.Println("Um, how did we fail to marshal this service instance:")
		fmt.Printf("%# v\n", pretty.Formatter(lastOp))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (si CreateServiceInstance) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Creating service instance %s for service %s plan %s\n", si.ServiceID, si.ServiceName, si.ServicePlan)
	instance := cf.ServiceCreationResponse{
		DashboardURL: si.DashboardURL,
	}
	json, err := json.Marshal(instance)
	if err != nil {
		fmt.Println("Um, how did we fail to marshal this service instance:")
		fmt.Printf("%# v\n", pretty.Formatter(instance))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}
	if si.FakeAsync {
		w.WriteHeader(http.StatusAccepted)
		w.Write(json)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(json)
}

func (di DeleteServiceInstance) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Deleting service instance %s for service %s plan %s\n", di.ServiceID, di.ServiceName, di.ServicePlan)
	if di.FakeAsync {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("{}"))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (sb CreateServiceBinding) Handle(w http.ResponseWriter, r *http.Request) {
	type serviceCredentials map[string]interface{}
	fmt.Printf(
		"Creating service binding %s for service %s plan %s instance %s\n",
		sb.ServiceBindingID,
		sb.ServiceName,
		sb.ServicePlan,
		sb.ServiceID,
	)

	c := make(serviceCredentials)
	e := json.Unmarshal([]byte(sb.Credentials), &c)
	if e != nil {
		fmt.Printf("Failed to load credentials: %s", sb.Credentials)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}

	serviceBinding := serviceBindingResponse{
		Credentials: c,
	}
	if sb.SyslogDrainUrl != "" {
		serviceBinding.SyslogDrainURL = sb.SyslogDrainUrl
	}
	json, err := json.Marshal(serviceBinding)
	if err != nil {
		fmt.Println("Um, how did we fail to marshal this binding:")
		fmt.Printf("%# v\n", pretty.Formatter(serviceBinding))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(json)
}

func (di DeleteServiceBinding) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Printf(
		"Delete service binding %s for service %s plan %s instance %s\n",
		di.ServiceBindingID,
		di.ServiceName,
		di.ServicePlan,
		di.ServiceID,
	)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (sd ServiceDashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Show dashboard for service %s plan %s\n", sd.ServiceName, sd.ServicePlan)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Dashboard"))
}
