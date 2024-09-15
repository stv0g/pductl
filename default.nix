{ buildGoModule }:
buildGoModule {
  pname = "pductl";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-Td9f39DAvl440K09yddMjnskxmjoCc9hYgxFASGG5hE=";

  meta = {
    mainProgram = "pdud";
  };
}
