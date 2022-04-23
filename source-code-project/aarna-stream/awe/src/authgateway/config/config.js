const crypto = require('crypto')
// Assert required variables are passed
new Array("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "MICROSOFT_GRAPH_CLIENT_ID", "MICROSOFT_GRAPH_CLIENT_SECRET", "MONGODB_HOST", "CAMUNDA_HOST", "CAMUNDA_USER", "CAMUNDA_USER_PASSWORD").forEach((secret) => {
    if (secret && !process.env[secret]) {
        throw new Error(`${secret} is missing from process.env`);
    }
});

const {
    NODE_ENV = "development",
    GOOGLE_CLIENT_ID = "",
    GOOGLE_CLIENT_SECRET = "",
    GOOGLE_CALLBACK_URL = (process.env.NODE_ENV === "development" ? "http" : "https") + "://localhost:5000/auth/google/callback",
    MONGODB_HOST = "",
    MICROSOFT_GRAPH_CLIENT_ID = "",
    MICROSOFT_GRAPH_CLIENT_SECRET = "",
    MICROSOFT_CALLBACK_URL = (process.env.NODE_ENV === "development" ? "http" : "https") + "://localhost:5000/auth/microsoft/callback",
    CAMUNDA_HOST = "",
    CAMUNDA_USER = "",
    CAMUNDA_USER_PASSWORD = "",
    UI_PROXY_TARGET = "",
    API_PROXY_TARGET = ""
} = process.env;

const CAMUNDA_AUTH = Buffer.from(`${CAMUNDA_USER}:${CAMUNDA_USER_PASSWORD}`).toString('base64');
const TWO_MINUTES_IN_MS = 2000 * 60;
const POLLING_INTERVAL = TWO_MINUTES_IN_MS;
const POLLING_MAX_ATTEMPTS = 10;

const UI_PROXY_OPTIONS = {
    target: UI_PROXY_TARGET,
    changeOrigin: true,
};

const API_PROXY_OPTIONS = {
    target: API_PROXY_TARGET,
    changeOrigin: true,
};
const SESSION_COOKIE = "sid";

//Changing the secret value will invalidate all existing sessions. In order to rotate the secret without invalidating sessions, provide an array of secrets, with the new secret as first element of the array, and including previous secrets as the later elements.
//so if we want the session to be valid even when the server restarts we need the provide a fixed set of secrets
const APP_SECRET = crypto.randomBytes(32).toString('base64');
const IN_PROD = NODE_ENV === "production";
const THIRTY_MINUTES_IN_MS = 10000 * 180;

const SESSION_OPTS = {
    cookie: {
        // domain, // current domain (Same-Origin, no CORS)
        httpOnly: true,
        maxAge: THIRTY_MINUTES_IN_MS,
        secure: IN_PROD,
    },
    name: SESSION_COOKIE,
    resave: false, // whether to save the session if it wasn't modified during the request
    rolling: true, // whether to (re-)set cookie on every response, this will reset the expiry on every request
    saveUninitialized: false, // whether to save empty sessions to the store
    secret: APP_SECRET,
};

const MONGO_URI = "mongodb://" + MONGODB_HOST + "/amcop_saas";
module.exports = {
    GOOGLE_CLIENT_ID,
    GOOGLE_CLIENT_SECRET,
    GOOGLE_CALLBACK_URL,
    MICROSOFT_GRAPH_CLIENT_ID,
    MICROSOFT_GRAPH_CLIENT_SECRET,
    MICROSOFT_CALLBACK_URL,
    SESSION_OPTS,
    MONGO_URI,
    UI_PROXY_OPTIONS,
    API_PROXY_OPTIONS,
    CAMUNDA_HOST,
    POLLING_INTERVAL,
    POLLING_MAX_ATTEMPTS,
    CAMUNDA_AUTH,
    APP_SECRET
}
