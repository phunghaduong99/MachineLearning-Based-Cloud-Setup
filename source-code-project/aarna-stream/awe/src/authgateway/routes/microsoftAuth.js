const express = require("express");
const router = express.Router();
const passport = require("passport");

router.get("/", passport.authenticate("microsoft"));

router.get(
    "/callback",
    passport.authenticate("microsoft", {
        failureRedirect: "/",
    }),
    function (req, res) {
        res.redirect("/");
    }
);

module.exports = router;
