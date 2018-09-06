package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/cloudfoundry-community/types-cf"
	"github.com/springernature/envvars-servicebroker/internal"
)

type response struct {
	header int
	body   []byte
}

type partialResponse struct {
	body []byte
}

func TestGetCatalog(t *testing.T) {
	bc := BrokerCatalog{}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodGet,
	}
	bc.ServeHTTP(w, r)
	raw, err := ioutil.ReadFile("../test/catalog.json")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	e := response{
		200,
		raw,
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return empty catalog for empty env vars",
			expected: e,
		},
	}

	for _, tc := range testCases {
		body := &cf.Catalog{}
		expectedBody := &cf.Catalog{}
		assertBodyAgainstExpectations(tc.body, body, tc.expected, expectedBody, tc.header, tc.name, t, true)
	}
}

func TestGetLastOperation(t *testing.T) {
	lo := LastOperation{}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodGet,
	}
	lo.ServeHTTP(w, r)

	lastOp := lastOperationResponse{
		State:       "succeeded",
		Description: "async in action",
	}
	l, _ := json.Marshal(lastOp)
	expectResp := response{
		200,
		l,
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return last operation",
			expected: expectResp,
		},
	}

	for _, tc := range testCases {
		body := &lastOperationResponse{}
		expectedBody := &lastOperationResponse{}
		assertBodyAgainstExpectations(tc.body, body, tc.expected, expectedBody, tc.header, tc.name, t, true)
	}
}

func TestCreateServiceInstance(t *testing.T) {
	si := CreateServiceInstance{
		DashboardURL: "amazing.dashboard.com",
		ServiceID:    "some service id",
	}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodPut,
	}

	si.Handle(w, r)

	creationResp := cf.ServiceCreationResponse{
		DashboardURL: "amazing.dashboard.com",
	}
	cr, _ := json.Marshal(creationResp)
	expectResp := response{
		201,
		cr,
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return a successful creation response",
			expected: expectResp,
		},
	}

	for _, tc := range testCases {
		body := &cf.ServiceCreationResponse{}
		expectedBody := &cf.ServiceCreationResponse{}
		assertBodyAgainstExpectations(tc.body, body, tc.expected, expectedBody, tc.header, tc.name, t, true)
	}
}

func TestDeleteServiceInstance(t *testing.T) {
	di := DeleteServiceInstance{
		ServiceID: "some service id",
	}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodDelete,
	}

	di.Handle(w, r)

	expectResp := response{
		200,
		[]byte("{}"),
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return a successful deletion response",
			expected: expectResp,
		},
	}

	for _, tc := range testCases {
		assertBodyAgainstExpectations(tc.body, tc.body, tc.expected, tc.expected.body, tc.header, tc.name, t, false)
	}
}

func TestCreateServiceBinding(t *testing.T) {
	sb := CreateServiceBinding{
		ServiceName:      "service name",
		ServicePlan:      "service plan",
		Credentials:      `{"port": 5514, "host": "syslog-app.snpaas.eu"}`,
		SyslogDrainUrl:   "syslog.drain.url",
		ServiceID:        "some service id",
		ServiceBindingID: "some service binding",
	}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodPut,
	}
	sb.Handle(w, r)

	cred := make(map[string]interface{})
	cred["port"] = 5514
	cred["host"] = "syslog-app.snpaas.eu"

	creationResp := serviceBindingResponse{
		Credentials:    cred,
		SyslogDrainURL: "syslog.drain.url",
	}
	cr, _ := json.Marshal(creationResp)
	expectResp := response{
		201,
		cr,
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return a successful binding response",
			expected: expectResp,
		},
	}

	for _, tc := range testCases {
		body := &serviceBindingResponse{}
		expectedBody := &serviceBindingResponse{}
		assertBodyAgainstExpectations(tc.body, body, tc.expected, expectedBody, tc.header, tc.name, t, true)
	}
}

func TestDeleteServiceBinding(t *testing.T) {
	id := "some service id"
	bId := "some service binding id"
	di := DeleteServiceBinding{
		ServiceID:        id,
		ServiceBindingID: bId,
	}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodDelete,
	}

	di.Handle(w, r)

	expectResp := response{
		200,
		[]byte("{}"),
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return a successful deletion response",
			expected: expectResp,
		},
	}

	for _, tc := range testCases {
		assertBodyAgainstExpectations(tc.body, tc.body, tc.expected, tc.expected.body, tc.header, tc.name, t, false)
	}
}

func TestShowServiceDashboard(t *testing.T) {
	sd := ServiceDashboard{}
	w := internal.CreateFakeWriter()
	r := &http.Request{
		Method: http.MethodGet,
	}

	sd.ServeHTTP(w, r)

	expectResp := response{
		200,
		[]byte("Dashboard"),
	}

	testCases := []struct {
		header   int
		body     []byte
		name     string
		expected response
	}{
		{
			header:   w.Status,
			body:     w.Bytes,
			name:     "Should return a successful deletion response",
			expected: expectResp,
		},
	}

	for _, tc := range testCases {
		assertBodyAgainstExpectations(tc.body, tc.body, tc.expected, tc.expected.body, tc.header, tc.name, t, false)
	}
}

func assertBodyAgainstExpectations(
	tcBody []byte,
	body interface{},
	tcExpected response,
	expectedBody interface{},
	header int,
	tcName string,
	t *testing.T,
	handleJson bool,
) {
	if handleJson {
		e := json.Unmarshal(tcBody, body)
		if e != nil {
			t.Fatalf("Error while marshalling body")
		}
		err := json.Unmarshal(tcExpected.body, expectedBody)
		if err != nil {
			t.Fatalf("Error while marshalling expected body")
		}

	}
	if header != tcExpected.header || !reflect.DeepEqual(body, expectedBody) {
		j, _ := json.Marshal(body)
		je, _ := json.Marshal(expectedBody)
		t.Errorf(
			"Test %s should return header |%d| and body |%s| but returned header |%d| and body |%s|",
			tcName,
			tcExpected.header,
			string(je),
			header,
			string(j),
		)
	}
}
