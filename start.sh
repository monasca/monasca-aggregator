#!/bin/sh

WAIT_AFTER_FAIL=${WAIT_AFTER_FAIL:-"0"}

/monasca-aggregator

sleep $WAIT_AFTER_FAIL
