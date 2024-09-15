# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  pkgs ? import <nixpkgs> { },
}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    golangci-lint
    reuse
    oapi-codegen
  ];
}
