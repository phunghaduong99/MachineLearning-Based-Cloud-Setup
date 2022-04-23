const express = require("express");
const router = express.Router();
const {createProxyMiddleware} = require("http-proxy-middleware");
const {API_PROXY_OPTIONS} = require("../config/config");
const apiProxy = createProxyMiddleware(API_PROXY_OPTIONS);

const checkUrlAuth = (req, res, next) => {
    if (req.user.role === "admin") return next();
    else if (req.user.role === "tenant" && (req.params.projectName === req.user.tenant)) {
        return next();
    } else {
        return res.status(403).send("unauthorized");
    }
};

const checkAdminAuth = (req, res, next) => {
    if (req.user.role === "admin") return next();
    return res.status(403).send("unauthorized");
};
router.all("/cluster-providers*", checkAdminAuth, apiProxy);
router.all("/controllers*", checkAdminAuth, apiProxy);
router.all("/projects/", checkUrlAuth, apiProxy);
router.all("/projects/:projectName*", checkUrlAuth, apiProxy);
router.all("/*", apiProxy);

module.exports = router;
