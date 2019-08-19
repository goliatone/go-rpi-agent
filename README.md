## Installation

To install the service run the `taskfile`:

```
$ run service:install <USER> <HOST> <UUID>
```

Where:
* USER: RPi user, e.g. `pi`
* HOST: RPi host, e.g. `raspberrypi.local`
* UUID: RPi instance ID, e.g. `20d520b5-1f94-412e-9dd6-d6b11aa89b06`

<!--
Integrate stats:
https://github.com/akhenakh/statgo

examples:
https://github.com/actuallyKasi/testProject/blob/2fc7760048ea98ea68bf026a58f9ac5f37132b27/metrics/api/dal.go

https://github.com/mikkergimenez/distmon/blob/master/proc/main.go

Implement SocketShell:
https://github.com/gravitational/console-demo


Reporter:
https://github.com/jondot/groundcontrol

Ticker: 

```go
refresh := 5
refreshes := time.NewTicker(time.Second * time.Duration(refresh)).C
go func() {
    for range refreshes {
        log.Print("Got empty zwave response...")
    }
}()
```

Upload to artifactory:
curl -H "X-JFrog-Art-Api:$ARTIFACTORY_APIKEY" -T ./bin/arm/rpi-agent.tar.gz "${ARTIFACTORY_URL}/${ARTIFACTORY_PATH}/rpi-agent.tar.gz"

macos dns:
dns-sd -B _rpi._tcp .

how to set TTL for record?
-->

## Metrics

TODO

- [ ] Take URL endpoint from flags
- [ ] Refactor to use external template to configure POST payload
- [ ] Add flag to run timer for POST payload
- [ ] Refactor so we collect data dynamically:
    - [ ] Each function should take output map and add its result to it, e.g.: getMac(&out)
    - [ ] These collectors should be a plugin interface to be loaded

- [ ] Is there a way to configure refresh rate of MDNS client?
- [ ] Create Prometheus endpoint for [metrics](https://prometheus.io/docs/concepts/metric_types/)
- [ ] Fluentbit [fluentd agent](https://fluentbit.io/)
    - [ ] arm info [here](https://fluentbit.io/documentation/0.12/installation/raspberry_pi.html)
- [ ] Add `ExecStop` and `ExecReload` to rpi service.