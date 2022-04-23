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
import {BrowserRouter as Router, Redirect, Route, Switch,} from "react-router-dom";
import "./App.css";
import AppBase from "./appbase/AppBase";
import Admin from "./admin/Admin";
import apiService from "./services/apiService";
import {UserContext} from "./UserContext";
import UpdatePassword from "./UpdatePassword";

function App() {
    const [user, setUser] = useState(null);
    const [passwordFormOpen, setPasswordFormOpen] = useState(false);
    const handlePasswordFormOpen = () => {
        setPasswordFormOpen(true);
    };
    useEffect(() => {
        //call user detail api only if rbac is enabled, otherwise set a default user
        if (window._env_.ENABLE_RBAC === 'true') {
            apiService
                .getUserDetails()
                .then((res) => {
                    setUser(res);
                })
                .catch(() => {
                });
        } else {
            setUser({
                email: "default@default.com",
                id: "123",
                tenant: "admin",
                role: "admin",
                displayName: "Default",
                provider: "default",
            })
        }

        const faviconUpdate =  () => {
            const favicon = document.getElementById("favicon");
            //update favicon for AMCOP
            if (process.env.REACT_APP_PRODUCT === "AMCOP")
                favicon.href = `${process.env.PUBLIC_URL}/amcop_favicon.ico`;
        };
        faviconUpdate();
    }, []);
    return (
        <UserContext.Provider value={{user, setUser}}>
            {user && (
                <>
                    <UpdatePassword formOpen={passwordFormOpen} setFormOpen={setPasswordFormOpen}/>
                    <Router>
                        <Switch>
                            {user.role === "admin" && (
                                <Route
                                    path="/app/admin"
                                    children={({match, ...others}) => {
                                        return (
                                            <Switch>
                                                <Redirect
                                                    exact
                                                    from={`${match.path}`}
                                                    to={`${match.path}/projects`}
                                                />
                                                <Route
                                                    path={`${match.path}`}
                                                    render={(props) => <Admin
                                                        handlePasswordFormOpen={handlePasswordFormOpen} {...props} />}
                                                />
                                            </Switch>
                                        );
                                    }}
                                />
                            )}
                            <Route
                                path="/app/projects/:projectName"
                                children={({match, ...others}) => {
                                    return (
                                        <Switch>
                                            <Redirect
                                                exact
                                                from={`${match.path}`}
                                                to={`${match.path}/dashboard`}
                                            />
                                            <Route
                                                path={`${match.path}`}
                                                render={(props) => <AppBase
                                                    handlePasswordFormOpen={handlePasswordFormOpen} {...props} />}
                                            />
                                        </Switch>
                                    );
                                }}
                            />
                            <Route
                                path="/"
                                render={() => {
                                    if (user.role === "admin")
                                        return <Redirect path="/" to={"/app/admin"}/>;
                                    else
                                        return (
                                            <Redirect path="/" to={`/app/projects/${user.tenant}`}/>
                                        );
                                }}
                            />
                        </Switch>
                    </Router>
                </>
            )}

        </UserContext.Provider>
    );
}

export default App;
