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
import {
  FormControl,
  FormHelperText,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from "@material-ui/core";
import * as Yup from "yup";
import { Formik } from "formik";
import LoadingButton from "../../common/LoadingButton";

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

const nameSchema = (fieldName) => {
  const name = fieldName || "";
  return Yup.string()
      .max(20, `${name} cannot exceed more than 20 characters`)
      .matches(
          /^[a-zA-Z0-9_-]+$/,
          `${name} can only contain letters, numbers, '-' and '_' and no spaces.`
      )
      .matches(
          /^[a-zA-Z0-9]/,
          `${name} must start with an alphanumeric character`
      )
      .matches(/[a-zA-Z0-9]$/, `${name} must end with an alphanumeric character`);
};

const getSchema = (existingUsers, isEdit) => {
  let schema;
  schema = Yup.object({
    firstName: Yup.string()
        .required("First name is required")
        .concat(nameSchema("First Name")),
    lastName: Yup.string().concat(nameSchema("Last Name")),
    tenant: Yup.string().required("Please select a tenant"),
    email: Yup.string()
        .email("Must be a valid email")
        .max(255)
        .required("Email is required")
        .test(
            "duplicate-test",
            "User with same email exists, please use a different email",
            (email) => {
              return existingUsers
                  ? existingUsers.findIndex((x) => x.email === email) === -1
                  : true;
            }
        ),
    password: !isEdit && Yup.string().max(50).min(8).required("Password is required"),
    confirmPassword: !isEdit && Yup.string()
        .required()
        .oneOf([Yup.ref("password"), null], "Passwords must match"),
  });
  return schema;
};

const UserForm = ({tenants, ...props}) => {
  const { onClose, item, open, onSubmit } = props;
  const isEdit = !!item;
  const buttonLabel = isEdit ? "OK" : "Add";
  const title = isEdit ? "Edit User" : "Add User";
  const handleClose = () => {
    onClose();
  };

  let initialValues = isEdit
      ? {
        firstName: item.firstName,
        lastName: item.lastName,
        tenant: item.tenant,
        email: item.email,
      }
      : {
        firstName: "",
        lastName: "",
        tenant: "",
        email: "",
        password: "",
        confirmPassword: "",
      };

  return (
      <Dialog
          maxWidth={"xs"}
          fullWidth
          onClose={handleClose}
          aria-labelledby="customized-dialog-title"
          open={open}
          disableBackdropClick
      >
        <DialogTitle id="simple-dialog-title">{title}</DialogTitle>
        <Formik
            initialValues={initialValues}
            onSubmit={(values) => {
              onSubmit(values);
            }}
            validationSchema={getSchema(props.existingUsers, isEdit)}
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
            } = props;

            return (
                tenants && (
                    <form noValidate onSubmit={handleSubmit}>
                      <DialogContent dividers>
                        <Grid container spacing={2}>
                          <Grid item xs={6}>
                            <TextField
                                style={{ width: "100%", marginBottom: "10px" }}
                                id="firstName"
                                label="First Name"
                                type="text"
                                value={values.firstName}
                                onChange={handleChange}
                                onBlur={handleBlur}
                                helperText={
                                  errors.firstName &&
                                  touched.firstName &&
                                  errors.firstName
                                }
                                required
                                error={errors.firstName && touched.firstName}
                            />
                          </Grid>
                          <Grid item xs={6}>
                            <TextField
                                style={{ width: "100%", marginBottom: "25px" }}
                                name="lastName"
                                value={values.lastName}
                                onChange={handleChange}
                                onBlur={handleBlur}
                                type="text"
                                id="lastName"
                                label="Last Name"
                                helperText={
                                  errors.lastName && touched.lastName && errors.lastName
                                }
                                error={errors.lastName && touched.lastName}
                            />
                          </Grid>
                          {!isEdit &&<>
                            <Grid item xs={6}>
                              <TextField
                                  style={{width: "100%", marginBottom: "25px"}}
                                  name="password"
                                  value={values.password}
                                  onChange={handleChange}
                                  onBlur={handleBlur}
                                  type="password"
                                  id="password"
                                  label="Password"
                                  required
                                  helperText={
                                    errors.password && touched.password && errors.password
                                  }
                                  error={errors.password && touched.password}
                              />
                            </Grid>
                            <Grid item xs={6}>
                              <TextField
                                  style={{width: "100%", marginBottom: "25px"}}
                                  name="confirmPassword"
                                  value={values.confirmPassword}
                                  onChange={handleChange}
                                  onBlur={handleBlur}
                                  type="password"
                                  id="confirmPassword"
                                  label="Confirm Password"
                                  required
                                  helperText={
                                    !errors.password &&
                                    errors.confirmPassword &&
                                    touched.confirmPassword &&
                                    errors.confirmPassword
                                  }
                                  error={
                                    !errors.password &&
                                    errors.confirmPassword &&
                                    touched.confirmPassword
                                  }
                              />
                            </Grid></>}
                          <Grid item xs={12}>
                            <TextField
                                disabled={isEdit}
                                style={{ width: "100%", marginBottom: "25px" }}
                                name="email"
                                value={values.email}
                                onChange={handleChange}
                                onBlur={handleBlur}
                                type="email"
                                id="email"
                                label="Email"
                                required
                                helperText={
                                  errors.email && touched.email && errors.email
                                }
                                error={errors.email && touched.email}
                            />
                          </Grid>
                          <Grid item xs={12}>
                            <FormControl
                                fullWidth
                                error={errors.tenant && touched.tenant}
                            >
                              <InputLabel htmlFor="select-tenant">Tenant</InputLabel>
                              <Select
                                  margin={"dense"}
                                  fullWidth
                                  name="tenant"
                                  labelId="select-tenant"
                                  value={values.tenant}
                                  onChange={handleChange}
                                  required
                              >
                                <MenuItem value="">
                                  <em>Select</em>
                                </MenuItem>
                                {tenants.map((tenant) => (
                                    <MenuItem
                                        key={tenant.metadata.name}
                                        value={tenant.metadata.name}
                                    >
                                      {tenant.metadata.name}
                                    </MenuItem>
                                ))}
                              </Select>
                              <FormHelperText>
                                {touched.tenant && errors.tenant}
                              </FormHelperText>
                            </FormControl>
                          </Grid>
                        </Grid>
                      </DialogContent>
                      <DialogActions>
                        <Button autoFocus onClick={handleClose} color="secondary" disabled={isSubmitting}>
                          Cancel
                        </Button>
                        <LoadingButton
                            type="submit"
                            buttonLabel={buttonLabel}
                            loading={isSubmitting}
                        />
                      </DialogActions>
                    </form>
                )
            );
          }}
        </Formik>
      </Dialog>
  );
};

UserForm.propTypes = {
  onClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
};

export default UserForm;
