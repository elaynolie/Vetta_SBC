# Vetta SBC

**Vetta SBC** is a Session Border Controller built on top of **OpenSIPS 3.6.7**, designed for carrier-grade SIP trunking, subscriber registration, and fraud-resistant call routing.

Hi, I'm **Roman Syrtlanov**, a VoIP engineer from Russia. This is my SBC project, built and battle-tested on production traffic.

---

## Tech Stack

- **[OpenSIPS 3.6](https://opensips.org/)** — core SIP signaling engine
- **[RTPengine](https://github.com/sipwise/rtpengine)** — media relay / RTP proxy
- **PostgreSQL** — persistent storage (subscribers, dialplan, routing tables)
- **RabbitMQ** — event bus / eventing
- **[Homer](https://github.com/sipcapture/homer)** — HEP-based SIP capture and traffic visualization

---

## Operating Modes

Vetta SBC can run in two distinct modes depending on your architecture:

| Mode | Description |
|---|---|
| `midregistrar_mode` | REGISTER requests are proxied/mid-registered to a downstream B2BUA (e.g. FreeSWITCH, Asterisk), which owns the subscriber database |
| `registrar_mode` | Vetta SBC itself handles the full registration and authentication lifecycle — no B2BUA required |

Both digest-authenticated and non-authenticated (trusted) carrier peers are supported.

---

## Transport Support

- IPv4 / IPv6
- TCP / UDP
- WS / WSS (WebRTC-ready)

---

## Security

By default, inbound calls to a carrier are only permitted from source addresses matching **dpid 1** (see the `dialplan` table).

On top of that, Vetta SBC implements an **allowed-countries mechanism**. After years in VoIP dealing with toll fraud — and the reality that firewall rules (iptables/nftables) can be flushed or misconfigured — this acts as an application-level safety net independent of the OS firewall.

Enable it with:

```bash
opensips-cli -x mi shv_set allowed_country str "PL|IT"
```

With this set, only traffic from source addresses affiliated with the listed countries (here, Poland and Italy) is accepted — everything else is rejected with a `503`. No surprise calls (or bills) to premium/high-fraud destinations.

---

## Custom Metrics

- Initialized in `startup_route`
- Native **Homer** integration for full SIP traffic capture and visualization

---

## Roadmap

This project is actively developed — commits are ongoing as features are refined and hardened in production.

## Contributing

⭐ If you find this project useful, a star is always appreciated.

## Commercial Inquiries

I'm open to discussing commercial deployments of Vetta SBC, including custom feature development for your specific use case. Feel free to reach out.

## Contact

- Roman Syrtlanov
- Telegram: [@rsyrtlanov](https://t.me/rsyrtlanov)


