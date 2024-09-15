# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, installShellFiles }:
buildGoModule {
  pname = "pductl";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-Td9f39DAvl440K09yddMjnskxmjoCc9hYgxFASGG5hE=";

  nativeBuildInputs = [ installShellFiles ];

  postInstall = ''
    installShellCompletion --cmd pductl \
      --bash <($out/bin/pductl completion bash) \
      --fish <($out/bin/pductl completion fish) \
      --zsh <($out/bin/pductl completion zsh)

    installShellCompletion --cmd pdud \
      --bash <($out/bin/pdud completion bash) \
      --fish <($out/bin/pdud completion fish) \
      --zsh <($out/bin/pdud completion zsh)
  '';

  meta = {
    mainProgram = "pdud";
  };
}
