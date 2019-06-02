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
-->

## Metrics

TODO
- [ ] Is there a way to configure refresh rate of MDNS client?
- [ ] Create Prometheus endpoint for [metrics](https://prometheus.io/docs/concepts/metric_types/)
- [ ] Fluentbit [fluentd agent](https://fluentbit.io/)
    - [ ] arm info [here](https://fluentbit.io/documentation/0.12/installation/raspberry_pi.html)
- [ ] Add `ExecStop` and `ExecReload` to rpi service.