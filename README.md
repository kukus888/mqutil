# mqutil

Mqutil is a package that can handle various tasks about IBM MQ. 

## Configuration
Create a mqutil.yml configuration file:

``` yml
ibmmq:
  connections:
    connectorType: docker
    containerName: "ibmmq-test"
  queueManagers:
    - qmName: QM1
      checkInterval: 10s
      retryTimes: 5
```

### Connections
Mqutil currently supports docker containers and local ibm mq.

Docker:
``` yml
ibmmq:
  connections:
    connectorType: docker # Specify the type
    containerName: "ibmmq-test" # Name of the container running locally
```

Local:
``` yml
ibmmq:
  connections:
    connectorType: local # Specify the local type
```
**Warning!** This will execute commands via `/bin/bash -c {{COMMAND}}`. Even though i tried my best to minimize malicious commands exectuion, it is still possible if `mqutil.yml` is misconfigured. 

### Queue Managers
Configure your Queue Managers in order to watch them. Mqutil supports automatic restart of QueueManagers in case of failure with `ibmmq.queueManagers.retryTimes` configuration attribute:
``` yml
ibmmq:
  queueManagers:
    - qmName: QM1 # Name of the QueueManager, has to be same as in IBM MQ
      checkInterval: 10s # Interval in which it will get QueueManager status from MQ
      retryTimes: 5 # How many times mqutil will try to restart the QueueManager (0 or empty for disable)
```

Further information can be retrieved by inspecting the `config.go` source file.

# TODO
Mqutil currently does not support any notifications of any kind. 