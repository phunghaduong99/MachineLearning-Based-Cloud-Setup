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
	"encoding/json"
	"net/http"
	"strings"

	"example.com/middleend/db"
	"example.com/middleend/localstore"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const DIG_INFO_COLLECTION = "diginfo"

type DigInfo struct {
	DigName     string   `json:"name"`
	VersionList []string `json:"versionList"`
}

type DigInfoKey struct {
	DigName string `json:"name"`
}

type IgpIntents struct {
	Metadata apiMetaData `json:"metadata"`
	Spec     AppIntents  `json:"spec"`
}

type AppIntents struct {
	Intent map[string]string `json:"intent"`
}

type DigpIntents struct {
	Intent []DigDeployedIntents `json:"intent"`
}
type DigDeployedIntents struct {
	GenericPlacementIntent string `json:"genericPlacementIntent"`
	Ovnaction              string `json:"ovnaction"`
}

// digpHandler implements the orchworkflow interface
type digpHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

// localStoreIntentHandler implements the orchworkflow interface
type localStoreDigHandler struct {
	orchInstance *OrchestrationHandler
}
type remoteStoreDigHandler struct {
	orchInstance *OrchestrationHandler
}

// Interface to creating the backend objects
// either in EMCO over REST or in middleend mongo
type digBackendStore interface {
	createDig(localstore.DeploymentIntentGroup, string, string, string) (interface{}, interface{})
	deleteDig(string, string, string, string) (interface{}, interface{})
	createIntents(localstore.Intent, string, string, string, string) (interface{}, interface{})
	deleteIntents(string, string, string, string, string) (interface{}, interface{})
	getDig(project string, compositeAppName string, version string, digName string) (interface{}, []byte, interface{})
	getAllDig(project string, compositeAppName string, version string) (interface{}, []byte, interface{})
	getIntents(project string, compositeAppName string, version string,
		digName string) (interface{}, []byte, interface{})
	getStatus(compositeAppName string, compositeAppVersion string, digName string) (digStatus, int)
}

func (h *localStoreDigHandler) getDig(project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {
	c := localstore.NewDeploymentIntentGroupClient()
	dig, err := c.GetDeploymentIntentGroup(digName, project, compositeAppName, version)
	log.Infof("Get Dig localStore in Composite app %s dig %s status: %s : value %s", compositeAppName,
		digName, err, dig)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, nil, err
		} else {
			return http.StatusNotFound, nil, err
		}
	}
	retval, _ := json.Marshal(dig)
	return http.StatusOK, retval, err
}

func (h *remoteStoreDigHandler) getDig(project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {

	orch := h.orchInstance
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		project + "/composite-apps/" + compositeAppName +
		"/" + version +
		"/deployment-intent-groups/" + digName
	retcode, retval, err := orch.apiGet(orchURL, compositeAppName+"_getdig")
	log.Infof("Get Dig in Composite app %s dig %s status: %d", compositeAppName,
		digName, retcode)
	return retcode, retval, err
}

func (h *localStoreDigHandler) getAllDig(project string, compositeAppName string, version string,
) (interface{}, []byte, interface{}) {

	c := localstore.NewDeploymentIntentGroupClient()
	gPIntent, err := c.GetAllDeploymentIntentGroups(project, compositeAppName, version)
	log.Infof("Get All DIG localStore in Composite app %s version %s status: %s", compositeAppName,
		version, err)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find") {
			return http.StatusNotFound, nil, err
		} else if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, nil, err
		} else {
			return http.StatusInternalServerError, nil, err
		}
	}
	retval, _ := json.Marshal(gPIntent)
	return http.StatusOK, retval, err
}

func (h *remoteStoreDigHandler) getAllDig(project string, compositeAppName string, version string,
) (interface{}, []byte, interface{}) {

	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		project + "/composite-apps/" + compositeAppName +
		"/" + version +
		"/deployment-intent-groups"
	retcode, retval, err := orch.apiGet(orchURL, compositeAppName+"_getdig")
	log.Infof("Get ALl Dig in Composite app %s version %s status: %d", compositeAppName, version,
		retcode)
	return retcode, retval, err
}

func (h *remoteStoreDigHandler) getIntents(project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {

	orch := h.orchInstance

	url := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		project + "/composite-apps/" + compositeAppName +
		"/" + version + "/deployment-intent-groups/" + digName + "/intents"
	retcode, retval, err := orch.apiGet(url, compositeAppName+"_getappPint")
	return retcode, retval, err
}

func (h *localStoreDigHandler) getIntents(project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {

	c := localstore.NewIntentClient()
	appIntent, err := c.GetAllIntents(project, compositeAppName, version, digName)
	log.Infof("Get All Intents localStore in Composite app %s version %s appIntent: %s", compositeAppName,
		version, appIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, nil, err
		} else {
			return http.StatusInternalServerError, nil, err
		}
	}
	retval, _ := json.Marshal(appIntent)
	return http.StatusOK, retval, err
}

func (h *localStoreDigHandler) createDig(g localstore.DeploymentIntentGroup, p string, ca string,
	v string) (interface{}, interface{}) {
	c := localstore.NewDeploymentIntentGroupClient()
	g.Spec.IsCheckedOut = true

	_, createErr := c.CreateDeploymentIntentGroup(g, p, ca, v)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "DeploymentIntent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}

	return http.StatusCreated, createErr
}

func (h *remoteStoreDigHandler) createDig(g localstore.DeploymentIntentGroup, p string, ca string,
	v string) (interface{}, interface{}) {
	orch := h.orchInstance
	digName := orch.DigData.Name
	jsonLoad, _ := json.Marshal(g)
	url := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups"
	resp, err := orch.apiPost(jsonLoad, url, digName)
	return resp, err
}
func (h *localStoreDigHandler) deleteDig(digName string, p string, ca string,
	v string) (interface{}, interface{}) {
	c := localstore.NewDeploymentIntentGroupClient()

	err := c.DeleteDeploymentIntentGroup(digName, p, ca, v)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Error getting appcontext") {
			return http.StatusNotFound, err
		} else if strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound, err
		} else if strings.Contains(err.Error(), "conflict") {
			return http.StatusConflict, err
		} else {
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusNoContent, err
}

func (h *remoteStoreDigHandler) deleteDig(digName string, p string, ca string,
	v string) (interface{}, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	resp, err := orch.apiDel(url, digName)
	return resp, err
}

func (h *localStoreDigHandler) createIntents(i localstore.Intent, p string, ca string, v string,
	digName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewIntentClient()
	_, createErr := c.AddIntent(i, p, ca, v, digName)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the intent") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Intent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}
	return http.StatusCreated, createErr
}

func (h *remoteStoreDigHandler) createIntents(i localstore.Intent, p string, ca string, v string,
	digName string) (interface{}, interface{}) {

	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(i)
	url := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/intents"
	status, err := orch.apiPost(jsonLoad, url, "PlacementIntent")
	return status, err
}
func (h *localStoreDigHandler) deleteIntents(i string, p string, ca string, v string,
	digName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewIntentClient()
	err := c.DeleteIntent(i, p, ca, v, digName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound, err
		} else if strings.Contains(err.Error(), "conflict") {
			return http.StatusConflict, err
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusNoContent, err
}

func (h *remoteStoreDigHandler) deleteIntents(i string, p string, ca string, v string,
	digName string) (interface{}, interface{}) {

	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/intents/PlacementIntent"
	status, err := orch.apiDel(url, digName)
	return status, err
}

func (h *remoteStoreDigHandler) getStatus(compositeAppName string, compositeAppVersion string, digName string) (digStatus, int) {
	orch := h.orchInstance
	vars := orch.Vars
	thisDigStatus := digStatus{}
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		vars["projectName"] + "/composite-apps/" + compositeAppName +
		"/" + compositeAppVersion +
		"/deployment-intent-groups/" + digName + "/status"
	thisDigStatus, httpStatus := h.getDigStatus(orchURL, vars["compositeAppName"]+"_digpStatus", [][]string{{"type", "cluster"}})

	// If clusterStatus is present Copy clusterStatus to RsyncStatus
	// When monitoring agent is not available, clusterStatus will be empty
	// Retry for rsyncStatus if clusterStatus is empty
	if httpStatus == http.StatusOK && thisDigStatus.Apps != nil && len(thisDigStatus.Apps) >  0 {
		for i := range thisDigStatus.Apps {
			for j := range thisDigStatus.Apps[i].Clusters {
				for k, resource := range thisDigStatus.Apps[i].Clusters[j].Resources {
					thisDigStatus.Apps[i].Clusters[j].Resources[k].RsyncStatus = resource.ClusterStatus
				}
			}
		}
	} else {
		thisDigStatus, httpStatus = h.getDigStatus(orchURL, vars["compositeAppName"]+"_digpStatus", [][]string{})
	}
	if httpStatus != http.StatusOK {
		return thisDigStatus, httpStatus
	}
	localDigStore := localStoreDigHandler{}
	localDigRetCode, _, _ := localDigStore.getDig(vars["projectName"],
		compositeAppName, compositeAppVersion, digName)
	if localDigRetCode == http.StatusOK {
		thisDigStatus.IsCheckedOut = true
	} else {
		thisDigStatus.IsCheckedOut = false
	}
	return thisDigStatus, http.StatusOK
}

func (h *remoteStoreDigHandler) getDigStatus(url string, statusKey string, arguments [][]string) (digStatus, int) {
	retCode, retVal, err := h.orchInstance.apiGetWithArguments(url, statusKey, arguments)
	if err != nil || retCode != http.StatusOK {
		log.Errorf("Failed to read DIG status. Err: %v, ReturnCode: %v", err, retCode)
		//Changing error code to StatusInternalServerError after logging actual response code from the backend
		return digStatus{}, http.StatusInternalServerError
	}
	status := digStatus{}
	err = json.Unmarshal(retVal, &status)
	if err != nil {
		log.Errorf("Failed to parse DIG status: %v", err)
		return digStatus{}, http.StatusInternalServerError
	}
	return status, http.StatusOK
}

func (h *localStoreDigHandler) getStatus(compositeAppName string, compositeAppVersion string, digName string) (digStatus, int) {
	thisDigStatus := digStatus{}
	actions := digActions{}
	actions.State = "checkedOut"
	thisDigStatus.States.Actions = append(thisDigStatus.States.Actions, actions)
	return thisDigStatus, http.StatusOK
}

func (h *digpHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	retcode := http.StatusOK
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		var digpList []localstore.DeploymentIntentGroup
		// This for the cases where the dig name is in the URL
		if orch.treeFilter != nil && orch.treeFilter.digName != "" {
			temp := localstore.DeploymentIntentGroup{}
			retcode, retval, err := orch.digStore.getDig(vars["projectName"],
				compositeAppMetadata.Name, compositeAppSpec.Version,
				orch.treeFilter.digName)
			log.Infof("Get Digp in composite app %s status: %d", compositeAppMetadata.Name, retcode)
			if retcode != http.StatusOK {
				log.Errorf("Failed to read digp")
				return nil, retcode
			}
			if err != nil {
				log.Errorf("Failed to read digp")
				return nil, retcode
			}
			json.Unmarshal(retval, &temp)
			digpList = append(digpList, temp)
		} else {
			retcode, retval, err := orch.digStore.getAllDig(vars["projectName"],
				compositeAppMetadata.Name, compositeAppSpec.Version)
			log.Infof("Get Digp in composite app %s status: %d", compositeAppMetadata.Name, retcode)
			if err != nil {
				log.Errorf("Failed to read digp")
				return nil, retcode
			}
			if retcode != http.StatusOK {
				log.Errorf("Failed to read digp")
				return nil, retcode
			}
			json.Unmarshal(retval, &digpList)
		}

		compositeAppValue.DigMap = make(map[string]*DigReadData, len(digpList))
		for k := range digpList {
			var Dig DigReadData
			// Get the DIG detailed status
			thisDigStatus, retcode := orch.digStore.getStatus(compositeAppValue.Metadata.Metadata.Name,
				compositeAppValue.Metadata.Spec.Version,
				digpList[k].MetaData.Name)
			if retcode != http.StatusOK {
				log.Errorf("Failed to read digp")
				//return nil, retcode
				continue
			}
			// Fetch the lastest state and populate the digpValue
			state := thisDigStatus.States.Actions[len(thisDigStatus.States.Actions)-1].State
			digpList[k].Spec.Status = state
			log.Debugf("DIG checkout state %s: %+v", digpList[k].MetaData.Name, thisDigStatus.IsCheckedOut)
			digpList[k].Spec.IsCheckedOut = thisDigStatus.IsCheckedOut
			Dig.DigpData = digpList[k]
			compositeAppValue.DigMap[digpList[k].MetaData.Name] = &Dig
		}
	}
	return nil, retcode
}

func (h *digpHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			vars["projectName"] + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version +
			"/deployment-intent-groups/"
		digpList := compositeAppValue.DigMap
		for digName, digValue := range digpList {
			retcode, retval, err := orch.digStore.getIntents(vars["projectName"], compositeAppMetadata.Name,
				compositeAppSpec.Version, digName)
			log.Infof("Get Dig int composite app %s Dig %s status %d \n", vars["compositeAppName"],
				digName, retcode)
			if err != nil {
				log.Errorf("Failed to read digp intents")
				return nil, retcode.(int)
			}
			if retcode != http.StatusOK {
				log.Errorf("Failed to read digp intents")
				return nil, retcode.(int)

			}
			err = json.Unmarshal(retval, &digValue.DigIntentsData)
			if err != nil {
				log.Errorf("Failed to read intents %s\n", err)
			}
		}
	}
	return nil, http.StatusOK
}

func (h *digpHandler) deleteObject() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		digpList := compositeAppValue.DigMap

		for digName, _ := range digpList {
			url := h.orchURL + digName + "/intents/PlacementIntent"
			log.Infof("delete intents %s", url)
			resp, err := orch.digStore.deleteIntents("PlacementIntent", vars["projectName"],
				compositeAppMetadata.Name, compositeAppSpec.Version, digName)
			if err != nil {
				return err //need to add the retcode
			}
			if resp != http.StatusNoContent {
				return resp
			}
			url = h.orchURL + digName + "/intents/GenericAction"
			log.Infof("delete intents %s", url)
			resp, err = orch.digStore.deleteIntents("GenericAction", vars["projectName"],
				compositeAppMetadata.Name, compositeAppSpec.Version, digName)
			log.Infof("Delete dig intents response: %d", resp)
			if err != nil {
				return err
			}
			if resp != http.StatusNoContent {
				return resp
			}
		}
	}
	return nil
}

func (h *digpHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		digpList := compositeAppValue.DigMap

		// loop through all the intents in the dig
		for digName := range digpList {
			url := h.orchURL + digName
			log.Infof("delete intents %s", url)
			resp, err := orch.digStore.deleteDig(digName, vars["projectName"], compositeAppMetadata.Name, compositeAppSpec.Version)
			if err != nil {
				return err //need to add the retcode
			}
			if resp != http.StatusNoContent {
				return resp
			}
			log.Infof("Delete dig response: %d", resp)
		}
	}
	return nil
}

func (h *digpHandler) createAnchor() interface{} {
	digData := h.orchInstance.DigData
	orch := h.orchInstance
	vars := orch.Vars

	var customData string
	if len(orch.Vars["operation"]) > 0 {
		customData = orch.Vars["operation"]
	} else {
		customData = "data1"
	}

	var originalVersion string
	if len(orch.Vars["originalversion"]) > 0 {
		originalVersion = orch.Vars["originalversion"]
	}
	digp := localstore.DeploymentIntentGroup{
		MetaData: localstore.DepMetaData{
			Name:        digData.Name,
			Description: digData.Description,
			UserData1:   customData,
			UserData2:   originalVersion},
		Spec: localstore.DepSpecData{
			Profile:           digData.CompositeProfile,
			Version:           digData.DigVersion,
			LogicalCloud:      digData.LogicalCloud,
			OverrideValuesObj: digData.Spec.OverrideValuesObj,
		},
	}

	resp, err := orch.digStore.createDig(digp, vars["projectName"], digData.CompositeAppName,
		digData.CompositeAppVersion)

	jsonLoad, _ := json.Marshal(digp)
	orch.response.payload[digData.Name] = jsonLoad
	orch.response.status[digData.Name] = resp.(int)
	if err != nil {
		return err //need to add the retcode
	}
	if resp != http.StatusCreated {
		return resp
	}
	log.Infof("Deployment intent group response: %d", resp)

	return nil
}

func (h *digpHandler) createObject() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	intentName := "PlacementIntent"
	digData := h.orchInstance.DigData
	gPintName := digData.CompositeAppName + "_gpint"
	nwCtlIntentName := digData.CompositeAppName + "_nwctlint"
	genK8sIntentName := digData.CompositeAppName + "_genk8sint"

	igp := localstore.Intent{
		MetaData: localstore.IntentMetaData{
			Name:        intentName,
			Description: "NA",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}
	igp.Spec.Intent = make(map[string]string)
	igp.Spec.Intent["genericPlacementIntent"] = gPintName
	igp.Spec.Intent["genericaction"] = genK8sIntentName
	// Add genericK8sIntent only if related payload is provided as part of DigData
	/*for _, appData := range orch.DigData.Spec.Apps {
		for _, _ = range appData.RsInfo {
			igp.Spec.Intent["genericaction"] = genK8sIntentName
			break
		}
	}*/

	intentData := h.orchInstance.DigData
	for _, app := range intentData.Spec.Apps {
		// Check if the application has any interfaces.
		// There is assumption that if an application must have same interfaces
		// specified in each cluster.
		if len(app.Interfaces) != 0 {
			igp.Spec.Intent["ovnaction"] = nwCtlIntentName
			h.orchInstance.DigData.NwIntents = true
			break
		}
	}

	status, err := orch.digStore.createIntents(igp, vars["projectName"], digData.CompositeAppName,
		digData.CompositeAppVersion, digData.Name)
	jsonLoad, _ := json.Marshal(igp)
	orch.response.payload[intentName] = jsonLoad
	orch.response.status[intentName] = status.(int)
	if err != nil {
		log.Fatalln(err)
	}
	if status != http.StatusCreated {
		return status
	}
	log.Infof("Placement intent %s status: %d ", intentName, status)

	return nil
}

func createDInents(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.createAnchor()
	if err != nil {
		return err //need to add the retcode
	}
	// 2. Create the Objects
	err = I.createObject()
	if err != nil {
		return err //need to add the retcode
	}
	return nil
}

func delDigp(I orchWorkflow) interface{} {
	// 1. Delete the object
	err := I.deleteObject()
	if err != nil {
		return err //need to add the retcode
	}
	// 2. Delete the Anchor
	err = I.deleteAnchor()
	if err != nil {
		return err //need to add the retcode
	}
	return nil
}

func (h *OrchestrationHandler) createDigData(w *http.ResponseWriter, storeType string) {
	// 1. Create DIG
	if storeType == "emco" {
		dStore := &remoteStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
	} else {
		dStore := &localStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
	}
	igHandler := &digpHandler{}
	igHandler.orchInstance = h
	igpStatus := createDInents(igHandler)
	if igpStatus != nil {
		if intval, ok := igpStatus.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		(*w).Write(h.response.payload[h.Vars["compositeAppName"]+"_digp"])
		return
	}

	// 2. Create intents
	h.Vars["deploymentIntentGroupName"] = h.DigData.Name // SANDEEP : is this gettign initalized anywhere ?
	if storeType == "emco" {
		bstore := &remoteStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
	} else {
		bstore := &localStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
	}
	intentHandler := &placementIntentHandler{}
	intentHandler.orchInstance = h
	intentStatus := addPlacementIntent(intentHandler)
	if intentStatus != nil {
		if intval, ok := intentStatus.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		(*w).Write(h.response.payload[h.Vars["compositeAppName"]+"_gpint"])
		return
	}

	// If the metadata contains network interface request then call the
	// network intent related part of the workflow.
	if h.DigData.NwIntents {
		nwHandler := &networkIntentHandler{}
		nwHandler.orchInstance = h
		nwIntentStatus := addNetworkIntent(nwHandler)
		if nwIntentStatus != nil {
			if intval, ok := nwIntentStatus.(int); ok {
				(*w).WriteHeader(intval)
			} else {
				(*w).WriteHeader(http.StatusInternalServerError)
			}
			(*w).Write(h.response.payload[h.Vars["compositeAppName"]+"_nwctlint"])
			return
		}
	}

	// If the metadata contains resource information, create resources and customizations
	h.createUpdateK8sResource(w, storeType)
}

// Checkout DIG to middleend collection for migrate
func (h *OrchestrationHandler) CheckoutDIGForMigrate(targetVersion string, w *http.ResponseWriter) {
	dStore := &remoteStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore

	bstore := &remoteStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore

	targetDIGExists := true

	// Check if DIG with targetVersion already exists
	retcode, _, err := h.digStore.getDig(h.Vars["projectName"],
		h.Vars["compositeAppName"], targetVersion, h.Vars["deploymentIntentGroupName"])
	log.Infof("Fetch DIG status: %d", retcode)
	if retcode != http.StatusOK {
		log.Errorf("Failed to read digp")
		targetDIGExists = false
	}
	if err != nil {
		log.Errorf("Failed to read digp")
		(*w).WriteHeader(retcode.(int))
		return
	}

	retCode, _ := h.DeleteDig("local")
	if retCode != http.StatusNoContent {
		(*w).WriteHeader(retCode)
		return
	}

	// If DIG with targetVersion exists, populate deployDigData, and create DIG in middleend collection
	if targetDIGExists {
		// As DIG for targetVersion already exists, invoking StoreDIG
		h.Vars["version"] = targetVersion
		appList := make([]string, 0)
		h.readDIGData(w, "emco", appList)
		h.createDigData(w, "middleend")
		(*w).WriteHeader(http.StatusCreated)
		return
	}

	// If DIG with targetVersion does not exists, populate deployDigData, and create DIG in middleend collection
	// Read sourceVersion DIG data
	dataPoints := []string{"projectHandler", "compAppHandler", "ProfileHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler", "genericK8sIntentHandler"}

	h.dataRead = &ProjectTree{}
	h.prepTreeReq()
	dStore = &remoteStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore

	bstore = &remoteStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore

	retcode = h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	sourceVersionData := h.dataRead
	log.Debugf("source DIG data: %+v", sourceVersionData)

	// Read targetVersion service data
	dataPoints = []string{"projectHandler", "compAppHandler", "ProfileHandler"}
	h.Vars["version"] = targetVersion
	h.prepTreeReq()
	h.dataRead = &ProjectTree{}
	dStore = &remoteStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore
	retcode = h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	targetVersionData := h.dataRead
	log.Debugf("target DIG data: %+v", targetVersionData)

	// Populate deployDigData
	var jsonData deployDigData

	for _, compositeAppValue := range targetVersionData.compositeAppMap {
		jsonData.CompositeAppName = compositeAppValue.Metadata.Metadata.Name
		jsonData.CompositeAppVersion = targetVersion

		for compAppProfile := range compositeAppValue.ProfileDataArray {
			jsonData.CompositeProfile = compAppProfile
			break
		}

		var meta []appsData

		for _, app := range compositeAppValue.AppsDataArray {
			appData := appsData{}
			appData.Metadata.FileName = app.App.Metadata.Name + ".tgz"
			appData.Metadata.Name = app.App.Metadata.Name
			appData.Metadata.Description = app.App.Metadata.Description
			meta = append(meta, appData)
		}

		for _, compositeAppValue := range sourceVersionData.compositeAppMap {
			Dig := compositeAppValue.DigMap
			for digName, digValue := range Dig {
				if h.Vars["deploymentIntentGroupName"] == digName {
					jsonData.Name = digName
					jsonData.Description = digValue.DigpData.MetaData.Description
					jsonData.DigVersion = digValue.DigpData.Spec.Version
					jsonData.LogicalCloud = digValue.DigpData.Spec.LogicalCloud
				}

				var appList []string
				for _, sapp := range compositeAppValue.AppsDataArray {
					appList = append(appList, sapp.App.Metadata.Name)
				}

				h.PopulateIntents(digValue, meta, appList)
				// Populate genK8sIntent
				for m, app := range meta {
					if genK8sData, ok := h.genK8sInfo[compositeAppValue.Metadata.Metadata.Name + "_genk8sint"]; ok {
						if val, ok := genK8sData.resData[app.Metadata.Name]; ok {
							meta[m].RsInfo = val
						}
					}
				}
			}
		}
		jsonData.Spec.Apps = meta
		jsonData.Spec.ProjectName = h.Vars["projectName"]
		// If override data is empty then add some dummy override data.
		if len(jsonData.Spec.OverrideValuesObj) == 0 {
			o := localstore.OverrideValues{}
			v := make(map[string]string)
			o.AppName = jsonData.Spec.Apps[0].Metadata.Name
			v["key"] = "value"
			o.ValuesObj = v
			jsonData.Spec.OverrideValuesObj = append(jsonData.Spec.OverrideValuesObj, o)
		}
		log.Debugf("json data for migrate checkout: +%v", jsonData)
	}
	h.DigData.NwIntents = true
	h.DigData = jsonData
	h.createDigData(w, "middleend")
	(*w).WriteHeader(http.StatusCreated)
	return
}

func (h *OrchestrationHandler) isAppExists(appName string, appList []string) bool {
	for _, app := range appList {
		if appName == app {
			return true
		}
	}
	return false
}

func (h *OrchestrationHandler) readDIGData(w *http.ResponseWriter, storeType string, appList []string) {
	var jsonData deployDigData

	dataPoints := []string{"projectHandler", "compAppHandler", "ProfileHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler", "genericK8sIntentHandler"}

	h.dataRead = &ProjectTree{}
	h.prepTreeReq()
	if storeType == "emco" {
		dStore := &remoteStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
		bstore := &remoteStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
	} else {
		dStore := &localStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
		bstore := &localStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
	}

	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	log.Debugf("readDIGData() Data Read : +%v", h.dataRead)

	for _, compositeAppValue := range h.dataRead.compositeAppMap {
		jsonData.CompositeAppName = compositeAppValue.Metadata.Metadata.Name
		jsonData.CompositeAppVersion = compositeAppValue.Metadata.Spec.Version

		for compAppProfile := range compositeAppValue.ProfileDataArray {
			jsonData.CompositeProfile = compAppProfile
			break
		}

		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			if h.Vars["deploymentIntentGroupName"] == digName {
				jsonData.Name = digName
				jsonData.Description = digValue.DigpData.MetaData.Description
				jsonData.DigVersion = digValue.DigpData.Spec.Version
				jsonData.LogicalCloud = digValue.DigpData.Spec.LogicalCloud

				var meta []appsData

				for _, app := range compositeAppValue.AppsDataArray {
					appData := appsData{}
					appData.Metadata.FileName = app.App.Metadata.Name + ".tgz"
					appData.Metadata.Name = app.App.Metadata.Name
					appData.Metadata.Description = app.App.Metadata.Description
					meta = append(meta, appData)
				}

				for _, profile := range compositeAppValue.ProfileDataArray {
					for _, appprofile := range profile.AppProfiles {
						for i := range meta {
							if meta[i].Metadata.Name == appprofile.Spec.AppName {
								meta[i].ProfileMetadata.FileName = appprofile.Metadata.Name
								meta[i].ProfileMetadata.Name = appprofile.Metadata.Name
							}
						}
					}
				}

				h.PopulateIntents(digValue, meta, appList)

				// Populate genK8sIntent
				for m, app := range meta {
					if genK8sData, ok := h.genK8sInfo[compositeAppValue.Metadata.Metadata.Name + "_genk8sint"]; ok {
						if val, ok := genK8sData.resData[app.Metadata.Name]; ok {
							meta[m].RsInfo = val
						}
					}
				}

				jsonData.Spec.Apps = meta
				jsonData.Spec.ProjectName = h.Vars["projectName"]
				jsonData.Spec.OverrideValuesObj = digValue.DigpData.Spec.OverrideValuesObj
				log.Debugf("json data: +%v", jsonData)
			}
		}
		h.DigData.NwIntents = true
		h.DigData = jsonData
	}
}

func (h *OrchestrationHandler) PopulateIntents(digValue *DigReadData, meta []appsData, appList []string) {
	// Populate the generic placement intents
	SourceGpintMap := digValue.GpintMap
	log.Debugf("SourceGpintMap: %s", SourceGpintMap)
	for _, gpintValue := range SourceGpintMap {
		for k := range gpintValue.AppIntentArray {
			for m, app := range meta {
				if len(appList) == 0 || h.isAppExists(app.Metadata.Name, appList) {
					if app.Metadata.Name == gpintValue.AppIntentArray[k].Spec.AppName {
						meta[m].Clusters = make([]ClusterInfo, 0)
						log.Infof("app name %s : %s %d", app.Metadata.Name, gpintValue.AppIntentArray[k].Spec.AppName, m)
						for i := range gpintValue.AppIntentArray[k].Spec.Intent.AllOfArray {
							var cluster ClusterInfo
							cluster.SelectedClusters = make([]SelectedCluster, 0)
							cluster.SelectedLabels = make([]SelectedLabel, 0)
							cluster.Provider = gpintValue.AppIntentArray[k].Spec.Intent.AllOfArray[i].ProviderName
							if len(gpintValue.AppIntentArray[k].Spec.Intent.AllOfArray[i].ClusterName) > 0 {
								cluster.SelectedClusters = append(cluster.SelectedClusters,
									SelectedCluster{Name: gpintValue.AppIntentArray[k].Spec.Intent.AllOfArray[i].ClusterName})
							}

							if len(gpintValue.AppIntentArray[k].Spec.Intent.AllOfArray[i].ClusterLabelName) > 0 {
								cluster.SelectedLabels = append(cluster.SelectedLabels,
									SelectedLabel{Name: gpintValue.AppIntentArray[k].Spec.Intent.AllOfArray[i].ClusterLabelName})
							}
							meta[m].Clusters = append(meta[m].Clusters, cluster)
							meta[m].PlacementCriterion = "allOf"
						}

						for i := range gpintValue.AppIntentArray[k].Spec.Intent.AnyOfArray {
							var cluster ClusterInfo
							cluster.SelectedClusters = make([]SelectedCluster, 0)
							cluster.SelectedLabels = make([]SelectedLabel, 0)
							cluster.Provider = gpintValue.AppIntentArray[k].Spec.Intent.AnyOfArray[i].ProviderName
							if len(gpintValue.AppIntentArray[k].Spec.Intent.AnyOfArray[i].ClusterName) > 0 {
								cluster.SelectedClusters = append(cluster.SelectedClusters,
									SelectedCluster{Name: gpintValue.AppIntentArray[k].Spec.Intent.AnyOfArray[i].ClusterName})
							}

							if len(gpintValue.AppIntentArray[k].Spec.Intent.AnyOfArray[i].ClusterLabelName) > 0 {
								cluster.SelectedLabels = append(cluster.SelectedLabels,
									SelectedLabel{Name: gpintValue.AppIntentArray[k].Spec.Intent.AnyOfArray[i].ClusterLabelName})
							}
							meta[m].Clusters = append(meta[m].Clusters, cluster)
							meta[m].PlacementCriterion = "anyOf"
						}
					}
				}
			}
		}
	}

	networkIntents := digValue.NwintMap
	log.Debugf("PopulateIntents() networkIntents: %s", networkIntents)
	for _, nwintValue := range networkIntents {
		for _, workloadIntents := range nwintValue.WrkintMap {
			appName := workloadIntents.Wrkint.Spec.AppName
			for m, app := range meta {
				log.Debugf("PopulateIntents() appName %s == app.Metadata.Name %s", appName, app.Metadata.Name)
				if len(appList) == 0 || h.isAppExists(app.Metadata.Name, appList) {
					if app.Metadata.Name == appName {
						meta[m].Interfaces = make([]NwInterfaces, len(workloadIntents.Interfaces))
						for i, nwinterface := range workloadIntents.Interfaces {
							meta[m].Interfaces[i].NetworkName = nwinterface.Spec.Name
							meta[m].Interfaces[i].IP = nwinterface.Spec.IPAddress
							meta[m].Interfaces[i].InterfaceName = nwinterface.Spec.Interface
						}
					}
				}
			}
		}
	}
}

// Checkout DIG information to middleend collection for update
func (h *OrchestrationHandler) CheckoutDIGForUpdate(w *http.ResponseWriter) {
	// Read data from EMCO and write to middleend
	appList := make([]string, 0)
	h.readDIGData(w, "emco", appList)
	h.createDigData(w, "middlend")

	(*w).WriteHeader(http.StatusCreated)
	return
}
func (h *OrchestrationHandler) ScaleOutDig(w http.ResponseWriter, r *http.Request) {
	// 1. Call Checkout DIG, set the query param to operation = update.
	q := r.URL.Query()
	q.Set("operation", "update")
	r.URL.RawQuery = q.Encode()
	h.CheckoutDIG(w, r)
	log.Debugf("1. Header value %s", w.Header())

	// 2. PUT the intet update
	q = r.URL.Query()
	q.Set("operation", "save")
	r.URL.RawQuery = q.Encode()
	h.DigUpdateHandler(w, r)
	log.Debugf("2. Header value %s", w.Header())

	// 3. Submit the dig update
	h.UpgradeDIG(w, r)
	log.Debugf("3. Header value %s", w.Header())
}

// Checkout DIG information to middleend collection
func (h *OrchestrationHandler) CheckoutDIG(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()

	filter := r.URL.Query().Get("operation")
	targetVersion := r.URL.Query().Get("targetVersion")

	var version string
	if targetVersion != "" {
		version = targetVersion
	} else {
		version = h.Vars["version"]
	}

	localDigStore := localStoreDigHandler{}
	// Check if checkout version already exists
	localDigRetCode, _, _ := localDigStore.getDig(h.Vars["projectName"],
		h.Vars["compositeAppName"], version, h.Vars["deploymentIntentGroupName"])
	if localDigRetCode == http.StatusOK {
		log.Infof("Checkout for DIG %s already exists", h.Vars["deploymentIntentGroupName"])
		w.WriteHeader(http.StatusOK)
		return
	}

	if filter == "migrate" {
		h.Vars["operation"] = "migrate"
		h.Vars["originalversion"] = h.Vars["version"]
		h.CheckoutDIGForMigrate(targetVersion, &w)
		return
	}
	if filter == "update" {
		h.Vars["operation"] = "update"
		h.Vars["originalversion"] = h.Vars["version"]
		h.CheckoutDIGForUpdate(&w)
		return
	}
	if filter != "migrate" && filter != "update" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *OrchestrationHandler) UpdateIntents(w *http.ResponseWriter) {
	var jsonData deployDigData

	// Fetch app intents data from middleend collection
	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler", "placementIntentHandler",
		"networkIntentHandler", "genericK8sIntentHandler"}

	h.dataRead = &ProjectTree{}
	h.prepTreeReq()
	bstore := &localStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore

	dStore := &localStoreDigHandler{}
	dStore.orchInstance = h
	h.digStore = dStore

	retCode := h.constructTree(dataPoints)
	if retCode != nil {
		if intval, ok := retCode.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Reset intent store to EMCO
	newbstore := &remoteStoreIntentHandler{}
	temp := &placementIntentHandler{}
	temp.orchInstance = h
	newbstore.orchInstance = temp.orchInstance
	temp.orchInstance.bstore = newbstore

	// Iterate and determine which app intents need to be played
	for _, compositeAppValue := range h.dataRead.compositeAppMap {
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			var meta []appsData
			for _, app := range compositeAppValue.AppsDataArray {
				appData := appsData{}
				appData.Metadata.Name = app.App.Metadata.Name
				if genK8sData, ok := h.genK8sInfo[compositeAppValue.Metadata.Metadata.Name + "_genk8sint"]; ok {
					if val, ok := genK8sData.resData[app.App.Metadata.Name]; ok {
						appData.RsInfo = val
					}
				}
				meta = append(meta, appData)
			}
			jsonData.Name = digName
			jsonData.Spec.Apps = meta
			h.DigData = jsonData

			// If the metadata contains resource information, create resources and customizations
			h.createUpdateK8sResource(w, "emco")

			// Update placement intents
			for gpintName, gpintValue := range digValue.GpintMap {
				for _, appIntent := range gpintValue.AppIntentArray {
					if appIntent.MetaData.UserData1 == "updated" {
						// Play updated changes on EMCO
						retcode, err := temp.orchInstance.bstore.deleteAppPIntent(appIntent.MetaData.Name, h.Vars["projectName"],
							h.Vars["compositeAppName"], h.Vars["version"], gpintName, digName)
						if err != nil {
							log.Errorf("%s", err)
							(*w).WriteHeader(retcode.(int))
							return
						}

						retcode, err = temp.orchInstance.bstore.createAppPIntent(appIntent, h.Vars["projectName"],
							h.Vars["compositeAppName"], h.Vars["version"], digName, gpintName)
						if err != nil {
							log.Errorf("%s", err)
							(*w).WriteHeader(retcode.(int))
							return
						}
					}
				}
			}

			// Update network intents
			networkIntents := digValue.NwintMap
			for _, nwintValue := range networkIntents {
				for _, workloadIntents := range nwintValue.WrkintMap {
					appName := workloadIntents.Wrkint.Spec.AppName
					if workloadIntents.Wrkint.Metadata.UserData1 == "updated" {
						for _, nwinterface := range workloadIntents.Interfaces {
							temp.orchInstance.bstore.deleteWorkloadIfIntent(nwinterface.Metadata.Name, appName+"_wlint",
								h.Vars["projectName"], h.Vars["compositeAppName"], h.Vars["version"], digName, nwintValue.Nwint.Metadata.Name)
						}
					}
					if workloadIntents.Wrkint.Metadata.UserData1 == "updated" {
						for _, nwinterface := range workloadIntents.Interfaces {
							cint := localstore.WorkloadIfIntent{
								Metadata: localstore.Metadata{
									Name:        nwinterface.Metadata.Name,
									Description: workloadIntents.Wrkint.Metadata.Description,
									UserData1:   workloadIntents.Wrkint.Metadata.UserData1,
									UserData2:   workloadIntents.Wrkint.Metadata.UserData2,
								},
								Spec: localstore.WorkloadIfIntentSpec{
									IfName:         nwinterface.Spec.Interface,
									NetworkName:    nwinterface.Spec.Name,
									DefaultGateway: nwinterface.Spec.DefaultGateway,
									IpAddr:         nwinterface.Spec.IPAddress,
								},
							}
							temp.orchInstance.bstore.createWorkloadIfIntent(cint, h.Vars["projectName"],
								h.Vars["compositeAppName"], h.Vars["version"], digName, nwintValue.Nwint.Metadata.Name,
								appName+"_wlint", false, nwinterface.Metadata.Name)
						}
					}
				}
			}
		}
	}
}

// Perform DIG upgrade
func (h *OrchestrationHandler) UpgradeDIG(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()

	// Read DIG from middleend, and determine type of operation
	localDigStore := localStoreDigHandler{}
	tempDIG := localstore.DeploymentIntentGroup{}
	retCode, retValue, err := localDigStore.getDig(h.Vars["projectName"],
		h.Vars["compositeAppName"], h.Vars["version"], h.Vars["deploymentIntentGroupName"])
	if retCode != http.StatusOK {
		log.Errorf("Failed to read digp")
		w.WriteHeader(retCode.(int))
		return
	}
	if err != nil {
		log.Errorf("Failed to read digp")
		w.WriteHeader(retCode.(int))
		return
	}

	json.Unmarshal(retValue, &tempDIG)

	targetDIGExists := true

	// Set digStore to EMCO
	newdStore := &remoteStoreDigHandler{}
	newdStore.orchInstance = h
	h.digStore = newdStore

	// Check if DIG with targetVersion already exists
	retCode, _, _ = h.digStore.getDig(h.Vars["projectName"],
		h.Vars["compositeAppName"], h.Vars["version"], h.Vars["deploymentIntentGroupName"])
	if retCode != http.StatusOK {
		log.Errorf("Failed to read digp")
		targetDIGExists = false
	}

	if targetDIGExists {
		// Fetch DIG state
		digStatus, retcode := newdStore.orchInstance.digStore.getStatus(h.Vars["compositeAppName"],
			h.Vars["version"],
			h.Vars["deploymentIntentGroupName"])
		if retcode != http.StatusOK {
			log.Errorf("Failed to read digp:md")
			w.WriteHeader(retcode)
			return
		}

		// Fetch the latest DIG state
		state := digStatus.States.Actions[len(digStatus.States.Actions)-1].State
		if tempDIG.MetaData.UserData1 == "update" && state != localstore.StateEnum.Instantiated {
			log.Errorf("DIG %s is not instantiated", h.Vars["deploymentIntentGroupName"])
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
	}

	// Create DIG with targetVersion, if not exists, else update intents
	if !targetDIGExists {
		appList := make([]string, 0)
		h.readDIGData(&w, "middleend", appList)
		h.createDigData(&w, "emco")
	} else {
		h.UpdateIntents(&w)
	}

	if tempDIG.MetaData.UserData1 == "migrate" {
		originalVersion := tempDIG.MetaData.UserData2
		// Approve DIG with targetVersion
		orchURL := "http://" + h.MiddleendConf.OrchService + "/v2/projects/" +
			h.Vars["projectName"] + "/composite-apps/" + h.Vars["compositeAppName"] +
			"/" + h.Vars["version"] +
			"/deployment-intent-groups/" + h.Vars["deploymentIntentGroupName"] + "/approve"

		var jsonLoad []byte
		retcode, err := h.apiPost(jsonLoad, orchURL, h.Vars["deploymentIntentGroupName"])
		if err != nil {
			log.Errorf("Failed to invoke dig approve: %s", err)
			w.WriteHeader(retcode.(int))
			return
		}

		// Invoke EMCO migrate API
		var temp localstore.MigrateJson
		temp.MetaData.Description = "Upgrade DIG"
		temp.Spec.TargetCompositeAppVersion = h.Vars["version"]
		temp.Spec.TargetDigName = h.Vars["deploymentIntentGroupName"]

		jsonLoad, _ = json.Marshal(temp)
		orchURL = "http://" + h.MiddleendConf.OrchService + "/v2/projects/" +
			h.Vars["projectName"] + "/composite-apps/" + h.Vars["compositeAppName"] +
			"/" + originalVersion +
			"/deployment-intent-groups/" + h.Vars["deploymentIntentGroupName"] + "/migrate"

		retcode, err = h.apiPost(jsonLoad, orchURL, h.Vars["deploymentIntentGroupName"])
		if err != nil {
			log.Errorf("Failed to invoke dig update: %s", err)
			w.WriteHeader(retcode.(int))
			return
		}

		if retcode != http.StatusAccepted {
			log.Errorf("Encountered error while migrating DIG %s", h.Vars["deploymentIntentGroupName"])
			w.WriteHeader(retcode.(int))
			return
		}

		// Append current version to the list of version for which migrate occurred
		h.UpdateDIGInfo()

		w.WriteHeader(retcode.(int))
	}

	if tempDIG.MetaData.UserData1 == "update" {
		// Invoke EMCO update API
		var jsonLoad []byte
		orchURL := "http://" + newdStore.orchInstance.MiddleendConf.OrchService + "/v2/projects/" +
			h.Vars["projectName"] + "/composite-apps/" + h.Vars["compositeAppName"] +
			"/" + h.Vars["version"] +
			"/deployment-intent-groups/" + h.Vars["deploymentIntentGroupName"] + "/update"

		retCode, err = newdStore.orchInstance.apiPost(jsonLoad, orchURL, h.Vars["deploymentIntentGroupName"])
		if err != nil {
			log.Errorf("Failed to invoke dig update: %s", err)
			w.WriteHeader(retCode.(int))
			return
		}

		if retCode != http.StatusAccepted {
			log.Errorf("Encountered error while updating DIG %s", h.Vars["deploymentIntentGroupName"])
			w.WriteHeader(retCode.(int))
			return
		}

		w.WriteHeader(retCode.(int))
	}

	// Delete checkout DIG
	retcode, _ := h.DeleteDig("local")
	if retcode != http.StatusNoContent {
		w.WriteHeader(retcode)
		return
	}
	return
}

// Get all DIGs
func (h *OrchestrationHandler) GetDigs(w *http.ResponseWriter, storeType string) {
	h.InitializeResponseMap()
	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler"}

	h.dataRead = &ProjectTree{}
	h.prepTreeReq()
	if storeType == "emco" {
		dStore := &remoteStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
	} else {
		dStore := &localStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
	}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

func (h *OrchestrationHandler) DeleteDig(filter string) (int, string) {
	var originalVersion string
	h.InitializeResponseMap()
	h.treeFilter = nil

	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler", "genericK8sIntentHandler"}

	// Initialize the project tree with target composite application.
	h.prepTreeReq()

	h.dataRead = &ProjectTree{}
	if filter == "local" {
		h.prepTreeReq()
		bstore := &localStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore

		dStore := &localStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
	} else {
		h.prepTreeReq()
		bstore := &remoteStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore

		dStore := &remoteStoreDigHandler{}
		dStore.orchInstance = h
		h.digStore = dStore
	}

	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			return intval, originalVersion
		} else {
			return http.StatusInternalServerError, originalVersion
		}
	}

	// Populate response
	for compositeAppName, value := range h.dataRead.compositeAppMap {
		for _, digValue := range h.dataRead.compositeAppMap[compositeAppName].DigMap {
			if value.Metadata.Spec.Version == h.Vars["version"] {
				log.Debugf("Found original version: %s", digValue.DigpData.MetaData.UserData2)
				originalVersion = digValue.DigpData.MetaData.UserData2
				break
			}
		}
	}

	// 1. Call DIG delete workflow
	log.Info("Start DIG delete workflow")
	deleteDataPoints := []string{"networkIntentHandler",
		"placementIntentHandler", "genericK8sIntentHandler",
		"digpHandler"}
	retcode = h.deleteTree(deleteDataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			return intval, originalVersion
		} else {
			return http.StatusInternalServerError, originalVersion
		}
	}
	return http.StatusNoContent, originalVersion
}

// Add DIG Info
func (h *OrchestrationHandler) AddDIGInfo() {
	var diginfo DigInfo

	digName := h.Vars["deploymentIntentGroupName"]
	key := DigInfoKey{DigName: digName}

	diginfo.DigName = digName
	diginfo.VersionList = append(diginfo.VersionList, h.Vars["version"])

	err := db.DBconn.Insert(DIG_INFO_COLLECTION, key, nil, "digmeta", diginfo)
	if err != nil {
		log.Errorf("Encountered error during add of dig info for %s: %s", h.Vars["deploymentIntentGroupName"], err)
		return
	}
	return
}

// Delete DIG Info
func (h *OrchestrationHandler) DeleteDIGInfo() {
	key := DigInfoKey{DigName: h.Vars["deploymentIntentGroupName"]}

	err := db.DBconn.Remove(DIG_INFO_COLLECTION, key)
	if err != nil {
		log.Errorf("Encountered error during delete of dig info for %s: %s", h.Vars["deploymentIntentGroupName"], err)
		return
	}
	return
}

// Fetch DIG Info
func (h *OrchestrationHandler) FetchDIGInfo(digName string) DigInfo {
	var diginfo DigInfo
	key := DigInfoKey{DigName: digName}
	exists := db.DBconn.CheckCollectionExists(DIG_INFO_COLLECTION)
	if exists {
		values, err := db.DBconn.Find(DIG_INFO_COLLECTION, key, "digmeta")
		if err != nil {
			log.Errorf("Encountered error while fetching DIG info for %s: %s", digName, err)
			return diginfo
		} else if len(values) == 0 {
			log.Infof("DIG info does not exists")
			return diginfo
		}
		err = db.DBconn.Unmarshal(values[0], &diginfo)
		log.Infof("DIG Info after Unmarshalling: %s", diginfo)
		if err != nil {
			log.Errorf("Unmarshalling DIG Info failed: %s", err)
			return diginfo
		}
	}
	return diginfo
}

// Update DIG Info collection
func (h *OrchestrationHandler) UpdateDIGInfo() {
	var diginfo DigInfo
	key := DigInfoKey{DigName: h.Vars["deploymentIntentGroupName"]}
	exists := db.DBconn.CheckCollectionExists(DIG_INFO_COLLECTION)
	if exists {
		values, err := db.DBconn.Find(DIG_INFO_COLLECTION, key, "digmeta")
		if err != nil {
			log.Errorf("Encountered error while fetching draft composite application: %s", err)
			return
		} else if len(values) == 0 {
			log.Infof("DIG info does not exists")
			return
		}

		err = db.DBconn.Unmarshal(values[0], &diginfo)
		log.Infof("DIG Info after Unmarshalling: %s", diginfo)
		if err != nil {
			log.Errorf("Unmarshalling DIG Info failed: %s", err)
			return
		}

		// Add current version to the list of versions of composite-app mapped to DIG
		diginfo.VersionList = append(diginfo.VersionList, h.Vars["version"])

		err = db.DBconn.Insert(DIG_INFO_COLLECTION, key, nil, "digmeta", diginfo)
		if err != nil {
			log.Errorf("Encountered error during update of dig info for %s: %s", h.Vars["deploymentIntentGroupName"], err)
			return
		}
	}
	return
}
