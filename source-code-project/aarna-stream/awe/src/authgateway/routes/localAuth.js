//TODO we can move all the auth related routes in a single file
const express = require("express");
const router = express.Router();
const Authentication = require('../controllers/authController');

router.post('/', Authentication.sign_in);

module.exports = router;