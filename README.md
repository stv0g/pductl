# PDU Controller & Prometheus exporter

A little Go tool to control and monitor Baytech PDUs via the serial console port.

Newer generations feature Ethernet connectivity. This is mainly for older models which can be found for quite affordable prices on the second hand market.

## Supported Models

- [Baytech MMP-14](https://www.baytech.net/product/mmp-modular/) - Per Outlet Switched and Metered

## Documentation

Please see [`pductl(3)`](./docs/pductl.md) and [`pdud(3)`](./docs/pdud.md).

## Roadmap

- Direct support for serial devices (rather than networked console ports)
- REST API

## Authors

- [Steffen Vogel](mailto:post@steffenvogel.de) ([@stv0g](https://github.com/stv0g))

## License

This code is license under the [Apache-2.0 licence](LICENSE).

- SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
- SPDX-License-Identifier: Apache-2.0
