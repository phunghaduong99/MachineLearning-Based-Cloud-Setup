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
import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import MuiDialogTitle from "@material-ui/core/DialogTitle";
import MuiDialogContent from "@material-ui/core/DialogContent";
import MuiDialogActions from "@material-ui/core/DialogActions";
import IconButton from "@material-ui/core/IconButton";
import CloseIcon from "@material-ui/icons/Close";
import Typography from "@material-ui/core/Typography";
import { TextField } from "@material-ui/core";
import * as Yup from "yup";
import { Formik } from "formik";
import FileUpload from "../../../common/FileUpload";
import LoadingButton from "../../../common/LoadingButton";

const styles = (theme) => ({
  root: {
    margin: 0,
    padding: theme.spacing(2),
  },
  closeButton: {
    position: "absolute",
    right: theme.spacing(1),
    top: theme.spacing(1),
    color: theme.palette.grey[500],
  },
});

const DialogTitle = withStyles(styles)((props) => {
  const { children, classes, onClose, ...other } = props;
  return (
    <MuiDialogTitle disableTypography className={classes.root} {...other}>
      <Typography variant="h6">{children}</Typography>
      {onClose ? (
        <IconButton className={classes.closeButton} onClick={onClose}>
          <CloseIcon />
        </IconButton>
      ) : null}
    </MuiDialogTitle>
  );
});

const DialogActions = withStyles((theme) => ({
  root: {
    margin: 0,
    padding: theme.spacing(1),
  },
}))(MuiDialogActions);

const DialogContent = withStyles((theme) => ({
  root: {
    padding: theme.spacing(2),
  },
}))(MuiDialogContent);

const getSchema = (existingClusters) => {
  let schema = {};
  schema = Yup.object({
    name: Yup.string()
      .required("Cluster Name is required")
      .test(
        "duplicate-test",
        "A cluster with the same name exists, please use a different name",
        (name) => {
          return existingClusters
            ? existingClusters.findIndex((x) => x.metadata.name === name) === -1
            : true;
        }
      )
      .matches(
        /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/,
        "Name can only contain letters, numbers, '-', '_' and no spaces. Name must start and end with an alphanumeric character"
      )
      .max(128, "Name cannot exceed 128 characters"),
    description: Yup.string(),
    isEdit: Yup.boolean().required(),
    file: Yup.mixed().when("isEdit", {
      is: false,
      then: Yup.mixed().required("A file is required"),
      otherwise: Yup.string(),
    }),
    userData: Yup.object().typeError("Invalid user data values, expected JSON"),
  });
  return schema;
};

const ClusterForm = (props) => {
  const { onClose, item, open, onSubmit } = props;
  const title = item ? "Edit Cluster" : "Onboard Cluster";
  const handleClose = () => {
    onClose();
  };
  let initialValues = item
    ? {
        name: item.metadata.name,
        description: item.metadata.description,
        file: undefined,
        isEdit: true,
      }
    : {
        name: "",
        description: "",
        file: undefined,
        isEdit: false,
        userData: undefined,
      };

  return (
    <Dialog
      maxWidth={"xs"}
      onClose={handleClose}
      aria-labelledby="customized-dialog-title"
      open={open}
      disableBackdropClick
    >
      <DialogTitle id="simple-dialog-title">{title}</DialogTitle>
      <Formik
        initialValues={initialValues}
        onSubmit={(values, actions) => {
          onSubmit(values, actions.setSubmitting);
        }}
        validationSchema={getSchema(
          props.existingClusters[props.providerIndex].clusters
        )}
      >
        {(props) => {
          const {
            values,
            touched,
            errors,
            isSubmitting,
            handleChange,
            handleBlur,
            handleSubmit,
            setFieldValue,
          } = props;
          return (
            <form
              encType="multipart/form-data"
              noValidate
              onSubmit={handleSubmit}
            >
              <DialogContent dividers>
                <TextField
                  style={{ width: "100%", marginBottom: "10px" }}
                  id="name"
                  label="Cluster name"
                  type="text"
                  value={values.name}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  helperText={errors.name && touched.name && errors.name}
                  required
                  error={errors.name && touched.name}
                />
                <TextField
                  style={{ width: "100%", marginBottom: "25px" }}
                  name="description"
                  value={values.description}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  id="description"
                  label="Description"
                  multiline
                  rowsMax={4}
                />
                <TextField
                  fullWidth
                  style={{ marginBottom: "25px" }}
                  id="userData"
                  label="User Data"
                  type="text"
                  name="userData"
                  onChange={handleChange}
                  onBlur={handleBlur}
                  multiline
                  rows={4}
                  variant="outlined"
                  error={errors.userData && touched.userData}
                  helperText={
                    errors.userData && touched.userData && errors["userData"]
                  }
                />
                <label
                  className="MuiFormLabel-root MuiInputLabel-root"
                  htmlFor="file"
                  id="file-label"
                >
                  Cluster config file
                  <span className="MuiFormLabel-asterisk MuiInputLabel-asterisk">
                     *
                  </span>
                </label>
                <FileUpload
                  setFieldValue={setFieldValue}
                  file={values.file}
                  name="file"
                />
                {touched.file && (
                  <p style={{ color: "#f44336" }}>{errors.file}</p>
                )}
              </DialogContent>
              <DialogActions>
                <Button
                  autoFocus
                  onClick={handleClose}
                  color="secondary"
                  disabled={isSubmitting}
                >
                  Cancel
                </Button>
                <LoadingButton
                  type="submit"
                  buttonLabel="OK"
                  loading={isSubmitting}
                />
              </DialogActions>
            </form>
          );
        }}
      </Formik>
    </Dialog>
  );
};

ClusterForm.propTypes = {
  onClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
  item: PropTypes.object,
};

export default ClusterForm;
