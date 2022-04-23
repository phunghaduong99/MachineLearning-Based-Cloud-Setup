const express = require("express");
const router = express.Router();

router.get("/", function (req, res, next) {
  if (!req.isAuthenticated()) {
    let login_view_data = { title: "AMCOP - Login" };
    const auth_flash_messages = req.flash("auth")[0];
    if (auth_flash_messages && auth_flash_messages.error) {
      login_view_data.auth_error = true;
      login_view_data.auth_error_message = auth_flash_messages.message;
      login_view_data.email = auth_flash_messages.email;
      login_view_data.password = auth_flash_messages.password;
    }
    res.render("login", login_view_data);
  } else {
    res.redirect("/");
  }
});

module.exports = router;
