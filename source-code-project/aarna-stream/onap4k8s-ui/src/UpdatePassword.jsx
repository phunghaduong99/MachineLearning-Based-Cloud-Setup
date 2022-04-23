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

import React, {useContext, useState} from "react";
import * as Yup from "yup";
import {Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField} from "@material-ui/core";
import LoadingButton from "./common/LoadingButton";
import {Formik, getIn} from "formik";
import apiService from "./services/apiService";
import {UserContext} from "./UserContext";
import Alert from "@material-ui/lab/Alert";

const schema = Yup.object({
    currentPassword: Yup.string().max(50).min(8).required("Current password is required"),
    newPassword: Yup.string().max(50).min(8)
        .required("New password is required").notOneOf([Yup.ref("currentPassword"), null], "New password must be different from the current password"),
    confirmPassword: Yup.string()
        .required("Confirm password is required")
        .oneOf([Yup.ref("newPassword"), null], "Passwords must match"),
});

export default function UpdatePassword({formOpen, setFormOpen}) {
    const {user} = useContext(UserContext);
    const [apiResponseError, setApiResponseError] = useState(null);
    const handleClose = () => {
        setFormOpen(false);
    };
    let initialValues = {currentPassword: "", newPassword: "", confirmPassword: ""}
    const handleFormSubmit = (values, isSubmitting) => {
        setApiResponseError(null);
        values.userId = user._id;
        apiService.updateUserPassword(values).then(() => {
            setFormOpen(false);
        }).catch(err => {
            if (err.response.data) {
                setApiResponseError(err.response);
            } else {
                setApiResponseError({data: "something went wrong, please try again later"})
            }
        }).finally(() => isSubmitting(false));
    }
    return (
        <Dialog
            open={formOpen}
            onClose={handleClose}
            disableBackdropClick
        >
            <DialogTitle>Change Password</DialogTitle>
            <Formik
                initialValues={initialValues}
                onSubmit={(values, {setSubmitting}) => {
                    handleFormSubmit(values, setSubmitting);
                }}
                validationSchema={schema}
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
                        <form noValidate onSubmit={handleSubmit}>
                            <DialogContent dividers>
                                {apiResponseError &&
                                <Alert style={{marginBottom: "15px"}} severity="error">{apiResponseError.data}</Alert>}
                                <TextField
                                    style={{width: "100%", marginBottom: "10px"}}
                                    id="currentPassword"
                                    label="Current Password"
                                    type="password"
                                    value={values.currentPassword}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    required
                                    helperText={
                                        getIn(touched, 'currentPassword') &&
                                        getIn(errors, 'currentPassword')}
                                    error={Boolean(getIn(touched, 'currentPassword') &&
                                        getIn(errors, 'currentPassword'))}
                                />
                                <TextField
                                    style={{width: "100%", marginBottom: "10px"}}
                                    id="newPassword"
                                    label="New Password"
                                    type="password"
                                    value={values.newPassword}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    required
                                    helperText={
                                        getIn(touched, 'newPassword') &&
                                        getIn(errors, 'newPassword')}
                                    error={Boolean(getIn(touched, 'newPassword') &&
                                        getIn(errors, 'newPassword'))}
                                />
                                <TextField
                                    style={{width: "100%", marginBottom: "10px"}}
                                    id="confirmPassword"
                                    label="Confirm Password"
                                    type="password"
                                    value={values.confirmPassword}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    required
                                    helperText={
                                        getIn(touched, 'confirmPassword') &&
                                        getIn(errors, 'confirmPassword')}
                                    error={Boolean(getIn(touched, 'confirmPassword') &&
                                        getIn(errors, 'confirmPassword'))}/>
                            </DialogContent>
                            <DialogActions>
                                <Button autoFocus onClick={handleClose}
                                        color="secondary"
                                        disabled={isSubmitting}>
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
    )
}