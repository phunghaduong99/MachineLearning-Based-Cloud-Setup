const UserModel = require("../models/User");
const mongoose = require("mongoose");

//return the currently logged-in user
exports.getCurrentUser = (req, res) => {
    UserModel.findById(req.user._id, '_id provider displayName tenant role createdAt email image').exec((err, user) => {
        res.status(200).json(user);
    });
}

exports.getAllUsers = (req, res) => {
    UserModel.find({},).exec((err, users) => {
        res.status(200).json(users);
    });
}

exports.deleteUser = async (req, res) => {
    if (req.user.role !== 'admin') {
        return res.status(401).send("unauthorized");
    } else {
        const user = await UserModel.findById(req.params.id);
        if (user === null) {
            res.status(404).json("user not found");
        } else {
            await UserModel.findOneAndRemove({_id: req.params.id})
            res.status(200).json({message: "user deleted", name: req.params.id});
        }
    }

}

exports.updateUser = async (req, res) => {
    if (req.user.role !== 'admin') {
        return res.status(401).send("unauthorized");
    } else {
        if (!mongoose.Types.ObjectId.isValid(req.params.id)) {
            res.status(404).json("user not found");
        } else{
            let updateData = {...req.body};
            if(updateData.firstName && !updateData.displayName){
                updateData.displayName = req.body.firstName;
            }
            let user = await UserModel.findByIdAndUpdate(req.params.id, updateData, {
                new: true
            });
            if (user === null) {
                res.status(404).json("user not found");
            } else {
                res.status(200).json( user);
            }
        }
    }
}

exports.updatePassword = async (req, res) => {
    const {currentPassword, newPassword} = req.body;
    UserModel.findById(req.params.id, function (err, user) {
        user.verifyPassword(currentPassword, async function (err, isMatch) {
            if (err) {
                res.status(500).send("something went wrong");
            }
            if (!isMatch) {
                res.status(400).send("invalid current password");
            } else {
                user.password = newPassword;
                await user.save();
                res.send("password changed");
            }
        });
    }).select('+password');
}