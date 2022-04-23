//=======================================================================
// Copyright (c) 2017-2020 Aarna Networks, Inc.
// All rights reserved.
// ======================================================================
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//           http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ========================================================================

package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"example.com/middleend/db"
	"example.com/middleend/localstore"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type deployServiceData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Spec        struct {
		ProjectName string     `json:"projectName"`
		Apps        []appsData `json:"appsData"`
	} `json:"spec"`
}

type deployDigData struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	CompositeAppName    string `json:"compositeApp"`
	CompositeProfile    string `json:"compositeProfile"`
	DigVersion          string `json:"version"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	NwIntents           bool   `json:"nwIntent,omitempty"`
	LogicalCloud        string `json:"logicalCloud"`
	Spec                DigSpec `json:"spec"`
}

type DigSpec struct {
	ProjectName string     `json:"projectName"`
	Apps        []appsData `json:"appsData"`
	OverrideValuesObj []localstore.OverrideValues `json:"override-values"`
}


// Exists is for mongo $exists filter
type Exists struct {
	Exists string `json:"$exists"`
}

// This is the json payload that the orchestration API expects.
type appsData struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		FileName    string `json:"filename"`
		FileContent string `json:"filecontent, omitempty"`
	} `json:"metadata"`
	ProfileMetadata struct {
		Name        string `json:"name"`
		FileName    string `json:"filename"`
		FileContent string `json:"filecontent, omitempty"`
	} `json:"profileMetadata"`
	BlueprintModels []struct {
		ArtifactName    string `json:"artifactName"`
		ArtifactVersion string `json:"artifactVersion"`
		Workflows       []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
		} `json:"workflows"`
	} `json:"blueprintModels"`
	Interfaces         []NwInterfaces `json:"interfaces"`
	PlacementCriterion string         `json:"placementCriterion"`
	Clusters           []ClusterInfo  `json:"clusters"`
	RsInfo             []ResourceInfo `json:"resourceData"`
}

type NwInterfaces struct {
	NetworkName   string `json:"networkName"`
	IP            string `json:"ip"`
	Subnet        string `json:"subnet"`
	InterfaceName string `json:"interfaceName"`
}

type ClusterInfo struct {
	Provider         string            `json:"clusterProvider"`
	SelectedClusters []SelectedCluster `json:"selectedClusters"`
	SelectedLabels   []SelectedLabel   `json:"selectedLabels"`
}

type SelectedCluster struct {
	Name string `json:"name"`
}

type SelectedLabel struct {
	Name string `json:"clusterLabel"`
}

type CompositeAppsInProject struct {
	Metadata    apiMetaData `json:"metadata" bson:"metadata"`
	Status      string      `json:"status" bson:"status"`
	ProjectName string      `json:"project,omitempty" bson:"project,omitempty"`
	Spec        struct {
		Version      string                              `json:"compositeAppVersion" bson:"compositeAppVersion"`
		AppsArray    []*Application                      `json:"apps,omitempty" bson:"apps,omitempty"`
		ProfileArray []*Profiles                         `json:"compositeProfiles,omitempty" bson:"compositeProfiles,omitempty"`
		DigArray     []*localstore.DeploymentIntentGroup `json:"deploymentIntentGroups,omitempty" bson:"deploymentIntentGroups,omitempty"`
	} `json:"spec" bson:"spec"`
}
type CompositeAppsInProjectShrunk struct {
	Metadata apiMetaData        `json:"metadata" bson:"metadata"`
	Spec     []CompositeAppSpec `json:"spec" bson:"spec"`
}

type CompositeAppSpec struct {
	Status       string                              `json:"status" bson:"status"`
	Version      string                              `json:"compositeAppVersion" bson:"compositeAppVersion"`
	AppsArray    []*Application                      `json:"apps,omitempty" bson:"apps,omitempty"`
	ProfileArray []*Profiles                         `json:"compositeProfiles,omitempty" bson:"compositeProfiles,omitempty"`
	DigArray     []*localstore.DeploymentIntentGroup `json:"deploymentIntentGroups,omitempty" bson:"deploymentIntentGroups,omitempty"`
}

type Profiles struct {
	Metadata appMetaData `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Spec     struct {
		ProfilesArray []ProfileMeta `json:"profile,omitempty" bson:"profile,omitempty"`
	} `json:"spec,omitempty" bson:"spec,omitempty"`
}

type DigsInProject struct {
	Metadata struct {
		Name                string `json:"name"`
		CompositeAppName    string `json:"compositeAppName"`
		CompositeAppVersion string `json:"compositeAppVersion"`
		Description         string `json:"description"`
		UserData1           string `userData1:"userData1"`
		UserData2           string `userData2:"userData2"`
	} `json:"metadata"`
	Spec struct {
		Status            string                      `json:"status, omitempty"`
		DigIntentsData    []DigDeployedIntents        `json:"deployedIntents"`
		Profile           string                      `json:"profile"`
		Version           string                      `json:"version"`
		Lcloud            string                      `json:"logicalCloud"`
		TargetVersion     string                      `json:"targetVersion"`
		OverrideValuesObj []localstore.OverrideValues `json:"overrideValues"`
		GpintArray        []*DigsGpint                `json:"GenericPlacementIntents,omitempty"`
		NwintArray        []*DigsNwint                `json:"networkCtlIntents,omitempty"`
		IsCheckedOut      bool                        `json:"is_checked_out"`
		Operation         string                      `json:operation,omitempty`
	} `json:"spec"`
}

type DigsGpint struct {
	Metadata localstore.GenIntentMetaData `json:"metadata,omitempty"`
	Spec     struct {
		AppIntentArray []PlacementIntentExport `json:"placementIntent,omitempty"`
	} `json:"spec,omitempty"`
}

type DigsNwint struct {
	Metadata apiMetaData `json:"metadata,omitempty"`
	Spec     struct {
		WorkloadIntentsArray []*WorkloadIntents `json:"WorkloadIntents,omitempty"`
	} `json:"spec,omitempty"`
}
type WorkloadIntents struct {
	Metadata apiMetaData `json:"metadata,omitempty"`
	Spec     struct {
		AppName    string        `json:"appName"`
		Interfaces []NwInterface `json:"interfaces,omitempty"`
	} `json:"spec,omitempty"`
}

// Project Tree
type ProjectTree struct {
	Metadata        ProjectMetadata
	compositeAppMap map[string]*CompositeAppTree
}

type treeTraverseFilter struct {
	compositeAppName      string
	compositeAppVersion   string
	digName               string
	compositeAppMultiPart bool
}

// Composite app tree
type CompositeAppTree struct {
	Metadata         CompositeApp
	Status           string
	AppsDataArray    map[string]*AppsData
	ProfileDataArray map[string]*ProfilesData
	DigMap           map[string]*DigReadData
}

type DigReadData struct {
	DigpData       localstore.DeploymentIntentGroup
	DigIntentsData DigpIntents
	GpintMap       map[string]*GpintData
	NwintMap       map[string]*NwintData
}

type GpintData struct {
	Gpint          localstore.GenericPlacementIntent
	AppIntentArray []localstore.AppIntent
}

type NwintData struct {
	Nwint     NetworkCtlIntent
	WrkintMap map[string]*WrkintData
}

type WrkintData struct {
	Wrkint     NetworkWlIntent
	Interfaces []NwInterface
}

type AppsData struct {
	App              Application
	CompositeProfile ProfileMeta
}

type ProfilesData struct {
	Profile     ProfileMeta
	AppProfiles []ProfileMeta
}

type ClusterMetadata struct {
	Metadata apiMetaData `json:"Metadata"`
}

type apiMetaData struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	UserData1   string `userData1:"userData1"`
	UserData2   string `userData2:"userData2"`
}

type appMetaData struct {
	Name         string `json:"name" bson:"name"`
	Description  string `json:"description" bson:"description"`
	UserData1    string `userData1:"userData1"`
	UserData2    string `userData2:"userData2"`
	ChartContent string `json:"chartContent" bson:"chartContent",omitempty`
	Status       string `json:"status" bson: "status",omitempty`
}

// The interface
type orchWorkflow interface {
	createAnchor() interface{}
	createObject() interface{}
	getObject() (interface{}, interface{})
	getAnchor() (interface{}, interface{})
	deleteObject() interface{}
	deleteAnchor() interface{}
}

// MiddleendConfig: The configmap of the middleend
type MiddleendConfig struct {
	OwnPort     string `json:"ownport"`
	Clm         string `json:"clm"`
	Dcm         string `json:"dcm"`
	Ncm         string `json:"ncm"`
	Gac         string `json:"gac"`
	OrchService string `json:"orchestrator"`
	OvnService  string `json:"ovnaction"`
	CfgService  string `json:"configSvc"`
	Mongo       string `json:"mongo"`
	LogLevel    string `json:"logLevel"`
	AppInstantiate    bool `json:"appInstantiate"`
	StoreName   string `json:"storeName"`
}

// OrchestrationHandler interface, handling the composite app APIs
type OrchestrationHandler struct {
	MiddleendConf                MiddleendConfig
	client                       http.Client
	meta                         []appsData
	DigData                      deployDigData
	resourceInfo                 ResourceInfo
	file                         map[string]*multipart.FileHeader
	dataRead                     *ProjectTree
	treeFilter                   *treeTraverseFilter
	guiDigViewJSON               guiDigView
	DigpReturnJSON               []DigsInProject
	CompositeAppReturnJSON       []CompositeAppsInProject
	CompositeAppReturnJSONShrunk []CompositeAppsInProjectShrunk
	ClusterProviders             []ClusterProvider
	DigStatusJSON                *digStatus
	Vars                         map[string]string
	bstore                       backendStore
	digStore                     digBackendStore
	genK8sInfo                   map[string]*GenericK8sIntentInfo
	response                     struct {
		lastKey    string
		payload    map[string][]byte
		status     map[string]int
		status_msg map[string]string
	}
}

type HealthcheckResponse struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

// NewAppHandler interface implementing REST callhandler
func NewAppHandler() *OrchestrationHandler {
	return &OrchestrationHandler{}
}

// GetHealth to check connectivity
func (h OrchestrationHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	healthcheckResponse := HealthcheckResponse{
		Name:   "amcop_middleend",
		Status: "pass"}
	retval, _ := json.Marshal(healthcheckResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}

func (h OrchestrationHandler) apiGet(url string, statusKey string) (interface{}, []byte, error) {
	// prepare and DEL API
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	if statusKey != "" {
	   h.response.payload[statusKey] = data
	   h.response.status[statusKey] = resp.StatusCode
	   h.response.status_msg[statusKey] = resp.Status
	}
	return resp.StatusCode, data, nil
}

func (h OrchestrationHandler) apiGetWithArguments(url string, statusKey string, arguments [][]string) (interface{}, []byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	q := request.URL.Query()
	for _, argument := range arguments {
		q.Add(argument[0], argument[1])
	}
	request.URL.RawQuery = q.Encode()
	resp, err := h.client.Do(request)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warningln(err)
	}
	if statusKey != "" {
		h.response.payload[statusKey] = data
		h.response.status[statusKey] = resp.StatusCode
		h.response.status_msg[statusKey] = resp.Status
	}
	return resp.StatusCode, data, nil
}

func (h OrchestrationHandler) apiGetMultiPart(url string, statusKey string) (interface{}, []byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("Accept", "multipart/form-data; charset=utf-8")
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	defer resp.Body.Close()

	var cSpecContent localstore.SpecFileContent
	var cz localstore.Customization
	_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	mr := multipart.NewReader(resp.Body, params["boundary"])
	for part, err := mr.NextPart(); err == nil; part, err = mr.NextPart() {
		value, _ := ioutil.ReadAll(part)
		log.Debugf("FormName is: %s", part.FormName())
		log.Debugf("Value: %s", value)
		if part.FormName() == "customization" {
			json.Unmarshal(value, &cz)
		}
		if part.FormName() == "file" {
			h.response.payload[statusKey] = value
			break
		} else if part.FormName() == "files" {
			if cz.Metadata.UserData2 == "Secret" {
				temp := string(value)
				cSpecContent.FileContents = strings.Split(temp, "\n")
			} else if cz.Metadata.UserData2 == "ConfigMap" {
				cSpecContent.FileContents = append(cSpecContent.FileContents, string(value))
			}
		}
	}

	if len(cSpecContent.FileContents) > 0 {
		h.response.payload[statusKey], _ = json.Marshal(cSpecContent)
	}

	h.response.status[statusKey] = resp.StatusCode
	h.response.status_msg[statusKey] = resp.Status

	return resp.StatusCode, h.response.payload[statusKey], nil
}

func (h OrchestrationHandler) apiDel(url string, statusKey string) (interface{}, error) {
	// prepare and DEL API
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return resp.StatusCode, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	h.response.payload[statusKey] = data
	h.response.status[statusKey] = resp.StatusCode
	h.response.status_msg[statusKey] = resp.Status

	return resp.StatusCode, nil
}

func (h OrchestrationHandler) apiPost(jsonLoad []byte, url string, statusKey string) (interface{}, error) {
	// prepare and POST API
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonLoad))
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return resp.StatusCode, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	if statusKey != "" {
		h.response.payload[statusKey] = data
		h.response.status[statusKey] = resp.StatusCode
		h.response.status_msg[statusKey] = resp.Status
	}
	return resp.StatusCode, nil
}

func (h OrchestrationHandler) apiPostMultipart(jsonLoad []byte,
	fh *multipart.FileHeader, url string, statusKey string, fileNames []string, fileContents []string) (interface{}, error) {
	// Open the file
	var file multipart.File
	var err error
	if fh != nil {
		file, err = fh.Open()
		if err != nil {
			return nil, err
		}
		// Close the file later
		defer file.Close()
	}
	// Buffer to store our request body as bytes
	var requestBody bytes.Buffer
	// Create a multipart writer
	multiPartWriter := multipart.NewWriter(&requestBody)
	// Initialize the file field. Arguments are the field name and file name
	// It returns io.Writer
	for i, fileName := range fileNames {
		var fileWriter io.Writer
		if h.Vars["multipartfiles"] != "true" {
			fileWriter, err = multiPartWriter.CreateFormFile("file", fileNames[0])
		} else {
			fileWriter, err = multiPartWriter.CreateFormFile("files", fileName)
		}
		if err != nil {
			return nil, err
		}
		// Copy the actual file content to the field field's writer
		if file != nil {
			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = io.Copy(fileWriter, strings.NewReader(fileContents[i]))
			if err != nil {
				return nil, err
			}
		}
	}
	// Populate other fields
	fieldWriter, err := multiPartWriter.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}

	_, err = fieldWriter.Write([]byte(jsonLoad))
	if err != nil {
		return nil, err
	}

	// We completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	multiPartWriter.Close()

	// By now our original request body should have been populated,
	// so let's just use it with our custom request
	log.Debugf("request body: %s", requestBody)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, err
	}
	// We need to set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	// Do the request
	resp, err := h.client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	h.response.status_msg[statusKey] = resp.Status
	h.response.status[statusKey] = resp.StatusCode
	h.response.payload[statusKey] = data

	return resp.StatusCode, nil
}

func (h *OrchestrationHandler) prepTreeReq() {
	// Initialise the project tree with target composite application.
	h.treeFilter = &treeTraverseFilter{}
	h.treeFilter.compositeAppName = h.Vars["compositeAppName"]
	h.treeFilter.compositeAppVersion = h.Vars["version"]
	h.treeFilter.digName = h.Vars["deploymentIntentGroupName"]
	h.treeFilter.compositeAppMultiPart, _ = strconv.ParseBool(h.Vars["multipart"])
}

// DelDig: Delete the deployment intent group tree
func (h *OrchestrationHandler) DelDig(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	filter := r.URL.Query().Get("operation")

	var originalVersion string
	var retCode int
	if filter == "deleteAll" {
		digInfo := h.FetchDIGInfo(h.Vars["deploymentIntentGroupName"])

		for _, version := range digInfo.VersionList {
			h.Vars["version"] = version
			retCode, _ = h.DeleteDig(filter)
			if retCode != http.StatusNoContent {
				w.WriteHeader(retCode)
				return
			}
		}

		// Clear DIG Info from diginfo collection
		h.DeleteDIGInfo()
	} else {
		retCode, originalVersion = h.DeleteDig(filter)
		if retCode != http.StatusNoContent {
			w.WriteHeader(retCode)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Original-Version", originalVersion)
	w.WriteHeader(http.StatusNoContent)
}

// Delete service workflow
func (h *OrchestrationHandler) DelSvc(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.treeFilter = nil
	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"ProfileHandler"}
	h.InitializeResponseMap()

	// Initialise the project tree with target composite application.
	h.prepTreeReq()

	h.dataRead = &ProjectTree{}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	log.Infof("tree %+v\n", h.dataRead)
	// Check if a dig is present in this composite application
	if len(h.dataRead.compositeAppMap[h.Vars["compositeAppName"]+"-"+h.Vars["version"]].DigMap) != 0 {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Non emtpy DIG in service\n"))
		return
	}

	// 1. Call Service delete workflow
	log.Info("Start Service delete workflow")
	deleteDataPoints := []string{"ProfileHandler",
		"compAppHandler"}
	retcode = h.deleteTree(deleteDataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Get DIG Status
func (h *OrchestrationHandler) GetDigStatus(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()
	// Get the DIG detailed status
	temp := &remoteStoreDigHandler{}
	temp.orchInstance = h
	thisDigStatus, respcode := temp.getStatus(h.Vars["compositeAppName"],
		h.Vars["version"], h.Vars["deploymentIntentGroupName"])
	if respcode != http.StatusOK {
		w.WriteHeader(respcode)
		return
	} else {
		h.DigStatusJSON = &thisDigStatus
		log.Infof("status %+v\n", h.DigStatusJSON)
		log.Infof("data  %+v\n", h.dataRead)

		// Fetch all versions for a given composite application
		retCode, versionList := h.GetCompAppVersions("")
		if retCode != http.StatusOK {
			w.WriteHeader(retCode)
			return
		}

		localDigStore := localStoreDigHandler{}
		for _, version := range versionList {
			localDigRetCode, _, _ := localDigStore.getDig(h.Vars["projectName"],
				h.Vars["compositeAppName"], version, h.Vars["deploymentIntentGroupName"])
			if localDigRetCode == http.StatusOK {
				thisDigStatus.IsCheckedOut = true
				h.DigStatusJSON.TargetVersion = version
				break
			}
		}

		// copy dig tree
		if len(h.DigStatusJSON.Apps) != 0 {
			h.copyNwToStatus()
			log.Infof("Desc %s", h.DigStatusJSON.Apps[0].Description)
		}
	}
	retval, _ := json.Marshal(h.DigStatusJSON)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}

//GetDigInEdit get all the deployment intents groups by iterating all composite apps in a project
func (h *OrchestrationHandler) GetDigInEdit(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()
	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler", "genericK8sIntentHandler"}

	h.dataRead = &ProjectTree{}
	h.prepTreeReq()
	bstore := &localStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore

	dStore := &localStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore

	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// copy dig tree
	retcode = h.copyDigTreeNew()
	if retcode != nil && retcode != http.StatusOK {
		log.Errorf("Error encountered during checkout of DIG: %s", h.Vars["deploymentIntentGroupName"])
		w.WriteHeader(retcode.(int))
		return
	}
	retval, _ := json.Marshal(h.guiDigViewJSON)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}

//GetAllDigs get all the deployment intents groups by iterating all composite apps in a project
func (h *OrchestrationHandler) GetAllDigs(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()
	h.GetDigs(&w, "emco")
	// copy dig tree
	h.copyDigTree()
	jsonResponse := h.DigpReturnJSON

	h.GetDigs(&w, "middleend")
	h.copyDigTree()

	// Update response
	for m, sdig := range jsonResponse {
		for _, tdig := range h.DigpReturnJSON {
			if sdig.Metadata.Name == tdig.Metadata.Name {
				jsonResponse[m].Spec.IsCheckedOut = true
				jsonResponse[m].Spec.TargetVersion = tdig.Metadata.CompositeAppVersion
				break
			}
		}
	}

	retval, _ := json.Marshal(jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}

// Fetches all composite application from middleend collection of mco, which are in checkout state
func (h *OrchestrationHandler) GetDraftCompositeApplication(key DraftCompositeAppKey, filter string) ([]CompositeAppsInProject, error) {
	var caList []CompositeAppsInProject

	/*var err error
	if key != (DraftCompositeAppKey{}) {
		jsonLoad, err = json.Marshal(key)
		if err != nil {
			log.Errorf("Marshalling of draft composite app key failed: %s", err)
			return nil, err
		}
	}*/

	exists := db.DBconn.CheckCollectionExists(h.MiddleendConf.StoreName)
	if exists {
		values, err := db.DBconn.Find(h.MiddleendConf.StoreName, key, "appmetadata")
		if err != nil {
			log.Errorf("Encountered error while fetching draft composite application: %s", err)
			return nil, err
		} else if len(values) == 0 {
			log.Infof("Draft composite applications does not exists")
		}

		log.Debugf("Draft composite app: %s", values)

		for _, value := range values {
			ca := CompositeAppsInProject{}
			log.Debugf("Draft composite app: %s", value)

			err = db.DBconn.Unmarshal(value, &ca)
			log.Debugf("Draft composite app after Unmarshalling: %s", ca)
			if err != nil {
				log.Errorf("Unmarshalling composite app failed: %s", err)
				return nil, err
			}

			if filter == "" {
				ca.Spec.ProfileArray = nil
				ca.Spec.AppsArray = nil
			}

			caList = append(caList, ca)
		}
		return caList, nil

	}
	return caList, nil
}

// GetSvc get the entire tree under project/<composite app>/<version> for a given composite app
// or fetches all composite apps under project
func (h *OrchestrationHandler) GetSvc(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.treeFilter = nil
	h.InitializeResponseMap()
	filter := r.URL.Query().Get("filter")
	status := r.URL.Query().Get("status")
	if filter != "" && filter != "depthAll" {
		log.Errorf("Invalid query argument provided")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//if any invalid app status is passed, ignore that
	if status != "" && status != "created" && status != "checkout" {
		status = ""
	}

	retCode, retval := h.GetCompApps(filter, status)
	if retCode != http.StatusOK {
		log.Errorf("Ecnountered error while fetching composite apps")
		w.WriteHeader(retCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}

func (h *OrchestrationHandler) GetCompositeAppData(AppName string, projectName string, filter string, status string) (string, int, interface{}){

        var objmap map[string]interface{}
	h.Vars = make(map[string]string)

	h.Vars["compositeAppName"] = AppName
        h.Vars["projectName"] = projectName
        h.Vars["version"] = "v1"
	compositeAppName := h.Vars["compositeAppName"]

	if filter != "" && filter != "depthAll" {
                log.Errorf("Invalid query argument provided")
                return "nil", 0, "nil"
        }

	if status != "" && status != "created" {
                status = ""
        }

        retCode, retval := h.GetCompApps("", "")
        if retCode != http.StatusOK {
                log.Errorf("Ecnountered error while fetching composite apps")
                return "nil", retCode, objmap["status"]
        }
        json.Unmarshal(retval, &objmap)
        log.Infof("composite:%v", objmap)
	h.Vars["compositeAppName"] = ""

	return compositeAppName, retCode, objmap["status"]
}


func (h *OrchestrationHandler) GetCompApps(filter string, status string) (int, []byte) {
	var retval []byte
	var err error
	bstore := &remoteStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore

	dStore := &remoteStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore
	var dataPoints []string
	if filter == "depthAll" {
		dataPoints = []string{"projectHandler", "compAppHandler", "ProfileHandler", "digpHandler"}
	} else {
		dataPoints = []string{"projectHandler"}
	}
	h.prepTreeReq()
	h.dataRead = &ProjectTree{}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		return http.StatusInternalServerError, retval
	}
	if h.treeFilter.compositeAppName != "" {
		h.copyCompositeAppTree(filter)
		if len(h.CompositeAppReturnJSON) == 1 && h.Vars["compositeAppName"] != "" {
			retval, _ = json.Marshal(h.CompositeAppReturnJSON[0])
		} else {
			retval, _ = json.Marshal(h.CompositeAppReturnJSON)
		}
	} else {
		h.createJSONResponse(filter, status)
		if len(h.CompositeAppReturnJSONShrunk) == 1 && h.Vars["compositeAppName"] != "" {
			retval, err = json.Marshal(h.CompositeAppReturnJSONShrunk[0])
		} else {
			retval, err = json.Marshal(h.CompositeAppReturnJSONShrunk)
		}
	}
	if err != nil {
		log.Errorf("Marshalling of CompositeAppReturnJSONShrunk failed: %s", err)
		retval = []byte("some error occurred")
		return http.StatusInternalServerError, retval
	}
	return http.StatusOK, retval
}

func (h *OrchestrationHandler) rollBackApp() {
	dataPoints := []string{"projectHandler", "compAppHandler", "ProfileHandler"}
	h.treeFilter = &treeTraverseFilter{}
	h.treeFilter.compositeAppName = h.Vars["compositeAppName"]
	h.treeFilter.compositeAppVersion = h.Vars["version"]

	h.dataRead = &ProjectTree{}
	/*
		retcode := h.constructTree(dataPoints)
		if retcode != nil {
			return
		}
	*/
	h.constructTree(dataPoints)
	log.Infof("tree %+v\n", h.dataRead)
	// 1. Call rollback workflow
	log.Infof("Start rollback workflow")
	deleteDataPoints := []string{"ProfileHandler",
		"compAppHandler"}
	retcode := h.deleteTree(deleteDataPoints)
	if retcode != nil {
		return
	}
	log.Infof("Rollback suucessful")
}

// CreateApp: Creates all applications and uploaded profiles for a composite application
func (h *OrchestrationHandler) CreateApp(w http.ResponseWriter, r *http.Request) {
	var jsonData deployServiceData
	h.Vars = mux.Vars(r)

	err := r.ParseMultipartForm(16777216)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Populate the multipart.FileHeader MAP. The key will be the
	// filename itself. The metadata Map will be keyed on the application
	// name. The metadata has a field file name, so later we can parse the metadata
	// Map, and fetch the file headers from this file Map with keys as the filename.
	h.file = make(map[string]*multipart.FileHeader)
	for _, v := range r.MultipartForm.File {
		fh := v[0]
		h.file[fh.Filename] = fh
	}

	jsn := ([]byte(r.FormValue("servicePayload")))
	err = json.Unmarshal(jsn, &jsonData)
	if err != nil {
		log.Info("Failed to parse json: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Vars["compositeAppName"] = strings.TrimSpace(jsonData.Name)
	h.Vars["description"] = jsonData.Description
	h.Vars["projectName"] = jsonData.Spec.ProjectName
	h.meta = jsonData.Spec.Apps
	h.Vars["version"] = "v1"

	// Sanity check. For each metadata there should be a
	// corresponding file in the multipart request. If it
	// not found we fail this API call.
	for i := range h.meta {
		switch {
		case h.file[h.meta[i].Metadata.FileName] == nil:
			t := fmt.Sprintf("File %s not in request", h.meta[i].Metadata.FileName)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(t))
			log.Error("app file not found\n")
			return
		case h.file[h.meta[i].ProfileMetadata.FileName] == nil:
			t := fmt.Sprintf("File %s not in request", h.meta[i].ProfileMetadata.FileName)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(t))
			log.Error("profile file not found\n")
			return
		default:
			log.Info("Good request")
		}
	}

	if len(h.meta) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request, no app metadata\n"))
		return
	}

	h.client = http.Client{}

	// These maps will get populated by the return status and responses of each V2 API
	// that is called during the execution of the workflow.
	h.InitializeResponseMap()

	// 1. create the composite application. the compAppHandler implements the
	// orchWorkflow interface.
	appHandler := &compAppHandler{}
	appHandler.orchInstance = h
	httpErr := createCompositeapp(appHandler)
	if httpErr != nil {
		h.rollBackApp()
		if intval, ok := httpErr.(int); ok {
			log.Errorf("CreateCompositeapp failed with error : %d", intval)
			w.WriteHeader(intval)
		} else {
			log.Infof("Encountered error for CreateCompositeapp")
			w.WriteHeader(http.StatusInternalServerError)
		}
		errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
		w.Write([]byte(errMsg))
		return
	}

	// 2. create the composite application profiles
	profileHandler := &ProfileHandler{}
	profileHandler.orchInstance = h
	httpErr = createProfile(profileHandler)
	if httpErr != nil {
		h.rollBackApp()
		if intval, ok := httpErr.(int); ok {
			log.Errorf("CreateProfile failed with error : %d", intval)
			w.WriteHeader(intval)
		} else {
			log.Infof("Encountered error for CreateProfile")
			w.WriteHeader(http.StatusInternalServerError)
		}
		errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(h.response.payload[h.Vars["compositeAppName"]+"_compapp"])
}

func (h *OrchestrationHandler) DIGApprove(namespace string, appname string, digname string) interface{}{

	url := "http://" + h.MiddleendConf.OrchService + "/v2/projects/" + namespace +
                        "/composite-apps/" +appname + "/v1/deployment-intent-groups/" +digname + "/approve"

	var payload []byte

	resp, err := h.apiPost(payload, url, "")
	if err != nil {
            return err
        }
        if resp != http.StatusAccepted {
           return  resp
        }
	log.Infof("Call Approve the Service Instance response: %d", resp)
	return nil
}

func (h *OrchestrationHandler) DIGInstantiate(namespace string, appname string, digname string) interface{}{

	url := "http://" + h.MiddleendConf.OrchService + "/v2/projects/" + namespace +
                        "/composite-apps/"+ appname +"/v1/deployment-intent-groups/" +digname + "/instantiate"
	var payload []byte
	resp , err := h.apiPost(payload, url, "")
	if err != nil {
            return err
        }
        if resp != http.StatusAccepted {
           return  resp
        }
	log.Infof("Call Instantiate the Service Instance response: %d", resp)
	return nil

}

func (h *OrchestrationHandler) createCluster(filename string, fh *multipart.FileHeader, clusterName string,
	jsonData ClusterMetadata) interface{} {
	url := "http://" + h.MiddleendConf.Clm + "/v2/cluster-providers/" + clusterName + "/clusters"
	jsonLoad, _ := json.Marshal(jsonData)

	var fileNames []string
	fileNames = append(fileNames, filename)
	var fileContents []string

	status, err := h.apiPostMultipart(jsonLoad, fh, url, clusterName, fileNames, fileContents)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return status
	}
	log.Infof("cluster creation %s status: %d", clusterName, status)
	return nil

}

func (h *OrchestrationHandler) CheckConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parse_err := r.ParseMultipartForm(16777216)
	if parse_err != nil {
		log.Errorf("multipart error: %s", parse_err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var fh *multipart.FileHeader
	for _, v := range r.MultipartForm.File {
		fh = v[0]
	}
	file, err := fh.Open()
	if err != nil {
		log.Errorf("Failed to open the file: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read the kconfig
	kubeconfig, _ := ioutil.ReadAll(file)

	jsonData := ClusterMetadata{}
	jsn := ([]byte(r.FormValue("metadata")))
	err = json.Unmarshal(jsn, &jsonData)
	if err != nil {
		log.Errorf("Failed to parse json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("metadata %+v\n", jsonData)

	// RESTConfigFromKubeConfig is a convenience method to give back
	// a restconfig from your kubeconfig bytes.
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Errorf("Error while reading the kubeconfig: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid configuration: Cluster has no server defined\n"))
		return
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("Failed to create clientset: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid configuration: Cluster has no server defined\n"))
		return
	}

	_, err = clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed to establish the connection: %s", err.Error())
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Cluster connectivity failed x509 certificate signed by unknown authority\n"))
		return
	}
	log.Infof("Successfully established the connection")
	h.client = http.Client{}
	h.InitializeResponseMap()
	status := h.createCluster(fh.Filename, fh, vars["cluster-provider-name"], jsonData)
	if status != nil {
		w.WriteHeader(status.(int))
		w.Write(h.response.payload[vars["cluster-provider-name"]])
		log.WithFields(log.Fields{
			"Statuscode": status,
			"status":     h.response.status_msg,
		}).Error(h.response.status_msg)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(h.response.payload[vars["cluster-provider-name"]])

	// Checking whether Monitor & Istio Service is Present or Not
	AppName_mon, retCode_mon, retval_mon := h.GetCompositeAppData("MonitorApp", "amcop-system", "", "")
	AppName_isops, retCode_isops, retval_isops := h.GetCompositeAppData("IstioOperatorApp", "amcop-system", "", "")
	AppName_isprofile, retCode_isprofile, retval_isprofile := h.GetCompositeAppData("IstioProfileApp", "amcop-system", "", "")

	// Creating the Logical Cloud
	if retval_mon == "created" || retval_isops == "created" {
                log.Infof("Creating the LC for Apps")
                lc_result := h.DeployLogicalCloud(w, vars["cluster-provider-name"], jsonData)
                if !lc_result {
                      log.Error("Logical Cloud Creation Failed..")
                }
        }else {
                log.Info("Monitor App & Istio App is not created by amcop-operator skipping LC Creation..")
        }

        // Creating the Monitor Service DIG
        if retCode_mon == http.StatusOK && retval_mon == "created"  {
                log.Infof("Creating the DIG for App: %s", AppName_mon)
                mon_result := h.DeployMonitorService(w, AppName_mon, "amcop-system", vars["cluster-provider-name"], jsonData)
                if !mon_result {
                        log.Error("Monitor Service Orchestration Failed..")
                }
        }else {
                log.Info("Monitor App is not created by amcop-operator..skipping")
        }
        // Creating the Istio Operator Service DIG
	if (retCode_isops == http.StatusOK && retval_isops == "created") {
                log.Infof("Creating the DIG for App: %s", AppName_isops)
                isops_result := h.DeployIstioOperator(w, AppName_isops, "amcop-system", vars["cluster-provider-name"], jsonData)
                if !isops_result {
                      log.Error("Istio Operator Service Orchestration Failed..")
                }
        } else {
		log.Info("Istio Operator App is not created by amcop-operator..skipping")
	}

        // Creating the Istio Profile Service DIG
	if retCode_isprofile == http.StatusOK && retval_isprofile == "created"  {
                log.Infof("Creating the DIG for App: %s", AppName_isprofile)
                isprofile_result := h.DeployIstioProfile(w, AppName_isprofile, "amcop-system", vars["cluster-provider-name"], jsonData)
                if !isprofile_result {
                        log.Error("Istio Profile Service Orchestration Failed..")
                }
        }else {
                log.Info("Istio Profile App is not created by amcop-operator..skipping")
        }

}

// CreateDraftCompositeApp: Creates checkout copy of given composite application
// POST middleend/projects/<projectName>/composite-apps/<compositeAppName>/v1/checkout
func (h *OrchestrationHandler) CreateDraftCompositeApp(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	version := h.Vars["version"]
	h.InitializeResponseMap()

	retCode, latestVersion := h.FetchLatestVersion()
	if retCode != http.StatusOK {
		log.Errorf("Encountered error while fetching latest version")
		w.WriteHeader(retCode)
		return
	}

	// Checkout of a given composite application is only permitted, if it is the latest version
	if latestVersion != version {
		log.Errorf("Checkout of composite application should be for latest version")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.treeFilter = nil
	h.Vars["multipart"] = "true"

	dataPoints := []string{"projectHandler", "compAppHandler", "ProfileHandler"}
	h.prepTreeReq()
	h.dataRead = &ProjectTree{}
	h.CompositeAppReturnJSON = []CompositeAppsInProject{}
	h.CompositeAppReturnJSONShrunk = []CompositeAppsInProjectShrunk{}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Gives original copy of composite application
	h.copyCompositeAppTree("depthAll")
	log.Debugf("jsonresponse: %+v", h.CompositeAppReturnJSON)

	// The logic below creates draft version of composite application, which will be stored in
	// middleend collection of mco database, for processing by GUI
	var key DraftCompositeAppKey
	for index, comApp := range h.CompositeAppReturnJSON {
		version := strings.SplitAfter(version, "v")
		newversion, err := strconv.Atoi(version[1])
		if err != nil {
			log.Errorf("Encountered error while processing composite app version: %s", err)
			return
		}

		newversion += 1
		h.CompositeAppReturnJSON[index].Spec.Version = "v" + strconv.Itoa(newversion)
		h.CompositeAppReturnJSON[index].Status = "checkout"

		// Construct the composite key to select the entry
		key = DraftCompositeAppKey{
			Cname:    comApp.Metadata.Name,
			Cversion: h.CompositeAppReturnJSON[index].Spec.Version,
			Project:  h.Vars["projectName"],
		}
		log.Infof("Updated composite app version: %s", h.CompositeAppReturnJSON[index].Spec.Version)

		// Check if composite application for given version already exists
		log.Debugf("DraftCompositeAppKey: %s", key)
		retval, err := h.GetDraftCompositeApplication(key, "")
		if err != nil {
			log.Errorf("Encountered error while fetching composite app from middleend collection: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(retval) > 0 {
			log.Infof("Draft Composite application already exists")
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	err := db.DBconn.Insert(h.MiddleendConf.StoreName, key, nil, "appmetadata", h.CompositeAppReturnJSON[0])
	if err != nil {
		log.Errorf("Encountered error during checkout of composite app: %s", err)
		return
	}
	retval, _ := json.Marshal(h.CompositeAppReturnJSON[0])
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(retval)
}

// GetSvcVersions fetches the list of versions for a given composite application
// GET middleend/projects/<projectName>/composite-apps/<compositeAppName>/versions
func (h *OrchestrationHandler) GetSvcVersions(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()

	filter := r.URL.Query().Get("state")

	retCode, versionList := h.GetCompAppVersions(filter)
	log.Infof("versionList: %s", versionList)
	if retCode != http.StatusOK {
		w.WriteHeader(retCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	retval, _ := json.Marshal(versionList)
	w.Write(retval)
}

func (h *OrchestrationHandler) GetCompAppVersions(filter string) (int, []string) {
	var versionList []string
	compAppName := h.Vars["compositeAppName"]
	h.Vars["compositeAppName"] = ""
	retCode, retval := h.GetCompApps("", "")
	if retCode != http.StatusOK {
		log.Errorf("Encountered error while fetching composite apps")
		return http.StatusInternalServerError, versionList
	}

	var compArray []CompositeAppsInProjectShrunk
	json.Unmarshal(retval, &compArray)

	log.Infof("composite:%v", compArray)

	for _, comApp := range compArray {
		if comApp.Metadata.Name == compAppName {
			for _, spec := range comApp.Spec {
				if filter != "" && filter == spec.Status {
					versionList = append(versionList, spec.Version)
				}

				if filter == "" {
					versionList = append(versionList, spec.Version)
				}
			}
			break
		}
	}
	h.Vars["compositeAppName"] = compAppName
	return http.StatusOK, versionList
}

// UpdateCompositeApp Updates an existing composite application
// POST /middleend/projects/<projectName>/composite-apps/<compositeAppName>/<version>/app
func (h *OrchestrationHandler) UpdateCompositeApp(w http.ResponseWriter, r *http.Request) {
	var jsonData appsData
	var newApp Application
	var newProfile ProfileMeta
	h.InitializeResponseMap()

	err := r.ParseMultipartForm(16777216)
	if err != nil {
		log.Errorf("Failed to parse multi part: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	vars := mux.Vars(r)
	jsn := []byte(r.FormValue("appsPayload"))
	err = json.Unmarshal(jsn, &jsonData)

	h.file = make(map[string]*multipart.FileHeader)
	for _, v := range r.MultipartForm.File {
		fh := v[0]
		h.file[fh.Filename] = fh
	}

	if err != nil {
		log.Errorf("Failed to parse json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if h.file[jsonData.Metadata.FileName] == nil {
		t := fmt.Sprintf("File %s not in request", jsonData.Metadata.FileName)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(t))
		log.Error("app file not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if h.file[jsonData.ProfileMetadata.FileName] == nil {
		t := fmt.Sprintf("File %s not in request", jsonData.ProfileMetadata.FileName)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(t))
		log.Error("profile file not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	newApp.Metadata.Name = strings.TrimSpace(jsonData.Metadata.Name)
	newApp.Metadata.Description = jsonData.Metadata.Description
	// Open the file
	file, err := h.file[jsonData.Metadata.FileName].Open()
	if err != nil {
		log.Errorf("Encountered error while processing multipart file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Close the file later
	defer file.Close()

	// Copy the app helm chart to application struct
	var appBuff bytes.Buffer
	io.Copy(&appBuff, file)
	newApp.Metadata.ChartContent = base64.StdEncoding.EncodeToString(appBuff.Bytes())

	log.Debugf("newApp is : %s", newApp)

	newProfile.Metadata.Name = strings.TrimSpace(jsonData.ProfileMetadata.Name)
	// Open the file
	file, err = h.file[jsonData.ProfileMetadata.FileName].Open()
	if err != nil {
		log.Errorf("Encountered error while processing multipart file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Close the file later
	defer file.Close()

	// Copy the profile helm chart to profile struct
	var profileBuff bytes.Buffer
	io.Copy(&profileBuff, file)
	newProfile.Metadata.ChartContent = base64.StdEncoding.EncodeToString(profileBuff.Bytes())
	newProfile.Spec.AppName = newApp.Metadata.Name

	log.Debugf("newProfile is : %s", newProfile)
	operation := r.URL.Query().Get("operation")

	var dboperation string
	if operation != "" && operation == "updateApp" {
		dboperation = "UpdateApplication"
	} else {
		dboperation = "AddApplication"
	}
	err = db.DBconn.Update(h.MiddleendConf.StoreName, dboperation, vars, newApp.Metadata.Name, newApp)
	if err != nil {
		log.Errorf("Encountered error during update of composite app apps: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if operation != "" && operation == "updateApp" {
		dboperation = "UpdateProfile"
	} else {
		dboperation = "AddProfile"
	}

	err = db.DBconn.Update(h.MiddleendConf.StoreName, dboperation, vars, newApp.Metadata.Name, newProfile)
	if err != nil {
		log.Errorf("Encountered error during update of composite app profile: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	retval, err := json.Marshal(jsonData)
	w.Write(retval)
}

// RemoveApp: removes an existing application from composite app
// DELETE /projects/{projectName}/composite-apps/{compositeAppName}/{version}/apps/{appName}
func (h *OrchestrationHandler) RemoveApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.InitializeResponseMap()
	dboperations := []string{"DeleteApplication", "DeleteProfile"}
	for _, dboperation := range dboperations {
		err := db.DBconn.Update(h.MiddleendConf.StoreName, dboperation, vars, "", "")
		if err != nil {
			log.Errorf("Encountered error during removing app in composite app : %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateService: Creates all applications and uploaded profiles for a versioned composite
// application, fetching all data from middleend collection
// POST /projects/{projectName}/composite-apps/{compositeAppName}/{version}/update
func (h *OrchestrationHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()
	key := DraftCompositeAppKey{
		Cversion: h.Vars["version"],
		Cname:    h.Vars["compositeAppName"],
		Project:  h.Vars["projectName"],
	}

	caList, err := h.GetDraftCompositeApplication(key, "depthAll")
	if err != nil {
		log.Errorf("Encountered error while fetching composite app from middleend collection: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(caList) == 0 {
		log.Errorf("Draft composite application does not exists, hence service cannot be created")
		w.WriteHeader(500)
		return
	}

	ca := caList[0]

	var meta []appsData

	for _, app := range ca.Spec.AppsArray {
		appData := appsData{}
		appData.Metadata.FileName = app.Metadata.Name + ".tgz"
		appData.Metadata.Name = app.Metadata.Name
		appData.Metadata.Description = app.Metadata.Description
		ccBytes, err := base64.StdEncoding.DecodeString(app.Metadata.ChartContent)
		if err != nil {
			log.Errorf("Encountered error while decoding filecontent: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		appData.Metadata.FileContent = string(ccBytes)
		meta = append(meta, appData)
	}

	for _, profile := range ca.Spec.ProfileArray {
		for _, appprofile := range profile.Spec.ProfilesArray {
			for m, _ := range meta {
				if meta[m].Metadata.Name == appprofile.Spec.AppName {
					meta[m].ProfileMetadata.FileName = appprofile.Metadata.Name
					meta[m].ProfileMetadata.Name = appprofile.Metadata.Name
					ccBytes, err := base64.StdEncoding.DecodeString(appprofile.Metadata.ChartContent)
					if err != nil {
						log.Errorf("Encountered error while decoding filecontent: %s", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					meta[m].ProfileMetadata.FileContent = string(ccBytes)
				}
			}
		}
	}

	h.meta = meta
	h.client = http.Client{}

	// 1. create the composite application. the compAppHandler implements the
	// orchWorkflow interface.
	appHandler := &compAppHandler{}
	appHandler.orchInstance = h
	httpErr := createCompositeapp(appHandler)
	if httpErr != nil {
		h.rollBackApp()
		if intval, ok := httpErr.(int); ok {
			log.Errorf("CreateCompositeapp failed with error : %d", intval)
			w.WriteHeader(intval)
		} else {
			log.Infof("Encountered error for CreateCompositeapp")
			w.WriteHeader(http.StatusInternalServerError)
		}
		errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
		w.Write([]byte(errMsg))
		return
	}

	// 2. create the composite application profiles
	profileHandler := &ProfileHandler{}
	profileHandler.orchInstance = h
	httpErr = createProfile(profileHandler)
	if httpErr != nil {
		h.rollBackApp()
		if intval, ok := httpErr.(int); ok {
			log.Errorf("CreateProfile failed with error : %d", intval)
			w.WriteHeader(intval)
		} else {
			log.Errorf("Encountered error for CreateProfile")
			w.WriteHeader(http.StatusInternalServerError)
		}
		errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusCreated)
	// Delete draft composite application from middleend collection
	err = db.DBconn.Delete(h.MiddleendConf.StoreName, h.Vars)
	if err != nil {
		log.Errorf("Encountered error during delete of composite app from middleend collection: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(h.response.payload[h.Vars["compositeAppName"]+"_compapp"])
}

// GetDashboardData get count of total composite-apps, deployment-intent-groups and clusters
func (h *OrchestrationHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Vars = vars
	h.InitializeResponseMap()
	//create the Dashboard client
	dStore := &remoteStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore
	dashboardClient := DashboardClient{h}
	retData, retcode := dashboardClient.getDashboardData()
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			log.Infof("Failed to get dashboard data : %d", intval)
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
			w.Write([]byte(errMsg))
		}
		return
	}

	var retval []byte
	retval, err := json.Marshal(retData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(retval)
}

// GetClusters get an a array of all the cluster providers and the clusters within them
func (h *OrchestrationHandler) GetClusters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Vars = vars
	h.InitializeResponseMap()
	dashboardClient := DashboardClient{h}
	retcode := dashboardClient.getAllClusters()
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			log.Infof("Failed to get clusterdata : %d", intval)
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
			w.Write([]byte(errMsg))
		}
		return
	}

	var retval []byte
	retval, err := json.Marshal(h.ClusterProviders)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(retval)
}

func (h *OrchestrationHandler) DeployIstioOperator(w http.ResponseWriter, appname string, namespace string, ClusterName string, jsonData ClusterMetadata) bool {

        // Final Result ByDefault Considered True
        var Result bool = true

        // Variable for DIG
        var appdata appsData
        var ClusterInfo ClusterInfo
        var SelectedCluster SelectedCluster

        h.InitializeResponseMap()

	// Creating the Service Instance for Monitor App
        log.Info("Creating the Service Instance for Istio-Operator App")
        appdata.Metadata.Name = "istio-operator"
        appdata.Metadata.Description = "Service Instance for Istio-Operator App"
        appdata.PlacementCriterion = "allOf"

        // CLuster Provider Metadata
        ClusterInfo.Provider = ClusterName

        // Target Cluster Metadata
        SelectedCluster.Name = jsonData.Metadata.Name

        // Array for clusterInfo and AppData
        ClusterInfo.SelectedClusters = append(ClusterInfo.SelectedClusters, SelectedCluster)
        appdata.Clusters = append(appdata.Clusters, ClusterInfo)


        // Initializing the DIG Struct with Payload
        dig := deployDigData{
                Name:             "operator-IstioOperator-" + jsonData.Metadata.Name,
                Description:      "operator-IstioOperator-" + jsonData.Metadata.Name,
                CompositeAppName: appname,
                CompositeProfile: "IstioOperatorApp_profile",
                CompositeAppVersion: "v1",
                DigVersion: "v1",
                LogicalCloud: "operator-logical-cloud-" + jsonData.Metadata.Name,
                Spec: DigSpec{
                        ProjectName: namespace,
                        Apps:        []appsData{appdata},
                        OverrideValuesObj: []localstore.OverrideValues{},
                },
        }

        h.DigData = dig
        log.Debugf("digData: %+v", dig)

	if len(h.DigData.Spec.Apps) == 0 {
                log.Errorf("Bad request, no app metadata\n with code:%d", http.StatusBadRequest)
                Result = false
        }
        h.DigData.NwIntents = false
        // Creating the DIG
        h.createDigData(&w, "emco")
        h.AddDIGInfo()

	if h.MiddleendConf.AppInstantiate {
                // Approve the service Instance for Monitor App
                resp_ap := h.DIGApprove(namespace, h.DigData.CompositeAppName, h.DigData.Name)
                if resp_ap != nil {
                        log.Errorf("Failed to Approve Service Instance for Istio Operator App: %d", resp_ap)
                        Result = false
                }

                // Instantiate the Service Instance for Monitor App
                time.Sleep(50 * time.Millisecond)
                resp_ins := h.DIGInstantiate(namespace, h.DigData.CompositeAppName, h.DigData.Name)
                if resp_ins != nil {
                        log.Errorf("Failed to Instantiate Service Instance for Istio Operator App: %d", resp_ins)
                        Result = false
            }
	}
        return Result
}
func (h *OrchestrationHandler) DeployIstioProfile(w http.ResponseWriter, appname string, namespace string, ClusterName string, jsonData ClusterMetadata) bool {

        // Final Result ByDefault Considered True
        var Result bool = true

        // Variable for DIG
        var appdata appsData
        var ClusterInfo ClusterInfo
        var SelectedCluster SelectedCluster

        h.InitializeResponseMap()

        // Creating the Service Instance for Monitor App
        log.Info("Creating the Service Instance for Istio-Profile App")
        appdata.Metadata.Name = "istio-profile"
        appdata.Metadata.Description = "Service Instance for Istio-Profile App"
        appdata.PlacementCriterion = "allOf"

        // CLuster Provider Metadata
        ClusterInfo.Provider = ClusterName

        // Target Cluster Metadata
        SelectedCluster.Name = jsonData.Metadata.Name

        // Array for clusterInfo and AppData
        ClusterInfo.SelectedClusters = append(ClusterInfo.SelectedClusters, SelectedCluster)
        appdata.Clusters = append(appdata.Clusters, ClusterInfo)


        // Initializing the DIG Struct with Payload
        dig := deployDigData{
                Name:             "operator-IstioProfile-" + jsonData.Metadata.Name,
                Description:      "operator-IstioProfile-" + jsonData.Metadata.Name,
                CompositeAppName: appname,
                CompositeProfile: "IstioProfileApp_profile",
                CompositeAppVersion: "v1",
                DigVersion: "v1",
                LogicalCloud: "operator-logical-cloud-" + jsonData.Metadata.Name,
                Spec: DigSpec{
                        ProjectName: namespace,
                        Apps:        []appsData{appdata},
                        OverrideValuesObj: []localstore.OverrideValues{},
                },
        }

	h.DigData = dig
        log.Debugf("digData: %+v", dig)

        if len(h.DigData.Spec.Apps) == 0 {
                log.Errorf("Bad request, no app metadata\n with code:%d", http.StatusBadRequest)
                Result = false
        }
        h.DigData.NwIntents = false
        // Creating the DIG
        h.createDigData(&w, "emco")
        h.AddDIGInfo()
	if h.MiddleendConf.AppInstantiate {
		// Approve the service Instance for Monitor App
                resp_ap := h.DIGApprove(namespace, h.DigData.CompositeAppName, h.DigData.Name)
                if resp_ap != nil {
                        log.Errorf("Failed to Approve Service Instance for Istio Profile App: %d", resp_ap)
                        Result = false
                }

                // Instantiate the Service Instance for Monitor App
                time.Sleep(50 * time.Millisecond)
                resp_ins := h.DIGInstantiate(namespace, h.DigData.CompositeAppName, h.DigData.Name)
                if resp_ins != nil {
                        log.Errorf("Failed to Instantiate Service Instance for Istio Profile App: %d", resp_ins)
                        Result = false
                }
	}


        return Result
}

func (h *OrchestrationHandler) DeployMonitorService(w http.ResponseWriter, appname string, namespace string, ClusterName string, jsonData ClusterMetadata) bool {

	// Final Result ByDefault Considered True
	var Result bool = true

	// Variable for DIG
	var appdata appsData
        var ClusterInfo ClusterInfo
        var SelectedCluster SelectedCluster

	h.InitializeResponseMap()

	// Creating the Service Instance for Monitor App
	log.Info("Creating the Service Instance for Monitor App")
	appdata.Metadata.Name = "monitor"
	appdata.Metadata.Description = "Service Instance for Monitor App"
	appdata.PlacementCriterion = "allOf"

	// CLuster Provider Metadata
	ClusterInfo.Provider = ClusterName

	// Target Cluster Metadata
	SelectedCluster.Name = jsonData.Metadata.Name

	// Array for clusterInfo and AppData
	ClusterInfo.SelectedClusters = append(ClusterInfo.SelectedClusters, SelectedCluster)
	appdata.Clusters = append(appdata.Clusters, ClusterInfo)


	// Initializing the DIG Struct with Payload
	dig := deployDigData{
		Name:             "operator-Monitor-" + jsonData.Metadata.Name,
		Description:      "operator-Monitor-" + jsonData.Metadata.Name,
		CompositeAppName: appname,
		CompositeProfile: "MonitorApp_profile",
		CompositeAppVersion: "v1",
		DigVersion: "v1",
		LogicalCloud: "operator-logical-cloud-" + jsonData.Metadata.Name,
		Spec: DigSpec{
			ProjectName: namespace,
			Apps:        []appsData{appdata},
			OverrideValuesObj: []localstore.OverrideValues{},
		},
	}

	h.DigData = dig
	log.Debugf("digData: %+v", dig)

        if len(h.DigData.Spec.Apps) == 0 {
		log.Errorf("Bad request, no app metadata\n with code:%d", http.StatusBadRequest)
		Result = false
        }
        h.DigData.NwIntents = false
	// Creating the DIG
        h.createDigData(&w, "emco")
        h.AddDIGInfo()
	log.Infof("state %s", h.MiddleendConf.AppInstantiate)
	if h.MiddleendConf.AppInstantiate {
		// Approve the service Instance for Monitor App
                resp_ap := h.DIGApprove(namespace, h.DigData.CompositeAppName, h.DigData.Name)
                if resp_ap != nil {
                        log.Errorf("Failed to Approve Service Instance for Monitor Agent App: %d", resp_ap)
                        Result = false
                }

                // Instantiate the Service Instance for Monitor App
                time.Sleep(50 * time.Millisecond)
                resp_ins := h.DIGInstantiate(namespace, h.DigData.CompositeAppName, h.DigData.Name)
                if resp_ins != nil {
                        log.Errorf("Failed to Instantiate Service Instance for Monitor Agent App: %d", resp_ins)
                        Result = false
                }
	}
	return Result

}

// GetClusterNetworks get an a array of all the cluster networks along with their rsync status
func (h *OrchestrationHandler) GetClusterNetworks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Vars = vars
	h.InitializeResponseMap()
	nwhandler := ncmHandler{}
	nwhandler.orchInstance = h
	respdata, retcode := nwhandler.getNetworks()
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			log.Infof("Failed to get cluster networks : %d", intval)
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
			w.Write([]byte(errMsg))
		}
		return
	}

	var retval []byte
	retval, err := json.Marshal(respdata)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(retval)
}

// SaveAppIntentsLocalStore
func (h *OrchestrationHandler) DigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Get the query filter
	var jsonData appsData
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()

	// Implementation using multipart form and set maxSize 16MB
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.file = make(map[string]*multipart.FileHeader)
	for _, v := range r.MultipartForm.File {
		fh := v[0]
		h.file[fh.Filename] = fh
	}

	jsn := []byte(r.FormValue("metadata"))
	err = json.Unmarshal(jsn, &jsonData)
	if err != nil {
		log.Error("Failed to parse json: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// FIXME
	tempDigData := deployDigData{}
	tempDigData.Name = h.Vars["deploymentIntentGroupName"]
	tempDigData.Spec.Apps = append(tempDigData.Spec.Apps, jsonData)

	h.DigData = tempDigData
	h.DigData.CompositeAppName = h.Vars["compositeAppName"]


	filter := r.URL.Query().Get("operation")
	if filter == "save" {
		bstore := &localStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
		intentHandler := &placementIntentHandler{}
		intentHandler.orchInstance = h
		h.Vars["update-intent"] = "yes"
		intentStatus := intentHandler.createObject()
		if intentStatus != nil {
			if intval, ok := intentStatus.(int); ok {
				w.WriteHeader(intval)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(h.response.payload[h.Vars["compositeAppName"]+"_gpint"])
			return
		}

		// If the metadata contains network interface request then call the
		// network intent related part of the workflow.
		h.DigData.NwIntents = true // FIXME
		if h.DigData.NwIntents {
			nwHandler := &networkIntentHandler{}
			nwHandler.orchInstance = h
			nwIntentStatus := nwHandler.createObject()
			if nwIntentStatus != nil {
				if intval, ok := nwIntentStatus.(int); ok {
					w.WriteHeader(intval)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				w.Write(h.response.payload[h.Vars["compositeAppName"]+"_nwctlint"])
				return
			}
		}

		// If the metadata contains genericK8sIntent info, process the same
		// Validate and process resource data
		if !h.processResourceData(&w, r) {
			log.Errorf("Unable to process resource data: %s", h.DigData.Name)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.createUpdateK8sResource(&w, "")
	}
}

//CreateDig CreateDig exported function which creates deployment intent group
func (h *OrchestrationHandler) CreateDig(w http.ResponseWriter, r *http.Request) {
	var jsonData deployDigData

	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()

	// Implementation using multipart form and set maxSize 16MB
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.file = make(map[string]*multipart.FileHeader)
	for _, v := range r.MultipartForm.File {
		fh := v[0]
		h.file[fh.Filename] = fh
	}

	jsn := []byte(r.FormValue("metadata"))
	err = json.Unmarshal(jsn, &jsonData)
	if err != nil {
		log.Error("Failed to parse json: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// If override data is empty then add some dummy override data.
	if len(jsonData.Spec.OverrideValuesObj) == 0 {
		o := localstore.OverrideValues{}
		v := make(map[string]string)
		o.AppName = jsonData.Spec.Apps[0].Metadata.Name
		v["key"] = "value"
		o.ValuesObj = v
		jsonData.Spec.OverrideValuesObj = append(jsonData.Spec.OverrideValuesObj, o)
	}

	h.DigData = jsonData
	log.Debugf("digData: %+v", jsonData)

	if len(h.DigData.Spec.Apps) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request, no app metadata\n"))
		return
	}
	h.DigData.NwIntents = false

	h.client = http.Client{}

	// Validate and process resource data
	if !h.processResourceData(&w, r) {
		log.Errorf("Unable to process resource data part of DIG: %s", h.DigData.Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.createDigData(&w, "emco")

	h.AddDIGInfo()
	w.Write(h.response.payload[h.DigData.Name])
}
