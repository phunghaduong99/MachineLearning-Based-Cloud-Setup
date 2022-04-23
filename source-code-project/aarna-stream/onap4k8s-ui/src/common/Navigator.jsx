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
import React, { useState } from "react";
import PropTypes from "prop-types";
import clsx from "clsx";
import Divider from "@material-ui/core/Divider";
import Drawer from "@material-ui/core/Drawer";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import {Link, useLocation, useRouteMatch} from "react-router-dom";
import amcop_logo from "../assets/icons/amcop_logo.png";
import emco_logo from "../assets/icons/emco_logo.png";
import Hidden from "@material-ui/core/Hidden";
import {makeStyles} from "@material-ui/styles";
import theme from "../theme/Theme";

const drawerWidth = 256;
const useAppStyles = makeStyles({
  drawer: {
    [theme.breakpoints.up("sm")]: {
      width: drawerWidth,
      flexShrink: 0,
    },
  },
  categoryHeader: {
    paddingTop: theme.spacing(2),
        paddingBottom: theme.spacing(2),
  },
  categoryHeaderPrimary: {
    color: theme.palette.common.white,
  },
  item: {
    paddingTop: 1,
        paddingBottom: 1,
        color: "rgba(255, 255, 255, 0.7)",
        "&:hover,&:focus": {
      backgroundColor: "rgba(255, 255, 255, 0.08)",
    },
  },
  itemCategory: {
    backgroundColor: "#232f3e",
        boxShadow: "0 -1px 0 #404854 inset",
        paddingTop: theme.spacing(2),
        paddingBottom: theme.spacing(2),
  },
  itemCategoryEmcoLogo: {
    paddingTop: "5px",
        paddingBottom: 0,
  },
  itemActiveItem: {
    color: theme.palette.primary.main,
  },
  itemPrimary: {
    fontSize: "inherit",
  },
  itemIcon: {
    minWidth: "auto",
        marginRight: theme.spacing(2),
  },
  divider: {
    marginTop: theme.spacing(2),
        backgroundColor: "#404854",
  },
  version: {
    fontSize: "15px",
        color: "#0096a6",
  },
  textLogo: {
    float: "left",
        paddingRight: "90px",
        paddingLeft: "5px",
        color: theme.palette.common.white,
  },
  emcoLogo: { width: "80px", marginLeft: "56px", marginRight: "10px" },
  amcopLogo: { width: "160px" },
});

function Navbar ({ menu: categories, ...props }) {
  let location = useLocation();
  const match = useRouteMatch();
  const { classes } = props;
  const [activeItem, setActiveItem] = useState(location.pathname);
  const setActiveTab = (itemId) => {
    setActiveItem(itemId);
  };
  if (location.pathname !== activeItem) {
    setActiveTab(location.pathname);
  }
  return (
    <Drawer
      PaperProps={props.PaperProps}
      variant={props.variant}
      open={props.open}
      onClose={props.onClose}
    >
      <List disablePadding>
        <Link style={{ textDecoration: "none" }} to="/">
          <ListItem
            className={clsx(
              classes.item,
              classes.itemCategory,
              process.env.REACT_APP_PRODUCT !== "AMCOP" &&
                classes.itemCategoryEmcoLogo
            )}
          >
            <ListItemText
              classes={{
                primary: classes.itemPrimary,
              }}
            >
              {process.env.REACT_APP_PRODUCT &&
              process.env.REACT_APP_PRODUCT === "AMCOP" ? (
                <img
                  className={classes.amcopLogo}
                  src={amcop_logo}
                  alt="AMCOP"
                />
              ) : (
                <img className={classes.emcoLogo} src={emco_logo} alt="EMCO" />
              )}
              <sub className={classes.version}>
                {process.env.REACT_APP_VERSION}
              </sub>
            </ListItemText>
          </ListItem>
        </Link>
        {categories.map(({ id, children }) => (
          <React.Fragment key={id}>
            {children.map(({ id: childId, icon, url }) => (
              <Link
                style={{ textDecoration: "none" }}
                to={{
                  pathname: `${match.url}${url}`,
                  activeItem: childId,
                }}
                key={childId}
              >
                <ListItem
                  button
                  className={clsx(
                    classes.item,
                    childId === "Dashboard" && classes.itemCategory,
                    activeItem.includes(url) && classes.itemActiveItem
                  )}
                >
                  <ListItemIcon className={classes.itemIcon}>
                    {icon}
                  </ListItemIcon>
                  <ListItemText
                    classes={{
                      primary: classes.itemPrimary,
                    }}
                  >
                    {childId}
                  </ListItemText>
                </ListItem>
              </Link>
            ))}

            <Divider className={classes.divider} />
          </React.Fragment>
        ))}
      </List>
    </Drawer>
  );
}

const Navigator = (props) =>{
  const classes = useAppStyles();
  return(
      <nav className={classes.drawer}>
        <Hidden smUp implementation="js">
          <Navbar
              PaperProps={{ style: { width: drawerWidth } }}
              variant="temporary"
              open={props.mobileOpen}
              onClose={props.handleDrawerToggle}
              menu={props.menu}
              classes={classes}
          />
        </Hidden>
        <Hidden xsDown implementation="js">
          <Navbar
              PaperProps={{ style: { width: drawerWidth } }}
              variant="permanent"
              menu={props.menu}
              classes={classes}
          />
        </Hidden>
      </nav>
  )
}


Navigator.propTypes = {
  menu: PropTypes.arrayOf(PropTypes.object).isRequired
};
export default Navigator;