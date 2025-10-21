#!/bin/bash
set -e

NAMESPACE="mongo"
REPLICA_COUNT=3

echo "Waiting for MongoDB pods to be ready..."
for i in $(seq 0 $((REPLICA_COUNT-1))); do
    POD="mongodb-$i"
    until kubectl -n $NAMESPACE get pod $POD -o jsonpath='{.status.containerStatuses[0].ready}' 2>/dev/null | grep -q "true"; do
        echo "Waiting for $POD to be ready..."
        sleep 5
    done
done

echo "Initializing MongoDB replica set..."
kubectl -n $NAMESPACE exec mongodb-0 -- mongosh --eval "
rs.initiate({
  _id: 'rs0',
  version: 1,
  members: [
    { _id: 0, host: 'mongodb-0.replica-set.mongo.svc.cluster.local:27017' },
    { _id: 1, host: 'mongodb-1.replica-set.mongo.svc.cluster.local:27017' },
    { _id: 2, host: 'mongodb-2.replica-set.mongo.svc.cluster.local:27017' }
  ]
}, {force: true})"

echo "Verifying replica set status..."
kubectl -n $NAMESPACE exec mongodb-0 -- mongosh --eval "
let i = 10;
while (i > 0) {
  const status = rs.status();
  if (status.ok === 1 && status.members) {
    const allHealthy = status.members.every(m => m.health === 1 && (m.state === 1 || m.state === 2));
    if (allHealthy) {
      print('Replica set is healthy');
      quit(0);
    }
  }
  sleep(500);
  i--;
}
print('Timeout waiting for replica set to initialize');
rs.status();
quit(1);"

echo "MongoDB Replica Set Initialized!"
echo "Connection string for your application:"
echo "mongodb://mongodb-0.replica-set.mongo.svc.cluster.local:27017,mongodb-1.replica-set.mongo.svc.cluster.local:27017,mongodb-2.replica-set.mongo.svc.cluster.local:27017/admin?replicaSet=rs0"
