const poll = ({task, validate, interval, maxAttempts}) => {
    console.log('Start to poll...');
    let attempts = 0;
    const executePoll = async (resolve, reject) => {
        console.log('executing polling task, attempt : ', attempts);
        try {
            const result = await task();
            attempts++;
            if (validate(result)) {
                return resolve(result);
            } else if (maxAttempts && attempts === maxAttempts) {
                console.log("max attempts reached, exiting polling");
                return reject(new Error('Exceeded max attempts'));
            } else {
                setTimeout(executePoll, interval, resolve, reject);
            }
        } catch (err) {
            //TODO : do we want to continue polling if there is some network error ?
            console.log("error executing poling task");
            return reject(new Error(err));
        }
    };
    return new Promise(executePoll);
};
module.exports = {poll};