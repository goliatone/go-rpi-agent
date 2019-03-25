## Installation

To install the service:

```
cp ./service/rpi-agent.service /etc/systemd/system/rpi-agent.service
sudo chmod 644 /etc/systemd/system/rpi-agent.service
```

If you make changes to the service then:

```
sudo systemctl daemon-reload
```

Or stop:

```
sudo systemctl stop rpi-agent.service
```

```
sudo systemctl start rpi-agent.service
sudo systemctl enable rpi-agent.service
```

Make directory to store configuration and device info:

```
sudo mkdir -p /usr/local/src/rpi-agent/metadata
```

Make `pi` owner of these directories:

```
sudo chown -R pi /usr/local/src/rpi-agent/
```

Create file with default `uuid`:

```
uuid -v4 -o /usr/local/src/rpi-agent/metadata/.device_uuid
```


<!--
Integrate stats:
https://github.com/akhenakh/statgo

examples:
https://github.com/actuallyKasi/testProject/blob/2fc7760048ea98ea68bf026a58f9ac5f37132b27/metrics/api/dal.go

https://github.com/mikkergimenez/distmon/blob/master/proc/main.go

Implement SocketShell:
https://github.com/gravitational/console-demo
-->

## Metrics

TODO 
- [ ] Create Prometheus endpoint for [metrics](https://prometheus.io/docs/concepts/metric_types/)
- [ ] Fluentbit [fluentd agent](https://fluentbit.io/)
    - [ ] arm info [here](https://fluentbit.io/documentation/0.12/installation/raspberry_pi.html)
- [ ] Add `ExecStop` and `ExecReload` to rpi service.