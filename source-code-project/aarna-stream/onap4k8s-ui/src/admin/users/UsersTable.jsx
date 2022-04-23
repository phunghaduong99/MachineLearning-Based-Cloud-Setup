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
import React, {useContext} from "react";
import {makeStyles, withStyles} from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import Paper from "@material-ui/core/Paper";
import IconButton from "@material-ui/core/IconButton";
import EditIcon from "@material-ui/icons/EditOutlined";
import DeleteDialog from "../../common/Dialogue";
import DeleteIcon from "@material-ui/icons/DeleteOutline";
import apiService from "../../services/apiService";
import {UserContext} from "../../UserContext";

const StyledTableCell = withStyles(() => ({
    body: {
        fontSize: 14,
    },
}))(TableCell);

const StyledTableRow = withStyles((theme) => ({
    root: {
        "&:nth-of-type(odd)": {
            backgroundColor: theme.palette.action.hover,
        },
    },
}))(TableRow);

const useStyles = makeStyles({
    table: {
        minWidth: 350,
    },
    cell: {
        color: "grey",
    },
});

export default function UsersTable(props) {
    const classes = useStyles();
    const [open, setOpen] = React.useState(false);
    const [index, setIndex] = React.useState(0);
    const {user} = useContext(UserContext);

    const handleCloseDeleteDialog = (el) => {
        if (el.target.innerText === "Delete") {
            apiService
                .deleteUser(props.data[index]._id)
                .then(() => {
                    props.data.splice(index, 1);
                    props.setUsersData([...props.data]);
                    props.setNotificationDetails({
                        show: true,
                        message: "User deleted",
                        severity: "success",
                    });
                })
                .catch((err) => {
                    let error_message = err.response ? "Error deleting user : " + err.response.data : "Error deleting user";
                    props.setNotificationDetails({
                        show: true,
                        message: `${error_message}`,
                        severity: "error",
                    });
                });
        }
        setOpen(false);
        setIndex(0);
    };
    const handleDelete = (index) => {
        setIndex(index);
        setOpen(true);
    };

    return (
        <React.Fragment>
            {props.data && props.data.length > 0 && (
                <>
                    <DeleteDialog
                        open={open}
                        onClose={handleCloseDeleteDialog}
                        title={"Delete User"}
                        content={`Are you sure you want to delete the user "${
                            props.data[index] ? props.data[index].displayName : ""
                        }" ?`}
                    />
                    <TableContainer component={Paper}>
                        <Table className={classes.table} size="small">
                            <TableHead>
                                <TableRow>
                                    <StyledTableCell>Name</StyledTableCell>
                                    <StyledTableCell>Role</StyledTableCell>
                                    <StyledTableCell>Tenant</StyledTableCell>
                                    <StyledTableCell>Email</StyledTableCell>
                                    <StyledTableCell>Actions</StyledTableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {props.data.map((row, index) => (
                                    <StyledTableRow key={row.email + "" + index}>
                                        <StyledTableCell>{row.displayName}</StyledTableCell>
                                        <StyledTableCell className={classes.cell}>
                                            {row.role}
                                        </StyledTableCell>
                                        <StyledTableCell className={classes.cell}>
                                            {row.tenant}
                                        </StyledTableCell>
                                        <StyledTableCell className={classes.cell}>
                                            {row.email}
                                        </StyledTableCell>
                                        <StyledTableCell className={classes.cell}>
                                            <IconButton
                                                disabled={row.email === user.email || row.role === "admin"}
                                                color="primary"
                                                onClick={() => props.onEditUser(index)}
                                                title="Edit"
                                            >
                                                <EditIcon/>
                                            </IconButton>
                                            <IconButton
                                                disabled={row.email === user.email || row.role === "admin"}
                                                color="secondary"
                                                onClick={() => handleDelete(index)}
                                                title="Delete"
                                            >
                                                <DeleteIcon/>
                                            </IconButton>
                                        </StyledTableCell>
                                    </StyledTableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </>
            )}
        </React.Fragment>
    );
}
