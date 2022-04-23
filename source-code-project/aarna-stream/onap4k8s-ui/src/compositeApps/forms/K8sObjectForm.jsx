import React, {useEffect, useState} from "react";
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import {
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  FormHelperText,
  Grid,
  InputLabel,
  MenuItem,
  Radio,
  Select,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
} from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import RadioGroup from "@material-ui/core/RadioGroup";
import FileUpload from "../../common/FileUpload";
import Paper from "@material-ui/core/Paper";
import {Formik, getIn} from "formik";
import * as Yup from "yup";
import IconButton from "@material-ui/core/IconButton";
import InfoOutlinedIcon from "@material-ui/icons/InfoOutlined";
import DeleteIcon from "@material-ui/icons/Delete";

const RESOURCE_FILE_SUPPORTED_FORMATS = [
  "json",
  "application/json",
  "application/x-yaml",
  ""
];

const getResourceSchema = () => {
  return Yup.object({
    rSpec: Yup.object({
      newObject: Yup.string(),
      resourceGVK: Yup.object().when('newObject', {
        is: (value) => value === "false",
        then: Yup.object({
          kind: Yup.string().required("Resource Type is required"),
          name: Yup.string().required("Name is required"),
          apiVersion: Yup.string().required("Api Version is required")
        })
      })
    }),
    cSpec: Yup.object({
      clusterSpecific: Yup.string(),
      clusterInfo: Yup
          .object()
          .when("clusterSpecific", {
            is: "true",
            then: Yup.object({
              scope: Yup.string(),
              clusterProvider: Yup.string().required("This field is required"),
              cluster: Yup.string(),
              clusterLabel: Yup.string(),
              mode: Yup.string(),
            }).required("This field is required")
          }),
    }).when('rSpec', {
      is: (value) => value.newObject === "true",
      then: Yup.object({
        file: Yup.mixed()
            .required("A YAML/JSON resource file is required")
            .test("fileFormat", "Unsupported file format",
                (value) =>
                    value && RESOURCE_FILE_SUPPORTED_FORMATS.includes(value.type)
            )
      }),
      otherwise: Yup.object({
        patchJson: Yup.array().of(Yup.object().typeError("Invalid patch value, expected JSON array"))
            .required('This field is required').typeError("Invalid patch value, expected JSON array"),
      }),
    }),
  });
}
const AddResourceDialog = ({item, clusters, placementCriterion, ...props}) => {
  const [selectedCluster, setSelectedCluster] = useState("");
  const [isCustomResource, setIsCustomResource] = useState(false);
  const handleSelectCluster = (e, setFieldValue) => {
    setSelectedCluster(e.target.value);
    const jsonVal = JSON.parse(e.target.value)
    if (jsonVal.cluster) {
      setFieldValue('cSpec.clusterInfo', {
        scope: "name",
        clusterProvider: jsonVal.clusterProvider,
        cluster: jsonVal.cluster,
        clusterLabel: "",
        mode: "allow"
      })
    } else {
      setFieldValue('cSpec.clusterInfo', {
        scope: "label",
        clusterProvider: jsonVal.clusterProvider,
        name: "",
        clusterLabel: jsonVal.clusterLabel,
        mode: "allow"
      })
    }
  }
  useEffect(() => {
    if (item) {
      let selectedClusterDropdownValue;
      if (item.cSpec.clusterInfo.cluster) {
        selectedClusterDropdownValue = `{"clusterProvider":"${item.cSpec.clusterInfo.clusterProvider}","clusterLabel":"${item.cSpec.clusterInfo.name}"}`
      } else {
        selectedClusterDropdownValue = `{"clusterProvider":"${item.cSpec.clusterInfo.clusterProvider}","clusterLabel":"${item.cSpec.clusterInfo.clusterLabel}"}`
      }
      setSelectedCluster(selectedClusterDropdownValue);
    } else {
      setSelectedCluster("");
    }
  }, [item])

  let initialValues = item ? {...item} : {
    rSpec: {newObject: "true", resourceGVK: {kind: "", name: "", apiVersion: ""}},
    cSpec: {
      clusterSpecific: 'false',
      patchJson: "",
      file: null,
      clusterInfo: {
        scope: "",
        clusterProvider: "",
        cluster: "",
        clusterLabel: "",
        mode: ""
      }
    }
  }
  const isView = !!item;

  const handleSelectResourceType = (event, setFieldValue) => {
    setFieldValue('rSpec.resourceGVK.kind', "")
    setIsCustomResource(event.target.checked);
  }

  return (
      <Dialog open={props.open}>
        <DialogTitle>{isView ? "View Resource" : "Add Resource"}</DialogTitle>
        <Formik
            initialValues={initialValues}
            onSubmit={(values) => {
              setSelectedCluster("");
              props.onSubmit(values);
            }}
            validationSchema={getResourceSchema()}
        >
          {(formikProps) => {
            const {
              values,
              touched,
              errors,
              setFieldValue,
              handleChange,
              handleBlur,
              handleSubmit,
            } = formikProps;
            return (
                <form noValidate onSubmit={handleSubmit}>
                  <DialogContent>
                    {!isView && <RadioGroup
                        row
                        name="rSpec.newObject"
                        value={values.rSpec.newObject}
                        onChange={handleChange}
                        style={{marginBottom: "20px"}}
                    >
                      <FormControlLabel
                          value={"true"}
                          control={<Radio/>}
                          label="New"/>
                      <FormControlLabel
                          value={"false"}
                          control={<Radio/>}
                          label="Patch"
                      />
                    </RadioGroup>}
                    <Grid container spacing={3} style={{marginBottom: "20px"}}>
                      {values.rSpec.newObject === "true" && (
                          <Grid item xs={12}>
                            <Typography>Resource File</Typography>
                            <FileUpload
                                disabled={isView}
                                setFieldValue={setFieldValue}
                                file={values.cSpec.file}
                                onBlur={handleBlur}
                                name={`cSpec.file`}
                                onChange={handleChange}
                                accept={".json, .yaml, .yml"}
                                helperText={errors.cSpec && errors.cSpec.file}
                                error={errors.cSpec && errors.cSpec.file && true}
                            />
                            {touched.cSpec && touched.cSpec.file && errors.cSpec && errors.cSpec.file && (
                                <p style={{color: "#f44336"}}>{errors.cSpec.file}</p>
                            )}
                          </Grid>
                      )}
                      {values.rSpec.newObject === "false" && (
                          <>
                            <Grid item xs={12}>

                                <FormControlLabel control={<Checkbox
                                    disabled={isView}
                                    checked={isCustomResource}
                                    onChange={(el) => handleSelectResourceType(el, setFieldValue)}
                                />} label="Custom Resource"/>

                            </Grid>
                            {isCustomResource ? <Grid item xs={4}>
                              <TextField
                                  disabled={isView}
                                  fullWidth
                                  name={'rSpec.resourceGVK.kind'}
                                  value={
                                    values.rSpec.resourceGVK.kind
                                  }
                                  label="Resource Name"
                                  type="text"
                                  onChange={handleChange}
                                  onBlur={handleBlur}
                                  required
                                  helperText={
                                    getIn(touched, 'rSpec.resourceGVK.kind') &&
                                    getIn(errors, 'rSpec.resourceGVK.kind')}
                                  error={Boolean(getIn(touched, 'rSpec.resourceGVK.kind') &&
                                      getIn(errors, 'rSpec.resourceGVK.kind'))}
                              />
                            </Grid> : <Grid item xs={4}>
                              <FormControl fullWidth
                                           error={Boolean(
                                               getIn(touched, 'rSpec.resourceGVK.kind') &&
                                               getIn(errors, 'rSpec.resourceGVK.kind'))}
                              >
                                <InputLabel id="resource-cluster-select">Resource Type</InputLabel>
                                <Select
                                    disabled={isView}
                                    margin={"dense"}
                                    fullWidth
                                    name={'rSpec.resourceGVK.kind'}
                                    labelId="resource-type-select"
                                    value={
                                      values.rSpec.resourceGVK.kind
                                    }
                                    onChange={handleChange}
                                >{
                                  ["Deployment", "ConfigMap", "Service", "Secret", "Pod", "StatefulSet", "DaemonSet", "PersistentVolumeClaim", "PersistentVolume", "StorageClass"].map((type) =>
                                      <MenuItem key={type} value={type}>{type}</MenuItem>)
                                }
                                </Select>
                                <FormHelperText>{
                                  getIn(touched, 'rSpec.resourceGVK.kind') &&
                                  getIn(errors, 'rSpec.resourceGVK.kind')
                                }</FormHelperText>
                              </FormControl>
                            </Grid>}
                            <Grid item xs={4}>
                              <TextField
                                  disabled={isView}
                                  fullWidth
                                  name={`rSpec.resourceGVK.apiVersion`}
                                  label="Api Version"
                                  type="text"
                                  value={
                                    values.rSpec.resourceGVK.apiVersion
                                  }
                                  onChange={handleChange}
                                  onBlur={handleBlur}
                                  required
                                  helperText={
                                    getIn(touched, 'rSpec.resourceGVK.apiVersion') &&
                                    getIn(errors, 'rSpec.resourceGVK.apiVersion')}
                                  error={Boolean(getIn(touched, 'rSpec.resourceGVK.apiVersion') &&
                                      getIn(errors, 'rSpec.resourceGVK.apiVersion'))}
                              />
                            </Grid>
                            <Grid item xs={4}>
                              <TextField
                                  disabled={isView}
                                  fullWidth
                                  name={`rSpec.resourceGVK.name`}
                                  label="Name"
                                  type="text"
                                  value={
                                    values.rSpec.resourceGVK.name
                                  }
                                  onChange={handleChange}
                                  onBlur={handleBlur}
                                  required
                                  helperText={
                                    getIn(touched, 'rSpec.resourceGVK.name') &&
                                    getIn(errors, 'rSpec.resourceGVK.name')}
                                  error={Boolean(
                                      getIn(touched, 'rSpec.resourceGVK.name') &&
                                      getIn(errors, 'rSpec.resourceGVK.name'))}
                              />
                            </Grid>
                            <Grid item xs={12}>
                              <TextField
                                  disabled={isView}
                                  fullWidth
                                  label="Resource Json Patch"
                                  name={`cSpec.patchJson`}
                                  type="text"
                                  value={isView ? JSON.stringify(values.cSpec.patchJson) : values.cSpec.patchJson}
                                  onChange={handleChange}
                                  onBlur={handleBlur}
                                  multiline
                                  rows={7}
                                  variant="outlined"
                                  required
                                  helperText={
                                    getIn(touched, 'cSpec.patchJson') &&
                                    getIn(errors, 'cSpec.patchJson')}
                                  error={Boolean(getIn(touched, 'cSpec.patchJson') &&
                                      getIn(errors, 'cSpec.patchJson'))}
                              />
                            </Grid>
                          </>
                      )}
                      <Grid item xs={12}>
                        <RadioGroup
                            row
                            name="cSpec.clusterSpecific"
                            value={values.cSpec.clusterSpecific}
                            onChange={handleChange}
                            style={{marginBottom: "20px"}}
                        >
                          <FormControlLabel
                              disabled={isView}
                              value={'false'}
                              control={<Radio/>}
                              label="All Clusters"
                          />
                          <FormControlLabel
                              disabled={!clusters || clusters.length < 1 || isView || (placementCriterion === "anyOf")}
                              value={'true'}
                              control={<Radio/>}
                              label="Cluster Specific"/>
                        </RadioGroup>
                      </Grid>
                      {values.cSpec.clusterSpecific === 'true' &&
                      <Grid item xs={12}>
                        <FormControl fullWidth
                                     error={Boolean(
                                         getIn(touched, 'cSpec.clusterInfo.clusterProvider') &&
                                         getIn(errors, 'cSpec.clusterInfo.clusterProvider'))}
                        >
                          <InputLabel id="resource-cluster-select">Select Cluster/Label</InputLabel>
                          <Select
                              disabled={isView}
                              margin={"dense"}
                              fullWidth
                              labelId="resource-cluster-select"
                              value={selectedCluster}
                              name="selectedCluster"
                              onChange={(e) => handleSelectCluster(e, setFieldValue)}
                          >
                            {clusters.map(cluster => {
                              if (cluster.selectedLabels) {
                                return cluster.selectedLabels.map(selectedLabel => <MenuItem
                                    id={cluster.clusterProvider + selectedLabel.clusterLabel}
                                    value={`{"clusterProvider":"${cluster.clusterProvider}","clusterLabel":"${selectedLabel.clusterLabel}"}`}>
                                  {cluster.clusterProvider + " : " + selectedLabel.clusterLabel}
                                </MenuItem>)
                              } else {
                                return cluster.selectedClusters.map(selectedCluster => <MenuItem
                                    id={cluster.clusterProvider + selectedCluster.name}
                                    value={`{"clusterProvider":"${cluster.clusterProvider}","cluster":"${selectedCluster.name}"}`}>
                                  {cluster.clusterProvider + " : " + selectedCluster.name}
                                </MenuItem>)
                              }
                            })}
                          </Select>
                          <FormHelperText>{
                            getIn(touched, 'cSpec.clusterInfo.clusterProvider') &&
                            getIn(errors, 'cSpec.clusterInfo.clusterProvider')
                          }</FormHelperText>
                        </FormControl>
                      </Grid>}
                    </Grid>
                  </DialogContent>
                  <DialogActions>
                    <Button autoFocus onClick={() => {
                      props.handleClose();
                      setSelectedCluster("");
                    }} color={isView ? "primary" : "secondary"}>
                      {isView ? "OK" : "Cancel"}
                    </Button>
                    {!isView && <Button
                        type="submit"
                        color="primary"
                    >OK</Button>}
                  </DialogActions>
                </form>);
          }}
        </Formik>
      </Dialog>
  )
}

function K8sObjectForm({formikProps, clusters, appName, placementCriterion, ...props}) {
  const [resourceFormOpen, setResourceFormOpen] = useState(false);
  const [itemToEdit, setItemToEdit] = useState(null);
  const handleAddResource = () => {
    setResourceFormOpen(true);
  }
  const handleCloseResourceDialog = () => {
    setResourceFormOpen(false);
    setItemToEdit(null);
  }

  const handleResourceFormSubmit = (values) => {
    if (values.rSpec.newObject === "false") {
      values.cSpec.patchType = "json";
      //we don't want any escape chars and newline/space strings
      values.cSpec.patchJson = JSON.parse(values.cSpec.patchJson)
    } else {
      //resource GVK is not required in case of a new resource
      values.rSpec.resourceGVK = {};
    }
    if (formikProps.values.apps[props.index].resourceData) {
      formikProps.setFieldValue(`apps[${props.index}].resourceData`, [...formikProps.values.apps[props.index].resourceData, values]);
    } else {
      formikProps.setFieldValue(`apps[${props.index}].resourceData`, [values]);
    }
    handleCloseResourceDialog();
  }
  const handleRemoveResource = (index) => {
    let existingResources = [...formikProps.values.apps[props.index].resourceData];
    existingResources.splice(index, 1);
    formikProps.setFieldValue(`apps[${props.index}].resourceData`, existingResources);
  }
  const handleEditResource = (resource) => {
    setItemToEdit(resource);
    setResourceFormOpen(true);
  }
  return (
      <>
        <AddResourceDialog open={resourceFormOpen} handleClose={handleCloseResourceDialog}
                           onSubmit={handleResourceFormSubmit} item={itemToEdit} clusters={clusters}
                           placementCriterion={placementCriterion}/>
        {formikProps.values.apps[props.index].resourceData && formikProps.values.apps[props.index].resourceData.length > 0 ?
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead
                    style={{backgroundColor: "rgb(234, 239, 241)"}}>
                  <TableRow>
                    <TableCell>Type</TableCell>
                    <TableCell>Name</TableCell>
                    <TableCell>Api Version</TableCell>
                    <TableCell>Kind</TableCell>
                    <TableCell>Cluster Specific</TableCell>
                    <TableCell>Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {formikProps.values.apps[props.index].resourceData.map((entry, index) => (
                      <TableRow key={index}>
                        <TableCell>{entry.rSpec.newObject === "true" ? "New" : "Patch"}</TableCell>
                        <TableCell>{entry.rSpec.resourceGVK.name || <i>NA</i>}</TableCell>
                        <TableCell>{entry.rSpec.resourceGVK.apiVersion || <i>NA</i>}</TableCell>
                        <TableCell>{entry.rSpec.resourceGVK.kind || <i>NA</i>}</TableCell>
                        <TableCell>{entry.cSpec.clusterSpecific}</TableCell>
                        <TableCell>
                          <IconButton
                              onClick={() => {
                                handleEditResource(entry)
                              }}
                              title="Edit">
                            <InfoOutlinedIcon color="primary"/>
                          </IconButton>
                          <IconButton
                              onClick={() => {
                                handleRemoveResource(index)
                              }}
                              title="Delete">
                            <DeleteIcon color="secondary"/>
                          </IconButton></TableCell>
                      </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer> :
            <Typography gutterBottom variant="h6">
              None
            </Typography>}
        <Button
            variant="outlined"
            color="primary"
            startIcon={<AddIcon/>}
            onClick={handleAddResource}
            fullWidth
            style={{marginTop: "20px"}}
        >Add
        </Button>
      </>
  );
}

export default K8sObjectForm;
