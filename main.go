package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/springernature/envvars-servicebroker/handler"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func getEnvVar(v string, def string) string {
	r := os.Getenv(v)
	if r == "" {
		return def
	}
	return r
}

func main() {
	appPort := getEnvVar("PORT", "3000") // default of martini
	appName := "some-service"
	appEnv, err := cfenv.Current()
	var appURL string
	if err == nil {
		appURL = fmt.Sprintf("https://%s", appEnv.ApplicationURIs[0])
		appName = appEnv.Name
	} else {
		appURL = "http://localhost:" + appPort
	}
	baseGUID := getEnvVar("SERVICE_BASE_GUID", "29140B3F-0E69-4C7E-8A35")
	serviceName := getEnvVar("SERVICE_NAME", appName)
	servicePlan := getEnvVar("SERVICE_PLAN", "shared")
	serviceDescription := getEnvVar("SERVICE_DESCRIPTION", "Shared service for "+serviceName)
	authUser := getEnvVar("SERVICE_AUTH_USER", "")
	authPassword := getEnvVar("SERVICE_AUTH_PASSWORD", "")
	syslogDrainUrl := getEnvVar("SYSLOG_DRAIN_URL", "")
	tags := getEnvVar("SERVICE_TAGS", "")
	serviceDashboardURL := getEnvVar("SERVICE_DASHBOARD_URL", fmt.Sprintf("%s/dashboard", appURL))
	metadataDisplayName := getEnvVar("SERVICE_METADATA_DISPLAYNAME", serviceName)
	metadataLongDescription := getEnvVar("SERVICE_METADATA_LONGDESC", serviceDescription)
	metadataImageURL := getEnvVar("SERVICE_METADATA_IMAGEURL", "")
	metadataProviderDisplayName := getEnvVar("SERVICE_METADATA_PROVIDERDISPLAYNAME", "")
	metadataDocumentationUrl := getEnvVar("SERVICE_METADATA_DOCURL", "")
	metadataSupportUrl := getEnvVar("SERVICE_METADATA_SUPPORTURL", "")
	credentials := getEnvVar("SERVICE_CREDENTIALS", "{}")
	// Each provision/deprovision request will support an async GET /last_operation request
	fakeAsync := getEnvVar("SERVICE_FAKE_ASYNC", "") == "true"

	fmt.Println("Running as", appURL)

	r := chi.NewRouter()

	if (authUser != "") && (authPassword != "") {
		// secure service broker with basic auth if both env variables are set
		r.Use(handler.New("Authorization Required", map[string][]string{
			authUser: {authPassword},
		}))
	}

	r.Use(middleware.Timeout(60 * time.Second))
	c := handler.BrokerCatalog{
		Tags:                        tags,
		SyslogDrainUrl:              syslogDrainUrl,
		BaseGUID:                    baseGUID,
		ServiceName:                 serviceName,
		ServiceDescription:          serviceDescription,
		MetadataDisplayName:         metadataDisplayName,
		MetadataImageURL:            metadataImageURL,
		MetadataLongDescription:     metadataLongDescription,
		MetadataProviderDisplayName: metadataProviderDisplayName,
		MetadataDocumentationUrl:    metadataDocumentationUrl,
		MetadataSupportUrl:          metadataSupportUrl,
		ServicePlan:                 servicePlan,
	}
	d := handler.ServiceDashboard{
		ServiceName: serviceName,
		ServicePlan: servicePlan,
	}
	lo := handler.LastOperation{}

	r.Route("/v2/service_instances/{service_id}", func(r chi.Router) {
		r.Get("/lastOperation", lo.ServeHTTP)
		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			serviceID := chi.URLParam(r, "service_id")
			csi := handler.CreateServiceInstance{
				DashboardURL: serviceDashboardURL,
				FakeAsync:    fakeAsync,
				ServiceName:  serviceName,
				ServicePlan:  servicePlan,
				ServiceID:    serviceID,
			}
			csi.Handle(w, r)
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			serviceID := chi.URLParam(r, "service_id")
			dsi := handler.DeleteServiceInstance{
				ServiceName: serviceName,
				ServicePlan: servicePlan,
				FakeAsync:   fakeAsync,
				ServiceID:   serviceID,
			}
			dsi.Handle(w, r)
		})
	})
	r.Route("/v2/service_instances/{service_id}/service_bindings/{service_binding_id}", func(r chi.Router) {
		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			serviceID := chi.URLParam(r, "service_id")
			serviceBID := chi.URLParam(r, "service_binding_id")
			csb := handler.CreateServiceBinding{
				ServiceName:      serviceName,
				ServicePlan:      servicePlan,
				SyslogDrainUrl:   syslogDrainUrl,
				Credentials:      credentials,
				ServiceID:        serviceID,
				ServiceBindingID: serviceBID,
			}
			csb.Handle(w, r)
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			serviceID := chi.URLParam(r, "service_id")
			serviceBID := chi.URLParam(r, "service_binding_id")
			dsb := handler.DeleteServiceBinding{
				ServiceName:      serviceName,
				ServicePlan:      servicePlan,
				ServiceID:        serviceID,
				ServiceBindingID: serviceBID,
			}
			dsb.Handle(w, r)
		})
	})

	r.Get("/v2/catalog", c.ServeHTTP)
	r.Get("/dashboard", d.ServeHTTP)

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
