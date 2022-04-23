const mongoose = require('mongoose')
const {MONGO_URI} = require('./config')

const connectDB = async () => {
    try {
        const conn = await mongoose.connect(MONGO_URI, {
            useNewUrlParser: true,
            useUnifiedTopology: true})
        console.log(`MongoDB Connected: ${conn.connection.host}`)
        return conn;
    }
    catch (err){
        console.error(err)
        process.exit(1)
    }
}

module.exports = connectDB