#!/bin/sh


helm install amf f5gc-amf-0.1.0.tgz -n f5gc
helm install smf f5gc-smf-0.1.0.tgz -n f5gc
helm install ausf f5gc-ausf-0.1.0.tgz -n f5gc
helm install mongo f5gc-mongodb-0.1.0.tgz -n f5gc
helm install nrf f5gc-nrf-0.1.0.tgz -n f5gc
helm install nssf f5gc-nssf-0.1.0.tgz -n f5gc
helm install pcf f5gc-pcf-0.1.0.tgz -n f5gc
helm install udm f5gc-udm-0.1.0.tgz -n f5gc
helm install udr f5gc-udr-0.1.0.tgz -n f5gc
helm install upf f5gc-upf-0.1.0.tgz -n f5gc
helm install webui f5gc-webui-0.1.0.tgz -n f5gc
