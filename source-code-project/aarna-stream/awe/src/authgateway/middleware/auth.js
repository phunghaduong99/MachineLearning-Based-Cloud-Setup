module.exports = {
    // for the routes for which we want to send a redirect request, e.g UI app
    ensureAuth: function (req, res, next) {
        if (req.isAuthenticated()) {
            return next();
        } else {
            //save the requested url in session so that after login the user can be redirected to it.
            if (req.session) {
                req.session.redirectUrl = req.headers.referer || req.originalUrl || req.url;
            }
            res.redirect("/login");
        }
    },
    //for api we dont want to send a redirect request
    ensureApiAuth: function (req, res, next) {
        if (req.isAuthenticated()) {
            return next();
        } else {
            res.status(401).send("unauthorized");
        }
    },
};
