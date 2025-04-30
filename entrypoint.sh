#!/bin/sh

set -e

APP_NAME="nurd-commerce"

/${APP_NAME} migrate
/${APP_NAME} api
