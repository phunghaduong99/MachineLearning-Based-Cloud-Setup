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
import React, {useEffect, useState} from "react";
import {Button, Grid} from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../../services/apiService";
import Spinner from "../../common/Spinner";
import UsersTable from "./UsersTable";
import UserForm from "./UserForm";
import Notification from "../../common/Notification";

const Users = () => {
    const [openUserForm, setOpenUserForm] = useState(false);
    const [usersData, setUsersData] = useState([]);
    const [isLoading, setIsloading] = useState(true);
    const [tenantsData, setTenantsData] = useState(null);
    const [notificationDetails, setNotificationDetails] = useState({});
    const [userToEdit, setUserToEdit] = useState(null);
    const handleClose = () => {
        setOpenUserForm(false);
    };

    const handleOpenUserForm = (user) => {
        if(user){
            setUserToEdit(user);
        } else {
            setUserToEdit(null);
        }
        if(!tenantsData || tenantsData.length < 1){
            apiService
                .getAllProjects()
                .then((res) => {
                    if(res && res.length > 0){
                        setTenantsData(res);
                        setOpenUserForm(true);
                    } else {
                        setNotificationDetails({
                            show: true,
                            message: "Please add at least one tenant before adding a user",
                            severity: "info",
                        });
                    }

                })
                .catch(() => {
                    setNotificationDetails({
                        show: true,
                        message: "something went wrong, please try again",
                        severity: "error",
                    });
                });
        } else {
            setOpenUserForm(true);
        }
    };
    const updateUser = (payload) =>{
        payload.userId = usersData[userToEdit]._id;
        apiService
            .updateUserDetails(payload)
            .then((res) => {
                setUsersData(existingData => {
                    existingData[userToEdit] = res;
                    return existingData;
                });
                setNotificationDetails({
                    show: true,
                    message: "User details updated",
                    severity: "success",
                });
            })
            .catch((err) => {
                let error_message = err.response ? "Error updating user : " + err.response.data : "Error updating user";
                setNotificationDetails({
                    show: true,
                    message: `${error_message}`,
                    severity: "error",
                });
            }).finally(() => {
            setOpenUserForm(false);
        });
    }

    const addUser = (payload) => {
        apiService
            .addUser(payload)
            .then((response) => {
                if (usersData && usersData.length > 0)
                    setUsersData([...usersData, response]);
                else setUsersData([response]);
                setNotificationDetails({
                    show: true,
                    message: "User added",
                    severity: "success",
                });
            })
            .catch((error) => {
                let error_message = error.response.data ? `Error adding user : ${error.response.data.error}` : "Error adding user;"
                setNotificationDetails({
                    show: true,
                    message: `${error_message}`,
                    severity: "error",
                });
            })
            .finally(() => {
                setOpenUserForm(false);
            });
    }
    const handleSubmitUserForm = (payload) => {
        if(userToEdit){
            updateUser(payload);
        } else{
            addUser(payload);
        }
    };
    useEffect(() => {
        apiService
            .getAllUsers()
            .then((response) => {
                setUsersData(response);
            })
            .catch((error) => {
                console.log(error);
            })
            .finally(() => {
                setIsloading(false);
            });
    }, []);

    return (
        <>
            {isLoading && <Spinner/>}
            {!isLoading && (
                <>
                    <Notification notificationDetails={notificationDetails}/>
                    <Button
                        variant="outlined"
                        color="primary"
                        startIcon={<AddIcon/>}
                        onClick={()=>handleOpenUserForm()}
                    >
                        Add User
                    </Button>
                    <UserForm
                        open={openUserForm}
                        onClose={handleClose}
                        onSubmit={handleSubmitUserForm}
                        existingUsers={userToEdit ? [] : usersData}
                        tenants={tenantsData}
                        item={usersData[userToEdit]}
                    />
                    <Grid container spacing={2} alignItems="center">
                        <Grid item xs style={{marginTop: "20px"}}>
                            <UsersTable
                                data={usersData} setUsersData={setUsersData}
                                setNotificationDetails={setNotificationDetails}
                                onEditUser={handleOpenUserForm}/>
                        </Grid>
                    </Grid>
                </>
            )}
        </>
    );
};
export default Users;
