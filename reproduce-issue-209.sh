#!/bin/bash

# Create MySQL object
cat <<EOF | kubectl apply -f -
apiVersion: mysql.nakamasato.com/v1alpha1
kind: MySQL
metadata:
  name: test-mysql
spec:
  host: localhost
  port: 3306
  adminUser:
    name: root
    type: raw
  adminPassword:
    name: password
    type: raw
EOF

# Create MySQLUser immediately
cat <<EOF | kubectl apply -f -
apiVersion: mysql.nakamasato.com/v1alpha1
kind: MySQLUser
metadata:
  name: test-user
spec:
  mysqlName: test-mysql
  username: testuser
  password:
    name: userpass
    type: raw
EOF

# Delete MySQL object immediately
kubectl delete mysql test-mysql

# Check if MySQLUser is orphaned
kubectl get mysqluser test-user -o yaml