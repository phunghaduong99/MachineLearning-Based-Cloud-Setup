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
import React, { useContext, useState } from "react";
import CssBaseline from "@material-ui/core/CssBaseline";
import Navigator from "../common/Navigator";
import Header from "./Header";
import CompositeApps from "../compositeApps/CompositeApps";
import CompositeApp from "../compositeApps/CompositeApp";
import theme from "../theme/Theme";
import DeploymentIntentGroups from "../deploymentIntentGroups/DeploymentIntentGroups";
import { Switch, Route } from "react-router-dom";
import DeploymentIntentGroup from "../deploymentIntentGroups/digView/DeploymentIntentGroup";
import DeploymentIntentGroupCheckout from "../deploymentIntentGroups/digView/DeploymentIntentGroupCheckout";
import Dashboard from "../dashboard/DashboardView";
import {tenantMenu} from "../config/uiConfig";
import PageNotFound from "../common/PageNotFound";
import LogicalClouds from "../logicalClouds/LogialClouds";
import { makeStyles } from "@material-ui/styles";
import { UserContext } from "../UserContext";

const useAppStyles = makeStyles({
  root: {
    display: "flex",
    minHeight: "100vh",
  },
  app: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
  },
  main: {
    flex: 1,
    padding: theme.spacing(3, 4, 6, 4),
    background: "#eaeff1",
  },
});

function AppBase(props) {
  const [mobileOpen, setMobileOpen] = useState(false);
  const projectName = props.match.params.projectName;
  const { user } = useContext(UserContext);
  const isAuthorized =
      user.role === "admin" || props.match.params.projectName === user.tenant;
  const classes = useAppStyles();
  const handleDrawerToggle = () => {
    setMobileOpen(() => !mobileOpen);
  };
  return (
      <>
        {projectName && (
            <div className={classes.root}>
              <CssBaseline />
              <Navigator menu={tenantMenu} handleDrawerToggle={handleDrawerToggle} mobileOpen={mobileOpen}/>
              <div className={classes.app}>
                <Header onDrawerToggle={handleDrawerToggle} onChangePasswordClick={props.handlePasswordFormOpen}/>
                {!isAuthorized && <PageNotFound />}
                {isAuthorized && (
                    <main className={classes.main}>
                      <Switch>
                        <Route
                            exact
                            path={`${props.match.url}/404`}
                            component={() => <PageNotFound />}
                        />
                        <Route exact path={`${props.match.url}/dashboard`}>
                          <Dashboard projectName={projectName} />
                        </Route>
                        <Route exact path={`${props.match.url}/services`}>
                          <CompositeApps projectName={projectName} />
                        </Route>
                        <Route exact path={`${props.match.url}/logical-clouds`}>
                          <LogicalClouds projectName={projectName} />
                        </Route>
                        <Route
                            exact
                            path={`${props.match.url}/services/:appname/:version`}
                            component={() => (
                                <CompositeApp projectName={projectName} />
                            )}
                        />
                        <Route
                            exact
                            path={`${props.match.url}/deployment-intent-groups`}
                        >
                          <DeploymentIntentGroups projectName={projectName} />
                        </Route>
                        <Route
                            exact
                            path={`${props.match.url}/deployment-intent-groups/:compositeAppName/:compositeAppVersion/:digName/status`}
                        >
                          <DeploymentIntentGroup projectName={projectName} />
                        </Route>
                        <Route
                            exact
                            path={`${props.match.url}/deployment-intent-groups/:compositeAppName/:compositeAppVersion/:digName/checkout`}
                        >
                          <DeploymentIntentGroupCheckout
                              projectName={projectName}
                          />
                        </Route>
                        <Route path="/" component={() => <PageNotFound />} />
                      </Switch>
                    </main>
                )}
              </div>
            </div>
        )}
      </>
  );
}

export default AppBase;
