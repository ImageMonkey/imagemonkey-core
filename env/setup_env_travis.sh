#! /bin/bash

echo "SENTRY_DSN=\"\"" >> /etc/environment
echo "IMAGEMONKEY_DB_CONNECTION_STRING=\"host=127.0.0.1 port=5433 user=monkey dbname=imagemonkey password=dbRuwMUo4Nfhs5hmMxhk sslmode=disable\"" >> /etc/environment
echo "JWT_SECRET=\"e0e8cb89320d6fd5b46eeb32c22cd3f5d657eb8eafcbed1cafe24a03a6ca47f7\"" >> /etc/environment
echo "X_CLIENT_ID=\"de61ac57c1889941a9200ecff2c8eeeb390350c9813e13e8d439516dd389127f\"" >> /etc/environment
echo "X_CLIENT_SECRET=\"ef2748970181a4d3b0e5892f755f60a1cb24980c66d880e971542e8f1aae8958\"" >> /etc/environment
echo "GITHUB_API_TOKEN=\"\"" >> /etc/environment
echo "GITHUB_PROJECT_OWNER=\"\"" >> /etc/environment
source /etc/environment
