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
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type logicalCloudData struct {
	Metadata apiMetaData      `json:"metadata"`
	Spec     logicalCloudSpec `json:"spec"`
}

// UserData contains the parameters needed for user
type UserData struct {
	UserName string `json:"userName"`
	Type     string `json:"type"`
}

// UserPermission contains the parameters needed for a user permission
type UserPermission struct {
	MetaData      UPMetaDataList `json:"metadata"`
	Specification UPSpec         `json:"spec"`
}

// UPMetaDataList contains the parameters needed for a user permission metadata
type UPMetaDataList struct {
	UserPermissionName string `json:"name"`
	Description        string `json:"description"`
	UserData1          string `json:"userData1"`
	UserData2          string `json:"userData2"`
}

// UPSpec contains the parameters needed for a user permission spec
type UPSpec struct {
	Namespace string   `json:"namespace"`
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// Quota contains the parameters needed for a Quota
type Quota struct {
	MetaData QMetaDataList `json:"metadata"`
	// Specification QSpec         `json:"spec"`
	Specification map[string]string `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type QMetaDataList struct {
	QuotaName   string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

type QuotaInfo struct {
	LimitsCPU                   string `json:"limits.cpu"`
	LimitsMemory                string `json:"limits.memory"`
	RequestsCPU                 string `json:"requests.cpu"`
	RequestsMemory              string `json:"requests.memory"`
	RequestsStorage             string `json:"requests.storage"`
	LimitsEphemeralStorage      string `json:"limits.ephemeral.storage"`
	PersistentVolumeClaims      string `json:"persistentvolumeclaims"`
	Pods                        string `json:"pods"`
	ConfigMaps                  string `json:"configmaps"`
	ReplicationControllers      string `json:"replicationcontrollers"`
	ResourceQuotas              string `json:"resourcequotas"`
	Services                    string `json:"services"`
	ServicesLoadBalancers       string `json:"services.loadbalancers"`
	ServicesNodePorts           string `json:"services.nodeports"`
	Secrets                     string `json:"secrets"`
	CountReplicationControllers string `json:"count/replicationcontrollers"`
	CountDeploymentsApps        string `json:"count/deployments.apps"`
	CountReplicasetsApps        string `json:"count/replicasets.apps"`
	CountStatefulSets           string `json:"count/statefulsets.apps"`
	CountJobsBatch              string `json:"count/jobs.batch"`
	CountCronJobsBatch          string `json:"count/cronjobs.batch"`
	CountDeploymentsExtensions  string `json:"count/deployments.extensions"`
}


// Logical cloud spec
type logicalCloudSpec struct {
	NameSpace string   `json:"namespace"`
	Level     string   `json:"level"`
	UserData  UserData `json:"user"`
}

type ClusterLabels struct {
	Metadata apiMetaData `json:"metadata"`
	Labels   []Labels    `json:"labels"`
}

type clusterReferenceFlat struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Userdata1   string `json:"userData1"`
		Userdata2   string `json:"userData2"`
	} `json:"metadata"`
	Spec struct {
		ClusterProvider string   `json:"clusterProvider"`
		ClusterName     string   `json:"cluster"`
		LoadbalancerIP  string   `json:"loadBalancerIP"`
		Certificate     string   `json:"certificate,omitempty"`
		LabelList       []Labels `json:"labels,omitempty"`
	} `json:"spec"`
}

type clusterReferenceNested struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"metadata"`
	Spec struct {
		ClusterProvidersList []ClusterProviders `json:"clusterProviders"`
	} `json:"spec"`
}

type LogicalCloudInfo struct {
	UserQuota map[string]string `json:"userQuota"`
	UserPermissions UPSpec `json:"userPermissions"`
	ClusterReferences clusterReferenceNested `json:"clusterReferences"`
}

type ClusterProviders struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"metadata"`
	Spec struct {
		ClustersList []Clusters `json:"clusters"`
	} `json:"spec"`
}
type Clusters struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"metadata"`
	Spec struct {
		Labels []Labels `json:"labels"`
	} `json:"spec"`
}

type Labels struct {
	LabelName string `json:"clusterLabel"`
}

type LogicalClouds struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Userdata1   string `json:"userData1"`
		Userdata2   string `json:"userData2"`
	} `json:"metadata"`
	Spec struct {
		Namespace string `json:"namespace"`
		Level     string `json:"level"`
		User      struct {
			UserName        string      `json:"userName"`
			Type            string      `json:"type"`
			UserPermissions interface{} `json:"userPermissions"`
		} `json:"user"`
	} `json:"spec"`
}

type userPermissions struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

type logicalCloudsPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Spec        LogicalCloudSpec
}

type LogicalCloudSpec struct {
	Namespace            string             `json:"namespace"`
	User                 UserData           `json:"user,omitempty"`
	Permissions          userPermissions    `json:"permissions,omitempty"`
	Quotas               QuotaInfo          `json:"quotas,omitempty"`
	ClusterProvidersList []ClusterProviders `json:"clusterProviders"`
}


// logicalCloudHandler implements the orchworkflow interface
type logicalCloudHandler struct {
	orchInstance *OrchestrationHandler
}

func (h *logicalCloudHandler) getLogicalClouds() ([]LogicalClouds, interface{}) {
	orch := h.orchInstance
	var lc []LogicalClouds
	projectName := orch.Vars["projectName"]
	url := "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
		projectName + "/logical-clouds"
	retcode, respval, err := orch.apiGet(url, projectName)
	log.Infof("Get LC status: %d", retcode)
	if err != nil {
		log.Errorf("Failed to LC for %s", projectName)
		return nil, retcode
	}
	if retcode != http.StatusOK {
		log.Errorf("Failed to LC for %s", projectName)
		return nil, retcode
	}
	json.Unmarshal(respval, &lc)
	return lc, nil
}

func (h *logicalCloudHandler) getLogicalCloudReferences(lcName string) (clusterReferenceNested, interface{}) {
	orch := h.orchInstance
	var lcRefList []clusterReferenceFlat
	projectName := orch.Vars["projectName"]
	url := "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
		projectName + "/logical-clouds/" + lcName + "/cluster-references"
	retcode, respval, err := orch.apiGet(url, lcName)
	log.Infof("Get LC references status: %d", retcode)
	if err != nil {
		log.Errorf("Failed to LC reference for %s", lcName)
		return clusterReferenceNested{}, retcode
	}
	if retcode != http.StatusOK {
		log.Errorf("Failed to LC reference for %s", lcName)
		return clusterReferenceNested{}, retcode
	}
	json.Unmarshal(respval, &lcRefList)

	// Fetch label information of all clusters belonging to cluster provider part of logical cloud
	clusterProviders := make(map[string]bool)
	for _, cluRef := range lcRefList {
		clusterProviders[cluRef.Spec.ClusterProvider] = true
	}

	// Build a map of cluster providers to clusters list
	var clusterProviderMap = make(map[string][]Clusters, len(lcRefList))

	for clusterProvider, _ := range clusterProviders {
		var clusterLabels []ClusterLabels
		url := "http://" + orch.MiddleendConf.Clm + "/v2/cluster-providers/" +
			clusterProvider + "/clusters?withLabels=true"
		retcode, respval, err := orch.apiGet(url, clusterProvider)
		if retcode != http.StatusOK {
			log.Errorf("Encountered error while fetching labels for cluster provider %s", clusterProvider)
			return clusterReferenceNested{}, retcode
		}
		if err != nil {
			log.Errorf("Failed while fetching labels for cluster provider %s", clusterProvider)
			return clusterReferenceNested{}, retcode
		}

		json.Unmarshal(respval, &clusterLabels)

		for _, ref := range lcRefList {
			var cluster Clusters
			cluster.Metadata.Name = ref.Spec.ClusterName
			cluster.Metadata.Description = "Cluster" + ref.Spec.ClusterName
			for _, cinfo := range clusterLabels {
				if ref.Spec.ClusterProvider == clusterProvider && ref.Spec.ClusterName == cinfo.Metadata.Name {
					cluster.Spec.Labels = cinfo.Labels
				}
			}
			if clusterProvider == ref.Spec.ClusterProvider {
				clusterProviderMap[clusterProvider] = append(clusterProviderMap[clusterProvider],
					cluster)
			}
		}
	}

	// parse through the output and fill int he reference nested structure
	// that is to be returned to the GUI
	var nestedRef clusterReferenceNested
	nestedRef.Metadata.Name = lcName
	nestedRef.Metadata.Description = "Cluster references for" + lcName

	for k, v := range clusterProviderMap {
		l := ClusterProviders{}
		l.Metadata.Name = k
		l.Metadata.Description = "cluster provider : " + k
		l.Spec.ClustersList = make([]Clusters, len(v))
		l.Spec.ClustersList = v
		nestedRef.Spec.ClusterProvidersList = append(nestedRef.Spec.ClusterProvidersList, l)
	}
	return nestedRef, nil
}

func (h *logicalCloudHandler) createLogicalCloud(w http.ResponseWriter, lcData logicalCloudsPayload) interface{} {
	orch := h.orchInstance
	if lcData.Spec.Namespace == "" {
		resp, err := h.createAdminLogicalCloud(lcData)
		if err != nil {
			log.Info("Error encountered during creation of Admin Logical Cloud: %s", err)
			return resp
		}
		if resp != nil {
			return resp
		}

	} else {
		resp, err := h.createStandardLogicalCloud(lcData)
		if err != nil {
			log.Info("Error encountered during creation of Standard Logical Cloud: %s", err)
			return resp
		}
		if resp != nil {
			return resp
		}
	}

	// Now Create the reference for each cluster in the logical cloud
        for _, clusterProvider := range lcData.Spec.ClusterProvidersList {
                for _, cluster := range clusterProvider.Spec.ClustersList {
                        clusterReferencePayload := clusterReferenceFlat{}
                        clusterReferencePayload.Metadata.Name = lcData.Name + "-" +
                                clusterProvider.Metadata.Name + "-" + cluster.Metadata.Name
                        clusterReferencePayload.Metadata.Description = "Cluster reference for cluster" +
                                clusterProvider.Metadata.Name + ":" + cluster.Metadata.Name
                        clusterReferencePayload.Metadata.Userdata1 = "NA"
                        clusterReferencePayload.Metadata.Userdata2 = "NA"
                        clusterReferencePayload.Spec.ClusterProvider = clusterProvider.Metadata.Name
                        clusterReferencePayload.Spec.ClusterName = cluster.Metadata.Name
                        clusterReferencePayload.Spec.LoadbalancerIP = "0.0.0.0"
                        jsonLoad, _ := json.Marshal(clusterReferencePayload)
                        url := "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
                                orch.Vars["projectName"] + "/logical-clouds/" + lcData.Name + "/cluster-references"
                        resp, err := orch.apiPost(jsonLoad, url, lcData.Name+"-"+cluster.Metadata.Name)
                        if err != nil {
                                w.WriteHeader(resp.(int))
                                return resp
                        }
                        if resp != http.StatusCreated {
                                w.WriteHeader(resp.(int))
                                return resp
                        }
                }
        }

	// Instantiate the cluster.
        url := "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
                orch.Vars["projectName"] + "/logical-clouds/" + lcData.Name + "/instantiate"
        var jsonLoad []byte
        resp, err := orch.apiPost(jsonLoad, url, lcData.Name+"-instantiate")
        if err != nil {
                w.WriteHeader(resp.(int))
                return resp
        }
        if resp != http.StatusCreated {
                w.WriteHeader(resp.(int))
                return resp
        }
        w.WriteHeader(resp.(int))

        // TODO: Workaround to create cluster config for logical cloud (Level 1)
        time.Sleep(30)
        for _, clusterProvider := range lcData.Spec.ClusterProvidersList {
                for _, cluster := range clusterProvider.Spec.ClustersList {
                        clusterReference := lcData.Name + "-" +
                                clusterProvider.Metadata.Name + "-" + cluster.Metadata.Name
                        url = "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
                                orch.Vars["projectName"] + "/logical-clouds/" + lcData.Name + "/cluster-references/" + clusterReference +
                                "/kubeconfig"
                        resp, _, err = orch.apiGet(url, lcData.Name+"-kubeconfig")
                        if err != nil {
                                w.WriteHeader(resp.(int))
                                return resp
                        }
                        if resp != http.StatusOK {
                                w.WriteHeader(resp.(int))
                                return resp
                        }
                }
        }
	w.WriteHeader(http.StatusCreated)
        w.Write(orch.response.payload[lcData.Name])
	return nil
}

func (h *logicalCloudHandler) createAdminLogicalCloud(lcData logicalCloudsPayload) (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars

	// Create the logical cloud
	apiPayload := logicalCloudData{
		Metadata: apiMetaData{
			Name:        lcData.Name,
			Description: lcData.Description,
			UserData1:   "data 1",
			UserData2:   "data 2"},
		Spec: logicalCloudSpec{
			Level: "0",
		},
	}

	jsonLoad, _ := json.Marshal(apiPayload)
	url := "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
		vars["projectName"] + "/logical-clouds"
	resp, err := orch.apiPost(jsonLoad, url, lcData.Name)
	log.Infof("Create Admin logical-cloud response: %d", resp)
	if err != nil {
		return resp, err
	}

	if resp != http.StatusCreated {
		return resp, nil
	}
	return nil, nil
}

func (h *logicalCloudHandler) createStandardLogicalCloud(lcData logicalCloudsPayload) (interface{}, interface{}) {
	orch := h.orchInstance
	vars := orch.Vars

	// Create the logical cloud
	apiPayload :=  logicalCloudData{
			Metadata: apiMetaData{
				Name:        lcData.Name,
				Description: lcData.Description,
				UserData1:   "data 1",
				UserData2:   "data 2"},
			Spec: logicalCloudSpec{
				UserData: UserData{
					UserName: lcData.Spec.User.UserName,
					Type:     lcData.Spec.User.Type},
				NameSpace: lcData.Spec.Namespace,
			},
		}
	jsonLoad, _ := json.Marshal(apiPayload)
	url := "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
		vars["projectName"] + "/logical-clouds"
	resp, err := orch.apiPost(jsonLoad, url, lcData.Name)
	if err != nil {
		return resp, err
	}

	log.Infof("Create Standard logical-cloud response: %d", resp)

	if resp != http.StatusCreated {
		return resp, nil
	}

	// Create User Permissions for Standard Logical Cloud
	userPerm := UserPermission{
		MetaData:      UPMetaDataList{
			UserPermissionName: lcData.Spec.User.UserName + "_permissions",
			Description: "User Permissions",
			UserData1: "UserData1",
            UserData2: "UserData2",
		},
		Specification: UPSpec{
			Namespace: lcData.Spec.Namespace,
			APIGroups: lcData.Spec.Permissions.APIGroups,
			Resources: lcData.Spec.Permissions.Resources,
			Verbs: lcData.Spec.Permissions.Verbs,
		},
	}

	jsonLoad, _ = json.Marshal(userPerm)
	url = "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
		vars["projectName"] + "/logical-clouds/" + lcData.Name + "/user-permissions"
	resp, err = orch.apiPost(jsonLoad, url, lcData.Name)
	if err != nil {
		return resp, err
	}
	if resp != http.StatusCreated {
		return resp, nil
	}

	// Create User Quotas for Standard Logical Cloud
	quotaInfo := make(map[string]string)
	quotaInfo["limits.cpu"] = lcData.Spec.Quotas.LimitsCPU
	quotaInfo["limits.memory"] = lcData.Spec.Quotas.LimitsMemory
	quotaInfo["requests.cpu"] = lcData.Spec.Quotas.RequestsCPU
	quotaInfo["requests.memory"] = lcData.Spec.Quotas.RequestsMemory
	quotaInfo["requests.storage"] = lcData.Spec.Quotas.RequestsStorage
	/*quotaInfo["limits.ephemeral-storage"] = lcData.Spec.Quotas.LimitsEphemeralStorage*/
	quotaInfo["persistentvolumeclaims"] = lcData.Spec.Quotas.PersistentVolumeClaims
	quotaInfo["pods"] = lcData.Spec.Quotas.Pods
	quotaInfo["configmaps"] = lcData.Spec.Quotas.ConfigMaps

	quotaInfo["replicationcontrollers"] = lcData.Spec.Quotas.ReplicationControllers
	quotaInfo["resourcequotas"] = lcData.Spec.Quotas.ResourceQuotas
	quotaInfo["services"] = lcData.Spec.Quotas.Services
	quotaInfo["services.loadbalancers"] = lcData.Spec.Quotas.ServicesLoadBalancers
	quotaInfo["services.nodeports"] = lcData.Spec.Quotas.ServicesNodePorts
	quotaInfo["secrets"] = lcData.Spec.Quotas.Secrets
	quotaInfo["count/replicationcontrollers"] = lcData.Spec.Quotas.CountReplicationControllers
	quotaInfo["count/deployments.apps"] = lcData.Spec.Quotas.CountDeploymentsApps
	quotaInfo["count/replicasets.apps"] = lcData.Spec.Quotas.CountReplicasetsApps
	quotaInfo["count/statefulsets.apps"] = lcData.Spec.Quotas.CountStatefulSets
	quotaInfo["count/jobs.batch"] = lcData.Spec.Quotas.CountJobsBatch
	quotaInfo["count/cronjobs.batch"] = lcData.Spec.Quotas.CountCronJobsBatch
	quotaInfo["count/cronjobs.batch"] = lcData.Spec.Quotas.CountDeploymentsExtensions
	quotas := Quota{
		MetaData:      QMetaDataList{
			QuotaName: lcData.Spec.User.UserName + "-quotas",
			Description: "User Quotas",
			UserData1: "UserData1",
			UserData2: "UserData2",
		},
		Specification: quotaInfo,
	}

	jsonLoad, _ = json.Marshal(quotas)
	url = "http://" + orch.MiddleendConf.Dcm + "/v2/projects/" +
		vars["projectName"] + "/logical-clouds/" + lcData.Name + "/cluster-quotas"
	resp, err = orch.apiPost(jsonLoad, url, lcData.Name)
	if err != nil {
		return resp, err
	}
	if resp != http.StatusCreated {
		return resp, nil
	}

	return nil, nil
}

func (h *OrchestrationHandler) DeployLogicalCloud(w http.ResponseWriter, ClusterName string, jsonData ClusterMetadata) bool {
        // Final Result ByDefault Considered True
        var Result bool = true
        // Variable for Logical Cloud
        var provider ClusterProviders
        var cluster Clusters
        h.InitializeResponseMap()


        // for provider Metadata
        provider.Metadata.Name = ClusterName
        provider.Metadata.Description = ""

        // for cluster Metadata
        cluster.Metadata.Name = jsonData.Metadata.Name
        cluster.Metadata.Description = ""

        // Array for ClusterProvider and Cluster List
        provider.Spec.ClustersList = append(provider.Spec.ClustersList, cluster)

        // Initialing the Logical Cloud Struct with Payload
        lcData := logicalCloudsPayload{
                Name:        "operator-logical-cloud-" + jsonData.Metadata.Name,
                Description: "operator-logical-cloud",
                Spec: LogicalCloudSpec{
                        ClusterProvidersList: []ClusterProviders{provider},
                },
        }

        h.client = http.Client{}

        lcHandler := &logicalCloudHandler{}
        lcHandler.orchInstance = h
        // Creating the Logical Cloud for Monitoring Service
        lcStatus := lcHandler.createLogicalCloud(w, lcData)
        if lcStatus != nil {
                if intval, ok := lcStatus.(int); ok {
                        log.Infof("logical cloud creation is successful: %d", intval)
                } else {
                        Result = false
                }
        }
        return Result
}

// CreateLogicalCloud, creates the logical clouds (level 0/level 1)
func (h *OrchestrationHandler) CreateLogicalCloud(w http.ResponseWriter, r *http.Request) {
	var lcData logicalCloudsPayload
	h.Vars = mux.Vars(r)
	h.InitializeResponseMap()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&lcData)
	if err != nil {
		log.Error("failed to parse json: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.client = http.Client{}

	lcHandler := &logicalCloudHandler{}
	lcHandler.orchInstance = h
	lcStatus := lcHandler.createLogicalCloud(w, lcData)
	if lcStatus != nil {
		if intval, ok := lcStatus.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(h.response.payload[lcData.Name])
		return
	}

}

// Get LC information
func (h *OrchestrationHandler) GetLogicalClouds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Vars = vars
	h.InitializeResponseMap()
	lcHandler := &logicalCloudHandler{}
	lcHandler.orchInstance = h
	// Get the logical cloud list
	lcList, retcode := lcHandler.getLogicalClouds()
	log.Infof("lcList: %+v", lcList)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			log.Infof("Failed to logical clouds : %d", intval)
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
			w.Write([]byte(errMsg))
		}
		return
	}

	lcInfoList := make([]LogicalCloudInfo, 0)
	for _, lc := range lcList {
		respdata, retcode := lcHandler.getLogicalCloudReferences(lc.Metadata.Name)
		log.Infof("lcReferences: %+v", respdata)
		if retcode != nil {
			if intval, ok := retcode.(int); ok {
				log.Infof("Failed to get lc references : %d", intval)
				if intval == http.StatusBadRequest { // FIXME:
					continue
				}
				w.WriteHeader(intval)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				errMsg := string(h.response.payload[h.response.lastKey]) + h.response.lastKey
				w.Write([]byte(errMsg))
			}
			return
		}

		var lcCloudInfo LogicalCloudInfo
		lcCloudInfo.ClusterReferences = respdata
		if lc.Spec.Namespace != "default" {
			// Fetch logical cloud permissions, if it is standard/privileged logical cloud
			var userPerm []UserPermission
			url := "http://" + h.MiddleendConf.Dcm + "/v2/projects/" +
				h.Vars["projectName"] + "/logical-clouds/" + lc.Metadata.Name + "/user-permissions"
			resp, data, err := h.apiGet(url, lc.Metadata.Name+"-permissions")
			json.Unmarshal(data, &userPerm)
			log.Infof("userPermissions: %+v", userPerm)
			if err != nil {
				w.WriteHeader(resp.(int))
				return
			}
			if resp != http.StatusOK {
				w.WriteHeader(resp.(int))
				return
			}
			if len(userPerm) > 0 {
				lcCloudInfo.UserPermissions.Namespace = userPerm[0].Specification.Namespace
				lcCloudInfo.UserPermissions.Verbs = make([]string, len(userPerm[0].Specification.Verbs))
				lcCloudInfo.UserPermissions.Verbs = userPerm[0].Specification.Verbs
				lcCloudInfo.UserPermissions.Resources = make([]string, len(userPerm[0].Specification.Resources))
				lcCloudInfo.UserPermissions.Resources = userPerm[0].Specification.Resources
				lcCloudInfo.UserPermissions.APIGroups = make([]string, len(userPerm[0].Specification.APIGroups))
				lcCloudInfo.UserPermissions.APIGroups = userPerm[0].Specification.APIGroups
			}

			// Fetch logical cloud quota info, if it is standard/privileged logical cloud
			var quotas []Quota
			url = "http://" + h.MiddleendConf.Dcm + "/v2/projects/" +
				h.Vars["projectName"] + "/logical-clouds/" + lc.Metadata.Name + "/cluster-quotas"
			resp, data, err = h.apiGet(url, lc.Metadata.Name+"-quotas")
			json.Unmarshal(data, &quotas)
			log.Infof("quotas: %+v", quotas)
			if err != nil {
				w.WriteHeader(resp.(int))
				return
			}
			if resp != http.StatusOK {
				w.WriteHeader(resp.(int))
				return
			}

			if len(quotas) > 0 {
				lcCloudInfo.UserQuota = make(map[string]string)
				for key, value  := range quotas[0].Specification {
					lcCloudInfo.UserQuota[key] = value
				}
			}
		}
		lcInfoList = append(lcInfoList, lcCloudInfo)
	}

	log.Infof("lcInfoList: %+v", lcInfoList)
	retval, err := json.Marshal(lcInfoList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debugf("retval: %+v", retval)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}
