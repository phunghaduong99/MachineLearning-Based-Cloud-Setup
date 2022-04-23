const User = require("../../models/User");
const LocalStrategy = require('passport-local').Strategy;
const jwt = require('jwt-simple');
const options = {usernameField: 'email'};
const APP_SECRET = require('../config').APP_SECRET

function tokenForUser(user) {
    const timestamp = new Date().getTime();
    return jwt.encode({sub: user.id, iat: timestamp, user}, APP_SECRET);
}

module.exports = function (passport) {
    passport.use(
        new LocalStrategy(options, function (email, password, done) {
            User.findOne({email: email}, function (err, user) {
                if (err) {
                    return done(err);
                }
                if (!user) {
                    return done(null, false);
                }

                // compare passwords
                user.verifyPassword(password, function (err, isMatch) {
                    if (err) {
                        return done(err);
                    }
                    if (!isMatch) {
                        return done(null, false);
                    }

                    return done(null, user, tokenForUser(user.getUserFields()));
                });
            }).select('+password');
        }))
}