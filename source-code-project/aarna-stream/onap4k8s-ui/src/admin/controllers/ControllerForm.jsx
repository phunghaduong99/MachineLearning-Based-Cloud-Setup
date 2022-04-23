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
import { Formik } from "formik";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Grid,
  TextField,
} from "@material-ui/core";
import * as Yup from "yup";
import LoadingButton from "../../common/LoadingButton";

ControllerForm.propTypes = {
  onClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
  item: PropTypes.object,
};

const getSchema = (existingControllers) => {
  let schema;
  schema = Yup.object({
    name: Yup.string()
      .required("Name is required")
      .matches(
        /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/,
        "Name can only contain letters, numbers, '-', '_' and no spaces. Name must start and end with an alphanumeric character"
      )
      .test(
        "duplicate-test",
        "A controller with same name exists, please use a different name",
        (name) => {
          return existingControllers
            ? existingControllers.findIndex((x) => x.metadata.name === name) ===
                -1
            : true;
        }
      )
      .max(128, "Name cannot exceed 128 characters"),
    description: Yup.string(),
    host: Yup.string().required("Host is required"),
    port: Yup.number()
      .typeError("Port must be a number")
      .required("Port is required"),
    type: Yup.string(),
    priority: Yup.number().typeError("Port must be a number"),
  });
  return schema;
};

function ControllerForm(props) {
  const { onClose, item, open, existingControllers } = props;
  let initialValues = item
    ? {
        name: item.metadata.name,
        description: item.metadata.description,
        host: item.spec.host,
        port: item.spec.port,
        type: item.spec.type,
        priority: item.spec.priority,
      }
    : {
        name: "",
        description: "",
        host: "",
        port: "",
        type: "",
        priority: "",
      };
  return (
    <Dialog
      maxWidth={"xs"}
      fullWidth
      onClose={() => onClose()}
      aria-labelledby="customized-dialog-title"
      open={open}
      disableBackdropClick
    >
      <DialogTitle id="simple-dialog-title">Register Controller</DialogTitle>
      <Formik
        initialValues={initialValues}
        onSubmit={(values) => {
          onClose(values);
        }}
        validationSchema={getSchema(existingControllers)}
      >
        {(props) => {
          const {
            touched,
            errors,
            values,
            isSubmitting,
            handleChange,
            handleBlur,
            handleSubmit,
          } = props;
          return (
            <form noValidate onSubmit={handleSubmit}>
              <DialogContent dividers>
                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <TextField
                      fullWidth
                      id="name"
                      label="Name"
                      type="text"
                      value={values.name}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      helperText={errors.name && touched.name && errors.name}
                      required
                      error={errors.name && touched.name}
                    />
                  </Grid>
                  <Grid item xs={12}>
                    <TextField
                      fullWidth
                      name="description"
                      value={values.description}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      id="description"
                      label="Description"
                      multiline
                      rowsMax={4}
                    />
                  </Grid>
                </Grid>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <TextField
                      fullWidth
                      name="host"
                      value={values.host}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      id="host"
                      label="Host"
                      helperText={errors.host && touched.host && errors["host"]}
                      required
                      error={errors.host && touched.host}
                    />
                  </Grid>
                  <Grid item xs={6}>
                    <TextField
                      fullWidth
                      name="port"
                      value={values.port}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      id="port"
                      label="Port"
                      helperText={errors.port && touched.port && errors.port}
                      required
                      error={errors.port && touched.port}
                    />
                  </Grid>
                </Grid>

                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <TextField
                      fullWidth
                      name="type"
                      value={values.type}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      id="type"
                      label="Type"
                      helperText={errors.type && touched.type && errors["type"]}
                      error={errors.type && touched.type}
                    />
                  </Grid>
                  <Grid item xs={6}>
                    <TextField
                      fullWidth
                      name="priority"
                      value={values.priority}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      id="priority"
                      label="Priority"
                      helperText={
                        errors.priority && touched.priority && errors.priority
                      }
                      error={errors.priority && touched.priority}
                    />
                  </Grid>
                </Grid>
              </DialogContent>
              <DialogActions>
                <Button autoFocus onClick={() => onClose()} color="secondary" disabled={isSubmitting}>
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
}

export default ControllerForm;
