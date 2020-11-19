#!/bin/bash -l
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#
###############################################################
#
#     CASM Test - Cray Inc.
#
#     TEST IDENTIFIER   : scsd_tavern_api_test
#
#     DESCRIPTION       : Automated test for verifying the HMS System
#                         Configuration Service (SCSD) API on Cray Shasta 
#                         systems.
#                         
#     AUTHOR            : Mitch Schooler
#
#     DATE STARTED      : 05/13/2020
#
#     LAST MODIFIED     : 09/16/2020
#
#     SYNOPSIS
#       This is a test wrapper for HMS System Configuration Service (SCSD)
#       API tests implemented in Tavern that launch via pytest within the
#       Continuous Test (CT) framework. All Tavern tests packaged in
#       the target CT test directory for SCSD are executed.
#
#     INPUT SPECIFICATIONS
#       Usage: scsd_tavern_api_test
#       
#       Arguments: None
#
#     OUTPUT SPECIFICATIONS
#       Plaintext is printed to stdout and/or stderr. The script exits
#       with a status of '0' on success and '1' on failure.
#
#     DESIGN DESCRIPTION
#       This test wrapper generates a Tavern configuration file based
#       on the target test system it is running on and then executes all
#       SCSD Tavern API CT tests using HMS's hms-pytest container which
#       includes pytest and other dependencies required to run Tavern.
#
#     SPECIAL REQUIREMENTS
#       Must be executed on the target test system on a fully-installed
#       NCN with the Continuous Test infrastructure in place.
#
#     UPDATE HISTORY
#       user       date         description
#       -------------------------------------------------------
#       schooler   05/19/2020   initial implementation
#       schooler   06/03/2020   updated source of vault BMC credentials
#       schooler   09/16/2020   use latest hms_common_file_generator
#
#     DEPENDENCIES
#       - hms-pytest wrapper script which is expected to be packaged
#         in /usr/bin on the NCN.
#       - hms_smoke_test_lib_ncn-resources_remote-resources.sh which
#         is expected to be packaged in
#         /opt/cray/tests/ncn-resources/hms/hms-test on the NCN.
#       - hms_pytest_ini_file_generator_ncn-resources_remote-resources.py
#         which is expected to be packaged in
#         /opt/cray/tests/ncn-resources/hms/hms-test on the NCN.
#       - hms_common_file_generator_ncn-resources_remote-resources.py
#         which is expected to be packaged in
#         /opt/cray/tests/ncn-resources/hms/hms-test on the NCN.
#       - SCSD Tavern API tests with names of the form test_*.tavern.yaml
#         which are expected to be packaged in
#         /opt/cray/tests/ncn-functional/hms/hms-scsd on the NCN.
#
#     BUGS/LIMITATIONS
#       None
#
###############################################################

# cleanup
function cleanup()
{
    echo "cleaning up..."
    if [[ -d ${SCSD_TEST_DIR}/.pytest_cache ]] ; then
        rm -rf ${SCSD_TEST_DIR}/.pytest_cache
    fi
    rm -f ${PYTEST_INI_PATH}
    rm -f ${COMMON_FILE_PATH}
}

# HMS path declarations
HMS_TEST_LIB="/opt/cray/tests/ncn-resources/hms/hms-test/hms_smoke_test_lib_ncn-resources_remote-resources.sh"
PYTEST_INI_GENERATOR="/opt/cray/tests/ncn-resources/hms/hms-test/hms_pytest_ini_file_generator_ncn-resources_remote-resources.py"
PYTEST_INI_PATH="/opt/cray/tests/ncn-functional/hms/hms-scsd/pytest.ini"
COMMON_FILE_GENERATOR="/opt/cray/tests/ncn-resources/hms/hms-test/hms_common_file_generator_ncn-resources_remote-resources.py"
COMMON_FILE_PATH="/opt/cray/tests/ncn-functional/hms/hms-scsd/common.yaml"
SCSD_TEST_DIR="/opt/cray/tests/ncn-functional/hms/hms-scsd"
API_TARGET="https://api-gw-service-nmn.local/apis"

# set SSL certificate checking to True for test execution from the NCN
VERIFY="True"
echo "VERIFY=${VERIFY}"

# set up signal handling
trap ">&2 echo \"recieved kill signal, exiting with status of '1'...\" ; \
    cleanup ; \
    exit 1" SIGHUP SIGINT SIGTERM

# verify that the hms-pytest wrapper script exists
PYTEST_PATH=$(which hms-pytest)
if [[ -z ${PYTEST_PATH} ]] ; then
    >&2 echo "ERROR: failed to locate command: hms-pytest"
    cleanup
    exit 1
fi

# source the HMS smoke test library file
if [[ -r ${HMS_TEST_LIB} ]] ; then
    . ${HMS_TEST_LIB}
else
    >&2 echo "ERROR: failed to source HMS smoke test library: ${HMS_TEST_LIB}"
    cleanup
    exit 1
fi

# verify that the Tavern test directory exists
if [[ ! -d ${SCSD_TEST_DIR} ]] ; then
    >&2 echo "ERROR: failed to locate Tavern test directory: ${SCSD_TEST_DIR}"
    cleanup
    exit 1
else
    TEST_DIR_FILES=$(ls ${SCSD_TEST_DIR})
    TEST_DIR_TAVERN_FILES=$(echo "${TEST_DIR_FILES}" | grep -E "^test_.*\.tavern\.yaml")
    if [[ -z "${TEST_DIR_TAVERN_FILES}" ]] ; then
        >&2 echo "ERROR: no Tavern tests in CT test directory: ${SCSD_TEST_DIR}"
        >&2 echo "${TEST_DIR_FILES}"
        cleanup
        exit 1
    fi
fi

# verify that the pytest.ini generator tool exists
if [[ ! -x ${PYTEST_INI_GENERATOR} ]] ; then
    >&2 echo "ERROR: failed to locate executable pytest.ini file generator: ${PYTEST_INI_GENERATOR}"
    cleanup
    exit 1
fi

# verify that the common file generator tool exists
if [[ ! -x ${COMMON_FILE_GENERATOR} ]] ; then
    >&2 echo "ERROR: failed to locate executable common file generator: ${COMMON_FILE_GENERATOR}"
    cleanup
    exit 1
fi

echo "Running scsd_tavern_api_test..."

# retrieve Keycloak authentication token for session
TOKEN=$(get_auth_access_token)
TOKEN_RET=$?
if [[ ${TOKEN_RET} -ne 0 ]] ; then
    cleanup
    exit 1
fi

# generate pytest.ini configuration file
GENERATE_PYTEST_INI_CMD="${PYTEST_INI_GENERATOR} --file ${PYTEST_INI_PATH}"
timestamp_print "Running '${GENERATE_PYTEST_INI_CMD}'..."
eval "${GENERATE_PYTEST_INI_CMD}"
GENERATE_PYTEST_INI_RET=$?
if [[ ${GENERATE_PYTEST_INI_RET} -ne 0 ]] ; then
    >&2 echo "ERROR: pytest.ini file generator failed with error code: ${GENERATE_PYTEST_INI_RET}"
    cleanup
    exit 1
else
    if [[ ! -r ${PYTEST_INI_PATH} ]] ; then
        >&2 echo "ERROR: failed to generate readable pytest.ini file"
        cleanup
        exit 1
    fi
fi

# generate Tavern common.yaml configuration file
BMC_ROOT_PW_FILE="/opt/cray/crayctl/ansible_framework/.vault.txt"
BMC_ROOT_PW_CMD="cat ${BMC_ROOT_PW_FILE}"
timestamp_print "Running '${BMC_ROOT_PW_CMD}'..."
BMC_ROOT_PW_ORIG=$(eval "${BMC_ROOT_PW_CMD}")
if [[ -z ${BMC_ROOT_PW_ORIG} ]] ; then
    >&2 echo "ERROR: failed to retrieve BMC root password from: '${BMC_ROOT_PW_CMD}'"
    # try back-up location for BMC credentials
    BMC_ROOT_PW_CMD_PREV="ansible-vault view /etc/vault/vault.yml | cat"
    timestamp_print "Running '${BMC_ROOT_PW_CMD_PREV}'..."
    ANSIBLE_VAULT_OUT=$(eval "${BMC_ROOT_PW_CMD_PREV}")
    BMC_ROOT_PW_ORIG=$(echo "${ANSIBLE_VAULT_OUT}" | grep vault_bmc_root_password | cut -d " " -f 2)
    if [[ -z ${BMC_ROOT_PW_ORIG} ]] ; then
        # non-fatal error for running the SCSD Tavern tests but will cause test failures
        >&2 echo "ERROR: failed to retrieve BMC root password from: '${BMC_ROOT_PW_CMD_PREV}'"
        #>&2 echo "${ANSIBLE_VAULT_OUT}"
        BMC_ROOT_PW_ORIG="unset"
    fi
fi
BMC_BASIC_AUTH_ORIG=$(echo -n "root:${BMC_ROOT_PW_ORIG}" | base64)
BMC_ROOT_PW_TMP="temporarybmcpassword"
BMC_BASIC_AUTH_TMP=$(echo -n "root:${BMC_ROOT_PW_TMP}" | base64)
GENERATE_COMMON_FILE_CMD="${COMMON_FILE_GENERATOR} \
--base_url ${API_TARGET} \
--file ${COMMON_FILE_PATH} \
--access_token ${TOKEN} \
--verify ${VERIFY} \
--bmc_basic_auth_orig ${BMC_BASIC_AUTH_ORIG} \
--bmc_password_orig ${BMC_ROOT_PW_ORIG} \
--bmc_basic_auth_new ${BMC_BASIC_AUTH_TMP} \
--bmc_password_new ${BMC_ROOT_PW_TMP}"
timestamp_print "Running '${GENERATE_COMMON_FILE_CMD}'..."
eval "${GENERATE_COMMON_FILE_CMD}"
GENERATE_COMMON_FILE_RET=$?
if [[ ${GENERATE_COMMON_FILE_RET} -ne 0 ]] ; then
    >&2 echo "ERROR: common file generator failed with error code: ${GENERATE_COMMON_FILE_RET}"
    cleanup
    exit 1
else
    if [[ ! -r ${COMMON_FILE_PATH} ]] ; then
        >&2 echo "ERROR: failed to generate readable Tavern common.yaml file"
        cleanup
        exit 1
    fi
fi

# execute Tavern tests in the hms-pytest container with pytest
PYTEST_CMD="${PYTEST_PATH} --tavern-global-cfg=${COMMON_FILE_PATH} ${SCSD_TEST_DIR}"
timestamp_print "Running '${PYTEST_CMD}'..."
eval "${PYTEST_CMD}"
TAVERN_RET=$?
if [[ ${TAVERN_RET} -ne 0 ]] ; then
    echo "FAIL: scsd_tavern_api_test ran with failures"
    cleanup
    exit 1
else
    echo "PASS: scsd_tavern_api_test passed!"
    cleanup
    exit 0
fi
