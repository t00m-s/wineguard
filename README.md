# wineguard

## Baby steps: first test

In terminal 1:

```sh
sudo go run ./cmd server
```

In terminal 2:

```sh
sudo go run ./cmd client
```

In terminal 3:

```sh
sudo ip addr add 10.0.69.0/24 dev tun0 &&\
sudo ip addr add 10.0.13.0/24 dev tun1 &&\
sudo ip link set dev tun0 up &&\
sudo ip link set dev tun1 up

sudo tcpdump -i tun1 -n
```

In terminal 4:

```sh
ping 10.0.69.69
```

Now you should see traffic sent to the 10.0.69/24 subnet come out of the tunnel
assigned to the 10.0.13/24 subnet
