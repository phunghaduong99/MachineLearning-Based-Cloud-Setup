const GoogleStrategy = require("passport-google-oauth2").Strategy;
const User = require("../../models/User");
const {GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_CALLBACK_URL} = require("../config");

const options = {
    clientID: GOOGLE_CLIENT_ID,
    clientSecret: GOOGLE_CLIENT_SECRET,
    callbackURL: GOOGLE_CALLBACK_URL,
}
module.exports = function (passport) {
    passport.use(new GoogleStrategy(options, async (request, accessToken, refreshToken, profile, done) => {
        const newUser = {
            provider: profile.provider,
            id: profile.id,
            displayName: profile.displayName,
            firstName: profile.name.givenName,
            lastName: profile.name.familyName,
            image: profile.photos[0].value, // not sure why this is an array ?
            email: profile.emails[0].value
        };
        try {
            let user = await User.findOne({id: profile.id});
            if (user) {
                done(null, user);
            } else {
                user = await User.create(newUser);
                return done(null, user);
            }
        } catch (err) {
            console.log(err);
        }
    }));
};
