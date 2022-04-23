#!/usr/bin/python

import json
import os, sys, time, requests
import sys

headers = {
    'Accept': 'application/json',
    'X-FromAppId': 'free5gc',
    'X-TransactionId': '1',
    'Content-Type': 'application/json',
}

k8_config = "~/.kube/config"
provider_name = "free5gc-cp"
cluster_name = "clu-43"
project_name = "free5gc-tenant"
comp_appname = "free5gc-test"
#dig_name = "vfw1_15"
dig_name = "free5gc-inst1"
lc_name = "free5gc-lc"

# total arguments
n = len(sys.argv)
print ("--------------------------")
print("Total arguments passed:", n)

if n < 5:
    print("Too few arguments passed") #TODO: Need to add help here
    print("required inputs in the mentioned order to the script: VM IP, middleend port, orch port, clm port, dcm port")
    print("Here are the arguments passed:")
    for i in range(1, n):
        print(sys.argv[i])
    exit(1)

print ("--------------------------")
print("Here are the arguments passed:")
for i in range(1, n):
    print(sys.argv[i])

print ("--------------------------")
amcop_deployment_ip = sys.argv[1]
middle_end_port = sys.argv[2]
orch_port = sys.argv[3]
clm_port = sys.argv[4]
dcm_port = sys.argv[5]

try:
    print("---------------Calling health check command---------------\n")
    url1 = 'http://%s:%s/middleend/healthcheck' % (amcop_deployment_ip, middle_end_port)
    res1 = requests.get(url1)
    if res1.status_code != 200:
        print("Health check command returned non 200 code\n")
        print("Health check command response code is: %s\n", res1.status_code)
        res1.raise_for_status()
        exit(1)
    print("Health check command executed successfully\n")
    print("Health check command response code is: %s\n", res1.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling health check command: %s' % e)
try:
    print("---------------Calling create cluster provider command---------------\n")
    #md = open("./metadata.json")
    md = open("./cp-create.json")
    url2 = 'http://%s:%s/v2/cluster-providers' % (amcop_deployment_ip, clm_port)
    time.sleep(1)
    res2 = requests.post(url2, headers=headers, data=md, verify=False)
    if res2.status_code != 200 and res2.status_code != 201 and res2.status_code != 409:
        print("create cluster provider command returned non 200 code\n")
        print("create cluster provider command response code is: %s\n", res2.status_code)
        res2.raise_for_status()
        exit(1)
    print("create cluster provider command response code is: %s\n", res2.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling create cluster provider command: %s' % e)

try:
    print("---------------Calling Onboard cluster command---------------\n")
    #f = open("clu.json", "rb")
    f = open("cluster-onboard.json", "rb")
    data = json.load(f)
    file = "./config_ci"

    files = {
	 'metadata': (None, json.dumps(data).encode('utf-8'), 'application/json'),
	 'file': (os.path.basename(file), open(file, 'rb'), 'application/octet-stream')
    }
    url3 = 'http://%s:%s/middleend/cluster-providers/%s/clusters' % (amcop_deployment_ip, middle_end_port, provider_name)
    time.sleep(1)
    res3 = requests.post(url3, files=files, verify=False)   
    #print res3.reason
    #res3.raise_for_status()
    if res3.status_code != 200 and res3.status_code != 201:
        print("create cluster command returned non 200 code\n")
        print("create cluster command response code is: %s\n", res3.status_code)
        res3.raise_for_status()
        exit(1)
    print("create cluster command response code is: %s", res3.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling create cluster command: %s' % e)
"""
try:
    print("---------------Calling Add network: emco-private-net to cluster command---------------\n")
    cluster_name = "clu-48"
    emco_net = open("./emco-private-net.json")
    # url_emco_net = 'http://%s:%s/v2/ncm/provider-1/clusters/%s/provider-networks' % (amcop_deployment_ip, middle_end_port, cluster_name)
    url_emco_net = 'http://%s:30480/v2/ncm/provider-1/clusters/%s/provider-networks' % (amcop_deployment_ip, cluster_name)
    time.sleep(2)
    res_emco_net = requests.post(url_emco_net, headers=headers, data=emco_net, verify=False)
    #print res_emco_net.reason
    #res_emco_net.raise_for_status()
    if res_emco_net.status_code != 200 and res_emco_net.status_code != 201:
        print("Add EMCO private network to cluster, command response code is: %s\n", res_emco_net.status_code)
        res_emco_net.raise_for_status()
        exit(1)
    print("Add EMCO private network to cluster, command response code is: %s\n", res_emco_net.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Add EMCO private network to cluster, command: %s' % e)

try:
    print("---------------Calling Add network: unprotected-private-net to cluster command---------------\n")
    cluster_name = "clu-48"
    unprotected_net = open("./unprotected-private-net.json")
    # url_unprotected_net = 'http://%s:%s/v2/ncm/provider-1/clusters/%s/provider-networks' % (amcop_deployment_ip, middle_end_port, cluster_name)
    url_unprotected_net = 'http://%s:30480/v2/ncm/provider-1/clusters/%s/provider-networks' % (amcop_deployment_ip, cluster_name)
    time.sleep(2)
    res_unprotected_net = requests.post(url_unprotected_net, headers=headers, data=unprotected_net, verify=False)
    #print res_unprotected_net.reason
    #res_unprotected_net.raise_for_status()
    if res_unprotected_net.status_code != 200 and res_unprotected_net.status_code != 201:
        print("Add unprotected private network to cluster, command response code is: %s\n", res_unprotected_net.status_code)
        res_unprotected_net.raise_for_status()
        exit(1)
    print("Add unprotected private network to cluster, command response code is: %s\n", res_unprotected_net.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Add unprotected private network to cluster, command: %s' % e)

try:
    print("---------------Calling Add network: protected-private-net to cluster command---------------\n")
    cluster_name = "clu-48"
    protected_net = open("./protected-private-net.json")
    # url_protected_net = 'http://%s:%s/v2/ncm/provider-1/clusters/%s/networks' % (amcop_deployment_ip, middle_end_port, cluster_name)
    url_protected_net = 'http://%s:30480/v2/ncm/provider-1/clusters/%s/networks' % (amcop_deployment_ip, cluster_name)
    time.sleep(2)
    res_protected_net = requests.post(url_protected_net, headers=headers, data=protected_net, verify=False)
    #print res_protected_net.reason
    #res_protected_net.raise_for_status()
    if res_protected_net.status_code != 200 and res_protected_net.status_code != 201:
        print("Add protected private network to cluster, command response code is: %s\n", res_protected_net.status_code)
        res_protected_net.raise_for_status()
        exit(1)
    print("Add protected private network to cluster, command response code is: %s\n", res_protected_net.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Add protected private network to cluster, command: %s' % e)

try:
    print("---------------Calling Apply network to the cluster command---------------\n")
    cluster_name = "clu-48"
    # url_net_apply = 'http://%s:%s/v2/ncm/provider-1/clusters/%s/apply' % (amcop_deployment_ip, middle_end_port, cluster_name)
    url_net_apply = 'http://%s:30480/v2/ncm/provider-1/clusters/%s/apply' % (amcop_deployment_ip, cluster_name)
    time.sleep(2)
    res_net_apply = requests.post(url_net_apply, headers=headers, verify=False)
    #print res_net_apply.reason
    #res_net_apply.raise_for_status()
    if res_net_apply.status_code != 204:
        print("Apply network to cluster, command response code is: %s\n", res_net_apply.status_code)
        res_net_apply.raise_for_status()
        exit(1)
    print("Apply network to cluster, command response code is: %s\n", res_net_apply.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Apply network to cluster, command: %s' % e)
"""
try:
    print("---------------Calling create project command---------------\n")
    #prj = open("./project.json")
    prj = open("./tenant-create.json")
    url4 = 'http://%s:%s/v2/projects' % (amcop_deployment_ip, orch_port)
    time.sleep(1)
    res4 = requests.post(url4, headers=headers, data=prj, verify=False)
    #print res4.reason
    #res4.raise_for_status()
    if res4.status_code != 200 and res4.status_code != 201 and res4.status_code != 409:
        print("create project command returned non 200 code\n")
        print("create project command response code is: %s\n", res4.status_code)
        res4.raise_for_status()
        exit(1)
    print("create project command response code is: %s\n", res4.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling create project command: %s' % e)    

try:
    print("---------------Calling create composite app command---------------\n")
    #f = open("tt.json", "rb")
    f = open("service-create.json", "rb")
    data = json.load(f)
    file1 = "./free5g304-helm-0.1.0.tgz"
    #file2 = "/home/gitciuser/packetgen.tgz"
    #file3 = "/home/gitciuser/firewall.tgz"
    file4 = "./profile.tar.gz"
    files = {
     'servicePayload': (None, json.dumps(data).encode('utf-8'), 'application/json'),
     'file1': (os.path.basename(file1), open(file1, 'rb'), 'application/octet-stream'),
     'file4': (os.path.basename(file4), open(file4, 'rb'), 'application/octet-stream')
    }
    url5 = 'http://%s:%s/middleend/projects/%s/composite-apps' % (amcop_deployment_ip, middle_end_port, project_name)
    time.sleep(5)
    res5 = requests.post(url5, files=files, verify=False)
    #print res5.reason
    #res5.raise_for_status()
    if res5.status_code != 200 and res5.status_code != 201 and res5.status_code != 409:
        print("create composite app command returned non 200 code\n")
        print("create composite app command response code is: %s\n", res5.status_code)
        res5.raise_for_status()
        exit(1)
    print("create composite app command response code is: %s\n", res5.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling create composite app command: %s' % e)
""" 
try:
    print("---------------Calling GET composite app command---------------\n")
    url6 = 'http://%s:%s/middleend/projects/%s/composite-apps?filter=depthAll' % (amcop_deployment_ip, middle_end_port, project_name)
    time.sleep(1)
    res6 = requests.get(url6)
    if res6.status_code != 200:
        print("GET composite app command returned non 200 code\n")
        print("GET composite app command response code is: %s\n", res6.status_code)
        res6.raise_for_status()
        exit(1)
    print("GET composite app command response code is: %s\n", res6.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling GET composite app command: %s' % e)  
"""
try:
    print("---------------Calling create logical cloud command---------------\n")
    #lc = open("./lc.json")
    lc = open("./lc-create.json")
    url7 = 'http://%s:%s/middleend/projects/%s/logical-clouds' % (amcop_deployment_ip, middle_end_port, project_name)
    time.sleep(1)
    res7 = requests.post(url7, headers=headers, data=lc, verify=False)
    #print res4.reason
    #res4.raise_for_status()
    if res7.status_code != 200 and res7.status_code != 201 and res7.status_code != 409 and res7.status_code != 202:
        print("create logical cloud command returned non 200 code\n")
        print("create logical cloud command response code is: %s\n", res7.status_code)
        res7.raise_for_status()
        exit(1)
    print("create logical cloud command response code is: %s\n", res7.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling create logical cloud command: %s' % e)


try:
    print("---------------Calling GET logical cloud command---------------\n")
    url8 = 'http://%s:%s/v2/projects/%s/logical-clouds' % (amcop_deployment_ip, dcm_port, project_name)
    time.sleep(1)
    res8 = requests.get(url8)
    if res8.status_code != 200:
        print("GET logical cloud command returned non 200 code\n")
        print("GET logical cloud command response code is: %s\n", res8.status_code)
        res8.raise_for_status()
        exit(1)
    print("GET logical cloud command response code is: %s\n", res8.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling GET logical cloud command: %s' % e)


try:
    print("---------------Calling create DIG command---------------\n")
    #dig = open("./dig.json")
    dig = open("./inst-create.json", "rb")
    data = json.load(dig)
    files = {
     'metadata': (None, json.dumps(data).encode('utf-8'), 'application/json'),
    }
    url9 = 'http://%s:%s/middleend/projects/%s/composite-apps/%s/v1/deployment-intent-groups' % (amcop_deployment_ip, middle_end_port, project_name, comp_appname)

    #url9 = 'http://%s:30480/middleend/projects/%s/composite-apps/%s/v1/deployment-intent-groups' % (amcop_deployment_ip, project_name, comp_appname)

    time.sleep(5)
    res9 = requests.post(url9, files=files, verify=False)
    #print res4.reason
    #res4.raise_for_status()
    #if res9.status_code != 200 and res9.status_code != 201 and res9.status_code != 409:
    if res9.status_code != 200 and res9.status_code != 201 and res9.status_code != 409:
        print("create DIG command returned non 200 code\n")
        print("create DIG command response code is: %s\n", res9.status_code)
        res9.raise_for_status()
        exit(1)
    print("create DIG command response code is: %s\n", res9.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling create DIG command: %s' % e)
"""
try:
    print("---------------Calling Verify DIG command---------------\n")
    url10 = 'http://%s:%s/middleend/projects/%s/composite-apps/%s/v1/deployment-intent-groups/%s' % (amcop_deployment_ip, middle_end_port, project_name, comp_appname, dig_name)
    time.sleep(1)
    res10 = requests.get(url10, headers=headers, verify=False)
    #print res4.reason
    #res4.raise_for_status()
    if res10.status_code != 200:
        print("Verify DIG command returned non 200 code\n")
        print("Verify DIG command response code is: %s\n", res10.status_code)
        res10.raise_for_status()
        exit(1)
    print("Verify DIG command response code is: %s\n", res10.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Verify DIG command: %s' % e)
"""

try:
    print("---------------Calling Approve DIG command---------------\n")
    url11 = 'http://%s:%s/v2/projects/%s/composite-apps/%s/v1/deployment-intent-groups/%s/approve' % (amcop_deployment_ip, orch_port, project_name, comp_appname, dig_name)
    #http://192.168.122.90:30480/v2/projects/TEST8/composite-apps/vFW1/v1/deployment-intent-groups/vfw1_1/approve
    time.sleep(1)
    res11 = requests.post(url11, headers=headers, verify=False)
    #print res4.reason
    #res4.raise_for_status()
    if res11.status_code != 200 and res11.status_code != 202:
        print("Approve DIG command returned non 200 code\n")
        print("Approve DIG command response code is: %s\n", res11.status_code)
        res11.raise_for_status()
        exit(1)
    print("Approve DIG command response code is: %s\n", res11.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Approve DIG command: %s' % e)

try:

    print("---------------Calling Instantiate DIG command---------------\n")
    time.sleep(5)
    url12 = 'http://%s:%s/v2/projects/%s/composite-apps/%s/v1/deployment-intent-groups/%s/instantiate' % (amcop_deployment_ip, orch_port, project_name, comp_appname, dig_name)
    #/v2/projects/TEST8/composite-apps/vFW1/v1/deployment-intent-groups/vfw1_1/instantiate
    time.sleep(1)
    res12 = requests.post(url12)
    #print res12
    #res12.raise_for_status()
    if res12.status_code != 200 and res11.status_code != 202:
        print("Instantiate DIG command returned non 200 code\n")
        print("Instantiate DIG command response code is: %s\n", res12.status_code)
        res12.raise_for_status()
        exit(1)
    print("Instantiate DIG command response code is: %s\n", res12.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception('Exception while calling Instantiate DIG command: %s' % e)




#time.sleep(5)
print("--------------------------------------------------------------\n")
print("Free5gc orchestration is successful")
print("--------------------------------------------------------------\n")
time.sleep(5)
print("\n")
print("--------------------------------------------------------------\n")
print("Starting Free5gc undeploy process")
print("--------------------------------------------------------------\n")
print("\n")
time.sleep(5)

try:
    print("---------------Terminating Free5gc DIG--------------\n")
    time.sleep(5)
    url13 = 'http://%s:%s/v2/projects/%s/composite-apps/%s/v1/deployment-intent-groups/%s/terminate' % (
        amcop_deployment_ip, orch_port, project_name, comp_appname, dig_name)
    time.sleep(1)
    res13 = requests.post(url13)
    if res13.status_code != 202:
        print("Terminate DIG command response code is: %s\n", res13.status_code)
        res13.raise_for_status()
        exit(1)
    print("Terminate DIG command response code is: %s\n", res13.status_code)
    print("--------------------------------------------------------------\n")
except Exception as e:
    raise Exception("Exception while terminating DIG command: %s " % e)
try:
    print("---------------Deleting Free5gc DIG---------------")
    time.sleep(5)
    url14 = 'http://%s:30480/middleend/projects/%s/composite-apps/%s/v1/deployment-intent-groups/%s?operation=deleteAll' % (
        amcop_deployment_ip, project_name, comp_appname, dig_name)
    time.sleep(1)
    res14 = requests.delete(url14)
    #print("Delete DIG status code is : %s", res14.status_code)
    if res14.status_code != 204:
        print("Delete DIG command response code is: %s\n", res14.status_code)
        res14.raise_for_status()
        exit(1)
    print("Delete DIG command response code is: %s\n", res14.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    print("Exception is %s", e)
    raise Exception("Exception while delete DIG command: %s " % e)

try:
    print("---------------Deleting Free5gc service---------------")
    url15 = 'http://%s:%s/middleend/projects/%s/composite-apps/%s/v1' % (
        amcop_deployment_ip, middle_end_port, project_name, comp_appname)
    time.sleep(1)
    res15 = requests.delete(url15)
    #print("Status code is : %s", res15.status_code)
    if res15.status_code != 204:
        print("Delete service command response code is: %s\n", res15.status_code)
        res15.raise_for_status()
        exit(1)
    print("Delete vFW service command response code is: %s\n", res15.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    raise Exception("Exception while deleting service: %s" % e)

try:
    print("---------------Terminating and Deleting logical cloud---------------")
    time.sleep(1)
    #url6 = 'http://%s:%s/v2/dcm/projects/%s/logical-clouds/lc2/terminate' % (amcop_deployment_ip, middle_end_port, project_name)
    url16 = 'http://%s:30480/v2/dcm/projects/%s/logical-clouds/%s/terminate' % (amcop_deployment_ip, project_name, lc_name)
    res16 = requests.post(url16)
    if res16.status_code != 200 and res16.status_code != 202:
        print("Terminate logical cloud command response code is: %s\n", res16.status_code)
        res16.raise_for_status()
        exit(1)
    print("Terminate logical cloud command response code is: %s\n", res16.status_code)

    time.sleep(1)
    #url7 = 'http://%s:%s/v2/dcm/projects/%s/logical-clouds/lc2/cluster-references/lc2-s3p-provider-1-clu-48' % (amcop_deployment_ip, middle_end_port, project_name)
    url17 = 'http://%s:30480/v2/dcm/projects/%s/logical-clouds/%s/cluster-references/free5gc-lc-free5gc-cp-clu-43' % (amcop_deployment_ip, project_name, lc_name)
    res17 = requests.delete(url17)
    if res17.status_code != 204:
        print("Delete logical cloud command-1 response code is: %s\n", res17.status_code)
        res17.raise_for_status()
        exit(1)
    print("Delete logical cloud command-1 response code is: %s\n", res17.status_code)
    print("--------------------------------------------------------------\n")

    time.sleep(1)
    #url8 = 'http://%s:%s/v2/dcm/projects/%s/logical-clouds/lc2' % (amcop_deployment_ip, middle_end_port, project_name)
    url18 = 'http://%s:30480/v2/dcm/projects/%s/logical-clouds/%s' % (amcop_deployment_ip, project_name, lc_name)
    res18 = requests.delete(url18)
    if res18.status_code != 204:
        print("Delete logical cloud command-2 response code is: %s\n", res18.status_code)
        res18.raise_for_status()
        exit(1)
    print("Delete logical cloud command-2 response code is: %s\n", res18.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    raise Exception("Exception while deleting logical cloud")
try:
    print("---------------Deleting Tenant---------------")
    time.sleep(1)
    #url19 = 'http://%s:%s/v2/projects/%s' % (amcop_deployment_ip, middle_end_port, project_name)
    url19 = 'http://%s:30480/v2/projects/%s' % (amcop_deployment_ip, project_name)
    res19 = requests.delete(url19)
    if res19.status_code != 204:
        print("Delete Tenant command response code is: %s\n", res19.status_code)
        res19.raise_for_status()
        exit(1)
    print("Delete Tenant command response code is: %s\n", res19.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    raise Exception("Exception while deleting Tenant")
"""
try:
    print("---------------Deleting added vFW networks---------------")
    time.sleep(1)
    #url20 = 'http://%s:%s/v2/ncm/%s/clusters/%s/terminate' % (amcop_deployment_ip, middle_end_port, provider_name, cluster_name)
    url20 = 'http://%s:30480/v2/ncm/%s/clusters/%s/terminate' % (amcop_deployment_ip, provider_name, cluster_name)
    res20 = requests.post(url20)
    if res20.status_code != 204:
        print("Terminate network command response code is: %s\n", res20.status_code)
        res20.raise_for_status()
        exit(1)
    print("Terminate network command response code is: %s\n", res20.status_code)
    print("--------------------------------------------------------------\n")

    time.sleep(1)
    #url21 = 'http://%s:%s/v2/ncm/%s/clusters/%s/provider-networks/unprotected-private-net' % (amcop_deployment_ip, middle_end_port, provider_name, cluster_name)
    url21 = 'http://%s:30480/v2/ncm/%s/clusters/%s/provider-networks/unprotected-private-net' % (amcop_deployment_ip, provider_name, cluster_name)
    res21 = requests.delete(url21)
    if res21.status_code != 204:
        print("Delete unprotected-private-net network command response code is: %s\n", res21.status_code)
        res21.raise_for_status()
        exit(1)
    print("Delete unprotected-private-net network command response code is: %s\n", res21.status_code)
    print("--------------------------------------------------------------\n")

    time.sleep(1)
    #url22 = 'http://%s:%s/v2/ncm/%s/clusters/%s/provider-networks/emco-private-net' % (amcop_deployment_ip, middle_end_port, provider_name, cluster_name)
    url22 = 'http://%s:30480/v2/ncm/%s/clusters/%s/provider-networks/emco-private-net' % (amcop_deployment_ip, provider_name, cluster_name)
    res22 = requests.delete(url22)
    if res22.status_code != 204:
        print("Delete emco-private-net network command response code is: %s\n", res22.status_code)
        res22.raise_for_status()
        exit(1)
    print("Delete emco-private-net network command response code is: %s\n", res22.status_code)
    print("--------------------------------------------------------------\n")

    time.sleep(1)
    #url23 = 'http://%s:%s/v2/ncm/%s/clusters/%s/networks/protected-private-net' % (amcop_deployment_ip, middle_end_port, provider_name, cluster_name)
    url23 = 'http://%s:30480/v2/ncm/%s/clusters/%s/networks/protected-private-net' % (amcop_deployment_ip, provider_name, cluster_name)
    res23 = requests.delete(url23)
    if res23.status_code != 204:
        print("Delete protected-private-net network command response code is: %s\n", res23.status_code)
        res23.raise_for_status()
        exit(1)
    print("Delete protected-private-net network command response code is: %s\n", res23.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    raise Exception("Exception while deleting vFW networks")
"""
try:
    print("---------------Deleting cluster---------------")
    time.sleep(1)
    #url24 = 'http://%s:%s/v2/cluster-providers/%s/clusters/%s' % (amcop_deployment_ip, middle_end_port, provider_name, cluster_name)
    url24 = 'http://%s:30480/v2/cluster-providers/%s/clusters/%s' % (amcop_deployment_ip, provider_name, cluster_name)
    res24 = requests.delete(url24)
    if res24.status_code != 204:
        print("Delete cluster command response code is: %s\n", res24.status_code)
        res24.raise_for_status()
        exit(1)
    print("Delete cluster command response code is: %s\n", res24.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    raise Exception("Exception while deleting cluster %s", cluster_name)

try:
    print("---------------Deleting provider---------------")
    time.sleep(1)
    #url25 = 'http://%s:%s/v2/cluster-providers/%s' % (amcop_deployment_ip, middle_end_port, provider_name)
    url25 = 'http://%s:30480/v2/cluster-providers/%s' % (amcop_deployment_ip, provider_name)
    res25 = requests.delete(url25)
    if res25.status_code != 204:
        print("Delete provider command response code is: %s\n", res25.status_code)
        res25.raise_for_status()
        exit(1)
    print("Delete provider command response code is: %s\n", res25.status_code)
    print("--------------------------------------------------------------\n")

except Exception as e:
    raise Exception("Exception while deleting provider %s", provider_name)






#python free5gc_orchestrate_python.py 192.168.122.97 30481 30415 30461 30477








