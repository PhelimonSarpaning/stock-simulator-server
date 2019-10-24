#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

DB_URI=""
SERVER_HOST=""
if [ "$1" == "prod" ]; then
  SERVER_HOST="ec2-34-221-86-219.us-west-2.compute.amazonaws.com"
  DB_URI="mockstarket-prod.c6ejpamhqiq5.us-west-2.rds.amazonaws.com"
elif [ "$1" == "dev" ]; then
  SERVER_HOST="ec2-35-164-117-217.us-west-2.compute.amazonaws.com"
  DB_URI="mockstarket-dev.c6ejpamhqiq5.us-west-2.rds.amazonaws.com"
fi;

. $DIR/secrets.sh "prod"

ssh -i $DIR/../mockstarket.pem ec2-user@$SERVER_HOST "psql \"host=$DB_URI port=5432 user=postgres password=$RDS_PASSWORD dbname=postgres\""
