import React, {useEffect, useState} from "react";
import {makeStyles} from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import {
    Button,
    CircularProgress,
    FormControl,
    FormHelperText,
    Grid,
    InputLabel,
    MenuItem,
    Select,
    TextField
} from "@material-ui/core";
import CardContent from "@material-ui/core/CardContent";
import Card from "@material-ui/core/Card";
import apiService from "../../services/apiService";
import {FieldArray, getIn} from "formik";
import Notification from "../../common/Notification";
import IconButton from "@material-ui/core/IconButton";
import DeleteIcon from "@material-ui/icons/Delete";
import AddIcon from "@material-ui/icons/Add";

function NetworkForm({formikProps, clusters, labels, ...props}) {
    const [notificationDetails, setNotificationDetails] = useState({});
    const [totalNetworks, setTotalNetworks] = useState([]);
    const [isLoading, setIsLoading] = useState(true);
    const useStyles = makeStyles({
        root: {
            minWidth: 275,
        },
        title: {
            fontSize: 14,
        },
        pos: {
            marginBottom: 12,
        },
    });

    const setInitState = (networkData) => {
        let interfaces = formikProps.values.apps[props.index].interfaces;
        if (interfaces) {
            let commonData = interfaces.filter((o1) =>
                networkData.some((o2) => o1.networkName === o2.name)
            );
            formikProps.setFieldValue(`apps[${props.index}].interfaces`, commonData);
        }
    };

    const initNetworkDataBySelectedClusters = (
        clusterProvider,
        networkData,
        providerIndex,
        selectedClusters
    ) => {

        let networksPromises = selectedClusters.map((cluster, clusterIndex) => apiService.getAllClusterNetworks({
                providerName: clusterProvider.clusterProvider,
                clusterName: cluster.name,
            })
        );

        return Promise.all(networksPromises).then(res => new Promise((resolve) => resolve(res)))
    };
    const initNetworkData = (combinedNetworkResponse) => {
        let networkData = [];
        combinedNetworkResponse.forEach(element => {
            element.forEach(element2 => {
                if (element2.spec.networks && element2.spec.networks.length > 0) {
                    element2.spec.networks.forEach((network) => {
                        //if two or more clusters have networks with same name, then add it only once
                        if (
                            networkData.findIndex(
                                (element) => element.name === network.metadata.name
                            ) !== -1
                        ) {
                            console.log(
                                `Provider Network : ${network.metadata.name} already exists`
                            );
                        } else {
                            networkData.push({
                                name: network.metadata.name,
                                subnets: network.spec.ipv4Subnets,
                            });
                        }
                    });
                }

                if (
                    element2.spec["providerNetworks"] &&
                    element2.spec["providerNetworks"].length > 0
                ) {
                    element2.spec["providerNetworks"].forEach((providerNetwork) => {
                        //if two or more clusters have provider networks with same name, then add it only once
                        if (
                            networkData.findIndex(
                                (element) => element.name === providerNetwork.metadata.name
                            ) !== -1
                        ) {
                            console.log(
                                `Network : ${providerNetwork.metadata.name} already exists`
                            );
                        } else {
                            networkData.push({
                                name: providerNetwork.metadata.name,
                                subnets: providerNetwork.spec.ipv4Subnets,
                            });
                        }
                    });
                }
            })
        })

        setInitState(networkData);
        if (networkData.length > 0) {
            setTotalNetworks(networkData);
        } else {
            setNotificationDetails({
                show: true,
                message: `No network available for selected cluster(s)`,
                severity: "warning",
            });
        }
        setIsLoading(false);
        // init(networkData);
    }
    const initNetworkDataBySelectedLabels = (
        clusterProvider,
        networkData,
        providerIndex,
        labels
    ) => {
        if (labels && labels.length > 0) {
            let clusterRequests = [];
            labels.forEach((label) => {
                clusterRequests.push(
                    apiService.getClustersByLabel(
                        clusterProvider.clusterProvider,
                        label.clusterLabel
                    )
                );
            });

            return Promise.all(clusterRequests).then((res) => {
                let overAllClusterList = [];
                res.forEach((clusterRes) => {
                    overAllClusterList = [...overAllClusterList, ...clusterRes];
                });
                //we need unique clusters so add the values in a set
                const overAllClustersSet = new Set(overAllClusterList);
                let selectedClusters = [];

                //initNetworkDataBySelectedClusters expects clusters data in this format as this function is used when passing individual clusters too
                overAllClustersSet.forEach((clusterName) => {
                    selectedClusters.push({name: clusterName});
                });

                return initNetworkDataBySelectedClusters(
                    clusterProvider,
                    networkData,
                    providerIndex,
                    selectedClusters
                );
            });
        }
    };

    useEffect(() => {
        let networkData = [];
        let prList = [];
        if (clusters && clusters.length > 0) {
            clusters && clusters.forEach((clusterProvider, providerIndex) => {
                if (formikProps.values.apps[props.index].placementType === "clusters") {
                    prList.push(initNetworkDataBySelectedClusters(
                        clusterProvider,
                        networkData,
                        providerIndex,
                        clusterProvider.selectedClusters
                    ));
                } else {
                    prList.push(initNetworkDataBySelectedLabels(
                        clusterProvider,
                        networkData,
                        providerIndex,
                        clusterProvider.selectedLabels
                    ));
                }
            })

            Promise.all(prList).then(res => {
                initNetworkData(res);
            }).catch(err => {
                console.error("error getting cluster networks" + err);
                setNotificationDetails({
                    show: true,
                    message: `Error getting cluster networks`,
                    severity: "error",
                });
            });
        }
    }, []);

    const getAvailableNetworks = (networkData, updatedFields) => {
        let availableNetworks = [];
        networkData.forEach((network) => {
            let match = false;
            updatedFields &&
            updatedFields.forEach((networkInterface) => {
                if (network.name === networkInterface.networkName) {
                    match = true;
                }
            });
            if (!match) availableNetworks.push(network);
        });
        return availableNetworks;
    };

    const classes = useStyles();
    return (
        <>
            <Notification notificationDetails={notificationDetails}/>
            <Grid
                key="networkForm"
                container
                spacing={3}
                style={{
                    height: "400px",
                    overflowY: "auto",
                    width: "100%",
                    marginTop: "10px",
                }}
            >
                {(!clusters || clusters.length < 1) && (
                    <Grid item xs={12}>
                        <Typography variant="h6">No clusters selected</Typography>
                    </Grid>
                )}
                {clusters && (
                    <Grid item xs={12}>
                        <Card className={classes.root}>
                            <CardContent>
                                <Grid container spacing={2}>
                                    <FieldArray name={`apps[${props.index}].interfaces`}
                                                render={(arrayHelpers) => {
                                                    const {values, errors, handleChange, handleBlur} =
                                                        formikProps;
                                                    const removeInterface = (interfaceIndex) => {
                                                        arrayHelpers.remove(interfaceIndex);
                                                    };
                                                    return (
                                                        <>
                                                            {
                                                                !isLoading && values.apps[props.index].interfaces && values.apps[props.index].interfaces.length > 0
                                                                    ? values.apps[props.index].interfaces.map((networkInterface, interfaceIndex) => (
                                                                        <Grid
                                                                            spacing={1}
                                                                            container
                                                                            item
                                                                            key={interfaceIndex}
                                                                            xs={12}>

                                                                            <Grid item xs={3}>
                                                                                <FormControl
                                                                                    fullWidth
                                                                                    required
                                                                                    error={
                                                                                        Boolean(getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].networkName`))
                                                                                    }
                                                                                >
                                                                                    <InputLabel
                                                                                        id="network-select-label">
                                                                                        Network
                                                                                    </InputLabel>
                                                                                    <Select
                                                                                        fullWidth
                                                                                        labelId="network-select-label"
                                                                                        id="network-select"
                                                                                        name={`apps[${props.index}].interfaces[${interfaceIndex}].networkName`}
                                                                                        value={
                                                                                            values.apps[props.index].interfaces[
                                                                                                interfaceIndex
                                                                                                ].networkName
                                                                                        }
                                                                                        onChange={handleChange}
                                                                                    >
                                                                                        {values.apps[props.index].interfaces[
                                                                                            interfaceIndex
                                                                                            ].networkName && (
                                                                                            <MenuItem
                                                                                                key={
                                                                                                    values.apps[props.index]
                                                                                                        .interfaces[interfaceIndex]
                                                                                                        .networkName
                                                                                                }
                                                                                                value={
                                                                                                    values.apps[props.index]
                                                                                                        .interfaces[interfaceIndex]
                                                                                                        .networkName
                                                                                                }
                                                                                            >
                                                                                                {
                                                                                                    values.apps[props.index]
                                                                                                        .interfaces[interfaceIndex]
                                                                                                        .networkName
                                                                                                }
                                                                                            </MenuItem>
                                                                                        )}
                                                                                        {getAvailableNetworks(totalNetworks, values.apps[props.index].interfaces) &&
                                                                                        getAvailableNetworks(totalNetworks, values.apps[props.index].interfaces).map((network) => (
                                                                                            <MenuItem
                                                                                                key={network.name}
                                                                                                value={network.name}
                                                                                            >
                                                                                                {network.name}
                                                                                            </MenuItem>
                                                                                        ))}
                                                                                    </Select>
                                                                                    <FormHelperText>{getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].networkName`)}</FormHelperText>
                                                                                </FormControl>
                                                                            </Grid>
                                                                            <Grid item xs={3}>
                                                                                <FormControl
                                                                                    fullWidth
                                                                                    required
                                                                                    error={
                                                                                        Boolean(getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].subnet`))
                                                                                    }
                                                                                >
                                                                                    <InputLabel
                                                                                        id="subnet-select-label">
                                                                                        Subnet
                                                                                    </InputLabel>
                                                                                    <Select
                                                                                        fullWidth
                                                                                        labelId="subnet-select-label"
                                                                                        id="subnet-select-label"
                                                                                        name={`apps[${props.index}].interfaces[${interfaceIndex}].subnet`}
                                                                                        value={
                                                                                            values.apps[props.index].interfaces[
                                                                                                interfaceIndex
                                                                                                ].subnet
                                                                                        }
                                                                                        onChange={handleChange}
                                                                                    >
                                                                                        {values.apps[props.index].interfaces[
                                                                                            interfaceIndex
                                                                                            ].networkName === ""
                                                                                            ? null
                                                                                            : totalNetworks
                                                                                                .filter(
                                                                                                    (network) =>
                                                                                                        network.name ===
                                                                                                        values.apps[props.index]
                                                                                                            .interfaces[
                                                                                                            interfaceIndex
                                                                                                            ].networkName
                                                                                                )[0]
                                                                                                .subnets.map((subnet) => (
                                                                                                    <MenuItem
                                                                                                        key={subnet.name}
                                                                                                        value={subnet.name}
                                                                                                    >
                                                                                                        {subnet.name}(
                                                                                                        {subnet.subnet})
                                                                                                    </MenuItem>
                                                                                                ))}
                                                                                    </Select>
                                                                                    <FormHelperText>{getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].subnet`)}</FormHelperText>
                                                                                </FormControl>
                                                                            </Grid>
                                                                            <Grid item xs={3}>
                                                                                <TextField
                                                                                    name={`apps[${props.index}].interfaces[${interfaceIndex}].ip`}
                                                                                    onBlur={handleBlur}
                                                                                    id="ip"
                                                                                    label="IP Address"
                                                                                    value={
                                                                                        values.apps[props.index].interfaces[
                                                                                            interfaceIndex
                                                                                            ].ip
                                                                                    }
                                                                                    onChange={handleChange}
                                                                                    helperText={
                                                                                        getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].ip`) || "blank for auto assign"
                                                                                    }
                                                                                    error={
                                                                                        Boolean(getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].ip`))
                                                                                    }
                                                                                />
                                                                            </Grid>
                                                                            <Grid item xs={3}>
                                                                                <TextField
                                                                                    style={{width: "75%"}}
                                                                                    name={`apps[${props.index}].interfaces[${interfaceIndex}].interfaceName`}
                                                                                    onBlur={handleBlur}
                                                                                    id="interfaceName"
                                                                                    label="Interface Name"
                                                                                    value={
                                                                                        values.apps[props.index].interfaces[
                                                                                            interfaceIndex
                                                                                            ].interfaceName
                                                                                    }
                                                                                    onChange={handleChange}
                                                                                    helperText={
                                                                                        getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].interfaceName`)
                                                                                    }
                                                                                    error={
                                                                                        Boolean(getIn(errors, `apps[${props.index}].interfaces[${interfaceIndex}].interfaceName`))
                                                                                    }
                                                                                />
                                                                                <IconButton
                                                                                    style={{
                                                                                        padding: "0 0 5px 0",
                                                                                        verticalAlign: "bottom",
                                                                                        marginLeft: "18px"
                                                                                    }}
                                                                                    color="secondary"
                                                                                    onClick={() => {
                                                                                        removeInterface(interfaceIndex);
                                                                                    }}
                                                                                >
                                                                                    <DeleteIcon
                                                                                        fontSize="small"/>
                                                                                </IconButton>
                                                                            </Grid>

                                                                        </Grid>)) : null
                                                            }
                                                            <Grid item xs={12}>
                                                                <Button
                                                                    variant="outlined"
                                                                    size="small"
                                                                    fullWidth
                                                                    color="primary"
                                                                    disabled={
                                                                        isLoading || (values.apps[props.index].interfaces && totalNetworks.length === values.apps[props.index].interfaces.length)
                                                                    }
                                                                    onClick={() => {
                                                                        arrayHelpers.push({
                                                                            networkName: "",
                                                                            ip: "",
                                                                            subnet: "",
                                                                            interfaceName: undefined
                                                                        })
                                                                    }}
                                                                    startIcon={
                                                                        isLoading ? (
                                                                            <CircularProgress
                                                                                style={{
                                                                                    width: "20px",
                                                                                    height: "20px"
                                                                                }}
                                                                            />
                                                                        ) : (
                                                                            <AddIcon/>
                                                                        )
                                                                    }
                                                                >
                                                                    Add Network Interface
                                                                </Button>
                                                            </Grid>
                                                        </>
                                                    )
                                                }}
                                    />
                                </Grid>
                            </CardContent>
                        </Card>
                    </Grid>
                )}
            </Grid>
        </>
    );
}

export default NetworkForm;
