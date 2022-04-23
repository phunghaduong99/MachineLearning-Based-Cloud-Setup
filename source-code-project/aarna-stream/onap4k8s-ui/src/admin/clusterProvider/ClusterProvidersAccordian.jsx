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
import React, { useState} from "react";
import { makeStyles } from "@material-ui/core/styles";
import Accordion from "@material-ui/core/Accordion";
import AccordionDetails from "@material-ui/core/AccordionDetails";
import AccordionSummary from "@material-ui/core/AccordionSummary";
import Typography from "@material-ui/core/Typography";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import apiService from "../../services/apiService";
import { Button } from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";
// import EditIcon from "@material-ui/icons/Edit";
import ClusterForm from "./clusters/ClusterForm";
import ClustersTable from "./clusters/ClusterTable";
import DeleteDialog from "../../common/Dialogue";
import Notification from "../../common/Notification";

//import ClusterProviderForm from "../clusterProvider/ClusterProviderForm";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
  },
  heading: {
    fontSize: theme.typography.pxToRem(15),
    flexBasis: "33.33%",
    flexShrink: 0,
  },
  secondaryHeading: {
    fontSize: theme.typography.pxToRem(15),
    color: theme.palette.text.secondary,
  },
}));
export default function ControlledAccordions({ data, setData }) {
  const classes = useStyles();
  const [expanded, setExpanded] = useState(false);
  const [open, setOpen] = React.useState(false);
  const [formOpen, setFormOpen] = useState(false);
  // const [openProviderForm, setOpenProviderForm] = useState(false);
  const [selectedRowIndex, setSelectedRowIndex] = useState(0);
  const [notificationDetails, setNotificationDetails] = useState({});
  const handleAccordionOpen = (providerName, providerIndex) => (event, isExpanded) => {
    if (!isExpanded) {
      setExpanded(isExpanded ? providerName : false);
    } else {
      apiService
        .getClusters(providerName)
        .then((response) => {
          data[providerIndex].clusters = response;
          setData([...data]);
        })
        .catch((error) => {
          console.log(error);
        })
        .finally(() => {
          getLabels(providerName, isExpanded, providerIndex);
          getAllNetworks(providerIndex);
        });
    }
  };
  const getLabels = (providerName, isExpanded, index) => {
    if (data[index].clusters && data[index].clusters.length > 0) {
      data[index].clusters.forEach((cluster) => {
        let request = {
          providerName: data[index].metadata.name,
          clusterName: cluster.metadata.name,
        };
        apiService
          .getClusterLabels(request)
          .then((res) => {
            cluster.labels = res;
          })
          .catch((err) => {
            console.log("error getting cluster label : ", err);
          })
          .finally(() => {
            setData([...data]);
            setExpanded(isExpanded ? providerName : false);
          });
      });
    } else setExpanded(isExpanded ? providerName : false);
  };

  const getAllNetworks = (providerRowIndex, clusterIndex) => {
    if (data[providerRowIndex].clusters && data[providerRowIndex].clusters.length > 0) {
      if (clusterIndex === undefined) {
        data[providerRowIndex].clusters.forEach((cluster) => {
          let request = {
            providerName: data[providerRowIndex].metadata.name,
            clusterName: cluster.metadata.name,
          };
          apiService
            .getAllClusterNetworks(request)
            .then((res) => {
              cluster.networks = res.spec.networks;
              cluster.providerNetworks = res.spec["providerNetworks"];
              cluster.networksStatus = res.spec.status;
          })
            .catch((err) => {
              console.log("error getting cluster networks : ", err);
            })
            .finally(() => {
          setData([...data]);
            });
        });
      } else {
        let request = {
          providerName: data[providerRowIndex].metadata.name,
          clusterName: data[providerRowIndex].clusters[clusterIndex].metadata.name,
        };
        apiService
          .getAllClusterNetworks(request)
          .then((res) => {
            if(res.spec.status.toLowerCase() === "instantiating" || res.spec.status.toLowerCase() === "terminating"){
              //if the status is "instantiating" or "terminating" then call the api again till status becomes instantiated or terminated
              setTimeout(()=>{getAllNetworks(providerRowIndex, clusterIndex)}, 2000);
            } else {
              data[providerRowIndex].clusters[clusterIndex].networks =
                  res.spec.networks;
              data[providerRowIndex].clusters[clusterIndex].providerNetworks =
                  res.spec["providerNetworks"];
              data[providerRowIndex].clusters[clusterIndex].networksStatus =
                  res.spec.status;
            }
          })
          .catch((err) => {
            console.log("error getting cluster networks : ", err);
          })
          .finally(() => {
            setData([...data]);
          });
      }
    }
  };

  const onAddCluster = (index) => {
    setSelectedRowIndex(index);
    setFormOpen(true);
  };
  const handleDelete = (index) => {
    setSelectedRowIndex(index);
    setOpen(true);
  };
  const handleSubmit = (values, setSubmitting) => {
    let metadata = {};
    if (values.userData) {
      metadata = JSON.parse(values.userData);
    }
    metadata.name = values.name;
    metadata.description = values.description;
    const formData = new FormData();
    formData.append("file", values.file);
    formData.append("metadata", `{"metadata":${JSON.stringify(metadata)}}`);
    formData.append("providerName", data[selectedRowIndex].metadata.name);
    apiService
      .addCluster(formData)
      .then((res) => {
        res.isNew = true;
        //a newly added cluster will have the below values, so we need not call getAllNetworks here
        res.networks = null;
        res.providerNetworks = null;
        res.networksStatus = "Created";
        if (
          !data[selectedRowIndex].clusters ||
          data[selectedRowIndex].clusters.length === 0
        ) {
          data[selectedRowIndex].clusters = [res];
        } else {
          const updatedClusters = [...data[selectedRowIndex].clusters];
          updatedClusters.push(res);
          data[selectedRowIndex].clusters = updatedClusters;
        }
        setData([...data]);
        setFormOpen(false);
        setNotificationDetails({
          show: true,
          message: `cluster added : ${values.name} `,
          severity: "success",
        });
      })
      .catch((err) => {
        let notificationMessage ;
        if (err.response.status === 403) {
          notificationMessage = err.response.data;
        } else {
          notificationMessage = "Error onboarding cluster : " + err;
        }
          setNotificationDetails({
            show: true,
            message: notificationMessage,
            severity: "error",
          });
          setSubmitting(false);
        console.log("error adding cluster : " + err);
      });
  };
  const handleFormClose = () => {
    setFormOpen(false);
  };
  const handleDeleteCluster = (providerRow, clusterRow) => {
    data[providerRow].clusters.splice(clusterRow, 1);
    setData([...data]);
  };
  const handleUpdateCluster = (providerRow, updatedData) => {
    data[providerRow].clusters = updatedData;
    setData([...data]);
  };
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      apiService
        .deleteClusterProvider(data[selectedRowIndex].metadata.name)
        .then(() => {
          console.log("Cluster Provider deleted");
          data.splice(selectedRowIndex, 1);
          let updatedData = data.slice();
          setData(updatedData);
          setExpanded(false);
        })
        .catch((err) => {
          console.log("Error deleting cluster provider : ", err);
          let notificationMessage =
            "Error deleting cluster provider : something went wrong";
          if (err.response.status === 409) {
            notificationMessage =
              "Error deleting cluster provider : remove clusters first";
          }
          setNotificationDetails({
            show: true,
            message: `${notificationMessage}`,
            severity: "error",
          });
        });
    }
    setOpen(false);
    setSelectedRowIndex(0);
  };
  // const handleEdit = (index) => {
  //   setSelectedRowIndex(index);
  //   setOpenProviderForm(true);
  // };
  // const handleCloseProviderForm = () => {
  //   setOpenProviderForm(false);
  // };
  // const handleSubmitProviderForm = (values) => {
  //   let request = {
  //     payload: { metatada: values },
  //     providerName: data[selectedRowIndex].metadata.name,
  //   };
  //   apiService
  //     .updateClusterProvider(request)
  //     .then((res) => {
  //       setData((data) => {
  //         data[selectedRowIndex].metadata = res.metadata;
  //         return data;
  //       });
  //     })
  //     .catch((err) => {
  //       console.log("error updating cluster provider. " + err);
  //     })
  //     .finally(() => {
  //       setOpenProviderForm(false);
  //     });
  // };
  return (
    <>
      <Notification notificationDetails={notificationDetails} />
      {data && data.length > 0 && (
        <div className={classes.root}>
          <ClusterForm
            open={formOpen}
            onClose={handleFormClose}
            onSubmit={handleSubmit}
            existingClusters={data}
            providerIndex={selectedRowIndex}
          />
          {/* <ClusterProviderForm
            open={openProviderForm}
            onClose={handleCloseProviderForm}
            onSubmit={handleSubmitProviderForm}
            item={data[selectedRowIndex]}
          /> */}
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Delete Cluster Provider"}
            content={`Are you sure you want to delete "${
              data[selectedRowIndex] ? data[selectedRowIndex].metadata.name : ""
            }" ?`}
          />
          {data.map((item, index) => (
            <Accordion
              TransitionProps={{ unmountOnExit: true }}
              key={item.metadata.name + "" + index}
              expanded={expanded === item.metadata.name}
              onChange={handleAccordionOpen(item.metadata.name, index)}
            >
              <AccordionSummary
                expandIcon={<ExpandMoreIcon />}
                id={`${index}-header`}
              >
                <Typography className={classes.heading}>
                  {item.metadata.name}
                </Typography>
                <Typography className={classes.secondaryHeading}>
                  {item.metadata.description}
                </Typography>
              </AccordionSummary>
              <div style={{ padding: "8px 16px 16px" }}>
                <Button
                  variant="outlined"
                  size="small"
                  color="primary"
                  onClick={() => {
                    onAddCluster(index);
                  }}
                >
                  Onboard Cluster
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  color="secondary"
                  style={{ float: "right", marginLeft: "10px" }}
                  startIcon={<DeleteIcon />}
                  onClick={() => {
                    handleDelete(index);
                  }}
                >
                  Delete Provider
                </Button>
                {/* 
                //edit cluster provider is not supported by the api yet
                <Button
                  variant="outlined"
                  size="small"
                  color="primary"
                  style={{ float: "right" }}
                  startIcon={<EditIcon />}
                  onClick={() => {
                    handleEdit(index);
                  }}
                >
                  Edit Provider
                </Button> */}
              </div>
              <AccordionDetails>
                {item.clusters && (
                  <ClustersTable
                    onRefreshNetworkData={getAllNetworks}
                    clustersData={item.clusters}
                    providerName={item.metadata.name}
                    parentIndex={index}
                    onDeleteCluster={handleDeleteCluster}
                    onUpdateCluster={handleUpdateCluster}
                  />
                )}
                {item.clusters == null && <span>No Clusters</span>}
              </AccordionDetails>
            </Accordion>
          ))}
        </div>
      )}
    </>
  );
}
