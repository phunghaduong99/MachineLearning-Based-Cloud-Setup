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
	"strconv"
	"strings"

	"example.com/middleend/localstore"
	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type PlacementIntentExport struct {
	Metadata localstore.MetaData          `json:"metadata"`
	Spec     AppPlacementIntentSpecExport `json:"spec"`
}

type AppPlacementIntentSpecExport struct {
	AppName string            `json:"appName"`
	Intent  arrayIntentExport `json:"intent"`
}
type arrayIntentExport struct {
	AllofCluster []AllofExport `json:"allof"`
	AnyofCluster []AnyofExport `json:"anyof"`
}
type AllofExport struct {
	ProviderName     string `json:"providerName"`
	ClusterName      string `json:"clusterName"`
	ClusterLabelName string `json:"clusterLabelName"`
}

type AnyofExport struct {
	ProviderName     string `json:"providerName"`
	ClusterName      string `json:"clusterName"`
	ClusterLabelName string `json:"clusterLabelName"`
}

// plamcentIntentHandler implements the orchworkflow interface
type placementIntentHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

type NetworkCtlIntent struct {
	Metadata apiMetaData `json:"metadata"`
}

type NetworkWlIntent struct {
	Metadata apiMetaData        `json:"metadata"`
	Spec     WorkloadIntentSpec `json:"spec"`
}

type WorkloadIntentSpec struct {
	AppName  string `json:"app"`
	Resource string `json:"workloadResource"`
	Type     string `json:"type"`
}

type NwInterface struct {
	Metadata apiMetaData   `json:"metadata"`
	Spec     InterfaceSpec `json:"spec"`
}

type InterfaceSpec struct {
	Interface      string `json:"interface"`
	Name           string `json:"name"`
	DefaultGateway string `json:"defaultGateway"`
	IPAddress      string `json:"ipAddress"`
	MacAddress     string `json:"macAddress"`
	SubNet         string `json:"subnet,omitempty"`
}

// networkIntentHandler implements the orchworkflow interface
type networkIntentHandler struct {
	ovnURL       string
	orchInstance *OrchestrationHandler
}

// genericK8sIntentHandler implements the orchworkflow interface
type genericK8sIntentHandler struct {
	ovnURL       string
	orchInstance *OrchestrationHandler
}

// localStoreIntentHandler implements the orchworkflow interface
type localStoreIntentHandler struct {
	orchInstance *OrchestrationHandler
}
type remoteStoreIntentHandler struct {
	orchInstance *OrchestrationHandler
}

// localStoreNwintHandler implements the orchworkflow interface
type localStoreNwintHandler struct {
	orchInstance *OrchestrationHandler
}
type remoteStoreNwintHandler struct {
	orchInstance *OrchestrationHandler
}

// Interface to creating the backend objects
// either in EMCO over REST or in middleend mongo
type backendStore interface {
	createGpint(localstore.GenericPlacementIntent, string, string, string, string) (interface{}, interface{})
	deleteGpint(string, string, string, string, string) (interface{}, interface{})
	createAppPIntent(localstore.AppIntent, string, string, string, string, string) (interface{}, interface{})
	deleteAppPIntent(ai string, p string, ca string, v string,
		gpintName string, digName string) (interface{}, interface{})
	getAllGPint(project string, compositeAppName string, version string, digName string) (interface{}, []byte, interface{})
	getAppPIntent(intentName string, gpintName string, project string, compositeAppName string, version string,
		digName string) (interface{}, []byte, interface{})
	createControllerIntent(cint localstore.NetControlIntent, p string, ca string, v string,
		digName string, exists bool, intentName string) (interface{}, interface{})
	getControllerIntents(p string, ca string, v string,
		digName string) (interface{}, []byte, interface{})
	deleteControllerIntent(p string, ca string, v string,
		digName string, intentName string) (interface{}, interface{})
	createWorkloadIntent(cint localstore.WorkloadIntent, p string, ca string, v string,
		digName string, nwControllerIntentName string, exists bool, intentName string) (interface{}, interface{})
	getWorkloadIntents(p string, ca string, v string,
		digName string, nwControllerIntentName string) (interface{}, []byte, interface{})
	deleteWorkloadIntent(workloadIntentName, p string, ca string, v string,
		digName string, nwControllerIntentName string) (interface{}, interface{})
	createWorkloadIfIntent(cint localstore.WorkloadIfIntent, p string, ca string, v string,
		digName string, nwControllerIntentName string, workloadIntentName string, exists bool, intentName string) (interface{}, interface{})
	getWorkloadIfIntents(p string, ca string, v string,
		digName string, nwControllerIntentName string, workloadIntentName string) (interface{}, []byte, interface{})
	deleteWorkloadIfIntent(ifaceName string, workloadIntentName, p string, ca string, v string,
		digName string, nwControllerIntentName string) (interface{}, interface{})
	createGenericK8sIntent(gki localstore.GenericK8sIntent, p string, ca string, v string, digName string, exists bool) (interface{}, interface{})
	deleteGenericK8sIntent(gkiName string, p string, ca string, v string, digName string) (interface{}, interface{})
	createResource(r localstore.Resource, t localstore.ResourceFileContent, fName, p, ca, cv, dig, gi string, exists bool) (interface{}, interface{})
	createCustomization(c localstore.Customization, t localstore.SpecFileContent, p, ca, cv, dig, gi, rs string, exists bool) (interface{}, interface{})
	getAllResources(p, ca, cv, dig, gi string) (interface{}, []byte, interface{})
	getResource(rName, p, ca, cv, dig, gi string) (interface{}, []byte, interface{})
	deleteResource(rName, p, ca, v, digName, gi string) (interface{}, interface{})
	getAllCustomization(p, ca, cv, dig, gki, rs string) (interface{}, []byte, interface{})
	getCustomization(c, p, ca, v, digName, gi, rs string) (interface{}, []byte, interface{})
	deleteCustomization(c, p, ca, v, digName, gi, rs string) (interface{}, interface{})
	getResourceContent(rName, p, ca, cv, dig, gi string)(interface{}, []byte, interface{})
	getCustomizationContent(c, p, ca, cv, dig, gi, rs string)(interface{}, []byte, interface{})
}

func (h *remoteStoreIntentHandler) createWorkloadIfIntent(wifint localstore.WorkloadIfIntent, p string, ca string, v string,
	digName string, nwControllerIntentName string, workloadIntentName string, exists bool, intentName string) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(wifint)
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" +
		nwControllerIntentName + "/workload-intents/" + workloadIntentName + "/interfaces"
	resp, err := orch.apiPost(jsonLoad, url, intentName)
	return resp, err
}

func (h *localStoreIntentHandler) createWorkloadIfIntent(wifint localstore.WorkloadIfIntent, p string, ca string, v string,
	digName string, nwControllerIntentName string, workloadIntentName string, exists bool, intentName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewWorkloadIfIntentClient()
	_, createErr := c.CreateWorkloadIfIntent(wifint, p, ca, v, digName, nwControllerIntentName, workloadIntentName, true)
	if createErr != nil {
		log.Error(":: Error creating workload interface ::", log.Fields{"Error": createErr})
		if strings.Contains(createErr.Error(), "does not exist") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "WorkloadIfIntent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}
	return http.StatusCreated, createErr
}

func (h *remoteStoreIntentHandler) getWorkloadIfIntents(p string, ca string, v string,
	digName string, nwControllerIntent string, workloadIntentName string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" +
		nwControllerIntent + "/workload-intents/" + workloadIntentName + "/interfaces"
	resp, retval, err := orch.apiGet(url, ca+"_getifaces")
	return resp, retval, err
}

func (h *localStoreIntentHandler) getWorkloadIfIntents(p string, ca string, v string,
	digName string, nwControllerIntent string, workloadIntentName string) (interface{}, []byte, interface{}) {
	// Get the local store handler.
	var retval []byte
	c := localstore.NewWorkloadIfIntentClient()
	interfaces, err := c.GetWorkloadIfIntents(p, ca, v, digName, nwControllerIntent, workloadIntentName)
	if err != nil {
		log.Error(":: Error getting workload interfaces ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err
		} else {
			return http.StatusInternalServerError, retval, err
		}
	}
	retval, _ = json.Marshal(interfaces)
	return http.StatusOK, retval, err
}

func (h *remoteStoreIntentHandler) deleteWorkloadIfIntent(ifaceName string, workloadIntentName string, p string, ca string, v string,
	digName string, nwControllerIntent string) (interface{}, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" +
		nwControllerIntent + "/workload-intents/" + workloadIntentName + "/interfaces/" + ifaceName
	resp, err := orch.apiDel(url, ca+"_delIface")
	return resp, err
}

func (h *localStoreIntentHandler) deleteWorkloadIfIntent(ifaceName string, workloadIntentName string, p string, ca string, v string,
	digName string, nwControllerIntent string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewWorkloadIfIntentClient()
	err := c.DeleteWorkloadIfIntent(ifaceName, p, ca, v, digName, nwControllerIntent, workloadIntentName)
	if err != nil {
		log.Error(":: Error deleting workloadIfIntent ::", log.Fields{"Error": err, "Name": ifaceName})
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

func (h *remoteStoreIntentHandler) createWorkloadIntent(wint localstore.WorkloadIntent, p string, ca string, v string,
	digName string, nwControllerIntentName string, exists bool, intentName string) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(wint)
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" +
		nwControllerIntentName + "/workload-intents"
	resp, err := orch.apiPost(jsonLoad, url, intentName)
	return resp, err
}

func (h *localStoreIntentHandler) createWorkloadIntent(wint localstore.WorkloadIntent, p string, ca string, v string,
	digName string, nwControllerIntentName string, exists bool, intentName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewWorkloadIntentClient()
	_, createErr := c.CreateWorkloadIntent(wint, p, ca, v, digName, nwControllerIntentName, true)
	if createErr != nil {
		log.Error(":: Error creating workload intent ::", log.Fields{"Error": createErr})
		if strings.Contains(createErr.Error(), "does not exist") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "WorkloadIntent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}
	return http.StatusCreated, createErr
}

func (h *remoteStoreIntentHandler) getWorkloadIntents(p string, ca string, v string,
	digName string, nwControllerIntent string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" +
		nwControllerIntent + "/workload-intents"
	resp, retval, err := orch.apiGet(url, ca+"_getWrkInt")
	return resp, retval, err
}

func (h *localStoreIntentHandler) getWorkloadIntents(p string, ca string, v string,
	digName string, nwControllerIntent string) (interface{}, []byte, interface{}) {
	// Get the local store handler.
	var retval []byte
	c := localstore.NewWorkloadIntentClient()
	workloadIntents, err := c.GetWorkloadIntents(p, ca, v, digName, nwControllerIntent)
	if err != nil {
		log.Error(":: Error getting workload intents ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err
		} else {
			return http.StatusInternalServerError, retval, err
		}
	}
	retval, _ = json.Marshal(workloadIntents)
	return http.StatusOK, retval, err
}

func (h *remoteStoreIntentHandler) deleteWorkloadIntent(workloadIntentName string, p string, ca string, v string,
	digName string, nwControllerIntent string) (interface{}, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" +
		nwControllerIntent + "/workload-intents/" + workloadIntentName
	resp, err := orch.apiDel(url, ca+"_delWrkInt")
	return resp, err
}

func (h *localStoreIntentHandler) deleteWorkloadIntent(workloadIntentName string, p string, ca string, v string,
	digName string, nwControllerIntent string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewWorkloadIntentClient()
	err := c.DeleteWorkloadIntent(workloadIntentName, p, ca, v, digName, nwControllerIntent)
	if err != nil {
		log.Error(":: Error deleting workload intent ::", log.Fields{"Error": err, "Name": workloadIntentName})
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

func (h *remoteStoreIntentHandler) createControllerIntent(cint localstore.NetControlIntent, p string, ca string, v string,
	digName string, exists bool, intentName string) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(cint)
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent"
	resp, err := orch.apiPost(jsonLoad, url, intentName)
	return resp, err
}

func (h *localStoreIntentHandler) createControllerIntent(cint localstore.NetControlIntent, p string, ca string, v string,
	digName string, exists bool, intentName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewNetControlIntentClient()
	_, createErr := c.CreateNetControlIntent(cint, p, ca, v, digName, true)
	if createErr != nil {
		log.Error(":: Error creating network control intent ::", log.Fields{"Error": createErr})
		if strings.Contains(createErr.Error(), "NetControlIntent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}
	return http.StatusCreated, createErr
}

func (h *remoteStoreIntentHandler) getControllerIntents(p string, ca string, v string,
	digName string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent"
	resp, retval, err := orch.apiGet(url, ca+"_getNwCtlInt")
	return resp, retval, err
}

func (h *localStoreIntentHandler) getControllerIntents(p string, ca string, v string,
	digName string) (interface{}, []byte, interface{}) {
	// Get the local store handler.
	var retval []byte
	c := localstore.NewNetControlIntentClient()
	ctlInents, err := c.GetNetControlIntents(p, ca, v, digName)
	if err != nil {
		log.Error(":: Error getting network control intents ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err
		} else {
			return http.StatusInternalServerError, retval, err
		}
	}
	retval, _ = json.Marshal(ctlInents)
	return http.StatusOK, retval, err
}

func (h *remoteStoreIntentHandler) deleteControllerIntent(nwIntentName string, p string, ca string, v string,
	digName string) (interface{}, interface{}) {
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/network-controller-intent/" + nwIntentName
	resp, err := orch.apiDel(url, ca+"_delnwCtlInt")
	return resp, err
}

func (h *localStoreIntentHandler) deleteControllerIntent(nwIntentName string, p string, ca string, v string,
	digName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewNetControlIntentClient()
	err := c.DeleteNetControlIntent(nwIntentName, p, ca, v, digName)
	if err != nil {
		log.Error(":: Error deleting network control intent ::", log.Fields{"Error": err, "Name": nwIntentName})
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

func (h *localStoreIntentHandler) getAllGPint(project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {
	var retval []byte
	c := localstore.NewGenericPlacementIntentClient()
	gPIntent, err := c.GetAllGenericPlacementIntents(project, compositeAppName, version, digName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "Unable to find") {
			return http.StatusNotFound, retval, err
		} else if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err
		} else {
			return http.StatusInternalServerError, retval, err
		}
	}
	log.Infof("Get All gpint localstore Composite app %s dig %s status: %s : value %s", compositeAppName,
		digName, gPIntent)
	retval, _ = json.Marshal(gPIntent)
	return http.StatusOK, retval, err
}

func (h *remoteStoreIntentHandler) getAllGPint(project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {

	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		project + "/composite-apps/" + compositeAppName +
		"/" + version +
		"/deployment-intent-groups/" + digName + "/generic-placement-intents"
	retcode, retval, err := orch.apiGet(orchURL, compositeAppName+"_gpint")
	log.Infof("Get Gpint in Composite app %s dig %s status: %d", compositeAppName,
		digName, retcode)
	return retcode, retval, err
}

func (h *remoteStoreIntentHandler) getAppPIntent(intentName string, gpintName string, project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {

	orch := h.orchInstance
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		project + "/composite-apps/" + compositeAppName +
		"/" + version + "/deployment-intent-groups/" + digName + "/generic-placement-intents"
	url := orchURL + "/" + gpintName + "/app-intents/" + intentName
	retcode, retval, err := orch.apiGet(url, compositeAppName+"_getappPint")
	return retcode, retval, err
}

func (h *localStoreIntentHandler) getAppPIntent(intentName string, gpintName string, project string, compositeAppName string, version string,
	digName string) (interface{}, []byte, interface{}) {
	var retval []byte
	c := localstore.NewAppIntentClient()
	appIntent, err := c.GetAppIntent(intentName, project, compositeAppName, version, gpintName, digName)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err
		} else {
			return http.StatusInternalServerError, retval, err
		}
	}
	retval, _ = json.Marshal(appIntent)
	return http.StatusOK, retval, err
}

func (h *localStoreIntentHandler) createGpint(g localstore.GenericPlacementIntent, p string, ca string,
	v string, digName string) (interface{}, interface{}) {
	c := localstore.NewGenericPlacementIntentClient()

	_, createErr := c.CreateGenericPlacementIntent(g, p, ca, v, digName)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the deploymentIntentGroupName") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Intent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}
	return http.StatusCreated, nil
}

func (h *remoteStoreIntentHandler) createGpint(g localstore.GenericPlacementIntent, p string, ca string,
	v string, digName string) (interface{}, interface{}) {
	orch := h.orchInstance
	gPintName := ca + "_gpint"
	jsonLoad, _ := json.Marshal(g)
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-placement-intents"
	resp, err := orch.apiPost(jsonLoad, url, gPintName)
	return resp, err
}

func (h *localStoreIntentHandler) deleteAppPIntent(appIntentName string, p string, ca string, v string,
	gpintName string, digName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewAppIntentClient()
	deleteErr := c.DeleteAppIntent(appIntentName, p, ca, v, gpintName, digName)
	if deleteErr != nil {
		log.Error(deleteErr.Error(), log.Fields{})
		if strings.Contains(deleteErr.Error(), "not found") {
			return http.StatusNotFound, deleteErr
		} else if strings.Contains(deleteErr.Error(), "conflict") {
			return http.StatusConflict, deleteErr
		} else {
			return http.StatusInternalServerError, deleteErr
		}
	}
	return http.StatusNoContent, deleteErr
}

func (h *remoteStoreIntentHandler) deleteAppPIntent(appIntentName string, p string, ca string, v string,
	gpintName string, digName string) (interface{}, interface{}) {

	orch := h.orchInstance
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-placement-intents/" + gpintName + "/app-intents/" + appIntentName
	status, err := orch.apiDel(url, gpintName)
	return status, err
}

func (h *localStoreIntentHandler) deleteGpint(gpintName string, p string, ca string,
	v string, digName string) (interface{}, interface{}) {
	c := localstore.NewGenericPlacementIntentClient()

	err := c.DeleteGenericPlacementIntent(gpintName, p, ca, v, digName)
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

	return http.StatusNoContent, nil
}

func (h *remoteStoreIntentHandler) deleteGpint(gpintName string, p string, ca string,
	v string, digName string) (interface{}, interface{}) {
	orch := h.orchInstance
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName + "/generic-placement-intents/" + gpintName
	resp, err := orch.apiDel(orchURL, gpintName)
	return resp, err
}

func (h *localStoreIntentHandler) createAppPIntent(pint localstore.AppIntent, p string, ca string, v string,
	digName string, gpintName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewAppIntentClient()
	_, createErr := c.CreateAppIntent(pint, p, ca, v, gpintName, digName)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the project") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the composite-app") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the intent") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "Unable to find the deploymentIntentGroupName") {
			return http.StatusNotFound, createErr
		} else if strings.Contains(createErr.Error(), "AppIntent already exists") {
			return http.StatusConflict, createErr
		} else {
			return http.StatusInternalServerError, createErr
		}
	}
	return http.StatusCreated, createErr
}

func (h *remoteStoreIntentHandler) createAppPIntent(pint localstore.AppIntent, p string, ca string, v string,
	digName string, gpintName string) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(pint)
	orchURL := "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-placement-intents/" + gpintName + "/app-intents"
	status, err := orch.apiPost(jsonLoad, url, ca+"_gpint")
	return status, err
}

func (h *localStoreIntentHandler) createGenericK8sIntent(gki localstore.GenericK8sIntent, p string, ca string,
	v string, digName string, exists bool) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewGenericK8sIntentClient()
	_, err := c.CreateGenericK8sIntent(gki, p, ca, v, digName, exists)
	if err != nil {
		log.Error(":: CreateGenericK8sIntent error ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "GenericK8sIntent already exists") {
			return http.StatusConflict, nil
		} else {
			return http.StatusInternalServerError, err.Error()
		}
	}
	return http.StatusCreated, err
}

func (h *remoteStoreIntentHandler) createGenericK8sIntent(gki localstore.GenericK8sIntent, p string, ca string,
	v string, digName string, exists bool) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(gki)
	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents"
	status, err := orch.apiPost(jsonLoad, url, ca+"_genk8sint")
	return status, err
}
func (h *localStoreIntentHandler) deleteGenericK8sIntent(gkiName string, p string, ca string,
	v string, digName string) (interface{}, interface{}) {
	// Get the local store handler.
	c := localstore.NewGenericK8sIntentClient()
	err := c.DeleteGenericK8sIntent(gkiName, p, ca, v, digName)
	if err != nil {
		log.Error(":: DeleteGenericK8sIntent failure ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound, err.Error()
		} else if strings.Contains(err.Error(), "conflict") {
			return http.StatusConflict, err.Error()
		} else {
			return http.StatusInternalServerError, err.Error()
		}
	}
	return http.StatusNoContent, err
}

func (h *remoteStoreIntentHandler) deleteGenericK8sIntent(gkiName string, p string, ca string,
	v string, digName string) (interface{}, interface{}) {
	orch := h.orchInstance
	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gkiName
	status, err := orch.apiDel(url, ca+"_genk8sint")
	return status, err
}

func (h *localStoreIntentHandler) createResource(r localstore.Resource, rc localstore.ResourceFileContent, fName string, p string, ca string,
	v string, digName string, gi string, exists bool) (interface{}, interface{}) {
	// Get the local store handler
	c := localstore.NewResourceClient()
	_, err := c.CreateResource(r, rc, p, ca, v, digName, gi, false)
	if err != nil {
		log.Error(":: Creation resource failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "resource already exists") {
			return http.StatusConflict, err.Error()
		} else {
			return http.StatusInternalServerError, err.Error()
		}
	}
	return http.StatusCreated, err
}

func (h *remoteStoreIntentHandler) createResource(r localstore.Resource, rc localstore.ResourceFileContent, fName string, p string, ca string,
	v string, digName string, gi string, exists bool) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(r)
	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources"

	var fileNames []string
	if fName == "" {
		fName = "resourceFile.yaml"
	}
	fileNames = append(fileNames, fName)
	var fileContents []string
	fileContents = append(fileContents, rc.FileContent)

	status, err := orch.apiPostMultipart(jsonLoad, nil, url, ca+"_"+r.Metadata.Name, fileNames, fileContents)
	return status, err
}

func (h *localStoreIntentHandler) createCustomization(cz localstore.Customization, t localstore.SpecFileContent, p string, ca string,
	v string, digName string, gi string, rs string, exists bool) (interface{}, interface{}) {
	// Get the local store handler
	c := localstore.NewCustomizationClient()
	_, err := c.CreateCustomization(cz, t, p, ca, v, digName, gi, rs, false)
	if err != nil {
		log.Error(":: Create customization failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Customization already exists") {
			return http.StatusConflict, err.Error()
		} else {
			return http.StatusInternalServerError, err.Error()
		}
	}
	return http.StatusCreated, err
}

func (h *remoteStoreIntentHandler) createCustomization(c localstore.Customization, t localstore.SpecFileContent, p string, ca string,
	v string, digName string, gi string, rs string, exists bool) (interface{}, interface{}) {
	orch := h.orchInstance
	jsonLoad, _ := json.Marshal(c)

	orch.Vars["multipartfiles"] = "true"

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rs + "/customizations"

	status, err := orch.apiPostMultipart(jsonLoad, nil, url, ca+"_"+c.Metadata.Name, t.FileNames, t.FileContents)
	orch.Vars["multipartfiles"] = "false"
	return status, err

}
func (h *localStoreIntentHandler) getAllResources(p string, ca string, v string, digName string, gi string) (interface{}, []byte, interface{}) {
	// Get the local store handler
	c := localstore.NewResourceClient()
	var retval []byte
	var brList []localstore.Resource

	ret, err := c.GetAllResources(p, ca, v, digName, gi)
	if err != nil {
		log.Error(":: GetAllResources failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err.Error()
		} else {
			return http.StatusInternalServerError, retval, err.Error()
		}
	}
	for _, br := range ret {
		brList = append(brList, localstore.Resource{Metadata: br.Metadata, Spec: br.Spec})
	}
	retval, _ = json.Marshal(brList)
	return http.StatusOK, retval, nil
}

func (h *remoteStoreIntentHandler) getAllResources(p string, ca string, v string, digName string, gi string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources"
	retcode, retval, err := orch.apiGet(url, ca+"_getGenk8sResources")
	return retcode, retval, err
}

func (h *localStoreIntentHandler) getResource(rName, p, ca, v, digName, gi string) (interface{}, []byte, interface{}) {
	// Get the local store handler
	c := localstore.NewResourceClient()

	var resource localstore.Resource
	var retval []byte
	resource, err := c.GetResource(rName, p, ca, v, digName, gi)
	if err != nil {
		log.Error(":: GetResource failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err.Error()
		} else {
			return http.StatusInternalServerError, retval, err.Error()
		}
	}
	retval, _ = json.Marshal(resource)
	return http.StatusOK, retval, nil
}

func (h *remoteStoreIntentHandler) getResource(rName, p, ca, v, digName, gi string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rName
	retcode, retval, err := orch.apiGet(url, ca+"_getGenk8sResource")
	return retcode, retval, err
}

func (h *localStoreIntentHandler) deleteResource(rName, p, ca, v, digName, gi string) (interface{}, interface{}) {
	// Get the local store handler
	c := localstore.NewResourceClient()

	err := c.DeleteResource(rName, p, ca, v, digName, gi)
	if err != nil {
		log.Error(":: DeleteResource failure ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound, err.Error()
		} else if strings.Contains(err.Error(), "conflict") {
			return http.StatusConflict, err.Error()
		} else {
			return http.StatusInternalServerError, err.Error()
		}
	}
	return http.StatusNoContent, nil
}

func (h *remoteStoreIntentHandler) deleteResource(rName, p, ca, v, digName, gi string) (interface{}, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rName
	retcode, err := orch.apiDel(url, ca+"_delGenk8sIntResource")
	return retcode, err
}

func (h *localStoreIntentHandler) getAllCustomization(p, ca, v, digName, gi, rs string) (interface{}, []byte, interface{}) {
	// Get the local store handler
	c := localstore.NewCustomizationClient()

	var czList []localstore.Customization
	var retval []byte

	ret, err := c.GetAllCustomization(p, ca, v, digName, gi, rs)
	if err != nil {
		log.Error(":: GetAllCustomization failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err.Error
		} else {
			return http.StatusInternalServerError, retval, err.Error()
		}
	}

	for _, cz := range ret {
		czList = append(czList, localstore.Customization{Metadata: cz.Metadata, Spec: cz.Spec})
	}
	retval, _ = json.Marshal(czList)
	return http.StatusOK, retval, nil
}

func (h *remoteStoreIntentHandler) getAllCustomization(p, ca, v, digName, gi, rs string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rs + "/customizations"
	retcode, retval, err := orch.apiGet(url, ca+"_getGenk8sIntCustomizations")
	return retcode, retval, err
}

func (h *localStoreIntentHandler) getCustomization(cz, p, ca, v, digName, gi, rs string) (interface{}, []byte, interface{}) {
	// Get the local store handler
	c := localstore.NewCustomizationClient()
	var retval []byte

	cusz, err := c.GetCustomization(cz, p, ca, v, digName, gi, rs)
	if err != nil {
		log.Error(":: GetCustomization failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			return http.StatusNotFound, retval, err.Error()
		} else {
			return http.StatusInternalServerError, retval, err.Error()
		}
	}
	retval, _ = json.Marshal(cusz)
	return http.StatusOK, retval, nil
}

func (h *remoteStoreIntentHandler) getCustomization(cz, p, ca, v, digName, gi, rs string) (interface{}, []byte, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rs + "/customizations/" + cz
	retcode, retval, err := orch.apiGet(url, ca+"_getGenk8sIntCustomization")
	return retcode, retval, err
}

func (h *localStoreIntentHandler) deleteCustomization(cz, p, ca, v, digName, gi, rs string) (interface{}, interface{}) {
	// Get the local store handler
	c := localstore.NewCustomizationClient()

	err := c.DeleteCustomization(cz, p, ca, v, digName, gi, rs)
	if err != nil {
		log.Error(":: DeleteCustomization failure ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound, err.Error()
		} else if strings.Contains(err.Error(), "conflict") {
			return http.StatusConflict, err.Error()
		}
	}
	return http.StatusNoContent, nil
}

func (h *remoteStoreIntentHandler) deleteCustomization(cz, p, ca, v, digName, gi, rs string) (interface{}, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rs + "/customizations/" + cz
	retcode, err := orch.apiDel(url, ca+"_delGenk8sIntCustomization")
	return retcode, err
}

func(h *localStoreIntentHandler) getResourceContent(rName, p, ca, v, digName, gi string)(interface{}, []byte, interface{}) {
	// Get the local store handler
	c := localstore.NewResourceClient()

	var retval []byte
	retBrContent, err := c.GetResourceContent(rName, p, ca, v, digName, gi)
	if err != nil {
		log.Error("Error encountered while fetching resource file content: %s", err)
		return http.StatusInternalServerError, retval, err.Error()
	}
	retval = []byte(retBrContent.FileContent)
	return http.StatusOK, retval, nil
}

func(h *remoteStoreIntentHandler) getResourceContent(rName, p, ca, v, digName, gi string)(interface{}, []byte, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rName
	retcode, retval, err := orch.apiGetMultiPart(url, ca+"_getGenk8sIntResourceContent")
	return retcode, retval, err
}

func(h *localStoreIntentHandler) getCustomizationContent(cz, p, ca, v, digName, gi, rs string)(interface{}, []byte, interface{}) {
	// Get the local store handler
	c := localstore.NewCustomizationClient()

	var retval []byte
	specFC, err := c.GetCustomizationContent(cz, p, ca, v, digName, gi, rs)
	if err != nil {
		log.Error("Error encountered while fetching customization file content: %s", err)
		return http.StatusInternalServerError, retval, err.Error()
	}
	/*if len(specFC.FileContents) > 0 {
		retval = []byte(specFC.FileContents[0])
	}*/
	retval, _ = json.Marshal(specFC)
	return http.StatusOK, retval, nil
}

func(h *remoteStoreIntentHandler) getCustomizationContent(c, p, ca, v, digName, gi, rs string)(interface{}, []byte, interface{}) {
	orch := h.orchInstance

	orchURL := "http://" + orch.MiddleendConf.Gac + "/v2/projects/" + p +
		"/composite-apps/" + ca + "/" + v +
		"/deployment-intent-groups/" + digName
	url := orchURL + "/generic-k8s-intents/" + gi + "/resources/" + rs + "/customizations/" + c
	retcode, retval, err := orch.apiGetMultiPart(url, ca+"_getGenk8sIntCustomizationContent")
	return retcode, retval, err
}

func (h *placementIntentHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	retcode := 200
	dataRead := h.orchInstance.dataRead
	project := vars["projectName"]
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		Apps := compositeAppValue.AppsDataArray
		for digName, digValue := range Dig {
			for gpintName, gpintValue := range digValue.GpintMap {
				for appName, _ := range Apps {
					var appPint localstore.AppIntent
					retcode, retval, err := orch.bstore.getAppPIntent(appName+"_pint", gpintName,
						project, compositeAppMetadata.Name, compositeAppSpec.Version, digName)
					log.Infof("Get Gpint App intent in Composite app %s dig %s Gpint %s status: %d",
						vars["compositeAppName"], digName, gpintName, retcode)
					if err != nil {
						log.Error("Failed to read app pint\n")
						return nil, retcode
					}
					if retcode != http.StatusOK {
						log.Error("Failed to read app pint\n")
						return nil, retcode
					}
					err = json.Unmarshal(retval, &appPint)
					if err != nil {
						log.Errorf("Failed to unmarshal json %s\n", err)
						return nil, http.StatusInternalServerError
					}
					gpintValue.AppIntentArray = append(gpintValue.AppIntentArray, appPint)
				}
			}
		}
	}
	return nil, retcode
}

func (h *placementIntentHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	retcode := 200

	dataRead := h.orchInstance.dataRead
	project := vars["projectName"]

	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			var gpintList []localstore.GenericPlacementIntent
			retcode, retval, err := orch.bstore.getAllGPint(project, compositeAppMetadata.Name,
				compositeAppSpec.Version, digName)
			log.Infof("Get Gpint in Composite app %s dig %s status: %d", vars["compositeAppName"],
				digName, retcode)
			if err != nil {
				log.Error("Failed to read gpint\n")
				return nil, retcode
			}
			if retcode != http.StatusOK {
				log.Error("Failed to read gpint\n")
				return nil, retcode
			}
			json.Unmarshal(retval, &gpintList)
			digValue.GpintMap = make(map[string]*GpintData, len(gpintList))
			for _, value := range gpintList {
				var GpintDataInstance GpintData
				GpintDataInstance.Gpint = value
				digValue.GpintMap[value.MetaData.Name] = &GpintDataInstance
			}
		}
	}
	return nil, retcode
}

func (h *placementIntentHandler) deleteObject() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		Apps := compositeAppValue.AppsDataArray

		// loop through all app intens in the gpint
		for digName, digValue := range Dig {
			for gpintName, _ := range digValue.GpintMap {
				for appName, _ := range Apps {
					// query based on app name.
					resp, err := orch.bstore.deleteAppPIntent(appName+"_pint", vars["projectName"],
						compositeAppMetadata.Name, compositeAppSpec.Version, gpintName, digName)
					if err != nil {
						return err
					}
					if resp != http.StatusNoContent {
						return resp
					}
					log.Infof("Delete gpint intents response: %d", resp)
				}
			}
		}
	}
	return nil
}

func (h placementIntentHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap

		// loop through all app intens in the gpint
		for digName, digValue := range Dig {
			for gpintName, _ := range digValue.GpintMap {
				log.Infof("Delete gpint  %s", h.orchURL)
				resp, err := orch.bstore.deleteGpint(gpintName, vars["projectName"],
					compositeAppMetadata.Name, compositeAppSpec.Version, digName)
				if err != nil {
					return err
				}
				if resp != http.StatusNoContent {
					return resp
				}
				log.Infof("Delete gpint response: %d", resp)
			}
		}
	}
	return nil
}

func (h *placementIntentHandler) createAnchor() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData
	gPintName := intentData.CompositeAppName + "_gpint"

	vars := orch.Vars
	projectName := vars["projectName"]
	version := vars["version"]
	digName := intentData.Name

	gpi := localstore.GenericPlacementIntent{
		MetaData: localstore.GenIntentMetaData{
			Name:        gPintName,
			Description: "Generic placement intent created from middleend",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}

	// POST the generic placement intent
	log.Infof("compositeAppName %s", intentData.CompositeAppName)
	resp, err := orch.bstore.createGpint(gpi, projectName, intentData.CompositeAppName, version, digName)
	jsonLoad, _ := json.Marshal(gpi)
	orch.response.payload[intentData.CompositeAppName+"_gpint"] = jsonLoad
	orch.response.status[intentData.CompositeAppName+"_gpint"] = resp.(int)
	if err != nil {
		return err
	}
	if resp != http.StatusCreated {
		return resp
	}
	log.Infof("Generic placement intent response: %d", resp)

	return nil
}

func (h *placementIntentHandler) createObject() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData
	vars := orch.Vars
	projectName := vars["projectName"]
	version := vars["version"]
	digName := vars["deploymentIntentGroupName"]

	for _, app := range intentData.Spec.Apps {
		appName := app.Metadata.Name
		intentName := appName + "_pint"
		genericAppIntentName :=  intentData.CompositeAppName + "_gpint"

		// Initialize the base structure and then add the cluster values,
		// we support only allof for now.
		var customData string
		if orch.Vars["update-intent"] == "yes" {
			customData = "updated"
		} else {
			customData = "data 1"
		}
		pint := localstore.AppIntent{
			MetaData: localstore.MetaData{
				Name:        intentName,
				Description: "NA",
				UserData1:   customData,
				UserData2:   "data2"},
			Spec: localstore.SpecData{
				AppName: appName,
				Intent:  localstore.IntentStruc{},
			},
		}

		for _, clusterProvider := range app.Clusters {
			if len(clusterProvider.SelectedClusters) > 0 {
				for _, cluster := range clusterProvider.SelectedClusters {
					if app.PlacementCriterion == "allOf" {
						allOfClusters := localstore.AllOf{}
						allOfClusters.ProviderName = clusterProvider.Provider
						allOfClusters.ClusterName = cluster.Name
						pint.Spec.Intent.AllOfArray = append(pint.Spec.Intent.AllOfArray, allOfClusters)
					} else {
						anyOfClusters := localstore.AnyOf{}
						anyOfClusters.ProviderName = clusterProvider.Provider
						anyOfClusters.ClusterName = cluster.Name
						pint.Spec.Intent.AnyOfArray = append(pint.Spec.Intent.AnyOfArray, anyOfClusters)
					}
				}
			} else {
				for _, label := range clusterProvider.SelectedLabels {
					if app.PlacementCriterion == "allOf" {
						allOfClusters := localstore.AllOf{}
						allOfClusters.ProviderName = clusterProvider.Provider
						allOfClusters.ClusterLabelName = label.Name
						pint.Spec.Intent.AllOfArray = append(pint.Spec.Intent.AllOfArray, allOfClusters)
					} else {
						anyOfClusters := localstore.AnyOf{}
						anyOfClusters.ProviderName = clusterProvider.Provider
						anyOfClusters.ClusterLabelName = label.Name
						pint.Spec.Intent.AnyOfArray = append(pint.Spec.Intent.AnyOfArray, anyOfClusters)
					}
				}
			}
		}
		log.Debugf("pint is: %+v", pint)
		status, err := orch.bstore.createAppPIntent(pint, projectName, intentData.CompositeAppName, version, digName, genericAppIntentName)
		jsonLoad, _ := json.Marshal(pint)
		orch.response.payload[genericAppIntentName] = jsonLoad
		orch.response.status[genericAppIntentName] = status.(int)
		if err != nil {
			log.Fatalln(err)
		}
		if status != http.StatusCreated {
			return status
		}
		log.Infof("Placement intent %s status: %d", intentName, status)
	}
	return nil
}

func addPlacementIntent(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.createAnchor()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.createObject()
	if err != nil {
		return err
	}
	return nil
}

func delGpint(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.deleteObject()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.deleteAnchor()
	if err != nil {
		return err
	}
	return nil
}

func (h *networkIntentHandler) createAnchor() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData

	nwCtlIntentName := intentData.CompositeAppName + "_nwctlint"

	nwIntent := localstore.NetControlIntent{
		Metadata: localstore.Metadata{
			Name:        nwCtlIntentName,
			Description: "Network Controller created from middleend",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}
	resp, err := orch.bstore.createControllerIntent(nwIntent, intentData.Spec.ProjectName, intentData.CompositeAppName,
		intentData.CompositeAppVersion, intentData.Name, false, nwCtlIntentName)
	jsonLoad, _ := json.Marshal(nwIntent)
	orch.response.payload[nwCtlIntentName] = jsonLoad
	orch.response.status[nwCtlIntentName] = resp.(int)
	if err != nil {
		return err
	}
	if resp != http.StatusCreated {
		return resp
	}
	log.Infof("Network controller intent response: %d", resp)

	return nil
}

func (h *networkIntentHandler) createObject() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData
	vars := orch.Vars
	projectName := vars["projectName"]
	version := vars["version"]
	genericAppIntentName := intentData.CompositeAppName + "_nwctlint"
	digName := vars["deploymentIntentGroupName"]

	for _, app := range intentData.Spec.Apps {
		// Check if the application has any interfaces.
		// There is assumption that if an application must have same interfaces
		// specified in each cluster.
		if len(app.Interfaces) == 0 {
			continue
		}

		appName := app.Metadata.Name
		workloadIntentName := appName + "_wlint"

		var customData string
		if orch.Vars["update-intent"] == "yes" {
			customData = "updated"
		} else {
			customData = "data 1"
		}

		wlIntent := localstore.WorkloadIntent{
			Metadata: localstore.Metadata{
				Name:        workloadIntentName,
				Description: "NA",
				UserData1:   customData,
				UserData2:   "data2"},
			Spec: localstore.WorkloadIntentSpec{
				AppName:          appName,
				WorkloadResource: intentData.DigVersion + "-" + appName,
				Type:             "Deployment",
			},
		}

		status, err := orch.bstore.createWorkloadIntent(wlIntent, projectName, intentData.CompositeAppName, 
			version, digName, genericAppIntentName, false, workloadIntentName)
		jsonLoad, _ := json.Marshal(wlIntent)
		orch.response.payload[workloadIntentName] = jsonLoad
		orch.response.status[workloadIntentName] = status.(int)
		if err != nil {
			log.Fatalln(err)
		}
		if status != http.StatusCreated {
			return status
		}
		log.Infof("Workload intent %s status: %d", workloadIntentName, status)

		// Create interfaces for each per app workload intent.
		for i, iface := range app.Interfaces {
			interfaceNum := strconv.Itoa(i + 1)
			interfaceName := app.Metadata.Name + "_interface" + interfaceNum

			var netInterfaceName string
			if len(iface.InterfaceName) > 0 {
				netInterfaceName = iface.InterfaceName
			} else {
				netInterfaceName = "net" + interfaceNum
			}

			nwiface := localstore.WorkloadIfIntent{
				Metadata: localstore.Metadata{
					Name:        interfaceName,
					Description: "NA",
					UserData1:   "data1",
					UserData2:   "data2"},
				Spec: localstore.WorkloadIfIntentSpec{
					IfName:         netInterfaceName,
					NetworkName:    iface.NetworkName,
					DefaultGateway: "false",
					IpAddr:         iface.IP,
				},
			}

			status, err := orch.bstore.createWorkloadIfIntent(nwiface, projectName, intentData.CompositeAppName,
				version, digName, genericAppIntentName, workloadIntentName, false, interfaceName)
			jsonLoad, _ := json.Marshal(nwiface)
			orch.response.payload[interfaceName] = jsonLoad
			orch.response.status[interfaceName] = status.(int)
			if err != nil {
				log.Fatalln(err)
			}
			if status != http.StatusCreated {
				return status
			}
			log.Infof("interface %s status: %d ", interfaceName, status)
		}
	}

	return nil
}

func (h *networkIntentHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			for nwintName, nwintValue := range digValue.NwintMap {
				var wrlintList []NetworkWlIntent
				retcode, retval, err := orch.bstore.getWorkloadIntents(projectName, compositeAppMetadata.Name,
					compositeAppSpec.Version, digName, nwintName)
				log.Infof("Get Wrkld intents in Composite app %s dig %s nw intent %s status: %d",
					compositeAppName, digName, nwintName, retcode)
				if err != nil {
					log.Error("Failed to read nw  workload int")
					return nil, retcode
				}
				if retcode != http.StatusOK {
					log.Error("Failed to read nw  workload int")
					return nil, retcode
				}
				json.Unmarshal(retval, &wrlintList)
				nwintValue.WrkintMap = make(map[string]*WrkintData, len(wrlintList))
				for _, wrlIntValue := range wrlintList {
					var WrkintDataInstance WrkintData
					WrkintDataInstance.Wrkint = wrlIntValue

					var ifaceList []NwInterface
					log.Infof("Get interface in Composite app %s dig %s nw intent %s wrkld intent %s status: %d",
						compositeAppName, digName, nwintName, wrlIntValue.Metadata.Name, retcode)
					retcode, retval, err := orch.bstore.getWorkloadIfIntents(projectName, compositeAppMetadata.Name,
						compositeAppSpec.Version, digName, nwintName, wrlIntValue.Metadata.Name)
					if err != nil {
						log.Error("Failed to read nw interface")
						return nil, retcode
					}
					if retcode != http.StatusOK {
						log.Error("Failed to read nw interface")
						return nil, retcode
					}
					json.Unmarshal(retval, &ifaceList)
					WrkintDataInstance.Interfaces = ifaceList
					nwintValue.WrkintMap[wrlIntValue.Metadata.Name] = &WrkintDataInstance
				}
			}
		}
	}
	return nil, retcode
}

func (h *networkIntentHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			var nwintList []NetworkCtlIntent

			retcode, retval, err := orch.bstore.getControllerIntents(projectName, compositeAppMetadata.Name, compositeAppSpec.Version, digName)
			log.Infof("Get Network Ctl intent in Composite app %s dig %s status: %d",
				compositeAppName, digName, retcode)
			if err != nil {
				log.Errorf("Failed to read nw int %s\n", err)
				return nil, retcode
			}
			if retcode != http.StatusOK {
				log.Error("Failed to read nw int")
				return nil, retcode
			}
			json.Unmarshal(retval, &nwintList)
			digValue.NwintMap = make(map[string]*NwintData, len(nwintList))
			for _, nwIntValue := range nwintList {
				var NwintDataInstance NwintData
				NwintDataInstance.Nwint = nwIntValue
				digValue.NwintMap[nwIntValue.Metadata.Name] = &NwintDataInstance
			}
		}
	}
	return nil, retcode
}

func (h *networkIntentHandler) deleteObject() interface{} {
	orch := h.orchInstance
	retcode := 200
	vars := orch.Vars
	projectName := vars["projectName"]
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" +
				projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName

			for nwintName, nwintValue := range digValue.NwintMap {
				for wrkintName, wrkintValue := range nwintValue.WrkintMap {
					// Delete the interfaces per workload intent.
					for _, value := range wrkintValue.Interfaces {
						retcode, err := orch.bstore.deleteWorkloadIfIntent(value.Metadata.Name, wrkintName,
							projectName, compositeAppMetadata.Name,
							compositeAppSpec.Version, digName, nwintName)
						if err != nil {
							return err
						}
						if retcode != http.StatusNoContent {
							return retcode
						}
						log.Infof("Delete nw interface response: %d", retcode)
					}
					// Delete the workload intents.
					url := h.ovnURL + "network-controller-intent/" + nwintName + "/workload-intents/" + wrkintName
					log.Infof("Delete app nw wl intent %s", url)
					retcode, err := orch.bstore.deleteWorkloadIntent(wrkintName, projectName, compositeAppMetadata.Name,
						compositeAppSpec.Version, digName, nwintName)
					log.Infof("Delete nw wl intent response: %d", retcode)
					if err != nil {
						return err
					}
					if retcode != http.StatusNoContent {
						return retcode
					}
				} // For workload intents in network controller intent.
			} // For network controller intents in Dig.
		} // For Dig.
	} // For composite app.
	return retcode
}

func (h networkIntentHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		if compositeAppValue.Status == "checkout" {
			continue
		}
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" +
				projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName
			for nwintName, _ := range digValue.NwintMap {
				// loop through all app intens in the gpint
				retcode, err := orch.bstore.deleteControllerIntent(nwintName, projectName, compositeAppMetadata.Name,
					compositeAppSpec.Version, digName)
				log.Infof("Delete nw controller intent response: %d", retcode)
				if err != nil {
					return err
				}
				if retcode != http.StatusNoContent {
					return retcode
				}
			}
		}
	}
	return retcode
}

func addNetworkIntent(I orchWorkflow) interface{} {
	//1. Add network controller Intent
	err := I.createAnchor()
	if err != nil {
		return err
	}

	//2. Add network workload intent
	err = I.createObject()
	if err != nil {
		return err
	}

	return nil
}

func delNwintData(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.deleteObject()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.deleteAnchor()
	if err != nil {
		return err
	}
	return nil
}

func (h genericK8sIntentHandler) createAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	version := vars["version"]
	genericK8sIntentName := compositeAppName + "_genk8sint"
	intentData := orch.DigData
	digName := intentData.Name

	// Validate if any genericK8sIntent specifications are provided
	var createGenK8sIntent bool
	for _, appData := range orch.DigData.Spec.Apps {
		for _, _ = range appData.RsInfo {
			createGenK8sIntent = true
		}
	}

	if createGenK8sIntent {
		gki := localstore.GenericK8sIntent{
			Metadata: localstore.Metadata{
				Name:        genericK8sIntentName,
				Description: "generic K8s intent",
				UserData1:   "data1",
				UserData2:   "data2"},
		}

		retcode, err := orch.bstore.createGenericK8sIntent(gki, projectName, compositeAppName, version, digName, false)
		log.Infof("Creation of generic K8s intent response: %s", retcode)
		if err != nil {
			return err
		}
		if retcode != http.StatusCreated && retcode != http.StatusConflict {
			return retcode
		}
	}
	return nil
}

func (h genericK8sIntentHandler) createObject() interface{} {
	// Create resource object
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	version := vars["version"]
	intentData := orch.DigData
	digName := intentData.Name

	for _, appData := range orch.DigData.Spec.Apps {
		for _, resObj := range appData.RsInfo {
			resourceName := compositeAppName + "_resource_" + uuid.New().String()
			resource := localstore.Resource{
				Metadata: localstore.Metadata{
					Name:        resourceName,
					Description: "NA",
					UserData1:   "data1",
					UserData2:   "data2"},
				Spec: localstore.ResourceSpec{
					AppName:     appData.Metadata.Name,
					NewObject:   resObj.ResourceSpec.NewObject,
					ResourceGVK: resObj.ResourceSpec.ResourceGVK,
				},
			}

			retcode, err := orch.bstore.createResource(resource, resObj.ResourceFile, resObj.ResourceFileName, projectName, compositeAppName,
				version, digName, compositeAppName+"_genk8sint", false)
			log.Infof("Creation of resource response: %s", retcode)
			if err != nil {
				return err
			}
			if retcode != nil && retcode.(int) != http.StatusCreated {
				return retcode.(int)
			}

			var cInfo localstore.ClusterInfo
			if resObj.CustomizationSpec.ClusterSpecific == "false" {
				cInfo = localstore.ClusterInfo{
					Scope:           "label",
					ClusterProvider: "xxx",
					ClusterName:     "dummy",
					ClusterLabel:    "dummy",
					Mode:            "allow",
				}
			} else {
				cInfo = localstore.ClusterInfo{
					Scope:           resObj.CustomizationSpec.ClusterInfo.Scope,
					ClusterProvider: resObj.CustomizationSpec.ClusterInfo.ClusterProvider,
					ClusterName:     resObj.CustomizationSpec.ClusterInfo.ClusterName,
					ClusterLabel:    resObj.CustomizationSpec.ClusterInfo.ClusterLabel,
					Mode:            resObj.CustomizationSpec.ClusterInfo.Mode,
				}
			}
			// Create customization object
			var cFile string
			if len(resObj.CustomFile.FileNames) > 0 {
				cFile = strings.Join(resObj.CustomFile.FileNames, ",")
			}
			customizationName := compositeAppName + "_custom_" + uuid.New().String()
			customization := localstore.Customization{
				Metadata: localstore.Metadata{
					Name:        customizationName,
					Description: "NA",
					UserData1:   cFile,
					UserData2:   resObj.ResourceSpec.ResourceGVK.Kind,
				},
				Spec: localstore.CustomizeSpec{
					ClusterSpecific: resObj.CustomizationSpec.ClusterSpecific,
					ClusterInfo:     cInfo,
					PatchType:       resObj.CustomizationSpec.PatchType,
					PatchJSON:       resObj.CustomizationSpec.PatchJSON,
				},
			}

			retcode, err = orch.bstore.createCustomization(customization, resObj.CustomFile, projectName, compositeAppName,
				version, digName, compositeAppName+"_genk8sint", resourceName, false)
			log.Infof("Creation of customization response: %s", retcode)
			if err != nil {
				return err
			}
			if retcode != http.StatusCreated {
				return retcode
			}
		}
	}
	return nil
}

func (h genericK8sIntentHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	version := vars["version"]
	digName := vars["deploymentIntentGroupName"]

	var brList []localstore.Resource
	retcode, retval, err := orch.bstore.getAllResources(projectName, compositeAppName,
		version, digName, compositeAppName+"_genk8sint")
	log.Infof("GetAllResources response: %s", retcode)
	if err != nil {
		return err, retcode
	}
	if retcode != http.StatusOK {
		return err, retcode
	}

	json.Unmarshal(retval, &brList)
	var genK8sInfo GenericK8sIntentInfo
	orch.genK8sInfo = make(map[string]*GenericK8sIntentInfo)
	genK8sInfo.listGenK8sData.resource = make([]localstore.Resource, len(brList))
	genK8sInfo.listGenK8sData.resource = brList
	log.Infof("resources: %+v", brList)
	h.orchInstance.genK8sInfo[compositeAppName+"_genk8sint"] = &genK8sInfo

	return nil, retcode
}

func (h genericK8sIntentHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	version := vars["version"]
	digName := vars["deploymentIntentGroupName"]

	for _, genK8sRes := range orch.genK8sInfo {
		genK8sRes.listGenK8sData.resMap = make(map[string][]localstore.Customization)
		genK8sRes.resData = make(map[string][]ResourceInfo)
		for _, res := range genK8sRes.listGenK8sData.resource {
			// Fetch resourceContent if any for given resource
			retcode, retval, err := orch.bstore.getResourceContent(res.Metadata.Name, projectName, compositeAppName,
				version, digName, compositeAppName+"_genk8sint")
			if err != nil {
				return err, retcode
			}
			if retcode != http.StatusOK {
				return err, retcode
			}
			var rFile localstore.ResourceFileContent
			rFile.FileContent = string(retval)

			var cList []localstore.Customization
			retcode, retvalue, err := orch.bstore.getAllCustomization(projectName, compositeAppName,
				version, digName, compositeAppName+"_genk8sint", res.Metadata.Name)
			log.Infof("GetAllCustomization response: %s", retcode)
			if err != nil {
				return err, retcode
			}
			if retcode != http.StatusOK {
				return err, retcode
			}
			json.Unmarshal(retvalue, &cList)
			log.Debugf("customization: %+v", cList)
			genK8sRes.listGenK8sData.resMap[res.Metadata.Name] = cList

			// Populate resData structure
			var resInfo ResourceInfo
			resInfo.ResourceSpec.ResourceGVK = res.Spec.ResourceGVK
			resInfo.ResourceSpec.NewObject = res.Spec.NewObject
			resInfo.ResourceFile.FileContent = rFile.FileContent

			// Iterate over all customization objects
			for _, cz := range cList {
				// Fetch customizationContent if any for given customization
				retcode, retval, err = orch.bstore.getCustomizationContent(cz.Metadata.Name, projectName, compositeAppName,
					version, digName, compositeAppName+"_genk8sint", res.Metadata.Name)
				if err != nil {
					return err, retcode
				}
				if retcode != http.StatusOK {
					return err, retcode
				}

				var cSpecContent localstore.SpecFileContent
				json.Unmarshal(retval, &cSpecContent)
				resInfo.CustomizationSpec = cz.Spec
				if len(cz.Metadata.UserData1) > 0 {
					cSpecContent.FileNames = strings.Split(cz.Metadata.UserData1, ",")
				}
				resInfo.CustomFile = cSpecContent
			}
			genK8sRes.resData[res.Spec.AppName] = append(genK8sRes.resData[res.Spec.AppName], resInfo)
		}
	}

	return nil, http.StatusOK
}

func (h genericK8sIntentHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	version := vars["version"]
	digName := vars["deploymentIntentGroupName"]

	// Delete genericK8sIntent belonging to DIG
	retcode, err := orch.bstore.deleteGenericK8sIntent(compositeAppName+"_genk8sint", projectName, compositeAppName,
		version, digName)
	log.Infof("delete genericGenK8sIntent response: %s", retcode)
	if err != nil {
		return err
	}
	if retcode != nil && retcode.(int) != http.StatusNoContent {
		return retcode.(int)
	}

	return nil
}

func (h genericK8sIntentHandler) deleteObject() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	projectName := vars["projectName"]
	compositeAppName := vars["compositeAppName"]
	version := vars["version"]
	digName := vars["deploymentIntentGroupName"]

	genK8sInfo := orch.genK8sInfo[compositeAppName+"_genk8sint"]
	log.Debugf("genK8sInfo: %+v", genK8sInfo)

	// Delete resources and customization objects belonging to DIG
	for resObj, czObjList := range genK8sInfo.listGenK8sData.resMap {
		for _, czObj := range czObjList {
			retcode, _ := orch.bstore.deleteCustomization(czObj.Metadata.Name, projectName,
				compositeAppName,
				version, digName, compositeAppName+"_genk8sint", resObj)
			log.Infof("deleteCustomization response: %s", retcode)
			if retcode != nil && retcode.(int) != http.StatusNoContent {
				return retcode.(int)
			}
		}
		retcode, _ := orch.bstore.deleteResource(resObj, orch.Vars["projectName"], compositeAppName,
			version, digName, compositeAppName+"_genk8sint")
		log.Infof("deleteResource response: %s", retcode)
		if retcode != nil && retcode.(int) != http.StatusNoContent {
			return retcode.(int)
		}
	}

	return nil
}

func addGenericK8sIntent(I orchWorkflow) interface{} {
	// 1. Add genericK8s intent
	err := I.createAnchor()
	if err != nil {
		return err
	}

	// 2. Add resource and customization object
	err = I.createObject()
	if err != nil {
		return err
	}

	return nil
}

func getGenericK8sIntent(I orchWorkflow) interface{} {
	// 1. Add genericK8s intent resource
	_, ret := I.getAnchor()
	if ret != nil {
		return ret
	}

	// 2. Get genericK8s intent customizations
	_, ret = I.getObject()
	if ret != nil {
		return ret
	}
	return nil
}
