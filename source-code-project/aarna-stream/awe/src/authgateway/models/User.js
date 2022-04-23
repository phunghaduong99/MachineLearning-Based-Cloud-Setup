const mongoose = require('mongoose')
const bcrypt = require('bcrypt-nodejs');

const UserSchema = new mongoose.Schema({
    provider: {
        type: String,
        required: true
    },
    id: {
        type: String,
        required: true
    },
    displayName: {
        type: String,
        required: true
    },
    firstName: {
        type: String,
        required: true
    },
    lastName: {
        type: String,
    },
    image: {
        type: String,
    },
    createdAt: {
        type: Date,
        default: Date.now
    },
    tenant: {
        type: String,
        default: null
    },
    role: {
        type: String,
        enum: ['admin', 'tenant', null],
        default: null
    },
    email: {
        type: String,
        required: true
    },
    password: {
        type: String,
        select: false,
        default: null
    }
})

// On Save Hook, encrypt password
// Before saving a model, run this function
UserSchema.pre('save', function (next) {
    // get access to the user model
    const user = this;
    // generate a salt then run callback
    bcrypt.genSalt(10, function (err, salt) {
        if (err) {
            return next(err);
        }

        // hash (encrypt) our password using the salt
        bcrypt.hash(user.password, salt, null, function (err, hash) {
            if (err) {
                return next(err);
            }

            // overwrite plain text password with encrypted password
            user.password = hash;
            next();
        });
    });
});

UserSchema.methods.verifyPassword = function (candidatePassword, callback) {
    bcrypt.compare(candidatePassword, this.password, function (err, isMatch) {
        if (err) {
            console.log("error in user password compare", err);
            return callback(err);
        }
        callback(null, isMatch);
    });
}

UserSchema.methods.getUserFields = function () {
    let userDetails = {}
    userDetails.email = this.email;
    userDetails.id = this.id;
    userDetails.tenant = this.tenant;
    userDetails.role = this.role;
    userDetails.createdAt = this.createdAt;
    userDetails.image = this.image;
    userDetails.displayName = this.displayName;
    userDetails.provider = this.provider;
    userDetails.firstName = this.firstName;
    userDetails.lastName = this.lastName;
    return userDetails;
}
module.exports = mongoose.model('User', UserSchema)