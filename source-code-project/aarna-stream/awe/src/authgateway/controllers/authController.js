const User = require("../models/User");
const passport = require("passport");
exports.signup = function (req, res, next) {
    //only admin can add users
    if (req.user.role !== 'admin') {
        return res.status(401).send("unauthorized");
    }
    const {password, email, firstName, lastName, tenant} = req.body;
    if (!email || !password || !firstName || !tenant) {
        return res
            .status(400)
            .send({error: "You must provide firstName, tenant, email and password"});
    }
    const displayName = req.body.displayName || firstName;
    // See if a user with the given email exists
    User.findOne({email: email}, async function (err, existingUser) {
        if (err) {
            return next(err);
        }

        // If a user with email does exist, return an error
        if (existingUser) {
            return res.status(422).send({error: "Email is in use"});
        }
        const newUser = {
            provider: "amcop",
            id: email,
            displayName: displayName,
            firstName: firstName,
            lastName: lastName,
            email: email,
            password: password,
            tenant: tenant,
            role: "tenant"
        };
        try {
            let user = await User.create(newUser);
            const {password, ...responseUser} = user._doc;
            res.json(responseUser);
        } catch (err) {
            console.log(err);
            return next(err);
        }

    });
};

exports.sign_in = function (req, res, next) {
    passport.authenticate("local", function (err, user, token) {
        if (err) {
            return next(err);
        }
        if (!user) {
            req.flash("auth", {
                error: true,
                message: "Invalid email or password",
                email: req.body.email,
                password: req.body.password,
            });
            return res.redirect("/login");
        }
        req.logIn(user, function (err) {
            if (err) {
                return next(err);
            }
            //redirect the user to the url originally requested
            let redirectionUrl = req.session.redirectUrl || "/app";
            return res.redirect(redirectionUrl);
        });
    })(req, res, next);
};
