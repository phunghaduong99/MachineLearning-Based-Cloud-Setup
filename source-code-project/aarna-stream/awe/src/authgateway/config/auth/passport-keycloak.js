const KeycloakStrategy = require("@exlinc/keycloak-passport");
const User = require("../../models/User");

// const User = require("../../models/User");
// Register the strategy with passport
const options = {
    // host: process.env.KEYCLOAK_HOST,
    // realm: process.env.KEYCLOAK_REALM,
    // clientID: process.env.KEYCLOAK_CLIENT_ID,
    host: "http://localhost:8080",
    realm: "amcop",
    clientID: "amcop",
    authorizationURL:
        "http://localhost:8080/auth/realms/amcop/protocol/openid-connect/auth",
    tokenURL:
        "http://localhost:8080/auth/realms/amcop/protocol/openid-connect/token",
    userInfoURL:
        "http://localhost:8080/auth/realms/amcop/protocol/openid-connect/userinfo",
    // clientSecret: process.env.KEYCLOAK_CLIENT_SECRET,
    clientSecret: "PKWm5STlCKTBTh8Y0XMF1YNnnvFy9VM3",
    callbackURL: `/auth/keycloak/callback`,
};

module.exports = function (passport) {
    passport.use(
        "keycloak",
        new KeycloakStrategy(
            options,
            (accessToken, refreshToken, profile, done) => {
                return done(null, profile);
            }
        )
    );
};
