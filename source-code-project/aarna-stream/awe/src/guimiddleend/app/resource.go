package app

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/middleend/localstore"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ResourceSpec consists of NewObject, ExistingResource
type ResourceSpec struct {
	NewObject   string                 `json:"newobject"`
	ResourceGVK localstore.ResourceGVK `json:"resourcegvk,omitempty"`
}

// ResourceInfo contains all the required information for creating a resource and applying the required
// customization
type ResourceInfo struct {
	ResourceSpec      ResourceSpec                   `json:"rspec"`
	CustomizationSpec localstore.CustomizeSpec       `json:"cspec"`
	ResourceFile      localstore.ResourceFileContent `json:"resFile,omitempty`
	ResourceFileName  string                         `json:"resFileName,omitempty"`
	CustomFile        localstore.SpecFileContent     `json: czFile,omitempty`
}

type GenericK8sIntentInfo struct {
	listGenK8sData GenericK8sIntentsData
	resData map[string][]ResourceInfo
}

type GenericK8sIntentsData struct {
	resource   []localstore.Resource
	resMap map[string][]localstore.Customization `json:"resourceInfo"`
}

// Validate and process resourceData
func (h *OrchestrationHandler) processResourceData(w *http.ResponseWriter, r *http.Request) bool {

	for i, appData := range h.DigData.Spec.Apps {
		var fileContent string
		var fileName string
		var resGVK localstore.ResourceGVK

		for j, resObj := range appData.RsInfo {
			var fileNameArray []string
			var contentArray []string
			var cmEnvContent string
			var count int

			// If newobject is true, then contentFile should be there
			if strings.ToLower(resObj.ResourceSpec.NewObject) == "true" {
				// Fetch the *fileheaders
				tag := appData.Metadata.Name + "_file" + strconv.Itoa(count)
				count += 1
				files := r.MultipartForm.File[tag]
				if len(files) == 0 {
					log.Errorf("Unable to fetch file from multipart request key %s", tag)
					return false
				}
				fileName = files[0].Filename
				resGVK, fileContent = h.FetchK8sFileContent(files)

				// Validate resourceGVK and fileContent
				if fileContent == "" || resGVK.APIVersion == "" || resGVK.Kind == "" || resGVK.Name == "" {
					log.Errorf("Unable to fetch fileContent or APIVersion/Kind/Name from resource object")
					return false
				}

				h.DigData.Spec.Apps[i].RsInfo[j].ResourceSpec.ResourceGVK.APIVersion = resGVK.APIVersion
				h.DigData.Spec.Apps[i].RsInfo[j].ResourceSpec.ResourceGVK.Kind = resGVK.Kind
				h.DigData.Spec.Apps[i].RsInfo[j].ResourceSpec.ResourceGVK.Name = resGVK.Name
				resObj = h.DigData.Spec.Apps[i].RsInfo[j]

				log.Infof("Resource kind is: %s", resGVK.Kind)
				if strings.ToLower(resGVK.Kind) == "configmap" || strings.ToLower(resGVK.Kind) == "secret" {
					byteContent, _ := base64.StdEncoding.DecodeString(fileContent)
					fileNameArray, contentArray, cmEnvContent = h.ProcessConfigMapSecret(byteContent)
					log.Debugf("fileNameArray: %+v", fileNameArray)
					log.Debugf("contentArray: %+v", contentArray)
					if cmEnvContent != "" {
						contentArray = append(contentArray, cmEnvContent)
						fileNameArray = append(fileNameArray, fileName)
					}
				}
			}

			if strings.ToLower(resObj.ResourceSpec.NewObject) == "true" &&
				strings.ToLower(resObj.ResourceSpec.ResourceGVK.Kind) != "configmap" && strings.ToLower(resObj.ResourceSpec.ResourceGVK.Kind) != "secret" {
				byteContent, _ := base64.StdEncoding.DecodeString(fileContent)
				h.DigData.Spec.Apps[i].RsInfo[j].ResourceFile.FileContent = string(byteContent)
				h.DigData.Spec.Apps[i].RsInfo[j].ResourceFileName = fileName
			}

			// Validation to ensure cluster info is present when customization spec clusterspecific field is set to true
			if strings.ToLower(resObj.CustomizationSpec.ClusterSpecific) == "true" && (localstore.ClusterInfo{}) == resObj.CustomizationSpec.ClusterInfo {
				log.Error(":: ClusterInfo missing when ClusterSpecific is true ::", log.Fields{})
				return false
			}

			if strings.ToLower(resObj.CustomizationSpec.ClusterSpecific) == "true" && strings.ToLower(resObj.CustomizationSpec.ClusterInfo.Scope) == "label" && resObj.CustomizationSpec.ClusterInfo.ClusterLabel == "" {
				log.Error(":: ClusterLabel missing when ClusterSpecific is true and  ClusterScope is label::", log.Fields{})
				(*w).WriteHeader(http.StatusBadRequest)
				return false
			}

			if strings.ToLower(resObj.CustomizationSpec.ClusterSpecific) == "true" && strings.ToLower(resObj.CustomizationSpec.ClusterInfo.Scope) == "name" && resObj.CustomizationSpec.ClusterInfo.ClusterName == "" {
				log.Error(":: ClusterName missing when ClusterSpecific is true and  ClusterScope is name::", log.Fields{})
				(*w).WriteHeader(http.StatusBadRequest)
				return false
			}

			// Set multipartfiles flag as true to ensure, we send the right payload as part of multipart upload
			if strings.ToLower(resObj.ResourceSpec.NewObject) == "true" &&
				strings.ToLower(resObj.ResourceSpec.ResourceGVK.Kind) == "configmap" || strings.ToLower(resObj.ResourceSpec.ResourceGVK.Kind) == "secret" {
				h.DigData.Spec.Apps[i].RsInfo[j].CustomFile = localstore.SpecFileContent{FileContents: contentArray, FileNames: fileNameArray}
			}
		}
	}
	return true
}

// Creates a GAC resource and apply the required customization
func (h *OrchestrationHandler) createUpdateK8sResource(w *http.ResponseWriter, storeType string) {
	if storeType == "emco" {
		bstore := &remoteStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
	} else {
		bstore := &localStoreIntentHandler{}
		bstore.orchInstance = h
		h.bstore = bstore
	}
	h.Vars["deploymentIntentGroupName"] = h.DigData.Name

	// Call workflow for creation of genericK8s intent, resource and customization EMCO objects
	gk8sHandler := &genericK8sIntentHandler{}
	gk8sHandler.orchInstance = h
	gk8sIntentStatus := addGenericK8sIntent(gk8sHandler)
	if gk8sIntentStatus != nil {
		if intval, ok := gk8sIntentStatus.(int); ok {
			(*w).WriteHeader(intval)
		} else {
			(*w).WriteHeader(http.StatusInternalServerError)
		}
	}
}

// Fetches all GAC resources belonging to a given composite application
// GET /projects/{projectName}/composite-apps/{compositeAppName}/{version}/deployment-intent-groups/{deploymentIntentGroupName}/resources
func (h *OrchestrationHandler) GetK8sResources(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	bstore := &remoteStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore
	h.InitializeResponseMap()

	// Call workflow for fetching genericK8s intent resource and customization EMCO objects
	gk8sHandler := &genericK8sIntentHandler{}
	gk8sHandler.orchInstance = h
	gk8sIntentStatus := getGenericK8sIntent(gk8sHandler)
	if gk8sIntentStatus != nil {
		if intval, ok := gk8sIntentStatus.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	compositeAppName := h.Vars["compositeAppName"]
	genK8sInfo := h.genK8sInfo[compositeAppName+"_genk8sint"]
	log.Debugf("genK8sInfo: %+v", genK8sInfo)

	var retval []byte
	retval, err := json.Marshal(genK8sInfo.listGenK8sData.resMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(retval)
}

// Deletes given GAC resource belonging to a given composite application
// DELETE /projects/{projectName}/composite-apps/{compositeAppName}/{version}/deployment-intent-groups/{deploymentIntentGroupName}/resources/{resourceName}
func (h *OrchestrationHandler) DeleteK8sResources(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	bstore := &remoteStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore
	h.InitializeResponseMap()

	retcode, _ := h.bstore.deleteResource(h.Vars["resourceName"], h.Vars["projectName"], h.Vars["compositeAppName"],
		h.Vars["version"], h.Vars["deploymentIntentGroupName"], h.Vars["compositeAppName"]+"_genk8sint")
	log.Infof("DeleteK8sResources response: %s", retcode.(int))

	w.WriteHeader(retcode.(int))
}

// Deletes given GAC resource customization belonging to a given composite application
// DELETE /projects/{projectName}/composite-apps/{compositeAppName}/{version}/deployment-intent-groups/{deploymentIntentGroupName}/resources/{resourceName}/customization/{customizationName}
func (h *OrchestrationHandler) DeleteK8sResourceCustomizations(w http.ResponseWriter, r *http.Request) {
	h.Vars = mux.Vars(r)
	bstore := &remoteStoreIntentHandler{}
	bstore.orchInstance = h
	h.bstore = bstore
	h.InitializeResponseMap()

	retcode, _ := h.bstore.deleteCustomization(h.Vars["customizationName"], h.Vars["projectName"], h.Vars["compositeAppName"],
		h.Vars["version"], h.Vars["deploymentIntentGroupName"], h.Vars["compositeAppName"]+"_genk8sint", h.Vars["resourceName"])
	log.Infof("DeleteK8sResourceCustomizations response: %s", retcode.(int))

	w.WriteHeader(retcode.(int))
}
