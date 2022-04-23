const MicrosoftStrategy = require("passport-microsoft").Strategy;
const User = require("../../models/User");
const {
    MICROSOFT_GRAPH_CLIENT_ID,
    MICROSOFT_GRAPH_CLIENT_SECRET,
    MICROSOFT_CALLBACK_URL,
} = require("../config.js");

const options = {
    clientID: MICROSOFT_GRAPH_CLIENT_ID,
    clientSecret: MICROSOFT_GRAPH_CLIENT_SECRET,
    callbackURL: MICROSOFT_CALLBACK_URL,
    scope: ["user.read"],
};

module.exports = function (passport) {
    passport.use(
        new MicrosoftStrategy(
            options,
            async (accessToken, refresh_token, params, profile, done) => {
                const newUser = {
                    provider: profile.provider,
                    id: profile.id,
                    displayName: profile.displayName,
                    firstName: profile.name.givenName,
                    lastName: profile.name.familyName,
                    image: profile.photos ? profile.photos[0].value : null, // not sure why this is an array ?
                    email: profile.emails[0].value,
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
            }
        )
    );
};
