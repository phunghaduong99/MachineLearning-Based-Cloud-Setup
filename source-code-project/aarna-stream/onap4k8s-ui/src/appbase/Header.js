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
import React, { useContext } from "react";
import PropTypes from "prop-types";
import AppBar from "@material-ui/core/AppBar";
import Hidden from "@material-ui/core/Hidden";
import MenuIcon from "@material-ui/icons/Menu";
import Toolbar from "@material-ui/core/Toolbar";
import { withStyles } from "@material-ui/core/styles";
import { withRouter } from "react-router-dom";
import {
  Typography,
  Grid,
  IconButton,
  Popover,
  Divider,
  Button,
  Avatar,
} from "@material-ui/core";
import Breadcrumbs from "@material-ui/core/Breadcrumbs";
import { UserContext } from "../UserContext";
import {ExitToApp as LogoutIcon, NavigateNext, LockOpen} from "@material-ui/icons";
const lightColor = "rgba(255, 255, 255, 0.7)";
const { ENABLE_RBAC } = window._env_;
const styles = (theme) => ({
  root: {
    boxShadow:
      "0 3px 4px 0 rgba(0,0,0,.2), 0 3px 3px -2px rgba(0,0,0,.14), 0 1px 8px 0 rgba(0,0,0,.12)",
  },
  secondaryBar: {
    zIndex: 0,
  },
  menuButton: {
    marginLeft: -theme.spacing(1),
  },
  iconButtonAvatar: {
    padding: 4,
  },
  link: {
    textDecoration: "none",
    color: lightColor,
    "&:hover": {
      color: theme.palette.common.white,
    },
  },
  button: {
    borderColor: lightColor,
  },
});

function Header(props) {
  const { classes, onDrawerToggle, location } = props;
  let breadcrumbs = [];
  let getHeaderName = () => {
    if (location.pathname === `${props.match.url}/projects`) {
      breadcrumbs.push({ name: "Tenants", path: "#" });
    } else if (location.pathname === `${props.match.url}/clusters`) {
      breadcrumbs.push({ name: "Clusters", path: "#" });
    } else if (location.pathname === `${props.match.url}/controllers`) {
      breadcrumbs.push({ name: "Controllers", path: "#" });
    } else if (location.pathname === `${props.match.url}/users`) {
      breadcrumbs.push({ name: "Users", path: "#" });
    } else if (location.pathname === `${props.match.url}/smo`) {
      breadcrumbs.push({ name: "SMO", path: "#" });
    } else if (location.pathname === `${props.match.url}/dashboard`) {
      breadcrumbs.push({ name: "Dashboard", path: "/dashboard" });
    } else if (location.pathname === `${props.match.url}/services`) {
      breadcrumbs.push({ name: "Services", path: "/services" });
    } else if (
      location.pathname ===
      `${props.match.url}/services/${props.match.params.appname}/${props.match.params.version}`
    ) {
      breadcrumbs.push({ name: "some", path: "/services" });
    } else if (
      location.pathname === `${props.match.url}/deployment-intent-groups`
    ) {
      breadcrumbs.push({ name: "Service Instances", path: "/services" });
    } else if (location.pathname.includes("services")) {
      let serviceNameWithVersion = location.pathname
        .slice(location.pathname.indexOf("services"))
        .slice(9);
      breadcrumbs.push({ name: "Services", path: "/services" });
      breadcrumbs.push({
        name: serviceNameWithVersion.slice(
          0,
          serviceNameWithVersion.indexOf("/")
        ),
        path: "/services",
      });
    } else if (location.pathname.includes("deployment-intent-groups")) {
      breadcrumbs.push({ name: "Service Instances", path: "#" });
      breadcrumbs.push({ name: "Service Instance Detail", path: "#" });
    } else if (location.pathname === `${props.match.url}/logical-clouds`) {
      breadcrumbs.push({ name: "Logical Clouds", path: "/logical-clouds" });
    }
  };

  // setHeaderName();
  getHeaderName();
  //set website title to current page
  breadcrumbs.forEach((breadcrumb, index) => {
    if (index === 0) {
      document.title = breadcrumb.name;
    } else {
      document.title = document.title + " - " + breadcrumb.name;
    }
  });

  const { user } = useContext(UserContext);
  const [anchorEl, setAnchorEl] = React.useState(null);
  const open = Boolean(anchorEl);
  const id = open ? "user-popover" : undefined;
  const handleClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <React.Fragment>
      {ENABLE_RBAC === 'true' &&
      <Popover
          id={id}
          open={open}
          anchorEl={anchorEl}
          onClose={handleClose}
          anchorOrigin={{
            vertical: "bottom",
            horizontal: "center",
          }}
          transformOrigin={{
            vertical: "top",
            horizontal: "center",
          }}
      >
        <Typography
            variant={"h6"}
            style={{textAlign: "center", padding: "7px"}}
        >
          {user.displayName}
        </Typography>
        <Typography
            variant={"subtitle2"}
            color="textSecondary"
            style={{textAlign: "center", padding: "0 0 7px 0"}}
        >
          {user.role}
        </Typography>
        <Divider variant="middle"/>
        <Button style={{display: "block", padding: "10px 10px"}} onClick={() => {
          handleClose();
          props.onChangePasswordClick()
        }} color="primary">
          <LockOpen style={{verticalAlign: "bottom"}}/>
          &nbsp;&nbsp;Change Password
        </Button>
        <Button style={{display: "block", padding: "10px 10px"}}
                href="/logout" color="primary">
          <LogoutIcon style={{verticalAlign: "bottom"}}/>
          &nbsp;&nbsp;Logout
        </Button>
      </Popover>}
      <AppBar
        className={classes.root}
        color="primary"
        position="sticky"
        elevation={0}
      >
        <Toolbar>
          <Grid
            container
            spacing={1}
            alignItems="center"
            justify="space-between"
          >
            <Hidden smUp implementation="js">
              <Grid item>
                <IconButton
                  color="inherit"
                  onClick={onDrawerToggle}
                  className={classes.menuButton}
                >
                  <MenuIcon />
                </IconButton>
              </Grid>
            </Hidden>
            <Grid item>
              <Breadcrumbs
                color="inherit"
                separator={<NavigateNext fontSize="small" />}
                aria-label="breadcrumb"
              >
                {breadcrumbs.map((breadcrumb, index) => (
                  <Typography
                    key={breadcrumb.name + index}
                    color="inherit"
                    href="/"
                    // onClick={handleClick}
                  >
                    {breadcrumb.name}
                  </Typography>
                ))}
              </Breadcrumbs>
            </Grid>
            {ENABLE_RBAC === 'true' && <Grid item>
              <IconButton color="inherit" onClick={handleClick}>
                <Avatar src={user.image || null} alt="profile_image"/>
              </IconButton>
            </Grid>}
          </Grid>
        </Toolbar>
      </AppBar>
    </React.Fragment>
  );
}

Header.propTypes = {
  classes: PropTypes.object.isRequired,
  onDrawerToggle: PropTypes.func.isRequired,
};

export default withStyles(styles)(withRouter(Header));
