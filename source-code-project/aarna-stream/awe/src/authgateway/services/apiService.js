const axios = require('axios');
const https = require('https');
const {CAMUNDA_HOST, CAMUNDA_AUTH} = require("../config/config");

const agent = new https.Agent({
    rejectUnauthorized: false
});
const axios_config = {
    headers: {
        'Authorization': `Basic ${CAMUNDA_AUTH}`,
        'Content-Type': 'application/json'
    },
    httpsAgent: agent,
}

const provisionTopology = (request) => {
    let config = {
        ...axios_config,
        method: 'post',
        url: `https://${CAMUNDA_HOST}/engine-rest/process-definition/key/PROCESS_PCC_INTERCONNECT_EDGE_CLUSTER_EXTERNAL/start`,
        data: {...request}
    };
    return axios(config)
        .then((res) => res.data);
};

const destroyTopology = (request) => {
    let config = {
        ...axios_config,
        method: 'post',
        url: `https://${CAMUNDA_HOST}/engine-rest/process-definition/key/PROCESS_DESTROY_PCC_INTERCONNECT_EDGE_EXTERNAL/start`,
        data: {...request}
    };
    return axios(config)
        .then((res) => res.data);
};

const getTopologyStatus = ({taskID}) => {
    let config = {
        ...axios_config,
        method: 'get',
        url: `https://${CAMUNDA_HOST}/engine-rest/history/process-instance/${taskID}`,
        httpsAgent: agent,
    };
    return axios(config)
        .then((res) => res.data);
}
const apiService = {provisionTopology, getTopologyStatus, destroyTopology}
module.exports = apiService


