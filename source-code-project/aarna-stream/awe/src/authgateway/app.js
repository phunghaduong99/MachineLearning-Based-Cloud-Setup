const createError = require("http-errors");
const express = require("express");
const path = require("path");
const logger = require("morgan");
const loginRouter = require("./routes/login");
const googleAuthRouter = require("./routes/googleAuth");
const microsoftAuthRouter = require("./routes/microsoftAuth");
const localAuthRouter = require("./routes/localAuth");
const apiRouter = require("./routes/api");
const emcoRouter = require("./routes/emco");
const app = express();
const session = require("express-session");
const MongoDBStore = require("connect-mongo");
const bodyParser = require("body-parser");
const {SESSION_OPTS, UI_PROXY_OPTIONS} = require("./config/config");
const connectDB = require("./config/db");
const {createProxyMiddleware} = require("http-proxy-middleware");
const {ensureAuth, ensureApiAuth} = require("./middleware/auth");
const flash = require("connect-flash");
const User = require("./models/User");

//passport configuration
const passport = require("passport");
passport.serializeUser((user, cb) => {
    cb(null, user);
});
passport.deserializeUser((id, done) => {
    User.findById(id, (err, user) => done(err, user));
});
require("./config/auth/passport-local")(passport);
require("./config/auth/passport-keycloak")(passport);
require("./config/auth/passport-microsoft")(passport);
require("./config/auth/passport-google")(passport);

//connect to the DB
const clientPromise = connectDB().then((conn) => conn.connection.getClient());

//session middleware, make sure this is added before passport middleware
app.use(
    session({
        ...SESSION_OPTS,
        store: MongoDBStore.create({clientPromise: clientPromise}),
    })
);
// passport middleware
app.use(passport.initialize());
app.use(passport.session());

//log all the api requests in dev
if (process.env.NODE_ENV === "development") {
    app.use(logger("dev"));
}

//proxy router
app.use("/v2", ensureApiAuth, emcoRouter);
app.use("/middleend", ensureApiAuth, emcoRouter);

// view engine setup
app.set("views", path.join(__dirname, "views"));
app.set("view engine", "jade");

// trust first proxy
app.set("trust proxy", 1);

app.use(express.json());
app.use(express.urlencoded({extended: false}));
app.use(express.static(path.join(__dirname, "public")));
app.use(bodyParser.urlencoded({extended: false}));

app.use(flash());

// proxy middlewares, this will be used when we want to add authentication for a service. We can proxy the request after authentication to the respective server.
// to add a proxy for multiple UI services, we need to set "PUBLIC_URL" and use the same while we proxy request for specific UI. e.g we can use /app as "PUBLIC_URL" for existing AMCOP UI and then proxy all the requests with /app to the AMCOP UI endpoint

const uiProxy = createProxyMiddleware(UI_PROXY_OPTIONS);
app.use("/app", ensureAuth, uiProxy);
app.use("/login", loginRouter);
app.get("/logout", (req, res) => {
    req.logout();
    req.session.destroy();
    res.redirect("/login");
});


//local auth
app.use("/auth/amcop", localAuthRouter);

//external OIDC
app.use("/auth/google", googleAuthRouter);
app.use("/auth/microsoft", microsoftAuthRouter);
// app.get("/auth/kc", passport.authenticate("keycloak"), (req, res) => {
//     res.status(200).send("hello");
// });
//
// app.get(
//     "/auth/keycloak/callback",
//     passport.authenticate("keycloak", {failureRedirect: "/login"}),
//     (req, res) => {
//         res.status(200).send(req.user);
//     }
// );
app.use("/api", apiRouter);

app.use("/", ensureAuth, (req, res) => {
    res.redirect("/app");
});

// catch 404 and forward to error handler
app.use(function (req, res, next) {
    next(createError(404));
});

//error handler
app.use(function (err, req, res, next) {
    // set locals, only providing error in development
    res.locals.message = err.message;
    res.locals.error = req.app.get("env") === "development" ? err : {};

    // render the error page
    res.status(err.status || 500);
    res.render("error");
});

module.exports = app;
