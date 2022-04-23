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
	log "github.com/sirupsen/logrus"
)

// CompositeApp application structure
type ProjectMetadata struct {
	Metadata apiMetaData `json:"metadata"`
}

// CompAppHandler , This implements the orchworkflow interface
type projectHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

func (h *projectHandler) getMiddleEndObject() (interface{}, interface{}) {
	respcode := 200
	orch := h.orchInstance
	vars := orch.Vars
	orch.CompositeAppReturnJSON = nil
	key := DraftCompositeAppKey{}
	if orch.treeFilter != nil && orch.treeFilter.compositeAppName != "" {
		key.Cname = orch.treeFilter.compositeAppName
		key.Project = vars["projectName"]
		key.Cversion = orch.treeFilter.compositeAppVersion
	}
	caList, err := orch.GetDraftCompositeApplication(key, "depthAll")
	if err != nil {
		log.Errorf("Encountered error while fetching composite app from middleend collection: %s", err)
		return nil, http.StatusInternalServerError
	}

	for _, ca := range caList {
		if orch.dataRead.Metadata.Metadata.Name != ca.ProjectName {
			continue
		}
		orch.CompositeAppReturnJSON = append(orch.CompositeAppReturnJSON, ca)
		var cappsDataInstance CompositeAppTree
		cappName := ca.Metadata.Name
		cappVersion := ca.Spec.Version
		cappsDataInstance.Metadata = CompositeApp{}
		cappsDataInstance.Metadata.Metadata.Name = ca.Metadata.Name
		cappsDataInstance.Metadata.Metadata.Description = ca.Metadata.Description
		cappsDataInstance.Metadata.Metadata.UserData1 = ca.Metadata.UserData1
		cappsDataInstance.Metadata.Metadata.UserData2 = ca.Metadata.UserData2
		cappsDataInstance.Status = ca.Status
		cappsDataInstance.Metadata.Spec.Version = ca.Spec.Version
		orch.dataRead.compositeAppMap[cappName+"-"+cappVersion] = &cappsDataInstance
	}
	return nil, respcode
}

func (h *projectHandler) getObject() (interface{}, interface{}) {
	_, respcode := h.getEMCOObject()
	_, respcode = h.getMiddleEndObject()
	return nil, respcode
}

func (h *projectHandler) getEMCOObject() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	var cappList []CompositeApp
	dataRead.compositeAppMap = make(map[string]*CompositeAppTree)
	if orch.treeFilter != nil && orch.treeFilter.compositeAppName != "" {
		temp := CompositeApp{}
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			vars["projectName"] + "/composite-apps/" + orch.treeFilter.compositeAppName + "/" +
			orch.treeFilter.compositeAppVersion
		log.Debugf("composite app URL project: %s", h.orchURL)
		respcode, respdata, err := orch.apiGet(h.orchURL, vars["projectName"]+"_getcapps")
		log.Infof("Get composite app status: %d", respcode)
		if err != nil {
			return nil, respcode
		}
		//need to change the retcode
		if respcode != http.StatusOK {
			return nil, respcode
		}
		json.Unmarshal(respdata, &temp)
		cappList = append(cappList, temp)
	} else {
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			vars["projectName"] + "/composite-apps"
		respcode, respdata, err := orch.apiGet(h.orchURL, vars["projectName"]+"_getcapps")
		log.Infof("Get composite app status: %d", respcode)
		if err != nil {
			return nil, respcode
		}
		if respcode != http.StatusOK {
			return nil, respcode
		}
		json.Unmarshal(respdata, &cappList)
	}

	for k, value := range cappList {
		log.Infof("Composite app: %+v", cappList[k])
		var cappsDataInstance CompositeAppTree
		cappName := value.Metadata.Name
		cappVersion := value.Spec.Version
		cappsDataInstance.Metadata = value
		cappsDataInstance.Status = "created"
		dataRead.compositeAppMap[cappName+"-"+cappVersion] = &cappsDataInstance
	}
	return nil, http.StatusOK
}

func (h *projectHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		vars["projectName"]
    log.Debugf("projectURL: %s", h.orchURL)
	respcode, respdata, err := orch.apiGet(h.orchURL, vars["projectName"]+"_getProject")
	log.Debugf("Get project status: %d", respcode)
	if err != nil {
		return nil, respcode
	}
	if respcode != http.StatusOK {
		return nil, respcode
	}

	json.Unmarshal(respdata, &dataRead.Metadata)
	return nil, respcode
}

func (h *projectHandler) deleteObject() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	dataRead := h.orchInstance.dataRead
	cappList := dataRead.compositeAppMap
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		vars["projectName"] + "/composite-apps"
	for compositeAppName, compositeAppValue := range cappList {
		url := h.orchURL + "/" + compositeAppName + "/" + compositeAppValue.Metadata.Spec.Version
		log.Debugf("Delete composite app %s", url)
		resp, err := orch.apiDel(url, compositeAppName+"_delcapp")
		log.Debugf("Delete composite app status: %d", resp)
		if err != nil {
			return err
		}
		if resp != http.StatusNoContent {
			return resp
		}
	}
	return nil
}

func (h *projectHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + vars["projectName"]
	log.Debugf("Delete Project %s", h.orchURL)
	resp, err := orch.apiDel(h.orchURL, vars["projectName"]+"_delProject")
	if err != nil {
		return err
	}
	if resp != http.StatusNoContent {
		return resp
	}
	log.Debugf("Delete Project status: %d", resp)
	return nil
}

func (h *projectHandler) createAnchor() interface{} {
	orch := h.orchInstance
	vars := orch.Vars

	projectCreate := ProjectMetadata{
		Metadata: apiMetaData{
			Name:        vars["projectName"],
			Description: vars["description"],
			UserData1:   "data 1",
			UserData2:   "data 2"},
	}

	jsonLoad, _ := json.Marshal(projectCreate)
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + vars["projectName"]
	resp, err := orch.apiPost(jsonLoad, h.orchURL, vars["projectName"])
	if err != nil {
		return err
	}
	if resp != http.StatusCreated {
		return resp
	}
	orch.Vars["version"] = "v1"
	log.Infof("projectHandler response: %d", resp)

	return nil
}

func (h *projectHandler) createObject() interface{} {
	return nil
}

func createProject(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.createAnchor()
	if err != nil {
		return err
	}
	return nil
}

func delProject(I orchWorkflow) interface{} {
	// 1. Delete the object
	err := I.deleteObject()
	if err != nil {
		return err
	}
	// 2. Delete the Anchor
	err = I.deleteAnchor()
	if err != nil {
		return err
	}
	return nil
}
