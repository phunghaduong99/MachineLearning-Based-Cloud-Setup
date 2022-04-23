const express = require("express");
const router = express.Router();
const passport = require("passport");

// @desc Auth with Google
// @route GET /auth/google
router.get(
    "/",
    passport.authenticate("google", {scope: ["email", "profile"]})
);

// @desc Google auth callback
// @route GET /auth/google/callback
router.get(
    "/callback",
    passport.authenticate("google", {
        failureRedirect: "/failure",
    }),
    function (req, res) {
        res.redirect("/");
    }
);

router.get("/failure", (req, res) => {
    res.send("Failed to authenticate..");
});

module.exports = router;
