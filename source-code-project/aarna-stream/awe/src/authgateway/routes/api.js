const express = require('express');
const router = express.Router();
const {ensureApiAuth} = require("../middleware/auth");
const userCtrl = require('../controllers/userController');
const Authentication = require("../controllers/authController");

router.get('/user/me', ensureApiAuth, userCtrl.getCurrentUser);
router.get('/users', ensureApiAuth, userCtrl.getAllUsers);
router.post("/user/add", ensureApiAuth, Authentication.signup);
router.delete("/user/:id", ensureApiAuth, userCtrl.deleteUser);
router.put("/user/:id/account/password", ensureApiAuth, userCtrl.updatePassword);
router.put("/user/:id", ensureApiAuth, userCtrl.updateUser);
router.all("*", ensureApiAuth, (req, res) => {
    res.status(404).send("not found")
});

module.exports = router;

