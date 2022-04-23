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

	pkgerrors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type DashboardClient struct {
	orchInstance *OrchestrationHandler
}

type DashboardData struct {
	CompositeAppCount          int `json:"compositeAppCount"`
	DeploymentIntentGroupCount int `json:"deploymentIntentGroupCount"`
	ClusterCount               int `json:"clusterCount"`
}

type ClusterProvider struct {
	Metadata apiMetaData         `json:"metadata"`
	Spec     ClusterProviderSpec `json:"spec"`
}

type ClusterProviderSpec struct {
	Clusters []Cluster `json:"clusters"`
}

type Cluster struct {
	Metadata apiMetaData `json:"metadata"`
}

type ClusterLabel struct {
	LabelName string `json:"labelName"`
}

// getClusterProviders fetches all the available cluster providers
func (h *DashboardClient) getClusterProviders() interface{} {
	var clusterProviderList []ClusterProvider
	orch := h.orchInstance
	url := "http://" + orch.MiddleendConf.Clm + "/v2/cluster-providers"
	respcode, respdata, err := orch.apiGet(url, "getClusterProviders")
	log.Infof("Get cluster providers status: %d", respcode)
	orch.response.lastKey = "getClusterProviders"
	if err != nil {
		return pkgerrors.New("Error getting ClusterProviders")
	}
	if respcode != http.StatusOK {
		//return pkgerrors.New("Error getting ClusterProvider")
		return respcode
	}
	json.Unmarshal(respdata, &clusterProviderList)
	orch.ClusterProviders = clusterProviderList
	return nil
}

//getClusters iterates thought all the cluster providers and gets the clusters in them
func (h *DashboardClient) getClusters() interface{} {
	orch := h.orchInstance
	for index, provider := range orch.ClusterProviders {
		var ClusterList []Cluster
		url := "http://" + orch.MiddleendConf.Clm + "/v2/cluster-providers/" + provider.Metadata.Name + "/clusters"
		orch.response.lastKey = "getClusters"
		respcode, respdata, err := orch.apiGet(url, "getClusters")
		if err != nil {
			return err //need to add the retcode
		}
		if respcode != http.StatusOK {
			//return pkgerrors.New("Error getting ClusterProvider")
			return respcode
		}
		json.Unmarshal(respdata, &ClusterList)
		orch.ClusterProviders[index].Spec.Clusters = ClusterList
		log.Infof("Get clusters status: %d", respcode)
	}
	return nil
}

func (h *DashboardClient) createCompositeAppTree() error {
	orch := h.orchInstance
	orch.treeFilter = nil
	orch.InitializeResponseMap()
	orch.prepTreeReq()
	dataPoints := []string{"projectHandler", "compAppHandler", "digpHandler"}
	orch.dataRead = &ProjectTree{}
	retcode := orch.constructTree(dataPoints)
	// Need to perform proper error handling
	if retcode != nil {
		pkgerrors.New("Error getting composite apps data")
	}
	return nil
}

func (h *DashboardClient) getAllClusters() interface{} {
	err := h.getClusterProviders()
	if err != nil {
		return err //need to add the retcode
	}
	err = h.getClusters()
	if err != nil {
		return err //need to add the retcode
	}

	return nil
}

//getDashboardData based on compositeapp data and clusters data,
//calculates the no of compositeapps (versions are not added to the count), deployment-intent-groups and clusters.
func (h *DashboardClient) getDashboardData() (DashboardData, interface{}) {
	orch := h.orchInstance
	err := h.createCompositeAppTree()
	if err != nil {
		return DashboardData{}, err
	}
	respcode := h.getClusterProviders()
	if respcode != nil {
		return DashboardData{}, respcode
	}
	respcode = h.getClusters()
	if err != nil {
		return DashboardData{}, respcode
	}

	dataRead := orch.dataRead
	orch.CompositeAppReturnJSONShrunk = nil
	var retData DashboardData
	retData.CompositeAppCount = 0
	retData.DeploymentIntentGroupCount = 0
	retData.ClusterCount = 0

	var compositeAppNameArray []string
	for compositeAppName := range dataRead.compositeAppMap {
		retData.DeploymentIntentGroupCount = retData.DeploymentIntentGroupCount + len(dataRead.compositeAppMap[compositeAppName].DigMap)
		compositeAppNameArray = append(compositeAppNameArray, dataRead.compositeAppMap[compositeAppName].Metadata.Metadata.Name)
	}
	//compositeAppNameArray can contain duplicate entries if there is more than 1 version of the compositeApp, but for count we want to ignore versions
	keys := make(map[string]bool)
	for _, entry := range compositeAppNameArray {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			retData.CompositeAppCount++
		}
	}

	//calculate total clusters
	for _, provider := range orch.ClusterProviders {
		retData.ClusterCount = retData.ClusterCount + len(provider.Spec.Clusters)
	}
	return retData, nil
}
